package contracts

import (
	_ "embed" // embed compiled smart contract

	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

var (
	//go:embed compiled_contracts/callee.json
	calleeJSON []byte

	// ERC20BurnableContract is the compiled ERC20Burnable contract
	CalleeContract evmtypes.CompiledContract
)

func init() {
	// err := json.Unmarshal(calleeJSON, &CalleeContract)
	// if err != nil {
	// 	// panic(err)
	// 	fmt.Println("ERROR HERE")
	// }
}
