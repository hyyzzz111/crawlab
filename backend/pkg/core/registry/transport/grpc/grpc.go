package grpc

import (
	"crawlab/pkg/core/registry"
	pb "crawlab/pkg/core/registry/transport/grpc/proto"
	"golang.org/x/net/context"
	"time"
)

//go:generate protoc --go_out=plugins=grpc:. proto/registry.proto
type RegistryGRPCService struct {
	innerRegistry registry.Registry
}

func (r *RegistryGRPCService) GetService(context context.Context, request *pb.GetRequest) (*pb.GetResponse, error) {
	services, err := r.innerRegistry.GetService(request.GetService())
	if err != nil {
		return nil, err
	}
	return &pb.GetResponse{Services: ToProtos(services)}, nil
}

func (r *RegistryGRPCService) Register(context context.Context, service *pb.Service) (*pb.EmptyResponse, error) {

	return &pb.EmptyResponse{}, r.innerRegistry.Register(ToService(service))
}

func (r *RegistryGRPCService) Deregister(context context.Context, service *pb.Service) (*pb.EmptyResponse, error) {
	return &pb.EmptyResponse{}, r.innerRegistry.Deregister(ToService(service))
}

func (r *RegistryGRPCService) ListServices(context context.Context, request *pb.ListRequest) (*pb.ListResponse, error) {
	services, err := r.innerRegistry.ListServices()
	if err != nil {
		return nil, err
	}

	return &pb.ListResponse{Services: ToProtos(services)}, nil
}

func (r *RegistryGRPCService) Watch(request *pb.WatchRequest, server pb.Registry_WatchServer) error {
	watcher, err := r.innerRegistry.Watch(func(options *registry.WatchOptions) {
		options.Service = request.Service
	})
	if err != nil {
		return err
	}
	result, err := watcher.Next()
	if err != nil {
		return err
	}
	return server.Send(&pb.Result{
		Action:    result.Action,
		Service:   ToProto(result.Service),
		Timestamp: time.Now().UnixNano(),
	})
}

func NewRegistryGRPCService(innerRegistry registry.Registry) *RegistryGRPCService {
	return &RegistryGRPCService{innerRegistry: innerRegistry}
}
