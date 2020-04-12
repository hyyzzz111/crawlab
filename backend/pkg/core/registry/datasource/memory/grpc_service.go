package memory

import (
	"crawlab/pkg/core/registry"
	gt "crawlab/pkg/core/registry/transport/grpc"
	pb "crawlab/pkg/core/registry/transport/grpc/proto"
	"google.golang.org/grpc"
)

func NewMemoryRPCRegistryService(opts ...registry.Option) *gt.RegistryGRPCService {

	return gt.NewRegistryGRPCService(GetRegistry(opts...))
}

func BindGRPCServices(server *grpc.Server, opts ...registry.Option) {
	pb.RegisterRegistryServer(server, NewMemoryRPCRegistryService(opts...))
}
