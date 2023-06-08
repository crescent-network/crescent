package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (app *App) MidBlocker(ctx sdk.Context, req abci.RequestMidBlock) abci.ResponseMidBlock {
	midBlockTxs, normalTxs := SplitMidBlockTxs(req.Txs, app.TxDecoder)

	idx := 0
	txResults := make([]*abci.ResponseDeliverTx, len(req.Txs))

	// run mid-block txs first
	for _, tx := range midBlockTxs {
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: tx})
		txResults[idx] = &res
		idx++
	}

	// run mid-block for each module
	events := app.mm.MidBlock(ctx)

	// run normal txs after mid-block
	for _, tx := range normalTxs {
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: tx})
		txResults[idx] = &res
		idx++
	}

	// mid-block events would be in end-block events
	return abci.ResponseMidBlock{DeliverTxs: txResults, Events: events}
}

func SplitMidBlockTxs(txs [][]byte, txDecoder sdk.TxDecoder) (midBlockTxs, normalTxs [][]byte) {
	for _, rawTx := range txs {
		tx, err := txDecoder(rawTx)
		if err != nil {
			normalTxs = append(normalTxs, rawTx)
			continue
		}
		midBlockTx := false

	loop:
		for _, msg := range tx.GetMsgs() {
			switch msg.(type) {
			case *exchangetypes.MsgPlaceBatchLimitOrder:
				midBlockTx = true
			case *exchangetypes.MsgPlaceMMBatchLimitOrder:
				midBlockTx = true
			// TODO: add cancel mm order msg
			default:
				midBlockTx = false
				break loop
			}
		}

		if midBlockTx {
			midBlockTxs = append(midBlockTxs, rawTx)
		} else {
			normalTxs = append(normalTxs, rawTx)
		}
	}
	return midBlockTxs, normalTxs
}
