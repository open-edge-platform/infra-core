// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package artifactservice

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	inv_errors "github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
)

const (
	maxIdleConns    = 10
	idleConnTimeout = 30

	defaultRsProxyRegistry = "rs-proxy.rs-proxy.svc.cluster.local:8081/"
	RsProxyRegistryEnv     = "RSPROXY_ADDRESS"
)

var (
	zlog = logging.GetLogger("ArtifactService")

	DefaultArtService ArtifactService = &artifactService{
		client: &http.Client{
			Transport: &http.Transport{
				Proxy:             http.ProxyFromEnvironment,
				ForceAttemptHTTP2: false,
				MaxIdleConns:      maxIdleConns,
				IdleConnTimeout:   idleConnTimeout,
			},
		},
	}
	rsProxyRegistry = defaultRsProxyRegistry

	once sync.Once
)

// GetRsProxyRegistryAddress returns the address of the Release Service proxy registry.
// It gets it from the environment variable RsProxyRegistryEnv, or returns the default address.
func GetRsProxyRegistryAddress() string {
	once.Do(func() {
		rsProxyRegistry = os.Getenv(RsProxyRegistryEnv)
		if rsProxyRegistry == "" {
			zlog.Debug().Msgf("RsProxyRegistryEnv %s not set, using default address %s",
				RsProxyRegistryEnv, defaultRsProxyRegistry)
			rsProxyRegistry = defaultRsProxyRegistry
		}
	})
	return rsProxyRegistry
}

// ArtifactService provides functionality for downloading artifacts from the Release Service.
type ArtifactService interface {
	// GetRepositoryTags retrieves all available tags for a specified repository.
	GetRepositoryTags(ctx context.Context, repository string) ([]string, error)
	// DownloadArtifacts retrieves the artifacts for a specified repository and tag.
	// It returns the raw data as a slice of Artifact structs.
	DownloadArtifacts(ctx context.Context, repository, tag string) (*[]Artifact, error)
}

type artifactService struct {
	client *http.Client
}

func GetRepositoryTags(ctx context.Context, repository string) ([]string, error) {
	return DefaultArtService.GetRepositoryTags(ctx, repository)
}

func DownloadArtifacts(ctx context.Context, repository, tag string) (*[]Artifact, error) {
	return DefaultArtService.DownloadArtifacts(ctx, repository, tag)
}

type Artifact struct {
	Name      string
	MediaType string
	Digest    string
	Size      int64
	Data      []byte
}

func (as *artifactService) GetRepositoryTags(ctx context.Context, repository string) ([]string, error) {
	reference := GetRsProxyRegistryAddress() + repository
	repoClient, err := getRepoClient(as.client, reference)
	if err != nil {
		zlog.InfraSec().Error().Err(err).Msg("Failed to get repository client")
		return []string{}, err
	}
	tags, err := registry.Tags(ctx, repoClient)
	if err != nil {
		zlog.InfraSec().Error().Err(err).Msg("Failed to get tags")
		return nil, inv_errors.Wrap(err)
	}
	return tags, nil
}

func (as *artifactService) DownloadArtifacts(ctx context.Context, repository, tag string) (*[]Artifact, error) {
	reference := GetRsProxyRegistryAddress() + repository
	repoClient, err := getRepoClient(as.client, reference)
	if err != nil {
		zlog.InfraSec().Error().Err(err).Msg("Failed to get repository client")
		return &[]Artifact{}, err
	}

	manifest, err := getOCIImageManifest(ctx, repoClient, tag)
	if err != nil {
		zlog.InfraSec().Error().Err(err).Msg("Error retrieving manifest")
		return &[]Artifact{}, err
	}

	artifacts := []Artifact{}
	for _, layer := range manifest.Layers {
		imgName, exists := layer.Annotations["org.opencontainers.image.title"]
		if !exists {
			err = inv_errors.Errorf("Missing OCI image title annotation")
			zlog.InfraErr(err).Msgf("")
			return &[]Artifact{}, err
		}

		imgContent, err := getOCIImageContent(ctx, repoClient, layer)
		if err != nil {
			zlog.InfraSec().Error().Err(err).Msg("Error retrieving image content")
			return &[]Artifact{}, err
		}

		artifact := Artifact{
			Name:      imgName,
			MediaType: layer.MediaType,
			Digest:    string(layer.Digest),
			Size:      layer.Size,
			Data:      imgContent,
		}

		artifacts = append(artifacts, artifact)
	}
	return &artifacts, nil
}

func getRepoClient(client *http.Client, reference string) (*remote.Repository, error) {
	repo, err := remote.NewRepository(reference)
	if err != nil {
		zlog.InfraSec().Error().Err(err).Msg("Failed to create repository")
		return &remote.Repository{}, inv_errors.Wrap(err)
	}

	// use a custom client to apply proxy settings
	repo.Client = client
	// rs-proxy only supports http
	repo.PlainHTTP = true

	return repo, nil
}

func getOCIImageManifest(ctx context.Context, repoClient *remote.Repository, tag string) (*ocispec.Manifest, error) {
	manifestDescriptor, body, err := oras.FetchBytes(ctx, repoClient, tag, oras.DefaultFetchBytesOptions)
	if err != nil {
		zlog.InfraSec().Error().Err(err).Msg("Failed to fetch manifest")
		return &ocispec.Manifest{}, inv_errors.Wrap(err)
	}

	zlog.InfraSec().Info().Msgf("Fetched manifest with MediaType: %s and Digest: %s", manifestDescriptor.MediaType,
		manifestDescriptor.Digest)

	var manifest ocispec.Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		zlog.InfraSec().Error().Err(err).Msg("Error unmarshalling JSON")
		return &ocispec.Manifest{}, inv_errors.Wrap(err)
	}
	return &manifest, nil
}

func getOCIImageContent(ctx context.Context, repoClient *remote.Repository, layer ocispec.Descriptor) ([]byte, error) {
	zlog.InfraSec().Info().Msgf("Fetching layer with MediaType: %s, Digest: %s, Size: %d bytes", layer.MediaType,
		layer.Digest, layer.Size)

	layerContent, err := repoClient.Fetch(ctx, layer)
	if err != nil {
		zlog.InfraSec().Error().Err(err).Msg("Failed to fetch layer")
		return []byte{}, inv_errors.Wrap(err)
	}
	defer layerContent.Close()

	body, err := io.ReadAll(layerContent)
	if err != nil {
		zlog.InfraSec().Error().Err(err).Msg("Error reading response body")
		return []byte{}, inv_errors.Wrap(err)
	}
	return body, nil
}
