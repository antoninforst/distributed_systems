// gRPC service support for proto.proto.
package grpc

import (
	"context"

	grpcgo "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const MutualExclusionService_RequestPermission_FullMethodName = "/mutualexclusion.MutualExclusionService/RequestPermission"

type MutualExclusionServiceClient interface {
	RequestPermission(ctx context.Context, in *Request, opts ...grpcgo.CallOption) (*Reply, error)
}

type mutualExclusionServiceClient struct{ connection grpcgo.ClientConnInterface }

func NewMutualExclusionServiceClient(connection grpcgo.ClientConnInterface) MutualExclusionServiceClient {
	return &mutualExclusionServiceClient{connection: connection}
}

func (c *mutualExclusionServiceClient) RequestPermission(ctx context.Context, in *Request, opts ...grpcgo.CallOption) (*Reply, error) {
	out := new(Reply)
	if err := c.connection.Invoke(ctx, MutualExclusionService_RequestPermission_FullMethodName, in, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

type MutualExclusionServiceServer interface {
	RequestPermission(context.Context, *Request) (*Reply, error)
}

type UnimplementedMutualExclusionServiceServer struct{}

func (UnimplementedMutualExclusionServiceServer) RequestPermission(context.Context, *Request) (*Reply, error) {
	return nil, status.Error(codes.Unimplemented, "method RequestPermission not implemented")
}

func RegisterMutualExclusionServiceServer(registrar grpcgo.ServiceRegistrar, server MutualExclusionServiceServer) {
	registrar.RegisterService(&MutualExclusionService_ServiceDesc, server)
}

func requestPermissionHandler(server any, ctx context.Context, decode func(any) error, interceptor grpcgo.UnaryServerInterceptor) (any, error) {
	request := new(Request)
	if err := decode(request); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(MutualExclusionServiceServer).RequestPermission(ctx, request)
	}
	info := &grpcgo.UnaryServerInfo{Server: server, FullMethod: MutualExclusionService_RequestPermission_FullMethodName}
	handler := func(ctx context.Context, req any) (any, error) {
		return server.(MutualExclusionServiceServer).RequestPermission(ctx, req.(*Request))
	}
	return interceptor(ctx, request, info, handler)
}

var MutualExclusionService_ServiceDesc = grpcgo.ServiceDesc{
	ServiceName: "mutualexclusion.MutualExclusionService",
	HandlerType: (*MutualExclusionServiceServer)(nil),
	Methods:     []grpcgo.MethodDesc{{MethodName: "RequestPermission", Handler: requestPermissionHandler}},
	Streams:     []grpcgo.StreamDesc{},
	Metadata:    "proto.proto",
}
