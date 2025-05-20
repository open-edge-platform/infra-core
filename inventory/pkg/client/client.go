// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostresource"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	telemetry_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/cert"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client/cache"
	inv_errors "github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/metrics"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tracing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/filters"
)

var zlog = logging.GetLogger("InfraAPIClient")

var (
	// Ping the server if it is idle for 60 seconds to ensure the connection is still active.
	clientKeepAliveTime = 60 * time.Second
	// Wait 1 second for the ping ack before assuming the connection is dead.
	clientKeepAliveTimeout = 1 * time.Second
)

// ResourceTenantIDCarrier type wrapper for FindResourcesResponse_ResourceTenantIDCarrier.
type ResourceTenantIDCarrier = inv_v1.FindResourcesResponse_ResourceTenantIDCarrier

const (
	BatchSize               = 100
	InsecureGrpc            = "insecureGRPC"
	InsecureGrpcDescription = "Flag to disable secure connectivity"
	CaCertPath              = "caCertPath"
	CaCertPathDescription   = "Path to the Certified Authority (CA) certificate. " +
		"It must be provided if InsecureGRPC is disabled."
	TLSCertPath                 = "tlsCertPath"
	TLSCertPathDescription      = "Path to the TLS certificate. It must be provided if InsecureGRPC is disabled."
	TLSKeyPath                  = "tlsKeyPath"
	TLSKeyPathDescription       = "Path to the TLS key. It must be provided if InsecureGRPC is disabled."
	InventoryAddress            = "inventoryAddress"
	InventoryAddressDescription = "Inventory service address to connect to. It should have the following " +
		"format <IP address>:<port>."
	InvCacheUUIDEnable                    = "invCacheUuidEnable"
	InvCacheUUIDEnableDescription         = "Flag to enable Inventory Client cache by UUID"
	InvCacheStaleTimeout                  = "invCacheStaleTimeout"
	InvCacheStaleTimeoutDescription       = "Flag to enable Inventory Client cache"
	InvCacheStaleTimeoutDefault           = 5 * time.Minute
	InvCacheStaleTimeoutOffset            = "invCacheStaleTimeoutOffset"
	InvCacheStaleTimeoutOffsetDefault     = 15
	InvCacheStaleTimeoutOffsetDescription = "Parameter to set the timeout offset percentage for the Inventory UUID cache"
)

type WatchEvents struct {
	Ctx   context.Context
	Event *inv_v1.SubscribeEventsResponse
}

type inventoryClient struct {
	cfg          *InventoryClientConfig
	connection   *grpc.ClientConn
	invAPI       inv_v1.InventoryServiceClient
	clientUUID   string
	streamCtx    context.Context
	streamCancel context.CancelFunc
	stream       inv_v1.InventoryService_SubscribeEventsClient
	uuidMutex    sync.RWMutex
	cache        *cache.InventoryCache
	cacheUUID    *cache.InventoryCache
}

// TenantAwareInventoryClient defines all the methods that inventoryClient must implement.
type TenantAwareInventoryClient interface {
	// Close unregisters the client from the inventory server and terminates the
	// gRPC connection. The client cannot be reused after this call. It is safe
	// to call this multiple times and from multiple goroutines.
	Close() error
	// List looks for inventory resources based on a filter definition
	// returning their objects. If no resources are found, an empty slice (of length 0) is returned.
	List(context.Context, *inv_v1.ResourceFilter) (*inv_v1.ListResourcesResponse, error)
	// ListAll looks for inventory resources based on the given filter and fieldMask
	// returning all objects that matches the filter. If no resources are found, an empty slice (of length 0) is returned.
	// Offset and limit set in the resource filter are ignored.
	ListAll(context.Context, *inv_v1.ResourceFilter) ([]*inv_v1.Resource, error)
	// Find looks for inventory resources based on a filter definition
	// returning their IDs. If no resources are found, an empty slice (of length 0) is returned.
	Find(context.Context, *inv_v1.ResourceFilter) (*inv_v1.FindResourcesResponse, error)
	// FindAll looks for inventory resources based on the given filter and fieldMask
	// returning all the ID that matches the filter. If no resources are found, an empty slice (of length 0) is returned.
	// Offset and limit set in the resource filter are ignored.
	FindAll(context.Context, *inv_v1.ResourceFilter) ([]*ResourceTenantIDCarrier, error)
	// Get retrieves a resource from inventory based on its ID.
	Get(ctx context.Context, tenantID, id string) (*inv_v1.GetResourceResponse, error)
	// Create creates a resource in inventory, providing its newly created ID in the response.
	Create(ctx context.Context, tenantID string, res *inv_v1.Resource) (*inv_v1.Resource, error)
	// Update updates a resource in inventory, given the resource ID, the fieldmask
	// to be applied on the resource fields, and the resource instance.
	Update(ctx context.Context, tenantID, id string,
		fm *fieldmaskpb.FieldMask, res *inv_v1.Resource) (*inv_v1.Resource, error)
	// Delete deletes a resource from inventory based on its ID.
	Delete(ctx context.Context, tenantID, id string) (*inv_v1.DeleteResourceResponse, error)
	// UpdateSubscriptions sets the resource kinds this clients will receive
	// events for.
	UpdateSubscriptions(ctx context.Context, tenantID string, kinds []inv_v1.ResourceKind) error
	// ListInheritedTelemetryProfiles lists inherited telemetry profiles given the inheritBy parameter.
	// The given filter parameter can then be added to filter the list of inherited telemetry.
	// orderBy can be specified to order result by a given field.
	// limit and offset parameters are used to paginate results.
	ListInheritedTelemetryProfiles(
		ctx context.Context,
		tenantID string,
		inheritBy *inv_v1.ListInheritedTelemetryProfilesRequest_InheritBy,
		filter string,
		orderBy string,
		limit, offset uint32,
	) (*inv_v1.ListInheritedTelemetryProfilesResponse, error)
	GetHostByUUID(ctx context.Context, tenantID string, uuid string) (*computev1.HostResource, error)
	GetTreeHierarchy(context.Context, *inv_v1.GetTreeHierarchyRequest) ([]*inv_v1.GetTreeHierarchyResponse_TreeNode, error)
	GetSitesPerRegion(context.Context, *inv_v1.GetSitesPerRegionRequest) (*inv_v1.GetSitesPerRegionResponse, error)
	// TestingOnlySetClient allows to set the internal inventory service client
	// API for testing purposes only.
	TestingOnlySetClient(inv_v1.InventoryServiceClient)
	// TestGetClientCache allows access to client cache for testing cache content.
	TestGetClientCache() *cache.InventoryCache
	// TestGetClientCacheUUID allows access to client cache for testing cache content.
	TestGetClientCacheUUID() *cache.InventoryCache
	// DeleteAllResources deletes resources of given kind and tenant
	DeleteAllResources(ctx context.Context, tenantID string, kind inv_v1.ResourceKind, enforce bool) error
}

