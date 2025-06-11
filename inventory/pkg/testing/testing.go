// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	"context"
	"embed"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/test/bufconn"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/migrate/migrations"
	_ "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/runtime" // initialize ent's runtime for test
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/server"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/store"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/utils/migrate"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/rego"
)

type ClientType string

const (
	APIClient       ClientType = "TestAPI"
	RMClient        ClientType = "TestRM"
	TCClient        ClientType = "TestTC"
	CacheUUIDClient ClientType = "TestCacheUuid"
	CacheClient     ClientType = "TestCache"
)

// BufconnLis TestClients TestClientsEvents are used only for testing,
// creating a bufconn between server and client, providing access to different
// test clients, and provide channels to inventory events. 	The TestClients maps
// are name -> InventoryClient which implies that names must be unique.
var (
	zlog              = logging.GetLogger("InfraInvTesting")
	TestClients       = make(map[ClientType]client.InventoryClient)
	TestClientsEvents = make(map[ClientType]chan *client.WatchEvents)
	termChan          = make(chan bool)
	wg                = sync.WaitGroup{}
	BufconnLis        *bufconn.Listener

	cacheStaleTime = 5 * time.Second
)

// Internal parameters for bufconn testing.
const (
	bufferSize    = util.Megabyte
	eventsBufSize = 128
	timeout       = 60 * time.Second
)

// This functions setups test server and test clients.
// policyPath is used for bootstrapping the OPA agent.
// certPath is used for boostrapping secure connections. Empty will fallback to insecure.
// migrPath is used to initialize the versioned migrations. Empty will fallback to automatic migrations.
func StartTestingEnvironment(policyPath, certPath, migrationsDir string) {
	// Bootstrap the policy bundle
	CreatePolicyBundle(policyPath)
	// Boostrap c/s connectivity using bufconn
	createBufConn()
	// Boostrap the DB
	createSchema(migrationsDir)
	// Bootstrap server
	createServer(policyPath, certPath)
	// Bootstrap the clients

	resourceKinds := collections.MapSlice[int32, inv_v1.ResourceKind](
		maps.Values(inv_v1.ResourceKind_value), func(v int32) inv_v1.ResourceKind { return inv_v1.ResourceKind(v) })

	for _, cc := range []struct {
		ct   ClientType
		inct inv_v1.ClientKind
	}{
		{APIClient, inv_v1.ClientKind_CLIENT_KIND_API},
		{RMClient, inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER},
		{TCClient, inv_v1.ClientKind_CLIENT_KIND_TENANT_CONTROLLER},
		{CacheClient, inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER},
		{CacheUUIDClient, inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER},
	} {
		if err := CreateClient(cc.ct, cc.inct, resourceKinds, certPath); err != nil {
			zlog.Fatal().Err(err).Msgf("Cannot create %s client", cc.ct)
		}
	}
}

// This function is used to stop the test environment.
func StopTestingEnvironment() {
	// Close all the registered clients
	for _, client := range TestClients {
		client.Close()
	}
	// stop the server after tests
	close(termChan)
	// wait until servers terminate
	wg.Wait()
}

// Helper function to create a client.
// clientName is the client name.
// clientKind identifies the type of client; clients can be retrieved by the depending using ClientKind on TestClients.
// resourceKinds indicates the type of resources this client is interested to receive events about.
// certPath is the certificate path.
func CreateClient(ct ClientType, clientKind inv_v1.ClientKind, resourceKinds []inv_v1.ResourceKind, certPath string) error {
	// Prevent duplicate
	if _, ok := TestClients[ct]; ok {
		return errors.Errorfc(codes.Internal, "Client %s already exists", ct)
	}
	if _, ok := TestClientsEvents[ct]; ok {
		return errors.Errorfc(codes.Internal, "Client Watcher %s already exists", ct)
	}
	dialOpts := []grpc.DialOption{
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return BufconnLis.Dial() }),
	}
	if ct == APIClient {
		dialOpts = append(dialOpts,
			grpc.WithUnaryInterceptor(client.TenantContextExtractingInterceptor()),
			grpc.WithUnaryInterceptor(client.TenantContextInsertingInterceptorTestingOnly()))
	}
	// Init the client and update the glob structures
	insecure := true
	if certPath != "" {
		insecure = false
	}
	events := make(chan *client.WatchEvents, eventsBufSize)
	clientCfg := client.InventoryClientConfig{
		Name:        string(ct),
		Address:     "bufconn",
		DialOptions: dialOpts,
		SecurityCfg: &client.SecurityConfig{
			Insecure: insecure,
			CaPath:   certPath + "/ca-cert.pem",
			CertPath: certPath + "/client-cert.pem",
			KeyPath:  certPath + "/client-key.pem",
		},
		Events:     events,
		ClientKind: clientKind,
		// Registering all kinds of used resources
		ResourceKinds: resourceKinds,
		Wg:            &wg,
		EnableTracing: true,
	}
	// enable cache only for cache based client
	if ct == CacheClient {
		clientCfg.ClientCache = client.InvClientCacheConfig{
			EnableCache: true,
			StaleTime:   cacheStaleTime,
		}
	}
	if ct == CacheUUIDClient {
		clientCfg.ClientCache = client.InvClientCacheConfig{
			EnableUUIDCache: true,
			StaleTime:       cacheStaleTime,
		}
	}

	// initialize client
	cli, err := client.NewInventoryClient(
		context.Background(),
		clientCfg,
	)
	if err != nil {
		return err
	}
	TestClients[ct] = cli
	TestClientsEvents[ct] = events
	zlog.Info().Msgf("Started Test Inventory client %s...\n", ct)

	return nil
}

