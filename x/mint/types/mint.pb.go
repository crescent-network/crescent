// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: crescent/mint/v1beta1/mint.proto

package types

import (
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"
	_ "google.golang.org/protobuf/types/known/durationpb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Params holds parameters for the mint module.
type Params struct {
	// mint_denom defines denomination of coin to be minted
	MintDenom string `protobuf:"bytes,1,opt,name=mint_denom,json=mintDenom,proto3" json:"mint_denom,omitempty"`
	// mint_pool_address defines the address where inflation will be minted. The default is FeeCollector,
	// but if it is set to FeeCollector, minted inflation could be mixed together with collected tx fees.
	// Therefore, it is recommended to specify a separate address depending on usage.
	MintPoolAddress string `protobuf:"bytes,2,opt,name=mint_pool_address,json=mintPoolAddress,proto3" json:"mint_pool_address,omitempty"`
	// block_time_threshold defines block time threshold to prevent from any inflationary manipulation attacks
	// it is used for maximum block duration when calculating block inflation
	BlockTimeThreshold time.Duration `protobuf:"bytes,3,opt,name=block_time_threshold,json=blockTimeThreshold,proto3,stdduration" json:"block_time_threshold"`
	// inflation_schedules defines a list of inflation schedules
	InflationSchedules []InflationSchedule `protobuf:"bytes,4,rep,name=inflation_schedules,json=inflationSchedules,proto3" json:"inflation_schedules"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_fe08af702efa1523, []int{0}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetMintDenom() string {
	if m != nil {
		return m.MintDenom
	}
	return ""
}

func (m *Params) GetMintPoolAddress() string {
	if m != nil {
		return m.MintPoolAddress
	}
	return ""
}

func (m *Params) GetBlockTimeThreshold() time.Duration {
	if m != nil {
		return m.BlockTimeThreshold
	}
	return 0
}

func (m *Params) GetInflationSchedules() []InflationSchedule {
	if m != nil {
		return m.InflationSchedules
	}
	return nil
}

// InflationSchedule defines the start and end time of the inflation period, and the amount of inflation during that
// period.
type InflationSchedule struct {
	// start_time defines the start date time for the inflation schedule
	StartTime time.Time `protobuf:"bytes,1,opt,name=start_time,json=startTime,proto3,stdtime" json:"start_time" yaml:"start_time"`
	// end_time defines the end date time for the inflation schedule
	EndTime time.Time `protobuf:"bytes,2,opt,name=end_time,json=endTime,proto3,stdtime" json:"end_time" yaml:"end_time"`
	// amount defines the total amount of inflation for the schedule
	Amount github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,3,opt,name=amount,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"amount"`
}

func (m *InflationSchedule) Reset()         { *m = InflationSchedule{} }
func (m *InflationSchedule) String() string { return proto.CompactTextString(m) }
func (*InflationSchedule) ProtoMessage()    {}
func (*InflationSchedule) Descriptor() ([]byte, []int) {
	return fileDescriptor_fe08af702efa1523, []int{1}
}
func (m *InflationSchedule) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *InflationSchedule) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_InflationSchedule.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *InflationSchedule) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InflationSchedule.Merge(m, src)
}
func (m *InflationSchedule) XXX_Size() int {
	return m.Size()
}
func (m *InflationSchedule) XXX_DiscardUnknown() {
	xxx_messageInfo_InflationSchedule.DiscardUnknown(m)
}

var xxx_messageInfo_InflationSchedule proto.InternalMessageInfo

func (m *InflationSchedule) GetStartTime() time.Time {
	if m != nil {
		return m.StartTime
	}
	return time.Time{}
}

func (m *InflationSchedule) GetEndTime() time.Time {
	if m != nil {
		return m.EndTime
	}
	return time.Time{}
}

func init() {
	proto.RegisterType((*Params)(nil), "crescent.mint.v1beta1.Params")
	proto.RegisterType((*InflationSchedule)(nil), "crescent.mint.v1beta1.InflationSchedule")
}

func init() { proto.RegisterFile("crescent/mint/v1beta1/mint.proto", fileDescriptor_fe08af702efa1523) }

