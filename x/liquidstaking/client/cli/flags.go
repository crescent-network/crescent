package cli

const (
	// TODO: add status flag
	//FlagName               = "name"
	//FlagSourceAddress      = "source-address"
	//FlagDestinationAddress = "destination-address"
	FlagType       = "type"
	FlagModuleName = "module-name"
)

//// flagSetLiquidValidators returns the FlagSet used for liquidStakings.
//func flagSetLiquidValidators() *flag.FlagSet {
//	fs := flag.NewFlagSet("", flag.ContinueOnError)
//
//	//fs.String(FlagName, "", "The liquidstaking name")
//	//fs.String(FlagSourceAddress, "", "The bech32 address of the source account")
//	//fs.String(FlagDestinationAddress, "", "The bech32 address of the destination account")
//
//	return fs
//}
//
//// flagSetAddress returns the FlagSet used for address.
//func flagSetAddress() *flag.FlagSet {
//	fs := flag.NewFlagSet("", flag.ContinueOnError)
//
//	fs.String(FlagType, "", "The Address Type, default 0 for ADDRESS_TYPE_32_BYTES or 1 for ADDRESS_TYPE_20_BYTES")
//	fs.String(FlagModuleName, "", "The module name to be used for address derivation, default is liquidstaking when type 0")
//
//	return fs
//}
