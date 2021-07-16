package app

// DONTCOVER

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	farmingkeeper "github.com/tendermint/farming/x/farming/keeper"
	"github.com/tendermint/farming/x/farming/types"
	farmingtypes "github.com/tendermint/farming/x/farming/types"
)

// DefaultConsensusParams defines the default Tendermint consensus params used in
// FarmingApp testing.
var DefaultConsensusParams = &abci.ConsensusParams{
	Block: &abci.BlockParams{
		MaxBytes: 200000,
		MaxGas:   2000000,
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
}

// Setup initializes a new FarmingApp. A Nop logger is set in FarmingApp.
func Setup(isCheckTx bool) *FarmingApp {
	db := dbm.NewMemDB()
	app := NewFarmingApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, MakeEncodingConfig(), EmptyAppOptions{})
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		genesisState := NewDefaultGenesisState()
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		app.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}

// CreateTestInput returns a simapp with custom FarmingKeeper to avoid
// messing with the hooks.
func CreateTestInput() (*FarmingApp, sdk.Context) {
	cdc := codec.NewLegacyAmino()
	farmingtypes.RegisterLegacyAminoCodec(cdc)

	app := Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	appCodec := app.AppCodec()

	blockedAddrs := map[string]bool{}

	app.FarmingKeeper = farmingkeeper.NewKeeper(
		appCodec,
		app.GetKey(types.StoreKey),
		app.GetSubspace(types.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.DistrKeeper,
		blockedAddrs,
	)

	return app, ctx
}

// createIncrementalAccounts is a strategy used by addTestAddrs() in order to generated addresses in ascending order.
func createIncrementalAccounts(accNum int) []sdk.AccAddress {
	var addresses []sdk.AccAddress
	var buffer bytes.Buffer

	// start at 100 so we can make up to 999 test addresses with valid test addresses
	for i := 100; i < (accNum + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6") // base address string

		buffer.WriteString(numString) // adding on final two digits to make addresses unique
		res, _ := sdk.AccAddressFromHex(buffer.String())
		bech := res.String()
		addr, _ := TestAddr(buffer.String(), bech)

		addresses = append(addresses, addr)
		buffer.Reset()
	}

	return addresses
}

// AddTestAddrs constructs and returns accNum amount of accounts with an
// initial balance of accAmt in random order
func AddTestAddrs(app *FarmingApp, ctx sdk.Context, accNum int, initCoins sdk.Coins) []sdk.AccAddress {
	testAddrs := createIncrementalAccounts(accNum)
	for _, addr := range testAddrs {
		fmt.Println(addr)
		// addTotalSupply(app, ctx, initCoins)
		// SaveAccount(app, ctx, addr, initCoins)
	}
	return testAddrs
}

// TODO: update the methods in accordance with SDK 0.43.0-rc0
// func addTotalSupply(app *FarmingApp, ctx sdk.Context, coins sdk.Coins) {
// 	for _, coin := range coins {
// 		prevSupply := app.BankKeeper.GetSupply(ctx, coin.Denom)
// 		app.BankKeeper.AllBalances(context.Context, *banktypes.QueryAllBalancesRequest)
// 		app.BankKeeper.SetSupply(ctx, banktypes.NewSupply(prevSupply.GetTotal().Add(coins...)))
// 	}
// }

// // setTotalSupply provides the total supply based on accAmt * totalAccounts.
// func setTotalSupply(app *FarmingApp, ctx sdk.Context, accAmt sdk.Int, totalAccounts int) {
// 	totalSupply := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), accAmt.MulRaw(int64(totalAccounts))))
// 	prevSupply := app.BankKeeper.GetSupply(ctx)
// 	app.BankKeeper.SetSupply(ctx, banktypes.NewSupply(prevSupply.GetTotal().Add(totalSupply...)))
// }

// // SaveAccount saves the provided account into the simapp with balance based on initCoins.
// func SaveAccount(app *FarmingApp, ctx sdk.Context, addr sdk.AccAddress, initCoins sdk.Coins) {
// 	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
// 	app.AccountKeeper.SetAccount(ctx, acc)
// 	err := app.BankKeeper.AddCoins(ctx, addr, initCoins)
// 	if err != nil {
// 		panic(err)
// 	}
// }

func TestAddr(addr string, bech string) (sdk.AccAddress, error) {
	res, err := sdk.AccAddressFromHex(addr)
	if err != nil {
		return nil, err
	}
	bechexpected := res.String()
	if bech != bechexpected {
		return nil, fmt.Errorf("bech encoding doesn't match reference")
	}

	bechres, err := sdk.AccAddressFromBech32(bech)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(bechres, res) {
		return nil, err
	}

	return res, nil
}

// EmptyAppOptions is a stub implementing AppOptions
type EmptyAppOptions struct{}

// Get implements AppOptions
func (ao EmptyAppOptions) Get(o string) interface{} {
	return nil
}
