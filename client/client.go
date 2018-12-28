package client

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strings"
)

const (
	PartitionRedirectKey = "partition_redirect"
	PartitionMapKey      = "partition_map"
)

func inter(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md := metadata.MD{} // to get panic
	co := grpc.Header(&md)
	ops := make([]grpc.CallOption, 0)
	ops = append(ops, co)
	ops = append(ops, opts...)
	// Logic before invoking the invoker
	// Calls the invoker to execute RPC
	err := invoker(ctx, method, req, reply, cc, ops...)
	// Logic after invoking the invoker

	partitions := strings.Join(md[PartitionRedirectKey], "")
	if errs == "" {
		return nil
	}

	return errors.FromString(errs)
	return err
}