// isRetryableStreamError checks if a registration error is recoverable and a
// new register retry should be performed.
func isRetryableStreamError(err error) bool {
	if errors.Is(err, io.EOF) {
		zlog.InfraSec().InfraErr(err).Msg("Inventory client stream gracefully disconnected")
		return true
	}

	if code := status.Code(err); code == codes.Unavailable {
		zlog.InfraSec().InfraErr(err).Msg("Inventory client stream unavailable")
		return true
	}

	return false
}

// registerRetryBackoffLoop runs a loop to retry register the inventory client.
// It returns an error in case the max elapsed time of expbackoff was attained,
// the inventory client was terminated, or if the stream error does not allow to
// retry the register.
func (client *inventoryClient) registerRetryBackoffLoop(expbackoff *backoff.ExponentialBackOff) error {
	for {
		// Try to register again
		zlog.InfraSec().Debug().Msgf("Client register retry, elapsed time %v", expbackoff.GetElapsedTime())
		err := client.register()

		// If register ok, break and return nil
		if err == nil {
			zlog.InfraSec().Info().Msg("Client register retry successful")
			return nil
		}

		// Checks if register error is retryable
		if !isRetryableStreamError(err) {
			zlog.InfraSec().InfraErr(err).Msg("Register retry loop finished, stream error not retryable")
			return err
		}

		// Gets next backoff time and checks if it is still valid
		d := expbackoff.NextBackOff()
		if d == backoff.Stop {
			err := inv_errors.Errorfc(codes.DeadlineExceeded, "maximum backoff time elapsed")
			zlog.InfraSec().InfraErr(err).Msg("Register retry loop terminated due to maximum time elapsed")
			return err
		}

		select {
		// Waits/sleeps during backoff time
		case <-time.After(d):
			zlog.InfraSec().Debug().Msgf("Client waited on next register retry for %v", d)

		// Waits for client context to be done
		case <-client.streamCtx.Done():
			err := inv_errors.Errorfc(codes.Canceled, "client context canceled")
			zlog.InfraSec().InfraErr(err).Msg("Register retry loop terminated due to client context cancellation")
			return err
		}
	}
}

// registerRetry is a helper method to be used to perform registration
// retries when the client stream context is closed.
// Once the subscription to events was interrupted or finished.
// It uses an exponential backoff timer to wait between retries.
// Its backoff mechanism is configured with InventoryClientConfig.RegisterMaxElapsedTime.
func (client *inventoryClient) registerRetry() error {
	// Checks if register retry is enabled, otherwise returns error.
	if !client.cfg.EnableRegisterRetry {
		err := inv_errors.Errorfc(codes.Internal, "register retry not enabled")
		zlog.InfraSec().InfraErr(err).Msg("could not retry register")
		return err
	}

	expbackoff := backoff.NewExponentialBackOff()
	expbackoff.MaxElapsedTime = client.cfg.RegisterMaxElapsedTime

	err := client.registerRetryBackoffLoop(expbackoff)
	if err != nil {
		return err
	}

	return nil
}

// streamClosedHandler is a helper function of inventory client.
// It invalidates the clientUUID once the subscription stream is closed.
// I.e., no client should make calls without a valid UUID.
func (client *inventoryClient) streamClosedHandler() {
	select {
	case <-client.stream.Context().Done():
		client.uuidMutex.Lock()
		client.clientUUID = ""
		client.uuidMutex.Unlock()
		zlog.InfraSec().Info().Msg("Inventory client stream disconnected, client unregistered")
	default:
		return
	}
}

func (client *inventoryClient) handleStreamErr(err error) error {
	zlog.InfraSec().Info().Msg("Handling Inventory client stream error")

	// Invalidate client UUID
	client.streamClosedHandler()

	// server canceled the stream, return error to end event loop
	if inv_errors.IsCanceled(err) {
		zlog.InfraSec().Info().Msg("Inventory client stream canceled")
		return err
	}

	select {
	// If stream context is done, go for retry (if enabled).
	case <-client.stream.Context().Done():
		err = client.registerRetry()
		// registerRetry returns error if retry fails or if retry is not enabled
		if err != nil {
			zlog.InfraSec().InfraErr(err).Msg("Inventory client disconnected")
			return err
		}
		// registerRetry went well, no error to report
		return nil

	// By default, if error happened and stream context is not done,
	// return it to end event loop
	default:
		zlog.InfraSec().InfraErr(err).Msg("Inventory client disconnected")
		return err
	}
}

