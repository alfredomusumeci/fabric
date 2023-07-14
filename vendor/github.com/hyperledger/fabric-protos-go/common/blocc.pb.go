// Code generated by protoc-gen-go. DO NOT EDIT.
// source: common/blocc.proto

package common

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ApprovalTxEnvelope struct {
	TxId                 []byte   `protobuf:"bytes,1,opt,name=txId,proto3" json:"txId,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ApprovalTxEnvelope) Reset()         { *m = ApprovalTxEnvelope{} }
func (m *ApprovalTxEnvelope) String() string { return proto.CompactTextString(m) }
func (*ApprovalTxEnvelope) ProtoMessage()    {}
func (*ApprovalTxEnvelope) Descriptor() ([]byte, []int) {
	return fileDescriptor_9d6377dfa67c5bc9, []int{0}
}

func (m *ApprovalTxEnvelope) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ApprovalTxEnvelope.Unmarshal(m, b)
}
func (m *ApprovalTxEnvelope) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ApprovalTxEnvelope.Marshal(b, m, deterministic)
}
func (m *ApprovalTxEnvelope) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ApprovalTxEnvelope.Merge(m, src)
}
func (m *ApprovalTxEnvelope) XXX_Size() int {
	return xxx_messageInfo_ApprovalTxEnvelope.Size(m)
}
func (m *ApprovalTxEnvelope) XXX_DiscardUnknown() {
	xxx_messageInfo_ApprovalTxEnvelope.DiscardUnknown(m)
}

var xxx_messageInfo_ApprovalTxEnvelope proto.InternalMessageInfo

func (m *ApprovalTxEnvelope) GetTxId() []byte {
	if m != nil {
		return m.TxId
	}
	return nil
}

func init() {
	proto.RegisterType((*ApprovalTxEnvelope)(nil), "common.ApprovalTxEnvelope")
}

func init() {
	proto.RegisterFile("common/blocc.proto", fileDescriptor_9d6377dfa67c5bc9)
}

var fileDescriptor_9d6377dfa67c5bc9 = []byte{
	// 148 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4a, 0xce, 0xcf, 0xcd,
	0xcd, 0xcf, 0xd3, 0x4f, 0xca, 0xc9, 0x4f, 0x4e, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62,
	0x83, 0x88, 0x29, 0x69, 0x70, 0x09, 0x39, 0x16, 0x14, 0x14, 0xe5, 0x97, 0x25, 0xe6, 0x84, 0x54,
	0xb8, 0xe6, 0x95, 0xa5, 0xe6, 0xe4, 0x17, 0xa4, 0x0a, 0x09, 0x71, 0xb1, 0x94, 0x54, 0x78, 0xa6,
	0x48, 0x30, 0x2a, 0x30, 0x6a, 0xf0, 0x04, 0x81, 0xd9, 0x4e, 0x61, 0x5c, 0x2a, 0xf9, 0x45, 0xe9,
	0x7a, 0x19, 0x95, 0x05, 0xa9, 0x45, 0x39, 0xa9, 0x29, 0xe9, 0xa9, 0x45, 0x7a, 0x69, 0x89, 0x49,
	0x45, 0x99, 0x50, 0x13, 0x8b, 0xf5, 0x20, 0x26, 0x46, 0xe9, 0xa5, 0x67, 0x96, 0x64, 0x94, 0x26,
	0x81, 0xb8, 0xfa, 0x48, 0x8a, 0xf5, 0x21, 0x8a, 0x75, 0x21, 0x8a, 0x75, 0xd3, 0xf3, 0xf5, 0x21,
	0xea, 0x93, 0xd8, 0xc0, 0x22, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x49, 0x47, 0x27, 0x5d,
	0xa6, 0x00, 0x00, 0x00,
}
