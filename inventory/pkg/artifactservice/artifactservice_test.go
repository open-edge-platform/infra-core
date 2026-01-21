// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//nolint:testpackage // testing internal functions
package artifactservice

import (
	"context"
	"flag"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	serverAddress = "127.0.0.1:61587"

	testDigest       = "TEST_DIGEST"
	testFile         = "TEST_FILE"
	testProfileName  = "edge-microvisor-toolkit"
	testTag          = "1.0.0"
	testImageVersion = "22.04.5"

	// OCI tags example.
	validTags = `
	{"name": "edge-microvisor-toolkit",
	"tags": ["tag1","tag2"]
	}`

	// OCI tags example without the mandatory "tags" key.
	noTagsList = `
	{"name": "edge-microvisor-toolkit",
	}`

	// OCI image manifest example from OCI repo, used by OSRM and DKAM to retrieve the image content.
	// Note: the size should match the string size of validENProfileManifest.
	validManifest = `
      {"schemaVersion": 2,
	  "mediaType": "application/vnd.oci.image.manifest.v1+json",
	  "config": {
		"mediaType": "application/vnd.intel.orch.en",
		"digest": "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
		"size": 2
	  },
	  "layers": [{
		  "mediaType": "application/vnd.oci.image.layer.v1.tar",
		  "digest": "` + testDigest + `",
		  "size": 718,
		  "annotations": {
			"org.opencontainers.image.title": "` + testFile + `"
      }}],
	  "annotations": {
		"org.opencontainers.image.created": "2024-09-09T14:43:10Z"
	  }}`

	// OCI image manifest example with corrupted digest in Layers.
	corruptedDigest = `
		{"schemaVersion":2,"mediaType":"application/vnd.oci.image.manifest.v1+json",
		"config":{"mediaType":"application/vnd.intel.orch.en",
		"digest":"sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a","size":2},
		"layers":[{
			"mediaType":"application/vnd.oci.image.layer.v1.tar",
			"digest":"corrupted digest",
			"size":24800,
		    "annotations": {
			  "org.opencontainers.image.title": "` + testFile + `"
		}}],
		"annotations":{"org.opencontainers.image.created":"2024-03-26T10:32:25Z"}}`

	// OCI image manifest example with no annotation in Layers.
	noAnnotationLayer = `
		{"schemaVersion":2,"mediaType":"application/vnd.oci.image.manifest.v1+json",
		"config":{"mediaType":"application/vnd.intel.orch.en",
		"digest":"sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a","size":2},
		"layers":[{
			"mediaType":"application/vnd.oci.image.layer.v1.tar",
			"digest":"` + testDigest + `",
			"size":24800
		}],
		"annotations":{"org.opencontainers.image.created":"2024-03-26T10:32:25Z"}}`

	// OCI image manifest example with corrupted json.
	corruptedJSON = `
      {"schemaVersion": 2,
	  "layers": [
		  "mediaType": "application/vnd.oci.image.layer.v1.tar",
      }}],
	  "annotations": {
	  }}`

	// OS profile manifest example.
	validENProfileManifest = `
      appVersion: apps/v1
      metadata:
        release: 24.11.0-dev
        version: ` + testTag + `
      spec:
        name: Edge Microvisor Toolkit
        type: OPERATING_SYSTEM_TYPE_IMMUTABLE
        provider: OS_PROVIDER_KIND_INFRA
        profileName: ` + testProfileName + `
        osImageUrl: files-edge-orch/microvisor/iso/edge-microvisor-toolkit:<build-commit>
        osImageSha256: 76423945c97fddd415fa17610c7472b07c46d6758d42f4f706f1bbe972f51155
        osImageVersion: ` + testImageVersion + `
        securityFeature: SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION
        platformBundle:
          kernelCmdlineParams: ""	
          artifact: edge-orch/edge-node/file/profile-scripts
          artifactVersion: 1.0.1`
)

func initArtifactTestServer(ociImgManifest, enProfileManifest, enProfileRepo, profileName, tag, digest string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/"+enProfileRepo+profileName+"/manifests/"+tag, func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(ociImgManifest))
	})
	mux.HandleFunc("/v2/"+enProfileRepo+profileName+"/blobs/"+digest, func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(enProfileManifest))
	})
	// return httptest.NewServer(mux)
	// Listen on a specific port
	lc := &net.ListenConfig{}
	listener, err := lc.Listen(context.Background(), "tcp", serverAddress)
	if err != nil {
		zlog.Fatal().Msgf("Failed to listen on %s: %v", serverAddress, err)
	}

	// Create an unstarted test server
	testServer := httptest.NewUnstartedServer(mux)

	// Use the custom listener
	testServer.Listener = listener
	testServer.Start()

	return testServer
}

func initTagsTestServer(tag, profileRepo, profileName string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/"+profileRepo+profileName+"/tags/list", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(tag))
	})

	// Listen on a specific port
	lc := &net.ListenConfig{}
	listener, err := lc.Listen(context.Background(), "tcp", serverAddress)
	if err != nil {
		zlog.Fatal().Msgf("Failed to listen on %s: %v", serverAddress, err)
	}

	// Create an unstarted test server
	testServer := httptest.NewUnstartedServer(mux)

	// Use the custom listener
	testServer.Listener = listener
	testServer.Start()

	return testServer
}