// eventContextTracing adds trace from stream header metadata to the context
// when tracing is enabled.
func (client *inventoryClient) eventContextTracing() context.Context {
	ctx := client.stream.Context()
	if client.cfg.EnableTracing {
		// Gets tracing info from stream header into metadata
		md, err := client.stream.Header()
		if err != nil {
			zlog.InfraErr(err).Msgf("could not read stream header metadata")
		}
		// Creates new context with tracing info from metadata
		ctx = metadata.NewIncomingContext(client.stream.Context(), md)
		// Sets a new span to the watch
		ctx = tracing.StartTraceFromRemote(ctx, client.cfg.Name, "watch")
		tracing.StopTrace(ctx)
	}
	return ctx
}

// eventHandler will listen for inventory events and enqueue them internal
// channel, which can be accessed with EventChannel. This function blocks until
// the context is canceled or the server closes the connection. It is safe
// to have a goroutine call this function and another goroutine calling Find,
// Get, Create or Update at the same time, but it is not safe to call eventHandler
// in different goroutines.
func (client *inventoryClient) eventHandler() {
	defer client.cfg.Wg.Done()
	defer close(client.cfg.Events) // Only the sender can safely close a channel.

	for {
		// Wait for next event.
		event, err := client.stream.Recv()
		// Checks stream error for retry register (if enabled).
		if err != nil {
			streamErr := client.handleStreamErr(err)
			if streamErr != nil {
				// Cannot retry register or failed doing it, need to stop event
				// loop handler.
				zlog.InfraSec().Info().Msg("event stream handler loop finished")
				return
			}
			// Tried register retry and succeeded, need to jump to next loop,
			// because event is nil.
			continue
		}
		// Adds tracing, if enabled, to the event context.
		ctx := client.eventContextTracing()

		if client.isClientCacheEnabled() {
			zlog.InfraSec().Debug().Msgf("subscribe event notify received, invalidating cache <%v>", event)
			client.getClientCache().InvalidateCacheEntryByResource(event.Resource)
		}
		if client.isClientUUIDCacheEnabled() {
			zlog.InfraSec().Debug().Msgf("subscribe event notify received, invalidating UUID cache <%v>", event)
			client.getClientCacheUUID().InvalidateCacheByEvent(event.GetEventKind(), event.GetResource())
		}
		// check if event notification is requested by cache or app.
		// if not by app then it is cache subscribed event notification.
		if !isEventAppSubscribed(client, event.Resource) {
			zlog.InfraSec().Debug().Msgf("subscribe notify received, not subscribed by app, ignoring <%v>", event)
			continue
		}
		// Put event in queue or drop. Non-blocking.
		select {
		case client.cfg.Events <- &WatchEvents{ctx, event}:
		default:
			zlog.Warn().Msgf("dropping event, queue is full: clientName=%s, clientUUID=%s",
				client.cfg.Name, client.clientUUID)
			zlog.Debug().Msgf("dropped event: clientName=%s, clientUUID=%s, event=%v",
				client.cfg.Name, client.clientUUID, event)
		}
	}
}

func (client *inventoryClient) Close() error {
	client.streamCancel()
	err := client.connection.Close()
	// Close might be called multiple times, we ignore this error.
	if s, ok := status.FromError(err); ok && s.Code() == codes.Canceled {
		err = nil
	}
	return inv_errors.Wrap(err)
}

// contextDoneHandler waits until the user-provided context ctx is done and initiates
// stream channel shutdown by calling Close.
func (client *inventoryClient) contextDoneHandler() {
	defer client.cfg.Wg.Done()
	<-client.streamCtx.Done()
	err := client.Close()
	zlog.Info().Err(err).Msg("stopping inventory client")
}

// connect creates a gRPC connection to a server.
func connect(
	address string,
	caPath, certPath, keyPath string,
	insec bool,
	opts ...grpc.DialOption,
) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn

	if insec {
		dialOpt := grpc.WithTransportCredentials(insecure.NewCredentials())
		opts = append(opts, dialOpt)
	} else {
		if caPath == "" || certPath == "" || keyPath == "" {
			err := inv_errors.Errorf("CaCertPath %s or TlsCerPath %s or TlsKeyPath %s were not provided",
				caPath, certPath, keyPath,
			)
			zlog.Fatal().Err(err).Msgf("CaCertPath %s or TlsCerPath %s or TlsKeyPath %s were not provided\n",
				caPath, certPath, keyPath,
			)
			return nil, err
		}
		// setting secure gRPC connection
		creds, err := cert.HandleCertPaths(caPath, keyPath, certPath, true)
		if err != nil {
			zlog.Fatal().Err(err).Msgf("an error occurred while loading credentials to server %v, %v, %v: %v\n",
				caPath, certPath, keyPath, err,
			)
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	// Adds keepalive options to the client connection.
	keepAliveOpts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                clientKeepAliveTime,
			Timeout:             clientKeepAliveTimeout,
			PermitWithoutStream: true,
		}),
	}
	opts = append(opts, keepAliveOpts...)

	// if testing, use a bufconn, otherwise TCP
	var err error
	if address == "bufconn" {
		// Use passthrough gRPC scheme to avoid error:
		// "failed to exit idle mode: dns resolver: missing address"
		conn, err = grpc.NewClient("passthrough://bufnet", opts...)
	} else {
		conn, err = grpc.NewClient(address, opts...)
	}
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msgf("Unable to dial connection to inventory client address %s", address)
		return nil, inv_errors.Wrap(err)
	}
	return conn, nil
}

// SecurityConfig security configuration for inventoryClient.
// CaPath, CertPath, KeyPath to be set if needed.
// Insecure determines whether to use TLS and requires the above fields to be set.
type SecurityConfig struct {
	CaPath   string
	KeyPath  string
	CertPath string
	Insecure bool
}

