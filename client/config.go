package client

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"go.admiral.io/admiral/client/defaults"
)

type Config struct {
	HostPort          string
	AuthToken         string
	ConnectionOptions ConnectionOptions
}

func (c *Config) CheckAndSetDefaults() error {
	if c.HostPort == "" {
		c.HostPort = net.JoinHostPort(defaults.DefaultHost, strconv.Itoa(defaults.DefaultPort))
	}

	if c.ConnectionOptions.DialTimeout == 0 {
		c.ConnectionOptions.DialTimeout = defaults.DefaultDialTimeout
	}
	if c.ConnectionOptions.KeepAliveTime == 0 {
		c.ConnectionOptions.KeepAliveTime = defaults.DefaultKeepAliveTime
	}
	if c.ConnectionOptions.KeepAliveTimeout == 0 {
		c.ConnectionOptions.KeepAliveTimeout = defaults.DefaultKeepAliveTimeout
	}

	if len(c.AuthToken) == 0 {
		return errors.New("oauth token is required")
	}
	c.ConnectionOptions.DialOptions = append(
		c.ConnectionOptions.DialOptions, grpc.WithPerRPCCredentials(tokenAuth{
			token: c.AuthToken,
		}),
	)

	if c.ConnectionOptions.EnableKeepAliveCheck {
		var kap = keepalive.ClientParameters{
			Time:                c.ConnectionOptions.KeepAliveTime,
			Timeout:             c.ConnectionOptions.KeepAliveTimeout,
			PermitWithoutStream: c.ConnectionOptions.KeepAlivePermitWithoutStream,
		}
		c.ConnectionOptions.DialOptions = append(c.ConnectionOptions.DialOptions, grpc.WithKeepaliveParams(kap))
	}

	return nil
}

type ConnectionOptions struct {
	TLSConfig                    *tls.Config
	DialOptions                  []grpc.DialOption
	DialTimeout                  time.Duration
	EnableKeepAliveCheck         bool
	KeepAliveTime                time.Duration
	KeepAliveTimeout             time.Duration
	KeepAlivePermitWithoutStream bool
}

type tokenAuth struct {
	token string
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": "Token " + t.token,
	}, nil
}

func (t tokenAuth) RequireTransportSecurity() bool {
	return true
}
