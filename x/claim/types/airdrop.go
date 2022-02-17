package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (a Airdrop) GetSourceAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(a.SourceAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (a Airdrop) GetTerminationAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(a.TerminationAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (r ClaimRecord) GetRecipient() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.Recipient)
	if err != nil {
		panic(err)
	}
	return addr
}

func (r ClaimRecord) GetClaimableCoinsForCondition(divisor int64) sdk.Coins {
	claimableCoins := sdk.Coins{}
	for _, c := range r.ClaimableCoins {
		claimableAmt := c.Amount.Quo(sdk.NewInt(divisor))
		claimableCoins = sdk.NewCoins(sdk.NewCoin(c.Denom, claimableAmt))
	}
	return claimableCoins
}
