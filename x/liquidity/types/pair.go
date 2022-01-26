package types

import (
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	farmingtypes "github.com/cosmosquad-labs/squad/x/farming/types"
)

func (pair Pair) GetEscrowAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(pair.EscrowAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewPair returns a new pair object.
func NewPair(id uint64, baseCoinDenom, quoteCoinDenom string) Pair {
	return Pair{
		Id:                       id,
		BaseCoinDenom:            baseCoinDenom,
		QuoteCoinDenom:           quoteCoinDenom,
		EscrowAddress:            PairEscrowAddr(id).String(),
		LastSwapRequestId:        0,
		LastCancelOrderRequestId: 0,
		LastPrice:                nil,
		CurrentBatchId:           1,
	}
}

// PairEscrowAddr returns a unique address of the pair's escrow.
func PairEscrowAddr(pairId uint64) sdk.AccAddress {
	return farmingtypes.DeriveAddress(
		AddressType,
		ModuleName,
		strings.Join([]string{PairEscrowAddrPrefix, strconv.FormatUint(pairId, 10)}, ModuleAddrNameSplitter))
}

// MustMarshalPair returns the pair bytes.
// It throws panic if it fails.
func MustMarshalPair(cdc codec.BinaryCodec, pair Pair) []byte {
	return cdc.MustMarshal(&pair)
}

// MustUnmarshalPair return the unmarshalled pair from bytes.
// It throws panic if it fails.
func MustUnmarshalPair(cdc codec.BinaryCodec, value []byte) Pair {
	pair, err := UnmarshalPair(cdc, value)
	if err != nil {
		panic(err)
	}

	return pair
}

// UnmarshalPair returns the pair from bytes.
func UnmarshalPair(cdc codec.BinaryCodec, value []byte) (pair Pair, err error) {
	err = cdc.Unmarshal(value, &pair)
	return pair, err
}