var fileDescriptor_fe08af702efa1523 = []byte{
	// 471 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x92, 0x4f, 0x6b, 0x13, 0x41,
	0x18, 0xc6, 0xb3, 0x69, 0x89, 0xcd, 0xe4, 0x50, 0x32, 0x56, 0x88, 0x91, 0x6e, 0x42, 0x0e, 0x12,
	0x84, 0xce, 0xd0, 0xa8, 0x17, 0x6f, 0x2e, 0x45, 0xe8, 0x45, 0xca, 0x1a, 0x41, 0xbc, 0x2c, 0xb3,
	0x3b, 0xd3, 0xcd, 0x92, 0x9d, 0x79, 0xc3, 0xce, 0x6c, 0xb5, 0x1f, 0x41, 0xbc, 0xf4, 0xe8, 0x47,
	0xea, 0xb1, 0x47, 0xf1, 0x10, 0x25, 0xf9, 0x06, 0x7e, 0x02, 0x99, 0xd9, 0x5d, 0x95, 0x56, 0xf0,
	0x94, 0xcc, 0xf3, 0x3e, 0xcf, 0xef, 0xfd, 0xc3, 0xa2, 0x71, 0x52, 0x08, 0x9d, 0x08, 0x65, 0xa8,
	0xcc, 0x94, 0xa1, 0x17, 0xc7, 0xb1, 0x30, 0xec, 0xd8, 0x3d, 0xc8, 0xaa, 0x00, 0x03, 0xf8, 0x41,
	0xe3, 0x20, 0x4e, 0xac, 0x1d, 0xc3, 0x83, 0x14, 0x52, 0x70, 0x0e, 0x6a, 0xff, 0x55, 0xe6, 0xa1,
	0x9f, 0x02, 0xa4, 0xb9, 0xa0, 0xee, 0x15, 0x97, 0xe7, 0x94, 0x97, 0x05, 0x33, 0x19, 0xa8, 0xba,
	0x3e, 0xba, 0x5d, 0x37, 0x99, 0x14, 0xda, 0x30, 0xb9, 0xaa, 0x0c, 0x93, 0xcf, 0x6d, 0xd4, 0x39,
	0x63, 0x05, 0x93, 0x1a, 0x1f, 0x22, 0x64, 0x3b, 0x46, 0x5c, 0x28, 0x90, 0x03, 0x6f, 0xec, 0x4d,
	0xbb, 0x61, 0xd7, 0x2a, 0x27, 0x56, 0xc0, 0x4f, 0x50, 0xdf, 0x95, 0x57, 0x00, 0x79, 0xc4, 0x38,
	0x2f, 0x84, 0xd6, 0x83, 0xb6, 0x73, 0xed, 0xdb, 0xc2, 0x19, 0x40, 0xfe, 0xb2, 0x92, 0xf1, 0x5b,
	0x74, 0x10, 0xe7, 0x90, 0x2c, 0x23, 0xdb, 0x2e, 0x32, 0x8b, 0x42, 0xe8, 0x05, 0xe4, 0x7c, 0xb0,
	0x33, 0xf6, 0xa6, 0xbd, 0xd9, 0x43, 0x52, 0x4d, 0x45, 0x9a, 0xa9, 0xc8, 0x49, 0x3d, 0x75, 0xb0,
	0x77, 0xbd, 0x1e, 0xb5, 0xbe, 0x7c, 0x1f, 0x79, 0x21, 0x76, 0x80, 0x79, 0x26, 0xc5, 0xbc, 0x89,
	0xe3, 0x08, 0xdd, 0xcf, 0xd4, 0x79, 0xee, 0xac, 0x91, 0x4e, 0x16, 0x82, 0x97, 0xb9, 0xd0, 0x83,
	0xdd, 0xf1, 0xce, 0xb4, 0x37, 0x9b, 0x92, 0x7f, 0x1e, 0x8e, 0x9c, 0x36, 0x89, 0x37, 0x75, 0x20,
	0xd8, 0xb5, 0x4d, 0x42, 0x9c, 0xdd, 0x2e, 0xe8, 0xc9, 0xa7, 0x36, 0xea, 0xdf, 0xf1, 0xe3, 0x77,
	0x08, 0x69, 0xc3, 0x0a, 0xe3, 0xb6, 0x71, 0x87, 0xe9, 0xcd, 0x86, 0x77, 0x76, 0x98, 0x37, 0x97,
	0x0d, 0x0e, 0x2d, 0xff, 0xe7, 0x7a, 0xd4, 0xbf, 0x64, 0x32, 0x7f, 0x31, 0xf9, 0x93, 0x9d, 0x5c,
	0xd9, 0xcd, 0xba, 0x4e, 0xb0, 0x76, 0x1c, 0xa2, 0x3d, 0xa1, 0x78, 0xc5, 0x6d, 0xff, 0x97, 0xfb,
	0xa8, 0xe6, 0xee, 0x57, 0xdc, 0x26, 0x59, 0x51, 0xef, 0x09, 0xc5, 0x1d, 0xf3, 0x15, 0xea, 0x30,
	0x09, 0xa5, 0x32, 0xee, 0xda, 0xdd, 0x80, 0xd8, 0xd4, 0xb7, 0xf5, 0xe8, 0x71, 0x9a, 0x99, 0x45,
	0x19, 0x93, 0x04, 0x24, 0x4d, 0x40, 0x4b, 0xd0, 0xf5, 0xcf, 0x91, 0xe6, 0x4b, 0x6a, 0x2e, 0x57,
	0x42, 0x93, 0x53, 0x65, 0xc2, 0x3a, 0x1d, 0xbc, 0xbe, 0xde, 0xf8, 0xde, 0xcd, 0xc6, 0xf7, 0x7e,
	0x6c, 0x7c, 0xef, 0x6a, 0xeb, 0xb7, 0x6e, 0xb6, 0x7e, 0xeb, 0xeb, 0xd6, 0x6f, 0xbd, 0x7f, 0xf6,
	0x37, 0xa9, 0xbe, 0xf9, 0x91, 0x12, 0xe6, 0x03, 0x14, 0xcb, 0xdf, 0x02, 0xbd, 0x78, 0x4e, 0x3f,
	0x56, 0x1f, 0xb9, 0x63, 0xc7, 0x1d, 0xb7, 0xd1, 0xd3, 0x5f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x5e,
	0x97, 0xd3, 0x74, 0x02, 0x03, 0x00, 0x00,
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.InflationSchedules) > 0 {
		for iNdEx := len(m.InflationSchedules) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.InflationSchedules[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintMint(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	n1, err1 := github_com_gogo_protobuf_types.StdDurationMarshalTo(m.BlockTimeThreshold, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdDuration(m.BlockTimeThreshold):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintMint(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x1a
	if len(m.MintPoolAddress) > 0 {
		i -= len(m.MintPoolAddress)
		copy(dAtA[i:], m.MintPoolAddress)
		i = encodeVarintMint(dAtA, i, uint64(len(m.MintPoolAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.MintDenom) > 0 {
		i -= len(m.MintDenom)
		copy(dAtA[i:], m.MintDenom)
		i = encodeVarintMint(dAtA, i, uint64(len(m.MintDenom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *InflationSchedule) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InflationSchedule) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *InflationSchedule) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.Amount.Size()
		i -= size
		if _, err := m.Amount.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintMint(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	n2, err2 := github_com_gogo_protobuf_types.StdTimeMarshalTo(m.EndTime, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdTime(m.EndTime):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintMint(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x12
	n3, err3 := github_com_gogo_protobuf_types.StdTimeMarshalTo(m.StartTime, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdTime(m.StartTime):])
	if err3 != nil {
		return 0, err3
	}
	i -= n3
	i = encodeVarintMint(dAtA, i, uint64(n3))
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintMint(dAtA []byte, offset int, v uint64) int {
	offset -= sovMint(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.MintDenom)
	if l > 0 {
		n += 1 + l + sovMint(uint64(l))
	}
	l = len(m.MintPoolAddress)
	if l > 0 {
		n += 1 + l + sovMint(uint64(l))
	}
	l = github_com_gogo_protobuf_types.SizeOfStdDuration(m.BlockTimeThreshold)
	n += 1 + l + sovMint(uint64(l))
	if len(m.InflationSchedules) > 0 {
		for _, e := range m.InflationSchedules {
			l = e.Size()
			n += 1 + l + sovMint(uint64(l))
		}
	}
	return n
}

func (m *InflationSchedule) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = github_com_gogo_protobuf_types.SizeOfStdTime(m.StartTime)
	n += 1 + l + sovMint(uint64(l))
	l = github_com_gogo_protobuf_types.SizeOfStdTime(m.EndTime)
	n += 1 + l + sovMint(uint64(l))
	l = m.Amount.Size()
	n += 1 + l + sovMint(uint64(l))
	return n
}

func sovMint(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMint(x uint64) (n int) {
	return sovMint(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMint
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
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MintDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
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
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MintDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MintPoolAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
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
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MintPoolAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlockTimeThreshold", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
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
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdDurationUnmarshal(&m.BlockTimeThreshold, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InflationSchedules", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
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
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.InflationSchedules = append(m.InflationSchedules, InflationSchedule{})
			if err := m.InflationSchedules[len(m.InflationSchedules)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMint(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMint
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
func (m *InflationSchedule) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMint
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
			return fmt.Errorf("proto: InflationSchedule: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InflationSchedule: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StartTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
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
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdTimeUnmarshal(&m.StartTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EndTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
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
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdTimeUnmarshal(&m.EndTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
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
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Amount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMint(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMint
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
func skipMint(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMint
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
					return 0, ErrIntOverflowMint
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
					return 0, ErrIntOverflowMint
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
				return 0, ErrInvalidLengthMint
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMint
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMint
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMint        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMint          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMint = fmt.Errorf("proto: unexpected end of group")
)