// Inventory Client Cache config.
type InvClientCacheConfig struct {
	// Enable/Disable Cache.
	EnableCache bool
	// Enable/Disable UUID->Host Cache.
	EnableUUIDCache bool
	// Cache Entry stale time in second.
	StaleTime time.Duration
	// Cache Entry stale time offset in percentage.
	StateTimeOffset int
	// Cache susbcription resources
	ResourceKinds []inv_v1.ResourceKind
}

// InventoryClientConfig comprises a set of inventory client configuration options.
type InventoryClientConfig struct {
	// Name allows registering this client with an unique name.
	Name string
	// Address is the inventory target to connect to.
	Address string
	// Dial options of this client. This might include also the interceptors
	DialOptions []grpc.DialOption
	// Events define the channel to receive subscription events from inventory.
	// Each event is received together with its incoming context, to be
	// used for tracing purposes.
	Events chan *WatchEvents
	// EnableRegisterRetry determines if a control loop to try to register in case the
	// subscription stream is closed.
	// To avoid race conditions, EnableRegisterRetry and AbortOnUnknownClientError should never be enabled together.
	EnableRegisterRetry bool
	// AbortOnUnknownClientError determines if the inventory client should abort on
	// UNKNOWN_CLIENT error received from Inventory. If AbortOnUnknownClientError is enabled,
	// the inventory client will fatal on UNKNOWN_CLIENT error received, causing a crash of the client's user.
	// To avoid race conditions, EnableRegisterRetry and AbortOnUnknownClientError should never be enabled together.
	AbortOnUnknownClientError bool
	// RegisterMaxElapsedTime is the max time allowed to retry registration procedures
	// in case RegisterRetry is enabled, if set to zero allows registration retries to run indefinitely.
	RegisterMaxElapsedTime time.Duration
	// ClientKind should be set to the appropriate enum value, depending on the type of application.
	ClientKind inv_v1.ClientKind
	// ResourceKinds is a list of resource kinds this client would like to receive
	// updates for.
	ResourceKinds []inv_v1.ResourceKind
	// EnableTracing enables tracing.
	EnableTracing bool
	// EnableMetrics enables client-side gRPC metrics.
	EnableMetrics bool
	// Wg will be unblocked upon termination of client.
	Wg *sync.WaitGroup
	// SecurityConfig security configuration for inventoryClient.
	SecurityCfg *SecurityConfig
	// Inventory Client Cache.
	ClientCache InvClientCacheConfig
}

// Return an error if user does not provide required input.
func validateClientInput(ctx context.Context, cfg InventoryClientConfig) error {
	if ctx == nil {
		zlog.InfraSec().InfraError("context is nil").Msg("")
		return inv_errors.Errorfc(codes.InvalidArgument, "context is nil")
	}
	if cfg.Wg == nil {
		zlog.InfraSec().InfraError("waitgroup is nil").Msg("")
		return inv_errors.Errorfc(codes.InvalidArgument, "waitgroup is nil")
	}
	if cfg.Events == nil {
		zlog.InfraSec().InfraError("events channel is nil").Msg("")
		return inv_errors.Errorfc(codes.InvalidArgument, "events channel is nil")
	}
	if cfg.EnableRegisterRetry && cfg.AbortOnUnknownClientError {
		zlog.InfraSec().InfraError("Both EnableRegisterRetry and AbortOnUnknownClientError cannot be enabled.").Msg("")
		return inv_errors.Errorfc(codes.InvalidArgument,
			"Both EnableRegisterRetry and AbortOnUnknownClientError cannot be enabled.")
	}
	return nil
}

