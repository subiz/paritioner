// Code generated by protoc-gen-go. DO NOT EDIT.
// source: partitioner.proto

package header

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Configuration struct {
	Version              string                       `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	Term                 int32                        `protobuf:"varint,2,opt,name=term,proto3" json:"term,omitempty"`
	Cluster              string                       `protobuf:"bytes,3,opt,name=cluster,proto3" json:"cluster,omitempty"`
	TotalPartitions      int32                        `protobuf:"varint,6,opt,name=total_partitions,json=totalPartitions,proto3" json:"total_partitions,omitempty"`
	NextTerm             int32                        `protobuf:"varint,7,opt,name=next_term,json=nextTerm,proto3" json:"next_term,omitempty"`
	Partitions           map[string]*WorkerPartitions `protobuf:"bytes,9,rep,name=partitions,proto3" json:"partitions,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Hosts                map[string]string            `protobuf:"bytes,10,rep,name=hosts,proto3" json:"hosts,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}                     `json:"-"`
	XXX_unrecognized     []byte                       `json:"-"`
	XXX_sizecache        int32                        `json:"-"`
}

func (m *Configuration) Reset()         { *m = Configuration{} }
func (m *Configuration) String() string { return proto.CompactTextString(m) }
func (*Configuration) ProtoMessage()    {}
func (*Configuration) Descriptor() ([]byte, []int) {
	return fileDescriptor_partitioner_b167a7339e76eecb, []int{0}
}
func (m *Configuration) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Configuration.Unmarshal(m, b)
}
func (m *Configuration) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Configuration.Marshal(b, m, deterministic)
}
func (dst *Configuration) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Configuration.Merge(dst, src)
}
func (m *Configuration) XXX_Size() int {
	return xxx_messageInfo_Configuration.Size(m)
}
func (m *Configuration) XXX_DiscardUnknown() {
	xxx_messageInfo_Configuration.DiscardUnknown(m)
}

var xxx_messageInfo_Configuration proto.InternalMessageInfo

func (m *Configuration) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *Configuration) GetTerm() int32 {
	if m != nil {
		return m.Term
	}
	return 0
}

func (m *Configuration) GetCluster() string {
	if m != nil {
		return m.Cluster
	}
	return ""
}

func (m *Configuration) GetTotalPartitions() int32 {
	if m != nil {
		return m.TotalPartitions
	}
	return 0
}

func (m *Configuration) GetNextTerm() int32 {
	if m != nil {
		return m.NextTerm
	}
	return 0
}

func (m *Configuration) GetPartitions() map[string]*WorkerPartitions {
	if m != nil {
		return m.Partitions
	}
	return nil
}

func (m *Configuration) GetHosts() map[string]string {
	if m != nil {
		return m.Hosts
	}
	return nil
}

