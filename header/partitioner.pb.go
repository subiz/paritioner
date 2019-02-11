// Code generated by protoc-gen-go. DO NOT EDIT.
// source: partitioner.proto

package partitioner

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

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
	return fileDescriptor_partitioner_d8610ba065ac326b, []int{0}
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
	return fileDescriptor_partitioner_d8610ba065ac326b, []int{1}
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
	return fileDescriptor_partitioner_d8610ba065ac326b, []int{2}
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
	return fileDescriptor_partitioner_d8610ba065ac326b, []int{3}
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
	return fileDescriptor_partitioner_d8610ba065ac326b, []int{4}
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
	proto.RegisterType((*Configuration)(nil), "partitioner.Configuration")
	proto.RegisterMapType((map[string]string)(nil), "partitioner.Configuration.HostsEntry")
	proto.RegisterMapType((map[string]*WorkerPartitions)(nil), "partitioner.Configuration.PartitionsEntry")
	proto.RegisterType((*WorkerPartitions)(nil), "partitioner.WorkerPartitions")
	proto.RegisterType((*Empty)(nil), "partitioner.Empty")
	proto.RegisterType((*Cluster)(nil), "partitioner.Cluster")
	proto.RegisterType((*WorkerHost)(nil), "partitioner.WorkerHost")
}

func init() { proto.RegisterFile("partitioner.proto", fileDescriptor_partitioner_d8610ba065ac326b) }

