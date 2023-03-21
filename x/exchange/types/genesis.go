package types

func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}

// DefaultGenesis returns the default genesis state for the module.
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams())
}

func (genState GenesisState) Validate() error {
	if err := genState.Params.Validate(); err != nil {
		return err
	}
	return nil
}
