// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: crescent/erc20/v1/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

// MsgConvertCoin defines a Msg to convert a native Cosmos coin to a ERC20 token
type MsgConvertCoin struct {
	// Cosmos coin which denomination is registered in a token pair. The coin
	// amount defines the amount of coins to convert.
	Coin types.Coin `protobuf:"bytes,1,opt,name=coin,proto3" json:"coin"`
	// recipient hex address to receive ERC20 token
	Receiver string `protobuf:"bytes,2,opt,name=receiver,proto3" json:"receiver,omitempty"`
	// cosmos bech32 address from the owner of the given Cosmos coins
	Sender string `protobuf:"bytes,3,opt,name=sender,proto3" json:"sender,omitempty"`
}

func (m *MsgConvertCoin) Reset()         { *m = MsgConvertCoin{} }
func (m *MsgConvertCoin) String() string { return proto.CompactTextString(m) }
func (*MsgConvertCoin) ProtoMessage()    {}
func (*MsgConvertCoin) Descriptor() ([]byte, []int) {
	return fileDescriptor_e8c16698ae8aba59, []int{0}
}
func (m *MsgConvertCoin) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgConvertCoin) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgConvertCoin.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgConvertCoin) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgConvertCoin.Merge(m, src)
}
func (m *MsgConvertCoin) XXX_Size() int {
	return m.Size()
}
func (m *MsgConvertCoin) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgConvertCoin.DiscardUnknown(m)
}

var xxx_messageInfo_MsgConvertCoin proto.InternalMessageInfo

func (m *MsgConvertCoin) GetCoin() types.Coin {
	if m != nil {
		return m.Coin
	}
	return types.Coin{}
}

func (m *MsgConvertCoin) GetReceiver() string {
	if m != nil {
		return m.Receiver
	}
	return ""
}

func (m *MsgConvertCoin) GetSender() string {
	if m != nil {
		return m.Sender
	}
	return ""
}

// MsgConvertCoinResponse returns no fields
type MsgConvertCoinResponse struct {
}

func (m *MsgConvertCoinResponse) Reset()         { *m = MsgConvertCoinResponse{} }
func (m *MsgConvertCoinResponse) String() string { return proto.CompactTextString(m) }
func (*MsgConvertCoinResponse) ProtoMessage()    {}
func (*MsgConvertCoinResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_e8c16698ae8aba59, []int{1}
}
func (m *MsgConvertCoinResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgConvertCoinResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgConvertCoinResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgConvertCoinResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgConvertCoinResponse.Merge(m, src)
}
func (m *MsgConvertCoinResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgConvertCoinResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgConvertCoinResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgConvertCoinResponse proto.InternalMessageInfo

// MsgConvertERC20 defines a Msg to convert a ERC20 token to a native Cosmos
// coin.
type MsgConvertERC20 struct {
	// ERC20 token contract address registered in a token pair
	ContractAddress string `protobuf:"bytes,1,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
	// amount of ERC20 tokens to convert
	Amount github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,2,opt,name=amount,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"amount"`
	// bech32 address to receive native Cosmos coins
	Receiver string `protobuf:"bytes,3,opt,name=receiver,proto3" json:"receiver,omitempty"`
	// sender hex address from the owner of the given ERC20 tokens
	Sender string `protobuf:"bytes,4,opt,name=sender,proto3" json:"sender,omitempty"`
}

func (m *MsgConvertERC20) Reset()         { *m = MsgConvertERC20{} }
func (m *MsgConvertERC20) String() string { return proto.CompactTextString(m) }
func (*MsgConvertERC20) ProtoMessage()    {}
func (*MsgConvertERC20) Descriptor() ([]byte, []int) {
	return fileDescriptor_e8c16698ae8aba59, []int{2}
}
func (m *MsgConvertERC20) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgConvertERC20) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgConvertERC20.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgConvertERC20) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgConvertERC20.Merge(m, src)
}
func (m *MsgConvertERC20) XXX_Size() int {
	return m.Size()
}
func (m *MsgConvertERC20) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgConvertERC20.DiscardUnknown(m)
}

var xxx_messageInfo_MsgConvertERC20 proto.InternalMessageInfo

func (m *MsgConvertERC20) GetContractAddress() string {
	if m != nil {
		return m.ContractAddress
	}
	return ""
}

