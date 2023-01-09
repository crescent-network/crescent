package wasmbinding

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	"github.com/crescent-network/crescent/v4/wasmbinding/bindings"
	liquiditykeeper "github.com/crescent-network/crescent/v4/x/liquidity/keeper"
	liquiditytypes "github.com/crescent-network/crescent/v4/x/liquidity/types"
)

// CustomMessageDecorator returns decorator for custom CosmWasm bindings messages
func CustomMessageDecorator(liquiditykeeper *liquiditykeeper.Keeper) func(wasmkeeper.Messenger) wasmkeeper.Messenger {
	return func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &CustomMessenger{
			wrapped:         old,
			liquidityKeeper: liquiditykeeper,
		}
	}
}

type CustomMessenger struct {
	wrapped         wasmkeeper.Messenger
	liquidityKeeper *liquiditykeeper.Keeper
}

var _ wasmkeeper.Messenger = (*CustomMessenger)(nil)

// DispatchMsg executes on the contractMsg.
func (m *CustomMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if msg.Custom != nil {
		var contractMsg bindings.CrescentMsg
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			return nil, nil, sdkerrors.Wrap(err, "unable to unmarshal crescent message")
		}
		if contractMsg.LimitOrder != nil {
			return m.limitOrder(ctx, contractAddr, contractMsg.LimitOrder)
		}
	}
	return m.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

func (m *CustomMessenger) limitOrder(ctx sdk.Context, contractAddr sdk.AccAddress, limitOrder *bindings.LimitOrder) ([]sdk.Event, [][]byte, error) {
	if err := DispatchLimitOrder(ctx, contractAddr, m.liquidityKeeper, limitOrder); err != nil {
		return nil, nil, sdkerrors.Wrap(err, "unable to dispatch limit order message")
	}
	return nil, nil, nil
}

func DispatchLimitOrder(ctx sdk.Context, contractAddr sdk.AccAddress, liquidityKeeeper *liquiditykeeper.Keeper, limitOrder *bindings.LimitOrder) error {
	if limitOrder == nil {
		return wasmvmtypes.InvalidRequest{Err: "limit order is null"}
	}

	msgServer := liquiditykeeper.NewMsgServerImpl(*liquidityKeeeper)

	ordererAddr, err := sdk.AccAddressFromBech32(limitOrder.Orderer)
	if err != nil {
		return err
	}

	msg := liquiditytypes.NewMsgLimitOrder(
		ordererAddr, limitOrder.PairId, limitOrder.Direction, limitOrder.OfferCoin,
		limitOrder.DemandCoinDenom, limitOrder.Price, limitOrder.Amount, limitOrder.OrderLifespan,
	)

	if err := msg.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "failed to validate MsgLimitOrder")
	}

	if _, err := msgServer.LimitOrder(sdk.WrapSDKContext(ctx), msg); err != nil {
		return sdkerrors.Wrap(err, "failed to dispatch limit order")
	}

	return nil
}