// Build at run time the fs in the output path.
func createFS(fs embed.FS, outputPath string) {
	dirEntries, err := fs.ReadDir(".")
	if err != nil {
		zlog.Fatal().Msgf("Cannot open embedded fs folder")
	}
	var content []byte
	for _, entry := range dirEntries {
		content, err = fs.ReadFile(entry.Name())
		if err != nil {
			zlog.Fatal().Msgf("Cannot read file %s", entry.Name())
		}
		err = os.WriteFile(outputPath+"/"+entry.Name(), content, 0o644) //nolint:gosec,mnd // used for testing only
		if err != nil {
			zlog.Fatal().Msgf("Cannot write file %s in %s", entry.Name(), outputPath+"/")
		}
	}
}

// Build at run time the policy bundle.
func CreatePolicyBundle(policyPath string) {
	// Write in the policy path the files needed for the bundle
	createFS(rego.RegoFolder, policyPath)
	// Finally, build the policy bundle
	args := []string{
		"build",
		policyPath,
		"-o", policyPath + "/policy_bundle.tar.gz",
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	zlog.Info().Msgf("Creating policy bundle in %v", args)
	_, err := exec.CommandContext(ctx, "opa", args...).CombinedOutput()
	if err != nil {
		zlog.Fatal().Msgf("Cannot create policy bundle in %s, %v", policyPath, err)
	}
}

// Create the bufconn listener used for the c/s communication.
func createBufConn() {
	// https://pkg.go.dev/google.golang.org/grpc/test/bufconn#Listener
	buffer := bufferSize
	BufconnLis = bufconn.Listen(buffer)
}

// Create the schema using the versioned migrations or the automatic migrations.
func createSchema(migrationsDir string) {
	// Write in the migrations the files needed for the atlas migrations
	createFS(migrations.MigrationsFolder, migrationsDir)
	// Create schema by run versioned schema migrations.
	dbURL := util.GetDBURL(util.LookupDBTestEnv())

	if out, merr := migrate.RunAtlasMigrations(dbURL, migrationsDir); merr != nil {
		zlog.Fatal().Err(merr).Msgf("Database migration failed. Aborting. Atlas output: %v", string(out))
	}
	// Clear all DB entries to provide a clean environment between tests.
	if err := clearDB(context.TODO(), dbURL); err != nil {
		zlog.Fatal().Err(err).Msg("Cannot clear DB")
	}
}

// clearDB opens a temporary connection to the database and deletes all known
// entries, but not the schema itself.
// Be aware that order of deletion statements below matters.
func clearDB(ctx context.Context, dbURL string) error { //nolint:cyclop,funlen // used for testing only
	c := store.ConnectEntDB(dbURL, "")
	defer c.Close()

	if _, err := c.WorkloadMember.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.EndpointResource.Delete().Exec(ctx); err != nil {
		return err
	}
	// Strong parent relation between IP and Nic resources
	if _, err := c.IPAddressResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.HostnicResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.HoststorageResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.HostusbResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.HostgpuResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.NetlinkResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.NetworkSegment.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.OuResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.ProviderResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.RegionResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.RepeatedScheduleResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.SingleScheduleResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.SiteResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.RemoteAccessConfiguration.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.OSUpdatePolicyResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.InstanceResource.Delete().Exec(ctx); err != nil {
		return err
	}
	// OS must be cleared after Instance
	if _, err := c.OperatingSystemResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.WorkloadResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.LocalAccountResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.CustomConfigResource.Delete().Exec(ctx); err != nil {
		return err
	}
	// Clear host after host components and instance
	if _, err := c.HostResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.TelemetryProfile.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.TelemetryGroupResource.Delete().Exec(ctx); err != nil {
		return err
	}
	if _, err := c.Tenant.Delete().Exec(ctx); err != nil {
		return err
	}
	return nil
}

// Helper function to create a server.
func createServer(policyPath, certPath string) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigChan
		close(termChan)
	}()
	wg.Add(1) // Add grpc server
	go func() {
		opts := server.Options{}
		opts.CaCertPath = certPath + "/ca-cert.pem"
		opts.TLSCertPath = certPath + "/server-cert.pem"
		opts.TLSKeyPath = certPath + "/server-key.pem"
		opts.InsecureGrpc = true
		opts.EnableAuditing = false
		if certPath != "" {
			opts.InsecureGrpc = false
		}
		// For testing purposes we use the same URL for both writer and reader
		dbURL := util.GetDBURL(util.LookupDBTestEnv())
		server.StartInventoryGrpcServer(termChan, nil, &wg, BufconnLis, dbURL, dbURL, policyPath+"/policy_bundle.tar.gz", opts)
	}()
	zlog.Info().Msgf("Started Inventory server...\n")
}

// SetEnvVariables sets the environment variables from the given map.
func SetEnvVariables(envVars map[string]string) error {
	for key, value := range envVars {
		err := os.Setenv(key, value)
		if err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", key, err)
		}
	}
	return nil
}
