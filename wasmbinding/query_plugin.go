package wasmbinding

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v4/wasmbinding/bindings"
)

// CustomQuerier dispatches custom CosmWasm bindings queries.
func CustomQuerier(qp *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var contractQuery bindings.CrescentQuery
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, sdkerrors.Wrap(err, "crescent query")
		}

		switch {
		case contractQuery.Pairs != nil:
			// TODO: not implemented yet

			return nil, nil
		case contractQuery.Pair != nil:
			// TODO: not implemented yet

			return nil, nil
		case contractQuery.Pools != nil:
			// TODO: not implemented yet

			return nil, nil
		case contractQuery.Pool != nil:
			// TODO: not implemented yet

			return nil, nil
		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown crescent query variant"}
		}
	}
}