// NewTenantAwareInventoryClient creates a new inventoryClient, establishes a connection to
// inventory and registers it.
// ctx is used for the initial connect and the later bidi stream channel to
// inventory, and can then be used to trigger client shutdown asynchronously. In
// addition, users can call InventoryClient.Close to terminate the gRPC
// connection. Both methods are equivalent and may be used at the same time.
func NewTenantAwareInventoryClient(
	ctx context.Context,
	cfg InventoryClientConfig,
) (TenantAwareInventoryClient, error) {
	// Handle required inputs
	if err := validateClientInput(ctx, cfg); err != nil {
		return nil, err
	}
	// User might not provide dial options
	if cfg.DialOptions == nil {
		cfg.DialOptions = make([]grpc.DialOption, 0)
	}
	if cfg.EnableTracing {
		cfg.DialOptions = append(cfg.DialOptions, grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	}
	if cfg.EnableMetrics {
		cliMetrics := metrics.GetClientMetricsWithLatency()
		// always prepend metrics gRPC interceptor as first element in the client interceptors' list,
		// so that we cover all subsequent interceptors in the measurements.
		cfg.DialOptions = append([]grpc.DialOption{
			grpc.WithChainUnaryInterceptor(cliMetrics.UnaryClientInterceptor()),
		}, cfg.DialOptions...)
	}

	// ToDo remove insec option as default connect mode
	conn, err := connect(
		cfg.Address,
		cfg.SecurityCfg.CaPath,
		cfg.SecurityCfg.CertPath,
		cfg.SecurityCfg.KeyPath,
		cfg.SecurityCfg.Insecure,
		cfg.DialOptions...)
	if err != nil {
		return nil, err
	}

	invSvcClient := inv_v1.NewInventoryServiceClient(conn)
	zlog.Debug().Msgf("Created inventory client to address: %s", cfg.Address)

	cl := &inventoryClient{
		cfg:        &cfg,
		connection: conn,
		invAPI:     invSvcClient,
	}
	cl.streamCtx, cl.streamCancel = context.WithCancel(ctx)

	// initialize cache.
	cl.newInventoryCache()

	// registering client and obtaining UUID
	err = cl.register()
	if err != nil {
		// stream is already cancel. Close the connection only
		cl.connection.Close()
		return nil, err
	}

	// Setup handler for user initiated shutdown.
	cl.cfg.Wg.Add(1)
	go cl.contextDoneHandler()

	// Setup inventory event handler, register retry inside of it.
	cl.cfg.Wg.Add(1)
	go cl.eventHandler()

	return cl, nil
}

// register registers the inventory client on a name and a list of resource kinds.
// It is meant to be used by any register retry go routine that can be called
// once the subscriptions stream context is closed by any unexpected reasons.
// Look at RegisterRetry method for a helper example.
func (client *inventoryClient) register() error {
	zlog := zlog.TraceCtx(client.streamCtx)
	zlog.InfraSec().Info().Msgf("Register inventory client: name %s, clientkind %v, prefixes %s",
		client.cfg.Name, client.cfg.ClientKind, client.cfg.ResourceKinds,
	)

	rKinds := []inv_v1.ResourceKind{}
	rKinds = append(rKinds, client.cfg.ResourceKinds...)
	if client.isClientCacheEnabled() || client.isClientUUIDCacheEnabled() {
		rKinds = append(rKinds, client.cfg.ClientCache.ResourceKinds...)
		rKinds = removeDuplicates(rKinds)
	}

	// Register client by setting up the stream channel.
	req := &inv_v1.SubscribeEventsRequest{
		Name:                    client.cfg.Name,
		Version:                 "0.1.0-dev", // TODO: pull version main.Version
		ClientKind:              client.cfg.ClientKind,
		SubscribedResourceKinds: rKinds,
	}
	stream, err := client.invAPI.SubscribeEvents(client.streamCtx, req)
	if err != nil {
		return inv_errors.Wrap(err)
	}
	// Get our UUID from the first response.
	resp, err := stream.Recv()
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msg("Unable to register inventory client")
		return inv_errors.Wrap(err)
	}
	if resp.ClientUuid == "" {
		zlog.InfraError("Server did not allocate an UUID unable to register inventory client").Msg("")
		return inv_errors.Errorfc(codes.Internal, "Server did not allocate an UUID unable to register inventory client")
	}
	// let's close the send half of the stream as we don't need it
	if err := stream.CloseSend(); err != nil {
		zlog.Warn().Msg("unable to close send")
	}
	client.uuidMutex.Lock()
	client.stream = stream
	client.clientUUID = resp.ClientUuid
	zlog.InfraSec().Info().Msgf("Registered inventory client with UUID: %s", resp.ClientUuid)
	client.uuidMutex.Unlock()

	return nil
}

func (client *inventoryClient) List(
	ctx context.Context,
	filter *inv_v1.ResourceFilter,
) (*inv_v1.ListResourcesResponse, error) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Debug().Msgf("List inventory resources filter: %s", filter.String())

	if err := client.clientIsRegistered(); err != nil {
		return nil, err
	}

	// check cache before querying inventory
	if client.isClientCacheEnabled() {
		if obj, err := client.getClientCache().GetResourceByFilter(filter); err == nil {
			return obj, nil
		}
	}

	object := inv_v1.ListResourcesRequest{
		ClientUuid: client.clientUUID,
		Filter:     filter,
	}
	objs, err := client.invAPI.ListResources(ctx, &object)
	if err != nil {
		zlog.Debug().Err(err).Msg("on List")
		return nil, inv_errors.Wrap(err)
	}

	if len(objs.Resources) == 0 {
		objs.Resources = make([]*inv_v1.GetResourceResponse, 0)
	} else if client.isClientCacheEnabled() {
		// store in cache
		client.getClientCache().StoreResourceByFilter(filter, objs)
	}

	return objs, nil
}

func (client *inventoryClient) ListAll(
	ctx context.Context,
	filter *inv_v1.ResourceFilter,
) ([]*inv_v1.Resource, error) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Debug().Msgf("List all inventory resources filter: %s", filter.String())
	if err := client.clientIsRegistered(); err != nil {
		return nil, err
	}

	filterRequest := &inv_v1.ResourceFilter{
		Resource: filter.GetResource(),
		Filter:   filter.GetFilter(),
		Limit:    BatchSize,
		Offset:   0,
		OrderBy:  filter.GetOrderBy(),
	}
	resources := make([]*inv_v1.Resource, 0, BatchSize) // Pre-allocate a slice of at least a batchSize
	hasNext := true
	firstRead := true
	err := error(nil)
	for hasNext {
		var objs *inv_v1.ListResourcesResponse
		objs, err = client.List(ctx, filterRequest)
		//nolint:gocritic // false-positive, no need for switch statement due to default option
		if firstRead && len(objs.GetResources()) == 0 {
			zlog.Debug().Msgf("no resources found for filter: %v", &filterRequest)
			break
		} else if !firstRead && len(objs.GetResources()) == 0 {
			zlog.Warn().Msgf("no resources found for filter (%v), expect to return an incoherent state", &filterRequest)
			break
		} else if err != nil {
			zlog.Debug().Err(err).Msg("on ListAll")
			// on errors, return partial result.
			// This covers also the case of interleaved deletes. In this case we could get a "not-found" error also when
			// getting a page different from the first.
			break
		}
		if firstRead {
			firstRead = false
		}
		for _, r := range objs.GetResources() {
			resources = append(resources, r.GetResource())
		}
		hasNext = objs.HasNext
		// Limit never changes, it's the number of entries returned
		filterRequest.Offset += BatchSize
	}

	return removeDuplicates(resources), err
}

func (client *inventoryClient) Find(
	ctx context.Context,
	filter *inv_v1.ResourceFilter,
) (*inv_v1.FindResourcesResponse, error) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Debug().Msgf("Find inventory resources filter: %s", filter.String())

	if err := client.clientIsRegistered(); err != nil {
		return nil, err
	}

	object := inv_v1.FindResourcesRequest{
		ClientUuid: client.clientUUID,
		Filter:     filter,
	}
	objs, err := client.invAPI.FindResources(ctx, &object)
	if err != nil {
		zlog.Debug().Err(err).Msg("on Find")
		return nil, inv_errors.Wrap(err)
	}

	if len(objs.Resources) == 0 {
		objs.Resources = make([]*ResourceTenantIDCarrier, 0)
	}

	return objs, nil
}