func (m *MsgConvertERC20) GetReceiver() string {
	if m != nil {
		return m.Receiver
	}
	return ""
}

func (m *MsgConvertERC20) GetSender() string {
	if m != nil {
		return m.Sender
	}
	return ""
}

// MsgConvertERC20Response returns no fields
type MsgConvertERC20Response struct {
}

func (m *MsgConvertERC20Response) Reset()         { *m = MsgConvertERC20Response{} }
func (m *MsgConvertERC20Response) String() string { return proto.CompactTextString(m) }
func (*MsgConvertERC20Response) ProtoMessage()    {}
func (*MsgConvertERC20Response) Descriptor() ([]byte, []int) {
	return fileDescriptor_e8c16698ae8aba59, []int{3}
}
func (m *MsgConvertERC20Response) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgConvertERC20Response) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgConvertERC20Response.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgConvertERC20Response) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgConvertERC20Response.Merge(m, src)
}
func (m *MsgConvertERC20Response) XXX_Size() int {
	return m.Size()
}
func (m *MsgConvertERC20Response) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgConvertERC20Response.DiscardUnknown(m)
}

var xxx_messageInfo_MsgConvertERC20Response proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgConvertCoin)(nil), "crescent.erc20.v1.MsgConvertCoin")
	proto.RegisterType((*MsgConvertCoinResponse)(nil), "crescent.erc20.v1.MsgConvertCoinResponse")
	proto.RegisterType((*MsgConvertERC20)(nil), "crescent.erc20.v1.MsgConvertERC20")
	proto.RegisterType((*MsgConvertERC20Response)(nil), "crescent.erc20.v1.MsgConvertERC20Response")
}

func init() { proto.RegisterFile("crescent/erc20/v1/tx.proto", fileDescriptor_e8c16698ae8aba59) }