func TestMain(m *testing.M) {
	// Only needed to suppress the error
	flag.String(
		"policyBundle",
		"/rego/policy_bundle.tar.gz",
		"Path of policy rego file",
	)
	flag.Parse()

	os.Setenv(RsProxyRegistryEnv, serverAddress+"/")
	run := m.Run() // run all tests
	os.Unsetenv(RsProxyRegistryEnv)
	os.Exit(run)
}

func Test_DownloadArtifacts(t *testing.T) {
	type args struct {
		ctx        context.Context
		repository string
		tag        string
	}
	tests := []struct {
		name              string
		args              args
		ociImgManifest    string
		enProfileManifest string
		contentSize       int
		digest            string
		wantErr           bool
	}{
		{
			name: "Happy path",
			args: args{
				ctx:        context.Background(),
				repository: testProfileName,
				tag:        testTag,
			},
			ociImgManifest:    validManifest,
			enProfileManifest: validENProfileManifest,
			digest:            testDigest,
			wantErr:           false,
		},
		{
			name: "Failed to get repository client",
			args: args{
				ctx:        context.Background(),
				repository: "invalidRepo",
				tag:        testTag,
			},
			ociImgManifest:    validManifest,
			enProfileManifest: validENProfileManifest,
			digest:            testDigest,
			wantErr:           true,
		},
		{
			name: "Corrupted OCI image manifest",
			args: args{
				ctx:        context.Background(),
				repository: testProfileName,
				tag:        testTag,
			},
			ociImgManifest:    corruptedJSON,
			enProfileManifest: validENProfileManifest,
			digest:            testDigest,
			wantErr:           true,
		},
		{
			name: "Corrupted digest in layer",
			args: args{
				ctx:        context.Background(),
				repository: testProfileName,
				tag:        testTag,
			},
			ociImgManifest:    corruptedDigest,
			enProfileManifest: validENProfileManifest,
			digest:            testDigest,
			wantErr:           true,
		},
		{
			name: "Missing annotation in layer",
			args: args{
				ctx:        context.Background(),
				repository: testProfileName,
				tag:        testTag,
			},
			ociImgManifest:    noAnnotationLayer,
			enProfileManifest: validENProfileManifest,
			digest:            testDigest,
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svr := initArtifactTestServer(tt.ociImgManifest, tt.enProfileManifest, tt.args.repository, "", tt.args.tag, tt.digest)
			defer svr.Close()

			artifacts, err := DownloadArtifacts(tt.args.ctx, tt.args.repository, tt.args.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadArtifacts() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				assert.Equal(t, tt.digest, (*artifacts)[0].Digest)
				assert.Equal(t, len(tt.enProfileManifest), int((*artifacts)[0].Size))
				assert.NotEmpty(t, (*artifacts)[0].Data)
			} else {
				assert.NotEqual(t, status.Code(err), codes.Unknown)
			}
		})
	}
}

func Test_GetRepositoryTags(t *testing.T) {
	type args struct {
		ctx        context.Context
		repository string
	}

	tests := []struct {
		name         string
		args         args
		tagList      string
		validTagList []string
		wantErr      bool
	}{
		{
			name: "Happy path",
			args: args{
				ctx:        context.Background(),
				repository: testProfileName,
			},
			tagList:      validTags,
			validTagList: []string{"tag1", "tag2"},
			wantErr:      false,
		},
		{
			name: "Failed to get repository client",
			args: args{
				ctx:        context.Background(),
				repository: "invalidRepo",
			},
			tagList:      validTags,
			validTagList: []string{"tag1", "tag2"},
			wantErr:      true,
		},
		{
			name: "Corrupted Json",
			args: args{
				ctx:        context.Background(),
				repository: testProfileName,
			},
			tagList:      corruptedJSON,
			validTagList: []string{"tag1", "tag2"},
			wantErr:      true,
		},
		{
			name: "Missing tags/list",
			args: args{
				ctx:        context.Background(),
				repository: testProfileName,
			},
			tagList:      noTagsList,
			validTagList: []string{"tag1", "tag2"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svr := initTagsTestServer(tt.tagList, tt.args.repository, "")
			defer svr.Close()

			tags, err := GetRepositoryTags(tt.args.ctx, tt.args.repository)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepositoryTags() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				assert.Equal(t, tt.validTagList, tags)
				assert.Equal(t, len(tt.validTagList), len(tags))
				assert.NotEmpty(t, tags)
			} else {
				assert.NotEqual(t, status.Code(err), codes.Unknown)
			}
		})
	}
}

// Test_RsProxyEnv tests the value of env var RsProxyRegistryEnv.
// This was initially set by sync.Once in the first call to GetRsProxyRegistryAddress().
// It was done as part of the call "os.Setenv(RsProxyRegistryEnv, serverAddress+"/")".
func Test_RsProxyEnv(t *testing.T) {
	t.Run("RsProxyEnv", func(t *testing.T) {
		rsProxyRegistryAddr := GetRsProxyRegistryAddress()
		assert.Equal(t, serverAddress+"/", rsProxyRegistryAddr)

		t.Setenv(RsProxyRegistryEnv, "test")
		assert.NotEqual(t, "test", GetRsProxyRegistryAddress())
		assert.Equal(t, serverAddress+"/", GetRsProxyRegistryAddress())

		t.Setenv(RsProxyRegistryEnv, "test2")
		assert.NotEqual(t, "test2", GetRsProxyRegistryAddress())
		assert.Equal(t, serverAddress+"/", GetRsProxyRegistryAddress())
	})
}