func (client *inventoryClient) FindAll(
	ctx context.Context,
	filter *inv_v1.ResourceFilter,
) ([]*ResourceTenantIDCarrier, error) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Debug().Msgf("Find all inventory resources filter: %s", filter.String())
	if err := client.clientIsRegistered(); err != nil {
		return nil, err
	}

	filterRequest := inv_v1.FindResourcesRequest{
		ClientUuid: client.clientUUID,
		Filter: &inv_v1.ResourceFilter{
			Resource: filter.GetResource(),
			Filter:   filter.GetFilter(),
			Limit:    BatchSize,
			Offset:   0,
			OrderBy:  filter.GetOrderBy(),
		},
	}
	resources := make([]*ResourceTenantIDCarrier, 0, BatchSize) // Pre-allocate a slice of at least a batchSize
	hasNext := true
	firstRead := true
	err := error(nil)
	for hasNext {
		var objs *inv_v1.FindResourcesResponse
		objs, err = client.invAPI.FindResources(ctx, &filterRequest)
		//nolint:gocritic // false-positive, no need for switch statement due to default option
		if firstRead && len(objs.GetResources()) == 0 {
			zlog.Debug().Msgf("no resources found for filter: %s", &filterRequest)
			break
		} else if !firstRead && len(objs.GetResources()) == 0 {
			zlog.Warn().Msgf("no resources found for filter (%s), expect to return an incoherent state", &filterRequest)
			break
		} else if err != nil {
			zlog.Debug().Err(err).Msg("on FindAll")
			// on errors, return partial result.
			// This covers also the case of interleaved deletes. In this case we could get a "not-found" error also when
			// getting a page different from the first.
			break
		}
		if firstRead {
			firstRead = false
		}
		resources = append(resources, objs.GetResources()...)
		hasNext = objs.HasNext
		// Limit never changes, it's the number of entries returned
		filterRequest.Filter.Offset += BatchSize
	}

	return removeDuplicates(resources), err
}

func (client *inventoryClient) Get(
	ctx context.Context,
	tenantID, resourceID string,
) (*inv_v1.GetResourceResponse, error) {
	zlog.Debug().Msgf("Get inventory resource tenantID: %s, ID: %s", tenantID, resourceID)

	if err := client.clientIsRegistered(); err != nil {
		return nil, err
	}

	// check cache before querying inventory
	if client.isClientCacheEnabled() {
		if obj, err := client.getClientCache().GetResourceByID(tenantID, resourceID); err == nil {
			return obj, nil
		}
	}

	object := inv_v1.GetResourceRequest{
		ClientUuid: client.clientUUID,
		ResourceId: resourceID,
		TenantId:   tenantID,
	}
	obj, err := client.invAPI.GetResource(ctx, &object)
	if err != nil {
		zlog.Debug().Err(err).Msg("on Get")
		return nil, inv_errors.Wrap(err)
	}

	// store in cache
	if client.isClientCacheEnabled() {
		client.getClientCache().StoreResourceByID(obj)
	}
	return obj, nil
}

func (client *inventoryClient) Create(
	ctx context.Context,
	tenantID string,
	res *inv_v1.Resource,
) (*inv_v1.Resource, error) {
	zlog.Debug().Msgf("Create inventory resource: tenantID: %s, resource: %s", tenantID, res.Resource)

	if err := client.clientIsRegistered(); err != nil {
		return nil, err
	}

	object := inv_v1.CreateResourceRequest{
		ClientUuid: client.clientUUID,
		Resource:   res,
		TenantId:   tenantID,
	}

	obj, err := client.invAPI.CreateResource(ctx, &object)
	if err != nil {
		zlog.Debug().Err(err).Msg("on Create")
		invErr := client.handleInventoryError(err)
		return nil, invErr
	}

	resID, err := util.GetResourceIDFromResource(obj)
	if err != nil {
		return nil, err
	}

	// invalidate cache entry
	if client.isClientCacheEnabled() {
		client.getClientCache().InvalidateCacheEntryByID(tenantID, resID)
	}

	// If we add any Host subresource or Instance we need to invalidate the linked host.
	if client.isClientUUIDCacheEnabled() {
		client.getClientCacheUUID().InvalidateCacheByEvent(inv_v1.SubscribeEventsResponse_EVENT_KIND_CREATED, res)
	}

	return obj, nil
}

func (client *inventoryClient) Update(
	ctx context.Context,
	tenantID, resourceID string,
	fieldmask *fieldmaskpb.FieldMask,
	res *inv_v1.Resource,
) (*inv_v1.Resource, error) {
	zlog.Debug().Msgf("Update inventory resource: tenantID: %s,  %s", tenantID, res.Resource)

	if err := client.clientIsRegistered(); err != nil {
		return nil, err
	}

	// invalidate cache entry
	if client.isClientCacheEnabled() {
		client.getClientCache().InvalidateCacheEntryByID(tenantID, resourceID)
	}

	object := inv_v1.UpdateResourceRequest{
		ClientUuid: client.clientUUID,
		ResourceId: resourceID,
		FieldMask:  fieldmask,
		Resource:   res,
		TenantId:   tenantID,
	}
	res, err := client.invAPI.UpdateResource(ctx, &object)
	if err != nil {
		zlog.Debug().Err(err).Msg("on Update")
		invErr := client.handleInventoryError(err)
		return nil, invErr
	}

	if client.isClientUUIDCacheEnabled() {
		client.getClientCacheUUID().InvalidateCacheByEvent(inv_v1.SubscribeEventsResponse_EVENT_KIND_UPDATED, res)
	}

	return res, nil
}

