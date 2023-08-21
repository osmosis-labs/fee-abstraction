// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: feeabstraction/absfee/v1beta1/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types/query"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

// Params defines the parameters for the feeabs module.
type GenesisState struct {
	Params Params      `protobuf:"bytes,1,opt,name=params,proto3" json:"params" yaml:"params"`
	Epochs []EpochInfo `protobuf:"bytes,2,rep,name=epochs,proto3" json:"epochs"`
	PortId string      `protobuf:"bytes,3,opt,name=port_id,json=portId,proto3" json:"port_id,omitempty"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_c85d049bd82004cc, []int{0}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetEpochs() []EpochInfo {
	if m != nil {
		return m.Epochs
	}
	return nil
}

func (m *GenesisState) GetPortId() string {
	if m != nil {
		return m.PortId
	}
	return ""
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "feeabstraction.absfee.v1beta1.GenesisState")
}

func init() {
	proto.RegisterFile("feeabstraction/absfee/v1beta1/genesis.proto", fileDescriptor_c85d049bd82004cc)
}

var fileDescriptor_c85d049bd82004cc = []byte{
	// 341 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x90, 0xc1, 0x4a, 0xc3, 0x40,
	0x10, 0x86, 0x13, 0x2b, 0x11, 0x53, 0xbd, 0x04, 0xc5, 0x52, 0x34, 0x2d, 0x05, 0x21, 0x2a, 0xcd,
	0xd2, 0x7a, 0x10, 0x3d, 0x16, 0x54, 0x7a, 0x93, 0xd6, 0x93, 0x17, 0x99, 0xa4, 0x93, 0x34, 0xd0,
	0x64, 0xd6, 0xec, 0xb6, 0xd8, 0xb7, 0xf0, 0xa5, 0x84, 0x1e, 0x7b, 0xf4, 0x54, 0xa4, 0x7d, 0x03,
	0x9f, 0x40, 0x92, 0x4d, 0x45, 0x2f, 0xf5, 0xb6, 0x0b, 0xdf, 0xff, 0xcd, 0xcc, 0x6f, 0x5e, 0x04,
	0x88, 0xe0, 0x09, 0x99, 0x82, 0x2f, 0x23, 0x4a, 0x18, 0x78, 0x22, 0x40, 0x64, 0x93, 0x96, 0x87,
	0x12, 0x5a, 0x2c, 0xc4, 0x04, 0x45, 0x24, 0x5c, 0x9e, 0x92, 0x24, 0xeb, 0xe4, 0x2f, 0xec, 0x2a,
	0xd8, 0x2d, 0xe0, 0xea, 0x41, 0x48, 0x21, 0xe5, 0x24, 0xcb, 0x5e, 0x2a, 0x54, 0x3d, 0x0e, 0x89,
	0xc2, 0x11, 0x32, 0xe0, 0x11, 0x83, 0x24, 0x21, 0x09, 0x59, 0xb6, 0x50, 0x56, 0xcf, 0x7d, 0x12,
	0x31, 0x09, 0xe6, 0x81, 0x40, 0xf6, 0x32, 0xc6, 0x74, 0xfa, 0x33, 0x9b, 0x43, 0x18, 0x25, 0x39,
	0xbc, 0x66, 0x37, 0xef, 0xca, 0x21, 0x85, 0x78, 0xed, 0x3d, 0xdb, 0xcc, 0x22, 0x27, 0x7f, 0xa8,
	0xd0, 0xc6, 0xbb, 0x6e, 0xee, 0xdd, 0xab, 0x3b, 0xfb, 0x12, 0x24, 0x5a, 0x8f, 0xa6, 0xa1, 0x5c,
	0x15, 0xbd, 0xae, 0x3b, 0xe5, 0xf6, 0xa9, 0xbb, 0xf1, 0x6e, 0xf7, 0x21, 0x87, 0x3b, 0x87, 0xb3,
	0x45, 0x4d, 0xfb, 0x5a, 0xd4, 0xf6, 0xa7, 0x10, 0x8f, 0x6e, 0x1a, 0x4a, 0xd1, 0xe8, 0x15, 0x2e,
	0xeb, 0xce, 0x34, 0xf2, 0xa9, 0xa2, 0xb2, 0x55, 0x2f, 0x39, 0xe5, 0xb6, 0xf3, 0x8f, 0xf5, 0x36,
	0x83, 0xbb, 0x49, 0x40, 0x9d, 0xed, 0x4c, 0xdc, 0x2b, 0xd2, 0xd6, 0x91, 0xb9, 0xc3, 0x29, 0x95,
	0xcf, 0xd1, 0xa0, 0x52, 0xaa, 0xeb, 0xce, 0x6e, 0xcf, 0xc8, 0xbe, 0xdd, 0x41, 0xa7, 0x3f, 0x5b,
	0xda, 0xfa, 0x7c, 0x69, 0xeb, 0x9f, 0x4b, 0x5b, 0x7f, 0x5b, 0xd9, 0xda, 0x7c, 0x65, 0x6b, 0x1f,
	0x2b, 0x5b, 0x7b, 0xba, 0x0e, 0x23, 0x39, 0x1c, 0x7b, 0xae, 0x4f, 0x31, 0xcb, 0xeb, 0x8e, 0x44,
	0x73, 0x04, 0x9e, 0x60, 0x01, 0x62, 0xf3, 0x77, 0x4b, 0x93, 0x2b, 0xf6, 0xca, 0xd4, 0x5a, 0x4c,
	0x4e, 0x39, 0x0a, 0xcf, 0xc8, 0x3b, 0xba, 0xfc, 0x0e, 0x00, 0x00, 0xff, 0xff, 0x28, 0x8c, 0x15,
	0xc6, 0x28, 0x02, 0x00, 0x00,
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.PortId) > 0 {
		i -= len(m.PortId)
		copy(dAtA[i:], m.PortId)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.PortId)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Epochs) > 0 {
		for iNdEx := len(m.Epochs) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Epochs[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.Epochs) > 0 {
		for _, e := range m.Epochs {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = len(m.PortId)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Epochs", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Epochs = append(m.Epochs, EpochInfo{})
			if err := m.Epochs[len(m.Epochs)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PortId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PortId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis = fmt.Errorf("proto: unexpected end of group")
)
