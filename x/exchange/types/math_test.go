package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/app/testutil"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestDepositAmount(t *testing.T) {
	price := utils.ParseDec("12.345")
	qty := sdk.NewDec(123456789)
	testutil.AssertEqual(t, utils.ParseDec("1524074060.205"), types.DepositAmount(true, price, qty))
	testutil.AssertEqual(t, utils.ParseDec("123456789"), types.DepositAmount(false, price, qty))
}

func TestQuoteAmount(t *testing.T) {
	price := utils.ParseDec("12.345")
	qty := sdk.NewDec(123456789)
	testutil.AssertEqual(t, utils.ParseDec("1524074060.205"), types.QuoteAmount(true, price, qty))
	testutil.AssertEqual(t, utils.ParseDec("1524074060.205"), types.QuoteAmount(false, price, qty))
}
