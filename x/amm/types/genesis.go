package types

func NewGenesisState(
	params Params, lastPoolId, lastPositionId uint64,
	poolRecords []PoolRecord, positions []Position, tickInfoRecords []TickInfoRecord) *GenesisState {
	return &GenesisState{
		Params:          params,
		LastPoolId:      lastPoolId,
		LastPositionId:  lastPositionId,
		PoolRecords:     poolRecords,
		Positions:       positions,
		TickInfoRecords: tickInfoRecords,
	}
}

// DefaultGenesis returns the default genesis state for the module.
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), 0, 0, nil, nil, nil)
}

func (genState GenesisState) Validate() error {
	if err := genState.Params.Validate(); err != nil {
		return err
	}
	return nil
}
