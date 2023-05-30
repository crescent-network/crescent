package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestDepositAmount(t *testing.T) {
	price := utils.ParseDec("12.345")
	qty := sdk.NewInt(123456789)
	require.Equal(t, "1524074061", types.DepositAmount(true, price, qty).String())
	require.Equal(t, "123456789", types.DepositAmount(false, price, qty).String())
}

func TestQuoteAmount(t *testing.T) {
	price := utils.ParseDec("12.345")
	qty := sdk.NewInt(123456789)
	require.Equal(t, "1524074061", types.QuoteAmount(true, price, qty).String())
	require.Equal(t, "1524074060", types.QuoteAmount(false, price, qty).String())
}
