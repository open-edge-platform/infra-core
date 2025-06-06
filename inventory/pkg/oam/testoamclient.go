// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package oam

import (
	"context"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"
)

var (
	// these are use only for testing, creating a bufconn between server and client
	// see.
	TestBufconn   *bufconn.Listener
	TestClient    Client
	TestReadyChan chan bool
)

type Client struct {
	connection *grpc.ClientConn
	HealthAPI  grpc_health_v1.HealthClient
	termChan   chan bool
	wg         *sync.WaitGroup
}

func NewGrpcClient(
	termChan chan bool,
	wg *sync.WaitGroup,
	servaddr string,
) Client {
	var conn *grpc.ClientConn

	err := error(nil)

	// used only for testing!
	// Notice: ato avoid DNS failure in client calls for bufconn testing:
	// https://github.com/grpc/grpc-go/blob/v1.64.0/internal/resolver/passthrough/passthrough.go
	// Package passthrough implements a pass-through resolver. It sends the target
	// name without scheme back to gRPC as resolved address.
	if servaddr == "bufconn" {
		conn, err = grpc.NewClient(
			"passthrough://bufnet",
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
				return TestBufconn.Dial()
			}),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			zlog.Fatal().Err(err).Msg("Unable to connect")
		}
	} else {
		zlog.Fatal().Err(err).Msg("Unsupported mode")
	}

	zlog.Info().Msg("Connected")

	// collect all info
	client := Client{
		connection: conn,
		HealthAPI:  grpc_health_v1.NewHealthClient(conn),
		termChan:   termChan,
		wg:         wg,
	}

	return client
}

func (client *Client) Close() {
	client.connection.Close()
}

// GetServingStatus retrieves service status based on its ID.
// Empty ID means global status.
func (client *Client) GetServingStatus(
	ctx context.Context,
	serviceID string,
) (*grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
	zlog.Info().Msgf("Get health of service ID: %s", serviceID)

	object := grpc_health_v1.HealthCheckRequest{
		Service: serviceID,
	}
	obj, err := client.HealthAPI.Check(ctx, &object)
	if err != nil {
		return nil, err
	}

	return &obj.Status, nil
}

// WatchServingStatus request streaming on serving status.
// Empty ID means global status.
func (client *Client) WatchServingStatus(
	ctx context.Context,
	serviceID string,
) (grpc_health_v1.Health_WatchClient, error) {
	zlog.Info().Msgf("Watch health of service ID: %s", serviceID)

	object := grpc_health_v1.HealthCheckRequest{
		Service: serviceID,
	}

	stream, err := client.HealthAPI.Watch(ctx, &object)
	if err != nil {
		return nil, err
	}

	return stream, nil
}
