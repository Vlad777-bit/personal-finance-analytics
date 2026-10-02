package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

type Client struct {
	connection *grpc.ClientConn
	service    authv1.AuthServiceClient
}

func New(
	ctx context.Context,
	address string,
	dialTimeout time.Duration,
) (*Client, error) {
	connection, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create Auth gRPC client: %w", err)
	}

	dialContext, cancel := context.WithTimeout(ctx, dialTimeout)
	defer cancel()

	connection.Connect()
	for state := connection.GetState(); state != connectivity.Ready; state = connection.GetState() {
		if !connection.WaitForStateChange(dialContext, state) {
			if closeErr := connection.Close(); closeErr != nil {
				return nil, fmt.Errorf(
					"connect to Auth gRPC server: %w",
					errors.Join(
						dialContext.Err(),
						fmt.Errorf("close connection: %w", closeErr),
					),
				)
			}

			return nil, fmt.Errorf(
				"connect to Auth gRPC server: %w",
				dialContext.Err(),
			)
		}
	}

	return &Client{
		connection: connection,
		service:    authv1.NewAuthServiceClient(connection),
	}, nil
}

func (c *Client) Close() error {
	if err := c.connection.Close(); err != nil {
		return fmt.Errorf("close Auth gRPC connection: %w", err)
	}

	return nil
}
