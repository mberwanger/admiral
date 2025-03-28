module github.com/mberwanger/admiral/client

go 1.24.1

require (
	github.com/mberwanger/admiral/server v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.71.0
)

replace github.com/mberwanger/admiral/server => ../server

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.6-20250307204501-0409229c3780.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
