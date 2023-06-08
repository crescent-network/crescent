package app

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
	liquidstakingtypes "github.com/crescent-network/crescent/v5/x/liquidstaking/types"
)

func TestSplitMidBlockTxs(t *testing.T) {
	txDecoder := MakeTestEncodingConfig().TxConfig.TxDecoder()
	txEncoder := MakeTestEncodingConfig().TxConfig.TxEncoder()
	txBuilder := MakeTestEncodingConfig().TxConfig.NewTxBuilder()

	getTx := func(msgs ...sdk.Msg) []byte {
		err := txBuilder.SetMsgs(msgs...)
		require.NoError(t, err)
		tx, err := txEncoder(txBuilder.GetTx())
		require.NoError(t, err)
		return tx
	}

	singleNormalTx := getTx(&banktypes.MsgSend{})
	multipleNormalTx := getTx(&liquidstakingtypes.MsgLiquidStake{}, &banktypes.MsgSend{})
	singleMidBlockTx := getTx(&exchangetypes.MsgPlaceBatchLimitOrder{})
	multipleMidBlockTx := getTx(&exchangetypes.MsgPlaceBatchLimitOrder{}, &exchangetypes.MsgPlaceMMBatchLimitOrder{})
	normalWithMidBlockTx := getTx(&banktypes.MsgSend{}, &exchangetypes.MsgPlaceBatchLimitOrder{})
	midBlockWithNormalTx := getTx(&exchangetypes.MsgPlaceBatchLimitOrder{}, &banktypes.MsgSend{})

	txs := [][]byte{
		singleNormalTx,
		multipleNormalTx,
		singleMidBlockTx,
		multipleMidBlockTx,
		normalWithMidBlockTx,
		midBlockWithNormalTx,
	}

	midBlockTxs, normalTxs := SplitMidBlockTxs(txs, txDecoder)

	require.Equal(t, len(txs), len(midBlockTxs)+len(normalTxs))
	require.EqualValues(t, midBlockTxs, [][]byte{singleMidBlockTx, multipleMidBlockTx})
	require.EqualValues(t, normalTxs, [][]byte{singleNormalTx, multipleNormalTx, normalWithMidBlockTx, midBlockWithNormalTx})

	midBlockTxs, normalTxs = SplitMidBlockTxs([][]byte{}, txDecoder)
	require.Len(t, midBlockTxs, 0)
	require.Len(t, normalTxs, 0)
	require.EqualValues(t, midBlockTxs, [][]byte(nil))
	require.EqualValues(t, normalTxs, [][]byte(nil))
}
