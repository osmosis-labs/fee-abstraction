// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: feeabstraction/feeabs/v1beta1/proposal.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type HostChainFeeAbsStatus int32

const (
	HostChainFeeAbsStatus_UNSPECIFIED HostChainFeeAbsStatus = 0
	HostChainFeeAbsStatus_UPDATED     HostChainFeeAbsStatus = 1
	HostChainFeeAbsStatus_OUTDATED    HostChainFeeAbsStatus = 2
	HostChainFeeAbsStatus_FROZEN      HostChainFeeAbsStatus = 3
)

var HostChainFeeAbsStatus_name = map[int32]string{
	0: "UNSPECIFIED",
	1: "UPDATED",
	2: "OUTDATED",
	3: "FROZEN",
}

var HostChainFeeAbsStatus_value = map[string]int32{
	"UNSPECIFIED": 0,
	"UPDATED":     1,
	"OUTDATED":    2,
	"FROZEN":      3,
}

func (x HostChainFeeAbsStatus) String() string {
	return proto.EnumName(HostChainFeeAbsStatus_name, int32(x))
}

func (HostChainFeeAbsStatus) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_c397b73ee3101036, []int{0}
}

type HostChainFeeAbsConfig struct {
	// ibc token is allowed to be used as fee token
	IbcDenom string `protobuf:"bytes,1,opt,name=ibc_denom,json=ibcDenom,proto3" json:"ibc_denom,omitempty" yaml:"allowed_token"`
	// token_in in cross_chain swap contract.
	OsmosisPoolTokenDenomIn string `protobuf:"bytes,2,opt,name=osmosis_pool_token_denom_in,json=osmosisPoolTokenDenomIn,proto3" json:"osmosis_pool_token_denom_in,omitempty"`
	// pool id
	PoolId uint64 `protobuf:"varint,3,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	// Host chain fee abstraction connection status
	Status HostChainFeeAbsStatus `protobuf:"varint,4,opt,name=status,proto3,enum=feeabstraction.feeabs.v1beta1.HostChainFeeAbsStatus" json:"status,omitempty"`
}

func (m *HostChainFeeAbsConfig) Reset()         { *m = HostChainFeeAbsConfig{} }
func (m *HostChainFeeAbsConfig) String() string { return proto.CompactTextString(m) }
func (*HostChainFeeAbsConfig) ProtoMessage()    {}
func (*HostChainFeeAbsConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_c397b73ee3101036, []int{0}
}
func (m *HostChainFeeAbsConfig) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *HostChainFeeAbsConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_HostChainFeeAbsConfig.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *HostChainFeeAbsConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HostChainFeeAbsConfig.Merge(m, src)
}
func (m *HostChainFeeAbsConfig) XXX_Size() int {
	return m.Size()
}
func (m *HostChainFeeAbsConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_HostChainFeeAbsConfig.DiscardUnknown(m)
}

var xxx_messageInfo_HostChainFeeAbsConfig proto.InternalMessageInfo

func (m *HostChainFeeAbsConfig) GetIbcDenom() string {
	if m != nil {
		return m.IbcDenom
	}
	return ""
}

func (m *HostChainFeeAbsConfig) GetOsmosisPoolTokenDenomIn() string {
	if m != nil {
		return m.OsmosisPoolTokenDenomIn
	}
	return ""
}

func (m *HostChainFeeAbsConfig) GetPoolId() uint64 {
	if m != nil {
		return m.PoolId
	}
	return 0
}

func (m *HostChainFeeAbsConfig) GetStatus() HostChainFeeAbsStatus {
	if m != nil {
		return m.Status
	}
	return HostChainFeeAbsStatus_UNSPECIFIED
}

type AddHostZoneProposal struct {
	// the title of the proposal
	Title string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	// the description of the proposal
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	// the host chain config
	HostChainConfig *HostChainFeeAbsConfig `protobuf:"bytes,3,opt,name=host_chain_config,json=hostChainConfig,proto3" json:"host_chain_config,omitempty"`
}

func (m *AddHostZoneProposal) Reset()         { *m = AddHostZoneProposal{} }
func (m *AddHostZoneProposal) String() string { return proto.CompactTextString(m) }
func (*AddHostZoneProposal) ProtoMessage()    {}
func (*AddHostZoneProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_c397b73ee3101036, []int{1}
}
func (m *AddHostZoneProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *AddHostZoneProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_AddHostZoneProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *AddHostZoneProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AddHostZoneProposal.Merge(m, src)
}
func (m *AddHostZoneProposal) XXX_Size() int {
	return m.Size()
}
func (m *AddHostZoneProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_AddHostZoneProposal.DiscardUnknown(m)
}

var xxx_messageInfo_AddHostZoneProposal proto.InternalMessageInfo

type DeleteHostZoneProposal struct {
	// the title of the proposal
	Title string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	// the description of the proposal
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	// the  ibc denom of this token
	IbcDenom string `protobuf:"bytes,3,opt,name=ibc_denom,json=ibcDenom,proto3" json:"ibc_denom,omitempty"`
}

func (m *DeleteHostZoneProposal) Reset()         { *m = DeleteHostZoneProposal{} }
func (m *DeleteHostZoneProposal) String() string { return proto.CompactTextString(m) }
func (*DeleteHostZoneProposal) ProtoMessage()    {}
func (*DeleteHostZoneProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_c397b73ee3101036, []int{2}
}
func (m *DeleteHostZoneProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *DeleteHostZoneProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_DeleteHostZoneProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *DeleteHostZoneProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteHostZoneProposal.Merge(m, src)
}
func (m *DeleteHostZoneProposal) XXX_Size() int {
	return m.Size()
}
func (m *DeleteHostZoneProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteHostZoneProposal.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteHostZoneProposal proto.InternalMessageInfo

type SetHostZoneProposal struct {
	// the title of the proposal
	Title string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	// the description of the proposal
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	// the host chain config
	HostChainConfig *HostChainFeeAbsConfig `protobuf:"bytes,3,opt,name=host_chain_config,json=hostChainConfig,proto3" json:"host_chain_config,omitempty"`
}

func (m *SetHostZoneProposal) Reset()         { *m = SetHostZoneProposal{} }
func (m *SetHostZoneProposal) String() string { return proto.CompactTextString(m) }
func (*SetHostZoneProposal) ProtoMessage()    {}
func (*SetHostZoneProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_c397b73ee3101036, []int{3}
}
func (m *SetHostZoneProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SetHostZoneProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SetHostZoneProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SetHostZoneProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SetHostZoneProposal.Merge(m, src)
}
func (m *SetHostZoneProposal) XXX_Size() int {
	return m.Size()
}
func (m *SetHostZoneProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_SetHostZoneProposal.DiscardUnknown(m)
}

var xxx_messageInfo_SetHostZoneProposal proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("feeabstraction.feeabs.v1beta1.HostChainFeeAbsStatus", HostChainFeeAbsStatus_name, HostChainFeeAbsStatus_value)
	proto.RegisterType((*HostChainFeeAbsConfig)(nil), "feeabstraction.feeabs.v1beta1.HostChainFeeAbsConfig")
	proto.RegisterType((*AddHostZoneProposal)(nil), "feeabstraction.feeabs.v1beta1.AddHostZoneProposal")
	proto.RegisterType((*DeleteHostZoneProposal)(nil), "feeabstraction.feeabs.v1beta1.DeleteHostZoneProposal")
	proto.RegisterType((*SetHostZoneProposal)(nil), "feeabstraction.feeabs.v1beta1.SetHostZoneProposal")
}

func init() {
	proto.RegisterFile("feeabstraction/feeabs/v1beta1/proposal.proto", fileDescriptor_c397b73ee3101036)
}

var fileDescriptor_c397b73ee3101036 = []byte{
	// 533 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xd4, 0x94, 0xcb, 0x6e, 0xd3, 0x4c,
	0x14, 0xc7, 0xe3, 0x26, 0x5f, 0xda, 0x4e, 0x3e, 0xd1, 0x30, 0x04, 0x1a, 0xb5, 0xe0, 0x44, 0x59,
	0x45, 0x88, 0xda, 0x2a, 0x17, 0x21, 0x2a, 0x36, 0x69, 0x2e, 0x22, 0x12, 0x6a, 0x22, 0x27, 0xd9,
	0x64, 0x63, 0xc6, 0xf6, 0xc4, 0x19, 0x61, 0xcf, 0xb1, 0x32, 0x93, 0x42, 0xdf, 0x80, 0x25, 0x8f,
	0xc0, 0x4b, 0xf0, 0x0e, 0x2c, 0xbb, 0x64, 0x85, 0x50, 0xf2, 0x06, 0xf0, 0x02, 0x68, 0x3c, 0x2e,
	0x0a, 0x20, 0x21, 0x21, 0xb1, 0x61, 0xe7, 0xff, 0xf1, 0xef, 0x9c, 0x33, 0xe7, 0x32, 0x83, 0xee,
	0xcd, 0x28, 0x25, 0x9e, 0x90, 0x0b, 0xe2, 0x4b, 0x06, 0xdc, 0xd6, 0xd2, 0x3e, 0x3f, 0xf6, 0xa8,
	0x24, 0xc7, 0x76, 0xb2, 0x80, 0x04, 0x04, 0x89, 0xac, 0x64, 0x01, 0x12, 0xf0, 0x9d, 0x1f, 0x69,
	0x4b, 0x4b, 0x2b, 0xa3, 0x0f, 0x2a, 0x21, 0x84, 0x90, 0x92, 0xb6, 0xfa, 0xd2, 0x4e, 0x07, 0xb5,
	0x10, 0x20, 0x8c, 0xa8, 0x9d, 0x2a, 0x6f, 0x39, 0xb3, 0x25, 0x8b, 0xa9, 0x90, 0x24, 0x4e, 0x32,
	0xe0, 0x76, 0x06, 0x90, 0x84, 0xd9, 0x84, 0x73, 0x90, 0x44, 0x05, 0x17, 0xfa, 0x6f, 0xe3, 0xab,
	0x81, 0x6e, 0x3e, 0x03, 0x21, 0xdb, 0x73, 0xc2, 0x78, 0x8f, 0xd2, 0x96, 0x27, 0xda, 0xc0, 0x67,
	0x2c, 0xc4, 0x8f, 0xd0, 0x2e, 0xf3, 0x7c, 0x37, 0xa0, 0x1c, 0xe2, 0xaa, 0x51, 0x37, 0x9a, 0xbb,
	0xa7, 0xd5, 0x2f, 0x9f, 0x6a, 0x95, 0x0b, 0x12, 0x47, 0x27, 0x0d, 0x12, 0x45, 0xf0, 0x8a, 0x06,
	0xae, 0x84, 0x97, 0x94, 0x37, 0x9c, 0x1d, 0xe6, 0xf9, 0x1d, 0x45, 0xe2, 0xa7, 0xe8, 0x10, 0x44,
	0x0c, 0x82, 0x09, 0x37, 0x01, 0x88, 0x34, 0xa0, 0xa3, 0xb8, 0x8c, 0x57, 0xb7, 0x54, 0x20, 0x67,
	0x3f, 0x43, 0x86, 0x00, 0xd1, 0x58, 0x01, 0xa9, 0x6f, 0x9f, 0xe3, 0x7d, 0xb4, 0x9d, 0x7a, 0xb1,
	0xa0, 0x9a, 0xaf, 0x1b, 0xcd, 0x82, 0x53, 0x54, 0xb2, 0x1f, 0xe0, 0xe7, 0xa8, 0x28, 0x24, 0x91,
	0x4b, 0x51, 0x2d, 0xd4, 0x8d, 0xe6, 0xb5, 0xfb, 0x0f, 0xad, 0xdf, 0x36, 0xcb, 0xfa, 0xa9, 0xa6,
	0x51, 0xea, 0xeb, 0x64, 0x31, 0x1a, 0xef, 0x0d, 0x74, 0xa3, 0x15, 0x04, 0x0a, 0x9a, 0x02, 0xa7,
	0xc3, 0x6c, 0x0e, 0xb8, 0x82, 0xfe, 0x93, 0x4c, 0x46, 0x54, 0xd7, 0xeb, 0x68, 0x81, 0xeb, 0xa8,
	0x14, 0x50, 0xe1, 0x2f, 0x58, 0xa2, 0x32, 0x65, 0x25, 0x6c, 0x9a, 0xf0, 0x0b, 0x74, 0x7d, 0x0e,
	0x42, 0xba, 0xbe, 0xca, 0xe8, 0xfa, 0x69, 0x03, 0xd3, 0x02, 0x4a, 0x7f, 0x7a, 0x50, 0xdd, 0x7c,
	0x67, 0x6f, 0x7e, 0x65, 0xd6, 0x86, 0x93, 0xc2, 0x9b, 0x77, 0xb5, 0x5c, 0x43, 0xa0, 0x5b, 0x1d,
	0x1a, 0x51, 0x49, 0xff, 0xda, 0xc9, 0x0f, 0x37, 0xa7, 0x9c, 0x4f, 0xff, 0x7f, 0x9f, 0x65, 0x96,
	0x54, 0x35, 0x6b, 0x44, 0xe5, 0xbf, 0xd6, 0xac, 0xbb, 0x83, 0x5f, 0x36, 0x5b, 0x6f, 0x01, 0xde,
	0x43, 0xa5, 0xc9, 0xd9, 0x68, 0xd8, 0x6d, 0xf7, 0x7b, 0xfd, 0x6e, 0xa7, 0x9c, 0xc3, 0x25, 0xb4,
	0x3d, 0x19, 0x76, 0x5a, 0xe3, 0x6e, 0xa7, 0x6c, 0xe0, 0xff, 0xd1, 0xce, 0x60, 0x32, 0xd6, 0x6a,
	0x0b, 0x23, 0x54, 0xec, 0x39, 0x83, 0x69, 0xf7, 0xac, 0x9c, 0x3f, 0x1d, 0x7d, 0x58, 0x99, 0xc6,
	0xe5, 0xca, 0x34, 0x3e, 0xaf, 0x4c, 0xe3, 0xed, 0xda, 0xcc, 0x5d, 0xae, 0xcd, 0xdc, 0xc7, 0xb5,
	0x99, 0x9b, 0x3e, 0x09, 0x99, 0x9c, 0x2f, 0x3d, 0xcb, 0x87, 0xd8, 0xce, 0x56, 0xfb, 0x28, 0x52,
	0x37, 0x7d, 0x46, 0xe9, 0xd1, 0xe6, 0x03, 0x70, 0xfe, 0xd8, 0x7e, 0x7d, 0xf5, 0x0a, 0xc8, 0x8b,
	0x84, 0x0a, 0xaf, 0x98, 0xde, 0xc3, 0x07, 0xdf, 0x02, 0x00, 0x00, 0xff, 0xff, 0x94, 0x71, 0x2d,
	0x92, 0x2b, 0x04, 0x00, 0x00,
}

func (m *HostChainFeeAbsConfig) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *HostChainFeeAbsConfig) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *HostChainFeeAbsConfig) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Status != 0 {
		i = encodeVarintProposal(dAtA, i, uint64(m.Status))
		i--
		dAtA[i] = 0x20
	}
	if m.PoolId != 0 {
		i = encodeVarintProposal(dAtA, i, uint64(m.PoolId))
		i--
		dAtA[i] = 0x18
	}
	if len(m.OsmosisPoolTokenDenomIn) > 0 {
		i -= len(m.OsmosisPoolTokenDenomIn)
		copy(dAtA[i:], m.OsmosisPoolTokenDenomIn)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.OsmosisPoolTokenDenomIn)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.IbcDenom) > 0 {
		i -= len(m.IbcDenom)
		copy(dAtA[i:], m.IbcDenom)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.IbcDenom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *AddHostZoneProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AddHostZoneProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *AddHostZoneProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.HostChainConfig != nil {
		{
			size, err := m.HostChainConfig.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintProposal(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *DeleteHostZoneProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DeleteHostZoneProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DeleteHostZoneProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.IbcDenom) > 0 {
		i -= len(m.IbcDenom)
		copy(dAtA[i:], m.IbcDenom)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.IbcDenom)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *SetHostZoneProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SetHostZoneProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SetHostZoneProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.HostChainConfig != nil {
		{
			size, err := m.HostChainConfig.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintProposal(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintProposal(dAtA []byte, offset int, v uint64) int {
	offset -= sovProposal(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *HostChainFeeAbsConfig) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.IbcDenom)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = len(m.OsmosisPoolTokenDenomIn)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	if m.PoolId != 0 {
		n += 1 + sovProposal(uint64(m.PoolId))
	}
	if m.Status != 0 {
		n += 1 + sovProposal(uint64(m.Status))
	}
	return n
}

func (m *AddHostZoneProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	if m.HostChainConfig != nil {
		l = m.HostChainConfig.Size()
		n += 1 + l + sovProposal(uint64(l))
	}
	return n
}

func (m *DeleteHostZoneProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = len(m.IbcDenom)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	return n
}

func (m *SetHostZoneProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	if m.HostChainConfig != nil {
		l = m.HostChainConfig.Size()
		n += 1 + l + sovProposal(uint64(l))
	}
	return n
}

func sovProposal(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozProposal(x uint64) (n int) {
	return sovProposal(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *HostChainFeeAbsConfig) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProposal
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: HostChainFeeAbsConfig: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: HostChainFeeAbsConfig: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IbcDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IbcDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OsmosisPoolTokenDenomIn", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OsmosisPoolTokenDenomIn = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolId", wireType)
			}
			m.PoolId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PoolId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
			m.Status = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Status |= HostChainFeeAbsStatus(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipProposal(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProposal
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *AddHostZoneProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProposal
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: AddHostZoneProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AddHostZoneProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field HostChainConfig", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.HostChainConfig == nil {
				m.HostChainConfig = &HostChainFeeAbsConfig{}
			}
			if err := m.HostChainConfig.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProposal(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProposal
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *DeleteHostZoneProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProposal
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: DeleteHostZoneProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DeleteHostZoneProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IbcDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IbcDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProposal(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProposal
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SetHostZoneProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProposal
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SetHostZoneProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SetHostZoneProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field HostChainConfig", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.HostChainConfig == nil {
				m.HostChainConfig = &HostChainFeeAbsConfig{}
			}
			if err := m.HostChainConfig.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProposal(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProposal
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipProposal(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowProposal
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthProposal
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupProposal
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthProposal
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthProposal        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowProposal          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupProposal = fmt.Errorf("proto: unexpected end of group")
)
