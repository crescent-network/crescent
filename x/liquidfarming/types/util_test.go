package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func TestPayingReserveAddress(t *testing.T) {
	config := sdk.GetConfig()
	addrPrefix := config.GetBech32AccountAddrPrefix()

	testAcc1 := types.DeriveBidReserveAddress(1)
	require.Equal(t, testAcc1, sdk.AccAddress(address.Module(types.ModuleName, []byte("DeriveBidReserveAddress|1"))))
	require.Equal(t, addrPrefix+"1h72q3pkvsz537kj08hyv20tun3apampxhpgad97t3ls47nukgtxqeq6eu2", testAcc1.String())

	testAcc2 := types.DeriveBidReserveAddress(22)
	require.Equal(t, testAcc2, sdk.AccAddress(address.Module(types.ModuleName, []byte("DeriveBidReserveAddress|22"))))
	require.Equal(t, addrPrefix+"1tepnmaep852l483ldnfxttgsua9j9ynpmelqmn3ywvwynr7s5acqr6sz4k", testAcc2.String())
}

func TestWithdrawnRewardsReserveAddress(t *testing.T) {
	config := sdk.GetConfig()
	addrPrefix := config.GetBech32AccountAddrPrefix()

	testAcc1 := types.WithdrawnRewardsReserveAddress(1)
	require.Equal(t, testAcc1, sdk.AccAddress(address.Module(types.ModuleName, []byte("WithdrawnRewardsReserveAddress|1"))))
	require.Equal(t, addrPrefix+"1f3x2x5dl6fdsttf5temu2tg03k7vc8w9tstc95y73sm73wyeav2qgs8rmt", testAcc1.String())

	testAcc2 := types.WithdrawnRewardsReserveAddress(22)
	require.Equal(t, testAcc2, sdk.AccAddress(address.Module(types.ModuleName, []byte("WithdrawnRewardsReserveAddress|22"))))
	require.Equal(t, addrPrefix+"19fl505lmma2c56ukdfxhjh02d0jv2kqp4rw3p0frvp09y33wt3rqc7u8gm", testAcc2.String())
}