var fileDescriptor_partitioner_d8610ba065ac326b = []byte{
	// 464 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xa4, 0x94, 0xc1, 0x6b, 0xd4, 0x40,
	0x14, 0xc6, 0x9b, 0x64, 0x93, 0x98, 0x17, 0xb4, 0xeb, 0xa3, 0x60, 0x8c, 0x28, 0x25, 0x20, 0x54,
	0x0f, 0x8b, 0xa4, 0x88, 0x4b, 0x3d, 0x48, 0xd9, 0x2e, 0xad, 0x9e, 0x24, 0x08, 0x5e, 0xc4, 0x92,
	0xcd, 0x8e, 0x1a, 0x9a, 0x66, 0xc2, 0x64, 0xb6, 0x98, 0xbb, 0x7f, 0x98, 0x7f, 0x9a, 0x99, 0x99,
	0xec, 0x66, 0x52, 0x76, 0x0b, 0xc5, 0xdb, 0x97, 0xb7, 0xef, 0xfb, 0xbd, 0x99, 0xef, 0x0d, 0x0b,
	0x8f, 0xab, 0x94, 0xf1, 0x9c, 0xe7, 0xb4, 0x24, 0x6c, 0x52, 0x31, 0xca, 0x29, 0xfa, 0x5a, 0x29,
	0xfa, 0x6b, 0xc1, 0xc3, 0x19, 0x2d, 0x7f, 0xe4, 0x3f, 0x57, 0x2c, 0x15, 0x35, 0x0c, 0xc0, 0xbd,
	0x21, 0xac, 0x6e, 0x65, 0x60, 0x1c, 0x1a, 0x47, 0x5e, 0xb2, 0xfe, 0x44, 0x84, 0x11, 0x27, 0xec,
	0x3a, 0x30, 0xdb, 0xb2, 0x9d, 0x48, 0x2d, 0xba, 0xb3, 0x62, 0x55, 0xb7, 0x3a, 0xb0, 0x54, 0x77,
	0xf7, 0x89, 0xaf, 0x60, 0xcc, 0x29, 0x4f, 0x8b, 0xcb, 0xcd, 0xb8, 0x3a, 0x70, 0xa4, 0x73, 0x5f,
	0xd6, 0x3f, 0x6f, 0xca, 0xf8, 0x0c, 0xbc, 0x92, 0xfc, 0xe6, 0x97, 0x92, 0xee, 0xca, 0x9e, 0x07,
	0xa2, 0xf0, 0x45, 0x4c, 0xf8, 0x04, 0xa0, 0x11, 0xbc, 0x43, 0xeb, 0xc8, 0x8f, 0x5f, 0x4f, 0xf4,
	0x6b, 0x0d, 0xce, 0x3f, 0xe9, 0xb9, 0xf3, 0x92, 0xb3, 0x26, 0xd1, 0xdc, 0xf8, 0x1e, 0xec, 0x5f,
	0xb4, 0xe6, 0x75, 0x00, 0x12, 0xf3, 0xf2, 0x0e, 0xcc, 0x85, 0xe8, 0x53, 0x04, 0xe5, 0x09, 0xbf,
	0xc1, 0xfe, 0x2d, 0x36, 0x8e, 0xc1, 0xba, 0x22, 0x4d, 0x97, 0x93, 0x90, 0x78, 0x0c, 0xf6, 0x4d,
	0x5a, 0xac, 0x88, 0x0c, 0xc9, 0x8f, 0x9f, 0x0f, 0x26, 0x7c, 0xa5, 0xec, 0x8a, 0xb0, 0x1e, 0x92,
	0xa8, 0xde, 0x13, 0x73, 0x6a, 0x84, 0x53, 0x80, 0x7e, 0xe4, 0x16, 0xf0, 0x81, 0x0e, 0xf6, 0x34,
	0x67, 0x14, 0xc3, 0xf8, 0x36, 0x18, 0x5f, 0x0c, 0x42, 0x73, 0xdb, 0xdb, 0xda, 0x7a, 0x10, 0x91,
	0x0b, 0xf6, 0xfc, 0xba, 0xe2, 0x4d, 0xf4, 0x14, 0xdc, 0x59, 0xb7, 0xb0, 0x47, 0x60, 0xe6, 0xcb,
	0x6e, 0x64, 0xab, 0xa2, 0xef, 0x00, 0x8a, 0x2b, 0xce, 0xa5, 0x2f, 0xda, 0x18, 0x2e, 0x5a, 0xf9,
	0xcc, 0xb5, 0x6f, 0xf3, 0x4c, 0x2c, 0xed, 0x99, 0xb4, 0x35, 0x11, 0x62, 0x30, 0x92, 0x5d, 0x52,
	0xc7, 0x7f, 0x4c, 0xf0, 0x67, 0x94, 0xb2, 0x65, 0x5e, 0xa6, 0x9c, 0x32, 0xfc, 0x00, 0xde, 0x39,
	0xe1, 0x6a, 0x0b, 0x78, 0x30, 0x5c, 0x8d, 0x1a, 0x15, 0x86, 0xbb, 0x17, 0x16, 0xed, 0xe1, 0x19,
	0x78, 0x09, 0x59, 0xa4, 0x45, 0x5a, 0x66, 0x04, 0x9f, 0x6c, 0x49, 0x5e, 0x5c, 0xe4, 0x6e, 0xc6,
	0x1b, 0x03, 0xdf, 0x81, 0x73, 0x9a, 0x65, 0xa4, 0xe2, 0xbb, 0x11, 0x38, 0xf8, 0x41, 0x05, 0xb9,
	0x87, 0x6f, 0x61, 0x74, 0x46, 0xca, 0xe6, 0x9e, 0xb6, 0xf8, 0x23, 0x38, 0xaa, 0xe7, 0xbf, 0x03,
	0x88, 0x4f, 0xc1, 0xbe, 0x20, 0x45, 0x41, 0x71, 0xba, 0x16, 0xdb, 0x29, 0xbb, 0x4e, 0xd8, 0x22,
	0xe6, 0xe0, 0x9e, 0x53, 0xba, 0x5c, 0x34, 0x04, 0x4f, 0x7a, 0x79, 0x5f, 0xcc, 0xc2, 0x91, 0x7f,
	0x35, 0xc7, 0xff, 0x02, 0x00, 0x00, 0xff, 0xff, 0x31, 0xba, 0xd4, 0x6a, 0x7f, 0x04, 0x00, 0x00,
}
