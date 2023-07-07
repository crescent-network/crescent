// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: crescent/amm/v1beta1/farming.proto

package types

import (
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"
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

type FarmingPlan struct {
	Id                 uint64                    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Description        string                    `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	FarmingPoolAddress string                    `protobuf:"bytes,3,opt,name=farming_pool_address,json=farmingPoolAddress,proto3" json:"farming_pool_address,omitempty"`
	TerminationAddress string                    `protobuf:"bytes,4,opt,name=termination_address,json=terminationAddress,proto3" json:"termination_address,omitempty"`
	RewardAllocations  []FarmingRewardAllocation `protobuf:"bytes,5,rep,name=reward_allocations,json=rewardAllocations,proto3" json:"reward_allocations"`
	StartTime          time.Time                 `protobuf:"bytes,6,opt,name=start_time,json=startTime,proto3,stdtime" json:"start_time"`
	EndTime            time.Time                 `protobuf:"bytes,7,opt,name=end_time,json=endTime,proto3,stdtime" json:"end_time"`
	IsPrivate          bool                      `protobuf:"varint,8,opt,name=is_private,json=isPrivate,proto3" json:"is_private,omitempty"`
	IsTerminated       bool                      `protobuf:"varint,9,opt,name=is_terminated,json=isTerminated,proto3" json:"is_terminated,omitempty"`
}

func (m *FarmingPlan) Reset()         { *m = FarmingPlan{} }
func (m *FarmingPlan) String() string { return proto.CompactTextString(m) }
func (*FarmingPlan) ProtoMessage()    {}
func (*FarmingPlan) Descriptor() ([]byte, []int) {
	return fileDescriptor_f1cf50e4f18be862, []int{0}
}
func (m *FarmingPlan) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FarmingPlan) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FarmingPlan.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FarmingPlan) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FarmingPlan.Merge(m, src)
}
func (m *FarmingPlan) XXX_Size() int {
	return m.Size()
}
func (m *FarmingPlan) XXX_DiscardUnknown() {
	xxx_messageInfo_FarmingPlan.DiscardUnknown(m)
}

var xxx_messageInfo_FarmingPlan proto.InternalMessageInfo

type FarmingRewardAllocation struct {
	PoolId        uint64                                   `protobuf:"varint,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	RewardsPerDay github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,2,rep,name=rewards_per_day,json=rewardsPerDay,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"rewards_per_day"`
}

func (m *FarmingRewardAllocation) Reset()         { *m = FarmingRewardAllocation{} }
func (m *FarmingRewardAllocation) String() string { return proto.CompactTextString(m) }
func (*FarmingRewardAllocation) ProtoMessage()    {}
func (*FarmingRewardAllocation) Descriptor() ([]byte, []int) {
	return fileDescriptor_f1cf50e4f18be862, []int{1}
}
func (m *FarmingRewardAllocation) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FarmingRewardAllocation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FarmingRewardAllocation.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FarmingRewardAllocation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FarmingRewardAllocation.Merge(m, src)
}
func (m *FarmingRewardAllocation) XXX_Size() int {
	return m.Size()
}
func (m *FarmingRewardAllocation) XXX_DiscardUnknown() {
	xxx_messageInfo_FarmingRewardAllocation.DiscardUnknown(m)
}

var xxx_messageInfo_FarmingRewardAllocation proto.InternalMessageInfo

func init() {
	proto.RegisterType((*FarmingPlan)(nil), "crescent.amm.v1beta1.FarmingPlan")
	proto.RegisterType((*FarmingRewardAllocation)(nil), "crescent.amm.v1beta1.FarmingRewardAllocation")
}

func init() {
	proto.RegisterFile("crescent/amm/v1beta1/farming.proto", fileDescriptor_f1cf50e4f18be862)
}