func (client *inventoryClient) Delete(
	ctx context.Context,
	tenantID, resourceID string,
) (*inv_v1.DeleteResourceResponse, error) {
	zlog.Debug().Msgf("Delete inventory resource tenantID: %s, ID: %s", tenantID, resourceID)

	if err := client.clientIsRegistered(); err != nil {
		return nil, err
	}

	// invalidate cache entry
	if client.isClientCacheEnabled() {
		client.getClientCache().InvalidateCacheEntryByID(tenantID, resourceID)
	}
	if client.isClientUUIDCacheEnabled() {
		// We don't re-use the manage event, for deleted case, we only need to clean whatever state is in the cache.
		// No need for smart invalidation
		client.getClientCacheUUID().InvalidateCacheUUIDByResourceID(tenantID, resourceID)
	}

	object := inv_v1.DeleteResourceRequest{
		ClientUuid: client.clientUUID,
		ResourceId: resourceID,
		TenantId:   tenantID,
	}
	obj, err := client.invAPI.DeleteResource(ctx, &object)
	if err != nil {
		zlog.Debug().Err(err).Msg("on Delete")
		invErr := client.handleInventoryError(err)
		return nil, invErr
	}
	return obj, nil
}

func (client *inventoryClient) UpdateSubscriptions(
	ctx context.Context,
	tenantID string,
	kinds []inv_v1.ResourceKind,
) error {
	zlog.Debug().Msgf("Update subscriptions: tenantID: %s, kinds: %s", tenantID, kinds)

	if err := client.clientIsRegistered(); err != nil {
		return err
	}

	req := &inv_v1.ChangeSubscribeEventsRequest{
		ClientUuid:              client.clientUUID,
		SubscribedResourceKinds: kinds,
	}
	_, err := client.invAPI.ChangeSubscribeEvents(ctx, req)
	if err != nil {
		zlog.Debug().Err(err).Msg("on change subs")
		invErr := client.handleInventoryError(err)
		return invErr
	}

	return nil
}

// handleInventoryError handles errors returned by inventory.
// In particular, it handles the UNKNOWN_CLIENT error.
// It's currently applied to CREATE, UPDATE and DELETE methods,
// as these are the only methods that can modify the inventory state.
func (client *inventoryClient) handleInventoryError(err error) error {
	if inv_errors.IsUnKnownClient(err) {
		if client.cfg.AbortOnUnknownClientError {
			// Hotfix
			// In summary, sometimes RMs don't re-register to inventory after redeployment or restart.
			// As a consequence, RMs keep getting UNKNOWN_CLIENT error for update operations.
			// If this option is enabled, we can let RMs crash (and restart), so that they can re-register on startup.
			zlog.InfraSec().Fatal().Msg("inventory client is unknown and abort on error enabled, aborting")
		} else {
			return inv_errors.Errorfc(codes.Unavailable,
				"inventory client is not registered: %s", err.Error())
		}
	}

	return inv_errors.Wrap(err)
}

// clientIsRegistered verifies if the client UUID is valid,
// i.e., if it is not invalid due to a subscription stream be closed.
func (client *inventoryClient) clientIsRegistered() error {
	client.uuidMutex.Lock()
	defer client.uuidMutex.Unlock()
	// Doing any rpc in this condition will fail. Short circuit and failfast
	if client.clientUUID == "" {
		// clientIsRegistered is called as pre-requisite in each API function. If retry is not enabled RMs
		// remain in a weird state getting unavailable even if it specified the abort on unknown.
		if client.cfg.AbortOnUnknownClientError {
			zlog.InfraSec().Fatal().Msg("inventory client is not registered and abort on error enabled, aborting")
		}
		zlog.InfraError("service unavailable - inventory client is not registered").Msg("")
		return inv_errors.Errorfc(codes.Unavailable, "inventory client is not registered")
	}
	return nil
}

func (client *inventoryClient) TestingOnlySetClient(c inv_v1.InventoryServiceClient) {
	client.invAPI = c
}

func removeDuplicates[T comparable](slice []T) []T {
	keys := make(map[T]struct{}, len(slice))
	noDupl := make([]T, 0, len(slice))
	for _, v := range slice {
		if _, ok := keys[v]; !ok {
			keys[v] = struct{}{}
			noDupl = append(noDupl, v)
		}
	}
	return noDupl
}

func isEventAppSubscribed(client *inventoryClient, r *inv_v1.Resource) bool {
	resKind := util.GetResourceKindFromResource(r)

	for _, v := range client.cfg.ResourceKinds {
		if v == resKind {
			return true
		}
	}
	return false
}

func (client *inventoryClient) ListInheritedTelemetryProfiles(
	ctx context.Context,
	tenantID string,
	inheritBy *inv_v1.ListInheritedTelemetryProfilesRequest_InheritBy,
	filter string,
	orderBy string,
	limit, offset uint32,
) (*inv_v1.ListInheritedTelemetryProfilesResponse, error) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Debug().Msgf("ListRenderedTelemetryProfiles: inheritBy=%v, filter=%s", inheritBy, filter)

	if err := client.clientIsRegistered(); err != nil {
		zlog.Debug().Err(err).Msg("on ListRenderedTelemetryProfiles")
		return nil, err
	}

	request := &inv_v1.ListInheritedTelemetryProfilesRequest{
		ClientUuid: client.clientUUID,
		InheritBy:  inheritBy,
		Filter: &inv_v1.ResourceFilter{
			Resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_TelemetryProfile{
					TelemetryProfile: &telemetry_v1.TelemetryProfile{},
				},
			},
			Limit:   limit,
			Offset:  offset,
			Filter:  filter,
			OrderBy: orderBy,
		},
		TenantId: tenantID,
	}

	obj, err := client.invAPI.ListInheritedTelemetryProfiles(ctx, request)
	if err != nil {
		zlog.Debug().Err(err).Msg("on ListInheritedTelemetryProfiles")
		return nil, inv_errors.Wrap(err)
	}

	return obj, nil
}

