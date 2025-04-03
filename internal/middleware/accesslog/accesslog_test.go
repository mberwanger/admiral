package accesslog

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	healthcheckv1 "go.admiral.io/admiral/api/healthcheck/v1"
	"go.admiral.io/admiral/internal/config"
)

func TestNew(t *testing.T) {
	tests := []struct {
		config *config.AccessLog
	}{
		{config: nil},
		{config: &config.AccessLog{
			StatusCodeFilters: []uint32{1},
		}},
	}

	for idx, tt := range tests {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			t.Parallel()

			m, err := New(tt.config, nil, nil)
			assert.NoError(t, err)
			assert.NotNil(t, m)
		})
	}
}

func fakeHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return &healthcheckv1.HealthcheckResponse{}, nil
}

func TestInterceptor(t *testing.T) {
	m := &mid{
		logger:      zaptest.NewLogger(t),
		statusCodes: []codes.Code{codes.OK},
	}

	interceptor := m.UnaryInterceptor()
	resp, err := interceptor(
		context.Background(),
		&healthcheckv1.HealthcheckRequest{},
		&grpc.UnaryServerInfo{FullMethod: "/foo/bar"},
		fakeHandler)

	assert.NotNil(t, resp)
	assert.NoError(t, err)
}

func TestStatusCodeFilter(t *testing.T) {
	tests := []struct {
		name              string
		statusCodeFilters []uint32
		wantLogLength     int
	}{
		{
			name:              "single matching status code",
			statusCodeFilters: []uint32{0}, // codes.OK
			wantLogLength:     1,
		},
		{
			name:              "non-matching status code",
			statusCodeFilters: []uint32{5}, // codes.FailedPrecondition
			wantLogLength:     0,
		},
		{
			name:              "multiple status codes with one match",
			statusCodeFilters: []uint32{0, 12, 13, 14}, // codes.OK, codes.NotFound, etc.
			wantLogLength:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup logger with observer
			core, recorded := observer.New(zapcore.DebugLevel)
			logger := zap.New(core)

			// Initialize middleware
			middleware, err := New(
				&config.AccessLog{
					StatusCodeFilters: tt.statusCodeFilters,
				},
				logger,
				nil,
			)
			require.NoError(t, err, "middleware creation failed")

			// Fake handler returning successful response
			fakeHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return &healthcheckv1.HealthcheckResponse{}, nil
			}

			// Execute middleware
			midFn := middleware.UnaryInterceptor()
			resp, err := midFn(
				context.Background(),
				nil,
				&grpc.UnaryServerInfo{FullMethod: "/foo/bar"},
				fakeHandler,
			)

			// Assertions
			assert.NotNil(t, resp, "response should not be nil")

			s := status.Convert(err)
			assert.Equal(t, codes.OK, s.Code(), "expected OK status code")

			assert.Equal(t,
				tt.wantLogLength,
				recorded.Len(),
				"log length mismatch for status code filter",
			)
		})
	}
}

func TestLogContent(t *testing.T) {
	tests := []struct {
		name              string
		statusCodeFilters []uint32
		wantLogLength     int
		wantStatusCode    int64
	}{
		{
			name:              "single status code match",
			statusCodeFilters: []uint32{0}, // codes.OK
			wantLogLength:     1,
			wantStatusCode:    0,
		},
		{
			name:              "multiple status codes with match",
			statusCodeFilters: []uint32{0, 12, 13, 14}, // codes.OK, codes.NotFound, etc.
			wantLogLength:     1,
			wantStatusCode:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup logger with observer
			core, recorded := observer.New(zapcore.DebugLevel)
			logger := zap.New(core)

			// Initialize middleware
			middleware, err := New(
				&config.AccessLog{
					StatusCodeFilters: tt.statusCodeFilters,
				},
				logger,
				nil,
			)
			require.NoError(t, err, "middleware creation failed")

			// Fake handler returning successful response
			fakeHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return &healthcheckv1.HealthcheckResponse{}, nil
			}

			// Execute middleware
			midFn := middleware.UnaryInterceptor()
			resp, err := midFn(
				context.Background(),
				nil,
				&grpc.UnaryServerInfo{FullMethod: "/foo/bar"},
				fakeHandler,
			)

			// Basic response validation
			require.NotNil(t, resp, "response should not be nil")
			s := status.Convert(err)
			require.Equal(t, codes.OK, s.Code(), "expected OK status code")

			// Validate log content
			assert.Equal(t, tt.wantLogLength, recorded.Len(), "unexpected log length")

			if tt.wantLogLength > 0 {
				logEntry := recorded.All()[0]
				logFields := logEntry.ContextMap()

				assert.Equal(t,
					tt.wantStatusCode,
					logFields["statusCode"],
					"status code mismatch in log",
				)

				assert.Equal(t,
					"foo",
					logFields["service"],
					"service field mismatch",
				)

				assert.NotNil(t, logFields["service"], "service field should be present")
				assert.Equal(t,
					"bar",
					logFields["method"],
					"method field mismatch",
				)
			}
		})
	}
}
