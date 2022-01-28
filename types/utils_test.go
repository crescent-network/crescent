package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmosquad-labs/squad/types"
	"github.com/stretchr/testify/require"
)

func TestGetShareValue(t *testing.T) {
	require.EqualValues(t, types.GetShareValue(sdk.NewInt(100), sdk.MustNewDecFromStr("0.9")), sdk.NewInt(90))
	require.EqualValues(t, types.GetShareValue(sdk.NewInt(100), sdk.MustNewDecFromStr("1.1")), sdk.NewInt(110))

	// truncated
	require.EqualValues(t, types.GetShareValue(sdk.NewInt(101), sdk.MustNewDecFromStr("0.9")), sdk.NewInt(90))
	require.EqualValues(t, types.GetShareValue(sdk.NewInt(101), sdk.MustNewDecFromStr("1.1")), sdk.NewInt(111))

	require.EqualValues(t, types.GetShareValue(sdk.NewInt(100), sdk.MustNewDecFromStr("0")), sdk.NewInt(0))
	require.EqualValues(t, types.GetShareValue(sdk.NewInt(0), sdk.MustNewDecFromStr("1.1")), sdk.NewInt(0))
}

func TestAddOrInit(t *testing.T) {
	strIntMap := make(types.StrIntMap)

	// Set when the key not existed on the map
	strIntMap.AddOrSet("a", sdk.NewInt(1))
	require.Equal(t, strIntMap["a"], sdk.NewInt(1))

	// Added when the key existed on the map
	strIntMap.AddOrSet("a", sdk.NewInt(1))
	require.Equal(t, strIntMap["a"], sdk.NewInt(2))
}
