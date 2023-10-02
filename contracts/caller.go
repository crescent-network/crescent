package contracts

import (
	_ "embed" // embed compiled smart contract

	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

var (
	//go:embed compiled_contracts/caller.json
	callerJSON []byte

	// ERC20BurnableContract is the compiled ERC20Burnable contract
	CallerContract evmtypes.CompiledContract
)

func init() {
	// err := json.Unmarshal(callerJSON, &CallerContract)
	// if err != nil {
	// 	panic(err)
	// }
}
