package wasmbinding

import (
	"encoding/json"
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v4/wasmbinding/bindings"
)

// CustomQuerier dispatches custom bindings queries.
func CustomQuerier(qp *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var contractQuery bindings.CrescentQuery
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, sdkerrors.Wrap(err, "crescent query")
		}

		switch {
		case contractQuery.Pairs != nil:
			pairs := qp.Pairs(ctx)

			bz, err := json.Marshal(&pairs)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}
			return bz, nil

		case contractQuery.Pair != nil:
			pair, err := qp.Pair(ctx, contractQuery.Pair.Id)
			if err != nil {
				return nil, err
			}

			bz, err := json.Marshal(&pair)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}
			return bz, nil

		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown crescent query variant"}
		}
	}
}