var fileDescriptor_e8c16698ae8aba59 = []byte{
	// 468 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x52, 0x4d, 0x6f, 0x13, 0x31,
	0x10, 0x8d, 0x93, 0x28, 0xa2, 0x2e, 0xa2, 0x60, 0xa1, 0x92, 0xae, 0xd0, 0xb6, 0x2c, 0x1f, 0x6a,
	0x8b, 0x6a, 0x37, 0xa9, 0xfa, 0x03, 0x48, 0x04, 0x12, 0x87, 0x0a, 0x69, 0x8f, 0x5c, 0x2a, 0xc7,
	0x19, 0x2d, 0xab, 0x12, 0x4f, 0x64, 0xbb, 0x4b, 0x7b, 0xe5, 0xc2, 0x81, 0x0b, 0x12, 0x67, 0xfe,
	0x07, 0x3f, 0xa1, 0xc7, 0x4a, 0x5c, 0x10, 0x87, 0x0a, 0x25, 0xfc, 0x10, 0xb4, 0xde, 0x4d, 0xda,
	0xe5, 0xa3, 0x39, 0xed, 0x7a, 0xde, 0x9b, 0xe7, 0x37, 0x6f, 0x4c, 0x03, 0x65, 0xc0, 0x2a, 0xd0,
	0x4e, 0x80, 0x51, 0xdd, 0x5d, 0x91, 0x75, 0x84, 0x3b, 0xe1, 0x63, 0x83, 0x0e, 0xd9, 0x9d, 0x19,
	0xc6, 0x3d, 0xc6, 0xb3, 0x4e, 0x70, 0x3f, 0x41, 0x4c, 0xde, 0x82, 0x90, 0xe3, 0x54, 0x48, 0xad,
	0xd1, 0x49, 0x97, 0xa2, 0xb6, 0x45, 0x43, 0x70, 0x37, 0xc1, 0x04, 0xfd, 0xaf, 0xc8, 0xff, 0xca,
	0x6a, 0xa8, 0xd0, 0x8e, 0xd0, 0x8a, 0x81, 0xb4, 0x20, 0xb2, 0xce, 0x00, 0x9c, 0xec, 0x08, 0x85,
	0xa9, 0x2e, 0xf0, 0xe8, 0x94, 0xde, 0x3a, 0xb0, 0x49, 0x1f, 0x75, 0x06, 0xc6, 0xf5, 0x31, 0xd5,
	0x6c, 0x8f, 0x36, 0x73, 0xbc, 0x4d, 0x36, 0xc8, 0xe6, 0x72, 0x77, 0x8d, 0x17, 0x02, 0x3c, 0x17,
	0xe0, 0xa5, 0x00, 0xcf, 0x89, 0xbd, 0xe6, 0xd9, 0xc5, 0x7a, 0x2d, 0xf6, 0x64, 0x16, 0xd0, 0x1b,
	0x06, 0x14, 0xa4, 0x19, 0x98, 0x76, 0x7d, 0x83, 0x6c, 0x2e, 0xc5, 0xf3, 0x33, 0x5b, 0xa5, 0x2d,
	0x0b, 0x7a, 0x08, 0xa6, 0xdd, 0xf0, 0x48, 0x79, 0x8a, 0xda, 0x74, 0xb5, 0x7a, 0x75, 0x0c, 0x76,
	0x8c, 0xda, 0x42, 0xf4, 0x95, 0xd0, 0x95, 0x4b, 0xe8, 0x79, 0xdc, 0xef, 0xee, 0xb2, 0x2d, 0x7a,
	0x5b, 0xa1, 0x76, 0x46, 0x2a, 0x77, 0x28, 0x87, 0x43, 0x03, 0xd6, 0x7a, 0x8b, 0x4b, 0xf1, 0xca,
	0xac, 0xfe, 0xac, 0x28, 0xb3, 0x17, 0xb4, 0x25, 0x47, 0x78, 0xac, 0x5d, 0x61, 0xa5, 0xc7, 0x73,
	0xa3, 0x3f, 0x2e, 0xd6, 0x9f, 0x24, 0xa9, 0x7b, 0x73, 0x3c, 0xe0, 0x0a, 0x47, 0xa2, 0x8c, 0xa5,
	0xf8, 0xec, 0xd8, 0xe1, 0x91, 0x70, 0xa7, 0x63, 0xb0, 0xfc, 0xa5, 0x76, 0x71, 0xd9, 0x5d, 0x19,
	0xaa, 0xf1, 0xdf, 0xa1, 0x9a, 0x95, 0xa1, 0xd6, 0xe8, 0xbd, 0x3f, 0x9c, 0xcf, 0xa6, 0xea, 0x7e,
	0xa9, 0xd3, 0xc6, 0x81, 0x4d, 0xd8, 0x07, 0x42, 0x97, 0xaf, 0x06, 0xfe, 0x80, 0xff, 0xb5, 0x6a,
	0x5e, 0x0d, 0x26, 0xd8, 0x5a, 0x48, 0x99, 0x67, 0xb7, 0xfd, 0xfe, 0xdb, 0xaf, 0xcf, 0xf5, 0x47,
	0x2c, 0x12, 0xff, 0x7a, 0x5c, 0x42, 0x15, 0x2d, 0x87, 0x7e, 0x6b, 0x1f, 0x09, 0xbd, 0x59, 0x09,
	0x39, 0xba, 0xf6, 0x1e, 0xcf, 0x09, 0xb6, 0x17, 0x73, 0xe6, 0x66, 0x9e, 0x7a, 0x33, 0x8f, 0xd9,
	0xc3, 0xeb, 0xcd, 0xf8, 0x5a, 0xef, 0xd5, 0xd9, 0x24, 0x24, 0xe7, 0x93, 0x90, 0xfc, 0x9c, 0x84,
	0xe4, 0xd3, 0x34, 0xac, 0x9d, 0x4f, 0xc3, 0xda, 0xf7, 0x69, 0x58, 0x7b, 0xbd, 0x7f, 0x75, 0x71,
	0xa5, 0xd0, 0x8e, 0x06, 0xf7, 0x0e, 0xcd, 0xd1, 0xa5, 0x72, 0xb6, 0x2f, 0x4e, 0x4a, 0x79, 0xbf,
	0xcb, 0x41, 0xcb, 0x3f, 0xf1, 0xbd, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0xa9, 0x21, 0x57, 0x58,
	0x67, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	// ConvertCoin mints a ERC20 representation of the native Cosmos coin denom
	// that is registered on the token mapping.
	ConvertCoin(ctx context.Context, in *MsgConvertCoin, opts ...grpc.CallOption) (*MsgConvertCoinResponse, error)
	// ConvertERC20 mints a native Cosmos coin representation of the ERC20 token
	// contract that is registered on the token mapping.
	ConvertERC20(ctx context.Context, in *MsgConvertERC20, opts ...grpc.CallOption) (*MsgConvertERC20Response, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) ConvertCoin(ctx context.Context, in *MsgConvertCoin, opts ...grpc.CallOption) (*MsgConvertCoinResponse, error) {
	out := new(MsgConvertCoinResponse)
	err := c.cc.Invoke(ctx, "/crescent.erc20.v1.Msg/ConvertCoin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) ConvertERC20(ctx context.Context, in *MsgConvertERC20, opts ...grpc.CallOption) (*MsgConvertERC20Response, error) {
	out := new(MsgConvertERC20Response)
	err := c.cc.Invoke(ctx, "/crescent.erc20.v1.Msg/ConvertERC20", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// ConvertCoin mints a ERC20 representation of the native Cosmos coin denom
	// that is registered on the token mapping.
	ConvertCoin(context.Context, *MsgConvertCoin) (*MsgConvertCoinResponse, error)
	// ConvertERC20 mints a native Cosmos coin representation of the ERC20 token
	// contract that is registered on the token mapping.
	ConvertERC20(context.Context, *MsgConvertERC20) (*MsgConvertERC20Response, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) ConvertCoin(ctx context.Context, req *MsgConvertCoin) (*MsgConvertCoinResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConvertCoin not implemented")
}
func (*UnimplementedMsgServer) ConvertERC20(ctx context.Context, req *MsgConvertERC20) (*MsgConvertERC20Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConvertERC20 not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_ConvertCoin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgConvertCoin)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ConvertCoin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crescent.erc20.v1.Msg/ConvertCoin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ConvertCoin(ctx, req.(*MsgConvertCoin))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ConvertERC20_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgConvertERC20)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ConvertERC20(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crescent.erc20.v1.Msg/ConvertERC20",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ConvertERC20(ctx, req.(*MsgConvertERC20))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "crescent.erc20.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ConvertCoin",
			Handler:    _Msg_ConvertCoin_Handler,
		},
		{
			MethodName: "ConvertERC20",
			Handler:    _Msg_ConvertERC20_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "crescent/erc20/v1/tx.proto",
}

func (m *MsgConvertCoin) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgConvertCoin) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgConvertCoin) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Sender) > 0 {
		i -= len(m.Sender)
		copy(dAtA[i:], m.Sender)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Sender)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Receiver) > 0 {
		i -= len(m.Receiver)
		copy(dAtA[i:], m.Receiver)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Receiver)))
		i--
		dAtA[i] = 0x12
	}
	{
		size, err := m.Coin.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *MsgConvertCoinResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgConvertCoinResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgConvertCoinResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgConvertERC20) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgConvertERC20) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgConvertERC20) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Sender) > 0 {
		i -= len(m.Sender)
		copy(dAtA[i:], m.Sender)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Sender)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Receiver) > 0 {
		i -= len(m.Receiver)
		copy(dAtA[i:], m.Receiver)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Receiver)))
		i--
		dAtA[i] = 0x1a
	}
	{
		size := m.Amount.Size()
		i -= size
		if _, err := m.Amount.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.ContractAddress) > 0 {
		i -= len(m.ContractAddress)
		copy(dAtA[i:], m.ContractAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ContractAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgConvertERC20Response) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgConvertERC20Response) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgConvertERC20Response) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgConvertCoin) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Coin.Size()
	n += 1 + l + sovTx(uint64(l))
	l = len(m.Receiver)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgConvertCoinResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgConvertERC20) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ContractAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = m.Amount.Size()
	n += 1 + l + sovTx(uint64(l))
	l = len(m.Receiver)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgConvertERC20Response) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgConvertCoin) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgConvertCoin: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgConvertCoin: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Coin", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Coin.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Receiver", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Receiver = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgConvertCoinResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgConvertCoinResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgConvertCoinResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgConvertERC20) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgConvertERC20: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgConvertERC20: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContractAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContractAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Amount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Receiver", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Receiver = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgConvertERC20Response) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgConvertERC20Response: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgConvertERC20Response: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