type WorkerPartitions struct {
	Partitions           []int32  `protobuf:"varint,7,rep,packed,name=partitions,proto3" json:"partitions,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *WorkerPartitions) Reset()         { *m = WorkerPartitions{} }
func (m *WorkerPartitions) String() string { return proto.CompactTextString(m) }
func (*WorkerPartitions) ProtoMessage()    {}
func (*WorkerPartitions) Descriptor() ([]byte, []int) {
	return fileDescriptor_partitioner_b167a7339e76eecb, []int{1}
}
func (m *WorkerPartitions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WorkerPartitions.Unmarshal(m, b)
}
func (m *WorkerPartitions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WorkerPartitions.Marshal(b, m, deterministic)
}
func (dst *WorkerPartitions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WorkerPartitions.Merge(dst, src)
}
func (m *WorkerPartitions) XXX_Size() int {
	return xxx_messageInfo_WorkerPartitions.Size(m)
}
func (m *WorkerPartitions) XXX_DiscardUnknown() {
	xxx_messageInfo_WorkerPartitions.DiscardUnknown(m)
}

var xxx_messageInfo_WorkerPartitions proto.InternalMessageInfo

func (m *WorkerPartitions) GetPartitions() []int32 {
	if m != nil {
		return m.Partitions
	}
	return nil
}

type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_partitioner_b167a7339e76eecb, []int{2}
}
func (m *Empty) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Empty.Unmarshal(m, b)
}
func (m *Empty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Empty.Marshal(b, m, deterministic)
}
func (dst *Empty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Empty.Merge(dst, src)
}
func (m *Empty) XXX_Size() int {
	return xxx_messageInfo_Empty.Size(m)
}
func (m *Empty) XXX_DiscardUnknown() {
	xxx_messageInfo_Empty.DiscardUnknown(m)
}

var xxx_messageInfo_Empty proto.InternalMessageInfo

type Cluster struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Cluster) Reset()         { *m = Cluster{} }
func (m *Cluster) String() string { return proto.CompactTextString(m) }
func (*Cluster) ProtoMessage()    {}
func (*Cluster) Descriptor() ([]byte, []int) {
	return fileDescriptor_partitioner_b167a7339e76eecb, []int{3}
}
func (m *Cluster) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Cluster.Unmarshal(m, b)
}
func (m *Cluster) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Cluster.Marshal(b, m, deterministic)
}
func (dst *Cluster) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Cluster.Merge(dst, src)
}
func (m *Cluster) XXX_Size() int {
	return xxx_messageInfo_Cluster.Size(m)
}
func (m *Cluster) XXX_DiscardUnknown() {
	xxx_messageInfo_Cluster.DiscardUnknown(m)
}

var xxx_messageInfo_Cluster proto.InternalMessageInfo

func (m *Cluster) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type WorkerHost struct {
	Cluster              string   `protobuf:"bytes,1,opt,name=cluster,proto3" json:"cluster,omitempty"`
	Id                   string   `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Term                 int32    `protobuf:"varint,3,opt,name=term,proto3" json:"term,omitempty"`
	Host                 string   `protobuf:"bytes,4,opt,name=host,proto3" json:"host,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *WorkerHost) Reset()         { *m = WorkerHost{} }
func (m *WorkerHost) String() string { return proto.CompactTextString(m) }
func (*WorkerHost) ProtoMessage()    {}
func (*WorkerHost) Descriptor() ([]byte, []int) {
	return fileDescriptor_partitioner_b167a7339e76eecb, []int{4}
}
func (m *WorkerHost) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WorkerHost.Unmarshal(m, b)
}
func (m *WorkerHost) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WorkerHost.Marshal(b, m, deterministic)
}
func (dst *WorkerHost) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WorkerHost.Merge(dst, src)
}
func (m *WorkerHost) XXX_Size() int {
	return xxx_messageInfo_WorkerHost.Size(m)
}
func (m *WorkerHost) XXX_DiscardUnknown() {
	xxx_messageInfo_WorkerHost.DiscardUnknown(m)
}

var xxx_messageInfo_WorkerHost proto.InternalMessageInfo

func (m *WorkerHost) GetCluster() string {
	if m != nil {
		return m.Cluster
	}
	return ""
}

func (m *WorkerHost) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *WorkerHost) GetTerm() int32 {
	if m != nil {
		return m.Term
	}
	return 0
}

func (m *WorkerHost) GetHost() string {
	if m != nil {
		return m.Host
	}
	return ""
}

func init() {
	proto.RegisterType((*Configuration)(nil), "header.Configuration")
	proto.RegisterMapType((map[string]string)(nil), "header.Configuration.HostsEntry")
	proto.RegisterMapType((map[string]*WorkerPartitions)(nil), "header.Configuration.PartitionsEntry")
	proto.RegisterType((*WorkerPartitions)(nil), "header.WorkerPartitions")
	proto.RegisterType((*Empty)(nil), "header.Empty")
	proto.RegisterType((*Cluster)(nil), "header.Cluster")
	proto.RegisterType((*WorkerHost)(nil), "header.WorkerHost")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// CoordinatorClient is the client API for Coordinator service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type CoordinatorClient interface {
	GetConfig(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (*Configuration, error)
	Rebalance(ctx context.Context, in *WorkerHost, opts ...grpc.CallOption) (Coordinator_RebalanceClient, error)
	Accept(ctx context.Context, in *WorkerHost, opts ...grpc.CallOption) (*Empty, error)
	Deny(ctx context.Context, in *WorkerHost, opts ...grpc.CallOption) (*Empty, error)
}

type coordinatorClient struct {
	cc *grpc.ClientConn
}

func NewCoordinatorClient(cc *grpc.ClientConn) CoordinatorClient {
	return &coordinatorClient{cc}
}

func (c *coordinatorClient) GetConfig(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (*Configuration, error) {
	out := new(Configuration)
	err := c.cc.Invoke(ctx, "/header.Coordinator/GetConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coordinatorClient) Rebalance(ctx context.Context, in *WorkerHost, opts ...grpc.CallOption) (Coordinator_RebalanceClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Coordinator_serviceDesc.Streams[0], "/header.Coordinator/Rebalance", opts...)
	if err != nil {
		return nil, err
	}
	x := &coordinatorRebalanceClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Coordinator_RebalanceClient interface {
	Recv() (*Configuration, error)
	grpc.ClientStream
}

type coordinatorRebalanceClient struct {
	grpc.ClientStream
}

func (x *coordinatorRebalanceClient) Recv() (*Configuration, error) {
	m := new(Configuration)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *coordinatorClient) Accept(ctx context.Context, in *WorkerHost, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/header.Coordinator/Accept", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coordinatorClient) Deny(ctx context.Context, in *WorkerHost, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/header.Coordinator/Deny", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CoordinatorServer is the server API for Coordinator service.
type CoordinatorServer interface {
	GetConfig(context.Context, *Cluster) (*Configuration, error)
	Rebalance(*WorkerHost, Coordinator_RebalanceServer) error
	Accept(context.Context, *WorkerHost) (*Empty, error)
	Deny(context.Context, *WorkerHost) (*Empty, error)
}

func RegisterCoordinatorServer(s *grpc.Server, srv CoordinatorServer) {
	s.RegisterService(&_Coordinator_serviceDesc, srv)
}

func _Coordinator_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Cluster)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoordinatorServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/header.Coordinator/GetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoordinatorServer).GetConfig(ctx, req.(*Cluster))
	}
	return interceptor(ctx, in, info, handler)
}

func _Coordinator_Rebalance_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(WorkerHost)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CoordinatorServer).Rebalance(m, &coordinatorRebalanceServer{stream})
}

type Coordinator_RebalanceServer interface {
	Send(*Configuration) error
	grpc.ServerStream
}

type coordinatorRebalanceServer struct {
	grpc.ServerStream
}

func (x *coordinatorRebalanceServer) Send(m *Configuration) error {
	return x.ServerStream.SendMsg(m)
}

func _Coordinator_Accept_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkerHost)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoordinatorServer).Accept(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/header.Coordinator/Accept",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoordinatorServer).Accept(ctx, req.(*WorkerHost))
	}
	return interceptor(ctx, in, info, handler)
}

func _Coordinator_Deny_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkerHost)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoordinatorServer).Deny(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/header.Coordinator/Deny",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoordinatorServer).Deny(ctx, req.(*WorkerHost))
	}
	return interceptor(ctx, in, info, handler)
}

var _Coordinator_serviceDesc = grpc.ServiceDesc{
	ServiceName: "header.Coordinator",
	HandlerType: (*CoordinatorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetConfig",
			Handler:    _Coordinator_GetConfig_Handler,
		},
		{
			MethodName: "Accept",
			Handler:    _Coordinator_Accept_Handler,
		},
		{
			MethodName: "Deny",
			Handler:    _Coordinator_Deny_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Rebalance",
			Handler:       _Coordinator_Rebalance_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "partitioner.proto",
}

// WorkerClient is the client API for Worker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type WorkerClient interface {
	GetConfig(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (*Configuration, error)
}

type workerClient struct {
	cc *grpc.ClientConn
}

func NewWorkerClient(cc *grpc.ClientConn) WorkerClient {
	return &workerClient{cc}
}

func (c *workerClient) GetConfig(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (*Configuration, error) {
	out := new(Configuration)
	err := c.cc.Invoke(ctx, "/header.Worker/GetConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WorkerServer is the server API for Worker service.
type WorkerServer interface {
	GetConfig(context.Context, *Cluster) (*Configuration, error)
}

func RegisterWorkerServer(s *grpc.Server, srv WorkerServer) {
	s.RegisterService(&_Worker_serviceDesc, srv)
}

func _Worker_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Cluster)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/header.Worker/GetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServer).GetConfig(ctx, req.(*Cluster))
	}
	return interceptor(ctx, in, info, handler)
}

var _Worker_serviceDesc = grpc.ServiceDesc{
	ServiceName: "header.Worker",
	HandlerType: (*WorkerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetConfig",
			Handler:    _Worker_GetConfig_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "partitioner.proto",
}

// HelloClient is the client API for Hello service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type HelloClient interface {
	Hello(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (*WorkerHost, error)
}

type helloClient struct {
	cc *grpc.ClientConn
}

func NewHelloClient(cc *grpc.ClientConn) HelloClient {
	return &helloClient{cc}
}

func (c *helloClient) Hello(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (*WorkerHost, error) {
	out := new(WorkerHost)
	err := c.cc.Invoke(ctx, "/header.Hello/Hello", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HelloServer is the server API for Hello service.
type HelloServer interface {
	Hello(context.Context, *Cluster) (*WorkerHost, error)
}

func RegisterHelloServer(s *grpc.Server, srv HelloServer) {
	s.RegisterService(&_Hello_serviceDesc, srv)
}

func _Hello_Hello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Cluster)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelloServer).Hello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/header.Hello/Hello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelloServer).Hello(ctx, req.(*Cluster))
	}
	return interceptor(ctx, in, info, handler)
}

var _Hello_serviceDesc = grpc.ServiceDesc{
	ServiceName: "header.Hello",
	HandlerType: (*HelloServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Hello",
			Handler:    _Hello_Hello_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "partitioner.proto",
}

// GoodbyeClient is the client API for Goodbye service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type GoodbyeClient interface {
	Goodbye(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (*WorkerHost, error)
}

type goodbyeClient struct {
	cc *grpc.ClientConn
}

func NewGoodbyeClient(cc *grpc.ClientConn) GoodbyeClient {
	return &goodbyeClient{cc}
}

func (c *goodbyeClient) Goodbye(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (*WorkerHost, error) {
	out := new(WorkerHost)
	err := c.cc.Invoke(ctx, "/header.Goodbye/Goodbye", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GoodbyeServer is the server API for Goodbye service.
type GoodbyeServer interface {
	Goodbye(context.Context, *Cluster) (*WorkerHost, error)
}

func RegisterGoodbyeServer(s *grpc.Server, srv GoodbyeServer) {
	s.RegisterService(&_Goodbye_serviceDesc, srv)
}

func _Goodbye_Goodbye_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Cluster)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoodbyeServer).Goodbye(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/header.Goodbye/Goodbye",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoodbyeServer).Goodbye(ctx, req.(*Cluster))
	}
	return interceptor(ctx, in, info, handler)
}

var _Goodbye_serviceDesc = grpc.ServiceDesc{
	ServiceName: "header.Goodbye",
	HandlerType: (*GoodbyeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Goodbye",
			Handler:    _Goodbye_Goodbye_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "partitioner.proto",
}

func init() { proto.RegisterFile("partitioner.proto", fileDescriptor_partitioner_b167a7339e76eecb) }

var fileDescriptor_partitioner_b167a7339e76eecb = []byte{
	// 464 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x9c, 0x54, 0x5d, 0x6b, 0xd4, 0x40,
	0x14, 0x6d, 0x36, 0x9b, 0xc4, 0xdc, 0xa5, 0xee, 0x7a, 0x51, 0x18, 0x57, 0x90, 0x25, 0x20, 0x54,
	0xc4, 0x50, 0x22, 0x6a, 0xa9, 0x0f, 0x22, 0xeb, 0xd2, 0x3e, 0x4a, 0x10, 0xfa, 0x66, 0x99, 0xdd,
	0xbd, 0xda, 0xd0, 0x34, 0xb3, 0x4c, 0x66, 0x8b, 0xf9, 0x59, 0xfe, 0x1f, 0x7f, 0x8c, 0x99, 0x99,
	0xfd, 0x48, 0x6a, 0x84, 0xe2, 0xdb, 0x99, 0x9b, 0x73, 0xce, 0xbd, 0x73, 0xcf, 0x10, 0x78, 0xb4,
	0xe2, 0x52, 0x65, 0x2a, 0x13, 0x05, 0xc9, 0x78, 0x25, 0x85, 0x12, 0xe8, 0x5f, 0x11, 0x5f, 0x92,
	0x8c, 0x7e, 0xb9, 0x70, 0x38, 0x15, 0xc5, 0xf7, 0xec, 0xc7, 0x5a, 0x72, 0xcd, 0x40, 0x06, 0xc1,
	0x2d, 0xc9, 0xb2, 0x86, 0xcc, 0x99, 0x38, 0x47, 0x61, 0xba, 0x3d, 0x22, 0x42, 0x5f, 0x91, 0xbc,
	0x61, 0xbd, 0xba, 0xec, 0xa5, 0x06, 0x6b, 0xf6, 0x22, 0x5f, 0x97, 0x35, 0x66, 0xae, 0x65, 0x6f,
	0x8e, 0xf8, 0x12, 0x46, 0x4a, 0x28, 0x9e, 0x5f, 0xee, 0x9a, 0x97, 0xcc, 0x37, 0xca, 0xa1, 0xa9,
	0x7f, 0xd9, 0x95, 0xf1, 0x19, 0x84, 0x05, 0xfd, 0x54, 0x97, 0xc6, 0x3d, 0x30, 0x9c, 0x07, 0xba,
	0xf0, 0x55, 0x77, 0x98, 0x01, 0x34, 0x1c, 0xc2, 0x89, 0x7b, 0x34, 0x48, 0x5e, 0xc4, 0x76, 0xfc,
	0xb8, 0x35, 0x7a, 0xbc, 0xb7, 0x9c, 0x15, 0x4a, 0x56, 0x69, 0x43, 0x88, 0xef, 0xc0, 0xbb, 0x12,
	0xa5, 0x2a, 0x19, 0x18, 0x87, 0x49, 0xb7, 0xc3, 0xb9, 0xa6, 0x58, 0xb1, 0xa5, 0x8f, 0x2f, 0x60,
	0x78, 0xc7, 0x16, 0x47, 0xe0, 0x5e, 0x53, 0xb5, 0xd9, 0x8e, 0x86, 0x18, 0x83, 0x77, 0xcb, 0xf3,
	0x35, 0x99, 0xd5, 0x0c, 0x12, 0xb6, 0x35, 0xbf, 0x10, 0xf2, 0x9a, 0xe4, 0x5e, 0x9f, 0x5a, 0xda,
	0x69, 0xef, 0xc4, 0x19, 0x9f, 0x00, 0xec, 0xbb, 0x75, 0x78, 0x3e, 0x6e, 0x7a, 0x86, 0x0d, 0x65,
	0x94, 0xc0, 0xe8, 0xae, 0x31, 0x3e, 0x6f, 0x6d, 0x29, 0xa8, 0xef, 0xe8, 0x35, 0xaf, 0x1f, 0x05,
	0xe0, 0xcd, 0x6e, 0x56, 0xaa, 0x8a, 0x9e, 0x42, 0x30, 0xdd, 0x24, 0xf4, 0x10, 0x7a, 0xd9, 0x72,
	0xd3, 0xb2, 0x46, 0xd1, 0x37, 0x00, 0xeb, 0xab, 0xe7, 0x6a, 0x26, 0xeb, 0xb4, 0x93, 0xb5, 0xba,
	0xde, 0x56, 0xb7, 0x7b, 0x17, 0x6e, 0xe3, 0x5d, 0xd4, 0x35, 0xbd, 0x3f, 0xd6, 0x37, 0x2c, 0x83,
	0x93, 0xdf, 0x0e, 0x0c, 0xa6, 0x42, 0xc8, 0x65, 0x56, 0x70, 0x25, 0x24, 0xbe, 0x85, 0xf0, 0x8c,
	0x94, 0x0d, 0x00, 0x87, 0xbb, 0x40, 0x6c, 0x97, 0xf1, 0x93, 0xce, 0x84, 0xa2, 0x03, 0x3c, 0x85,
	0x30, 0xa5, 0x39, 0xcf, 0x79, 0xb1, 0x20, 0xc4, 0xf6, 0xaa, 0xf5, 0xe4, 0xff, 0x54, 0x1e, 0x3b,
	0xf8, 0x1a, 0xfc, 0x4f, 0x8b, 0x05, 0xad, 0x54, 0xa7, 0xf0, 0x70, 0x5b, 0xb3, 0xab, 0x3a, 0xc0,
	0x57, 0xd0, 0xff, 0x4c, 0x45, 0x75, 0x2f, 0x72, 0xf2, 0x11, 0x7c, 0xfb, 0xf9, 0x3f, 0x2f, 0x96,
	0xbc, 0x07, 0xef, 0x9c, 0xf2, 0x5c, 0xe8, 0xe7, 0x64, 0xc1, 0x5f, 0xda, 0x8e, 0x41, 0x6a, 0xe1,
	0x07, 0x08, 0xce, 0x84, 0x58, 0xce, 0x2b, 0xc2, 0xe3, 0x3d, 0xbc, 0x9f, 0x78, 0xee, 0x9b, 0x1f,
	0xc2, 0x9b, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0xaa, 0x96, 0xc6, 0xa9, 0x25, 0x04, 0x00, 0x00,
}