func (client *inventoryClient) GetHostByUUID(ctx context.Context, tenantID, uuid string) (*computev1.HostResource, error) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Info().Msgf("GetHostByUUID: tenantID=%s, uuid=%v", tenantID, uuid)

	if err := client.clientIsRegistered(); err != nil {
		zlog.Debug().Err(err).Msg("on GetHostByUUID")
		return nil, err
	}

	if client.isClientUUIDCacheEnabled() {
		if obj, err := client.getClientCacheUUID().GetHostByUUID(tenantID, uuid); err == nil {
			return obj, nil
		}
	}

	filter := &inv_v1.ResourceFilter{
		Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
		Filter: filters.NewBuilder().
			And(filters.ValEq(hostresource.FieldUUID, uuid)).
			And(filters.ValEq(hostresource.FieldTenantID, tenantID)).Build(),
		Offset: 0,
		Limit:  1,
	}
	objs, err := client.List(ctx, filter)
	if err != nil {
		zlog.Debug().Err(err).Msg("on GetHostByUUID")
		return nil, inv_errors.Wrap(err)
	}

	err = util.CheckListOutputIsSingular(objs.GetResources())
	if err != nil {
		zlog.InfraSec().InfraErr(err).
			Msgf("Obtained non-singular Host resource: UUID=%s, totalElem=%v", uuid, objs.GetTotalElements())
		return nil, err
	}
	host := objs.GetResources()[0].GetResource().GetHost()

	if client.isClientUUIDCacheEnabled() {
		client.getClientCacheUUID().StoreHostByUUID(uuid, host)
	}
	return host, nil
}

func (client *inventoryClient) GetTreeHierarchy(ctx context.Context, request *inv_v1.GetTreeHierarchyRequest) (
	[]*inv_v1.GetTreeHierarchyResponse_TreeNode, error,
) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Debug().Msgf("GetHierarchy: request=%v", request)

	if err := client.clientIsRegistered(); err != nil {
		zlog.Debug().Err(err).Msg("on GetHierarchy")
		return nil, err
	}
	// Populate the client UUID
	request.ClientUuid = client.clientUUID
	tree, err := client.invAPI.GetTreeHierarchy(ctx, request)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msg("on GetHierarchy")
		return nil, err
	}
	return tree.Tree, nil
}

func (client *inventoryClient) GetSitesPerRegion(ctx context.Context, request *inv_v1.GetSitesPerRegionRequest) (
	*inv_v1.GetSitesPerRegionResponse, error,
) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Debug().Msgf("GetSitesPerRegion: request=%v", request)

	if err := client.clientIsRegistered(); err != nil {
		zlog.Debug().Err(err).Msg("on GetSitesPerRegion")
		return nil, err
	}
	// Populate the client UUID
	request.ClientUuid = client.clientUUID
	resp, err := client.invAPI.GetSitesPerRegion(ctx, request)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msg("on GetSitesPerRegion")
		return nil, err
	}
	return resp, nil
}

// isClientCacheEnabled checks if client config has cache enabled.
func (client *inventoryClient) isClientCacheEnabled() bool {
	return client.cfg.ClientCache.EnableCache
}

// isClientCacheEnabled checks if client config has cache enabled.
func (client *inventoryClient) isClientUUIDCacheEnabled() bool {
	return client.cfg.ClientCache.EnableUUIDCache
}

// newInventoryCache initializes client cache if enabled in config.
func (client *inventoryClient) newInventoryCache() {
	if client.cfg.ClientCache.EnableUUIDCache && client.cacheUUID == nil {
		client.cacheUUID = cache.NewInventoryCache(client.cfg.ClientCache.StaleTime, client.cfg.ClientCache.StateTimeOffset)
		// The assumption here is that GetCacheUUIDSubscriptionResourceKind will always be a subset of
		// GetCacheSusbcriptionResourceKind.
		// If this doesn't apply, we need to rewrite this code
		client.cfg.ClientCache.ResourceKinds = client.cache.GetCacheUUIDSubscriptionResourceKind()
	}
	if client.cfg.ClientCache.EnableCache && client.cache == nil {
		client.cache = cache.NewInventoryCache(client.cfg.ClientCache.StaleTime, client.cfg.ClientCache.StateTimeOffset)
		client.cfg.ClientCache.ResourceKinds = client.cache.GetCacheSusbcriptionResourceKind()
	}
}

// getClientCache returns cache object.
func (client *inventoryClient) getClientCache() *cache.InventoryCache {
	return client.cache
}

// getClientCache returns cache object.
func (client *inventoryClient) getClientCacheUUID() *cache.InventoryCache {
	return client.cacheUUID
}

// TestGetClientCache returns cache object.
func (client *inventoryClient) TestGetClientCache() *cache.InventoryCache {
	return client.cache
}

// TestGetClientCacheUUID returns cacheUUID object.
func (client *inventoryClient) TestGetClientCacheUUID() *cache.InventoryCache {
	return client.cacheUUID
}

func (client *inventoryClient) DeleteAllResources(
	ctx context.Context, tenantID string, kind inv_v1.ResourceKind, enforce bool,
) error {
	_, err := client.invAPI.DeleteAllResources(ctx, &inv_v1.DeleteAllResourcesRequest{
		ClientUuid:   client.clientUUID,
		ResourceKind: kind,
		TenantId:     tenantID,
		Enforce:      enforce,
	})
	return err
}
