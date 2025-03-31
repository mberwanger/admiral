package client

import (
	"compress/gzip"
	"context"
	"fmt"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	ggzip "google.golang.org/grpc/encoding/gzip"

	"go.admiral.io/admiral/client/metadata"
	applicationv1 "go.admiral.io/admiral/server/api/application/v1"
	clusterv1 "go.admiral.io/admiral/server/api/cluster/v1"
)

func init() {
	if err := ggzip.SetLevel(gzip.BestSpeed); err != nil {
		panic(err)
	}
}

type Client struct {
	config     Config
	conn       *grpc.ClientConn
	grpc       serviceClient
	closedFlag *int32
}

type serviceClient struct {
	applicationv1.ApplicationAPIClient
	clusterv1.ClusterAPIClient
}

func New(ctx context.Context, cfg Config) (client *Client, err error) {
	if err = cfg.CheckAndSetDefaults(); err != nil {
		return nil, err
	}

	client = &Client{
		config:     cfg,
		closedFlag: new(int32),
	}
	if err := client.dialGRPC(ctx, cfg.HostPort); err != nil {
		return nil, fmt.Errorf("failed to connect to addr %v due to '%s'", cfg.HostPort, err)
	}
	return client, nil
}

func (c *Client) dialGRPC(ctx context.Context, hostPort string) error {
	dialContext, cancel := context.WithTimeout(ctx, c.config.ConnectionOptions.DialTimeout)
	defer cancel()

	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts,
		grpc.WithChainUnaryInterceptor(
			metadata.UnaryClientInterceptor,
		),
		grpc.WithChainStreamInterceptor(
			metadata.StreamClientInterceptor,
		),
	)

	dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(c.config.ConnectionOptions.TLSConfig)))
	dialOpts = append(dialOpts, c.config.ConnectionOptions.DialOptions...)

	conn, err := grpc.DialContext(dialContext, hostPort, dialOpts...)
	if err != nil {
		return err
	}

	c.conn = conn
	c.grpc = serviceClient{
		ApplicationAPIClient: applicationv1.NewApplicationAPIClient(c.conn),
		ClusterAPIClient:     clusterv1.NewClusterAPIClient(c.conn),
	}

	return nil
}

func (c *Client) GetConnection() *grpc.ClientConn {
	return c.conn
}

func (c *Client) Close() error {
	if c.setClosed() && c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) isClosed() bool {
	return atomic.LoadInt32(c.closedFlag) == 1
}

func (c *Client) setClosed() bool {
	return atomic.CompareAndSwapInt32(c.closedFlag, 0, 1)
}

func (c *Client) CreateApplication(ctx context.Context, request *applicationv1.CreateApplicationRequest) (*applicationv1.CreateApplicationResponse, error) {
	response, err := c.grpc.CreateApplication(ctx, request)
	if err != nil {
		return nil, err
	}

	return response, nil
}
