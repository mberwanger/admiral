package metadata

import (
	"context"

	goversion "github.com/caarlos0/go-version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const versionKey = "version"

// defaultMetadata returns the default metadata which will be added to all outgoing calls.
func defaultMetadata() map[string]string {
	return map[string]string{
		versionKey: goversion.GetVersionInfo().String(),
	}
}

// addMetadataToContext returns a new context copied from ctx with the given
// raw metadata added. Metadata already set on the given context for any key
// will not be overridden, but new key/value pairs will always be added.
func addMetadataToContext(ctx context.Context, raw map[string]string) context.Context {
	md := metadata.New(raw)
	if existingMd, ok := metadata.FromOutgoingContext(ctx); ok {
		for key, vals := range existingMd {
			md.Set(key, vals...)
		}
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// DisableInterceptors can be set on the client context with context.WithValue(ctx, DisableInterceptors{}, struct{}{})
// to stop the client interceptors from adding any metadata to the context (useful for testing).
type DisableInterceptors struct{}

// StreamClientInterceptor intercepts a GRPC client stream call and adds
// default metadata to the context.
func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if disable := ctx.Value(DisableInterceptors{}); disable == nil {
		ctx = addMetadataToContext(ctx, defaultMetadata())
	}
	return streamer(ctx, desc, cc, method, opts...)
}

// UnaryClientInterceptor intercepts a GRPC client unary call and adds default
// metadata to the context.
func UnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if disable := ctx.Value(DisableInterceptors{}); disable == nil {
		ctx = addMetadataToContext(ctx, defaultMetadata())
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

// StreamServerInterceptor intercepts a GRPC client stream call and adds
// default metadata to the context.
func StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if disable := stream.Context().Value(DisableInterceptors{}); disable == nil {
		header := metadata.New(defaultMetadata())
		if err := grpc.SendHeader(stream.Context(), header); err != nil {
			return err
		}
	}
	return handler(srv, stream)
}

// ClientVersionFromContext can be called from a GRPC server method to return
// the client version that was added to the GRPC metadata by
// StreamClientInterceptor or UnaryClientInterceptor on the client.
func ClientVersionFromContext(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	versionList := md.Get(versionKey)
	if len(versionList) != 1 {
		return "", false
	}
	return versionList[0], true
}