var fileDescriptor_f1cf50e4f18be862 = []byte{
	// 517 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x53, 0x4d, 0x6f, 0xd3, 0x3c,
	0x1c, 0x6f, 0xba, 0x3e, 0x7d, 0x71, 0x9f, 0x81, 0x30, 0x95, 0x16, 0x2a, 0x91, 0x46, 0xe5, 0x92,
	0x4b, 0xed, 0xbd, 0x88, 0x33, 0x5a, 0x87, 0x90, 0xb8, 0x95, 0x68, 0x27, 0x2e, 0x91, 0x13, 0x7b,
	0xc1, 0x5a, 0x12, 0x47, 0xb6, 0xd7, 0xd1, 0x6f, 0xb1, 0x6f, 0x81, 0xc4, 0x27, 0xe9, 0xb1, 0x47,
	0x4e, 0x0c, 0xda, 0x2f, 0x82, 0x62, 0xa7, 0x51, 0x85, 0xe0, 0xc0, 0xa9, 0xf5, 0xef, 0xcd, 0xf6,
	0xcf, 0xff, 0x80, 0x69, 0x22, 0x99, 0x4a, 0x58, 0xa1, 0x31, 0xc9, 0x73, 0xbc, 0x3c, 0x8b, 0x99,
	0x26, 0x67, 0xf8, 0x86, 0xc8, 0x9c, 0x17, 0x29, 0x2a, 0xa5, 0xd0, 0x02, 0x8e, 0xf6, 0x1a, 0x44,
	0xf2, 0x1c, 0xd5, 0x9a, 0xf1, 0x28, 0x15, 0xa9, 0x30, 0x02, 0x5c, 0xfd, 0xb3, 0xda, 0xb1, 0x97,
	0x08, 0x95, 0x0b, 0x85, 0x63, 0xa2, 0x58, 0x13, 0x97, 0x08, 0x5e, 0xd4, 0xfc, 0x24, 0x15, 0x22,
	0xcd, 0x18, 0x36, 0xab, 0xf8, 0xee, 0x06, 0x6b, 0x9e, 0x33, 0xa5, 0x49, 0x5e, 0x5a, 0xc1, 0x74,
	0x73, 0x04, 0x86, 0xef, 0xec, 0xf6, 0x8b, 0x8c, 0x14, 0xf0, 0x09, 0x68, 0x73, 0xea, 0x3a, 0xbe,
	0x13, 0x74, 0xc2, 0x36, 0xa7, 0xd0, 0x07, 0x43, 0xca, 0x54, 0x22, 0x79, 0xa9, 0xb9, 0x28, 0xdc,
	0xb6, 0xef, 0x04, 0x83, 0xf0, 0x10, 0x82, 0xa7, 0x60, 0x54, 0x9f, 0x3f, 0x2a, 0x85, 0xc8, 0x22,
	0x42, 0xa9, 0x64, 0x4a, 0xb9, 0x47, 0x46, 0x0a, 0x6b, 0x6e, 0x21, 0x44, 0x76, 0x69, 0x19, 0x88,
	0xc1, 0x73, 0xcd, 0x2a, 0x94, 0x54, 0x01, 0x8d, 0xa1, 0x63, 0x0d, 0x07, 0xd4, 0xde, 0x10, 0x03,
	0x28, 0xd9, 0x3d, 0x91, 0x34, 0x22, 0x59, 0x26, 0x12, 0xc3, 0x29, 0xf7, 0x3f, 0xff, 0x28, 0x18,
	0x9e, 0xcf, 0xd0, 0x9f, 0xea, 0x42, 0xf5, 0x9d, 0x42, 0x63, 0xbb, 0x6c, 0x5c, 0xf3, 0xce, 0xfa,
	0xfb, 0xa4, 0x15, 0x3e, 0x93, 0xbf, 0xe1, 0x0a, 0x5e, 0x01, 0xa0, 0x34, 0x91, 0x3a, 0xaa, 0x1a,
	0x72, 0xbb, 0xbe, 0x13, 0x0c, 0xcf, 0xc7, 0xc8, 0xd6, 0x87, 0xf6, 0xf5, 0xa1, 0xeb, 0x7d, 0x7d,
	0xf3, 0x7e, 0x15, 0xf4, 0xf0, 0x38, 0x71, 0xc2, 0x81, 0xf1, 0x55, 0x0c, 0x7c, 0x03, 0xfa, 0xac,
	0xa0, 0x36, 0xa2, 0xf7, 0x0f, 0x11, 0x3d, 0x56, 0x50, 0x13, 0xf0, 0x12, 0x00, 0xae, 0xa2, 0x52,
	0xf2, 0x25, 0xd1, 0xcc, 0xed, 0xfb, 0x4e, 0xd0, 0x0f, 0x07, 0x5c, 0x2d, 0x2c, 0x00, 0x5f, 0x81,
	0x63, 0xae, 0xa2, 0x7d, 0x43, 0x8c, 0xba, 0x03, 0xa3, 0xf8, 0x9f, 0xab, 0xeb, 0x06, 0x9b, 0x7e,
	0x71, 0xc0, 0xc9, 0x5f, 0xae, 0x0f, 0x4f, 0x40, 0xcf, 0x3c, 0x52, 0xf3, 0xc6, 0xdd, 0x6a, 0xf9,
	0x9e, 0x42, 0x05, 0x9e, 0xda, 0x4e, 0x54, 0x54, 0x32, 0x19, 0x51, 0xb2, 0x72, 0xdb, 0xa6, 0xdf,
	0x17, 0xc8, 0x8e, 0x18, 0xaa, 0x46, 0xac, 0xa9, 0xf7, 0x4a, 0xf0, 0x62, 0x7e, 0x5a, 0x9d, 0xff,
	0xeb, 0xe3, 0x24, 0x48, 0xb9, 0xfe, 0x74, 0x17, 0xa3, 0x44, 0xe4, 0xb8, 0x9e, 0x47, 0xfb, 0x33,
	0x53, 0xf4, 0x16, 0xeb, 0x55, 0xc9, 0x94, 0x31, 0xa8, 0xf0, 0xb8, 0xde, 0x63, 0xc1, 0xe4, 0x5b,
	0xb2, 0x9a, 0x7f, 0x58, 0xff, 0xf4, 0x5a, 0xeb, 0xad, 0xe7, 0x6c, 0xb6, 0x9e, 0xf3, 0x63, 0xeb,
	0x39, 0x0f, 0x3b, 0xaf, 0xb5, 0xd9, 0x79, 0xad, 0x6f, 0x3b, 0xaf, 0xf5, 0xf1, 0xe2, 0x30, 0xb6,
	0x7e, 0xe3, 0x59, 0xc1, 0xf4, 0xbd, 0x90, 0xb7, 0x0d, 0x80, 0x97, 0xaf, 0xf1, 0x67, 0xf3, 0x31,
	0x99, 0x7d, 0xe2, 0xae, 0xe9, 0xf9, 0xe2, 0x57, 0x00, 0x00, 0x00, 0xff, 0xff, 0x5d, 0x34, 0x54,
	0x85, 0x69, 0x03, 0x00, 0x00,
}

