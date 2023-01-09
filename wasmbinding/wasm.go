package wasmbinding

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	liquiditykeeper "github.com/crescent-network/crescent/v4/x/liquidity/keeper"
)

func RegisterCustomPlugin(liquiditykeeper *liquiditykeeper.Keeper) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(liquiditykeeper)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})

	messengerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(liquiditykeeper),
	)

	return []wasm.Option{
		queryPluginOpt,
		messengerDecoratorOpt,
	}
}