func (m *FarmingPlan) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FarmingPlan) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FarmingPlan) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.IsTerminated {
		i--
		if m.IsTerminated {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x48
	}
	if m.IsPrivate {
		i--
		if m.IsPrivate {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x40
	}
	n1, err1 := github_com_gogo_protobuf_types.StdTimeMarshalTo(m.EndTime, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdTime(m.EndTime):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintFarming(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x3a
	n2, err2 := github_com_gogo_protobuf_types.StdTimeMarshalTo(m.StartTime, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdTime(m.StartTime):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintFarming(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x32
	if len(m.RewardAllocations) > 0 {
		for iNdEx := len(m.RewardAllocations) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.RewardAllocations[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintFarming(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x2a
		}
	}
	if len(m.TerminationAddress) > 0 {
		i -= len(m.TerminationAddress)
		copy(dAtA[i:], m.TerminationAddress)
		i = encodeVarintFarming(dAtA, i, uint64(len(m.TerminationAddress)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.FarmingPoolAddress) > 0 {
		i -= len(m.FarmingPoolAddress)
		copy(dAtA[i:], m.FarmingPoolAddress)
		i = encodeVarintFarming(dAtA, i, uint64(len(m.FarmingPoolAddress)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintFarming(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if m.Id != 0 {
		i = encodeVarintFarming(dAtA, i, uint64(m.Id))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *FarmingRewardAllocation) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FarmingRewardAllocation) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FarmingRewardAllocation) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.RewardsPerDay) > 0 {
		for iNdEx := len(m.RewardsPerDay) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.RewardsPerDay[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintFarming(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if m.PoolId != 0 {
		i = encodeVarintFarming(dAtA, i, uint64(m.PoolId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintFarming(dAtA []byte, offset int, v uint64) int {
	offset -= sovFarming(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *FarmingPlan) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Id != 0 {
		n += 1 + sovFarming(uint64(m.Id))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovFarming(uint64(l))
	}
	l = len(m.FarmingPoolAddress)
	if l > 0 {
		n += 1 + l + sovFarming(uint64(l))
	}
	l = len(m.TerminationAddress)
	if l > 0 {
		n += 1 + l + sovFarming(uint64(l))
	}
	if len(m.RewardAllocations) > 0 {
		for _, e := range m.RewardAllocations {
			l = e.Size()
			n += 1 + l + sovFarming(uint64(l))
		}
	}
	l = github_com_gogo_protobuf_types.SizeOfStdTime(m.StartTime)
	n += 1 + l + sovFarming(uint64(l))
	l = github_com_gogo_protobuf_types.SizeOfStdTime(m.EndTime)
	n += 1 + l + sovFarming(uint64(l))
	if m.IsPrivate {
		n += 2
	}
	if m.IsTerminated {
		n += 2
	}
	return n
}

func (m *FarmingRewardAllocation) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PoolId != 0 {
		n += 1 + sovFarming(uint64(m.PoolId))
	}
	if len(m.RewardsPerDay) > 0 {
		for _, e := range m.RewardsPerDay {
			l = e.Size()
			n += 1 + l + sovFarming(uint64(l))
		}
	}
	return n
}

func sovFarming(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozFarming(x uint64) (n int) {
	return sovFarming(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *FarmingPlan) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowFarming
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
			return fmt.Errorf("proto: FarmingPlan: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FarmingPlan: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			m.Id = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFarming
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Id |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFarming
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
				return ErrInvalidLengthFarming
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthFarming
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FarmingPoolAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFarming
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
				return ErrInvalidLengthFarming
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthFarming
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.FarmingPoolAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TerminationAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFarming
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
				return ErrInvalidLengthFarming
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthFarming
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TerminationAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RewardAllocations", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFarming
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
				return ErrInvalidLengthFarming
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthFarming
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RewardAllocations = append(m.RewardAllocations, FarmingRewardAllocation{})
			if err := m.RewardAllocations[len(m.RewardAllocations)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StartTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFarming
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
				return ErrInvalidLengthFarming
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthFarming
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdTimeUnmarshal(&m.StartTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EndTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFarming
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
				return ErrInvalidLengthFarming
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthFarming
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdTimeUnmarshal(&m.EndTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IsPrivate", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFarming
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.IsPrivate = bool(v != 0)
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IsTerminated", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFarming
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.IsTerminated = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipFarming(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthFarming
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
func (m *FarmingRewardAllocation) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowFarming
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
			return fmt.Errorf("proto: FarmingRewardAllocation: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FarmingRewardAllocation: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolId", wireType)
			}
			m.PoolId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFarming
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
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RewardsPerDay", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFarming
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
				return ErrInvalidLengthFarming
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthFarming
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RewardsPerDay = append(m.RewardsPerDay, types.Coin{})
			if err := m.RewardsPerDay[len(m.RewardsPerDay)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipFarming(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthFarming
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
func skipFarming(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowFarming
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
					return 0, ErrIntOverflowFarming
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
					return 0, ErrIntOverflowFarming
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
				return 0, ErrInvalidLengthFarming
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupFarming
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthFarming
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthFarming        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowFarming          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupFarming = fmt.Errorf("proto: unexpected end of group")
)