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
	for _, poolRecord := range genState.PoolRecords {
		if err := poolRecord.Validate(); err != nil {
			return err
		}
	}
	for _, position := range genState.Positions {
		if err := position.Validate(); err != nil {
			return err
		}
	}
	for _, tickInfoRecord := range genState.TickInfoRecords {
		if err := tickInfoRecord.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (record PoolRecord) Validate() error {
	if err := record.Pool.Validate(); err != nil {
		return err
	}
	if err := record.State.Validate(); err != nil {
		return err
	}
	return nil
}

func (record TickInfoRecord) Validate() error {
	if err := record.TickInfo.Validate(); err != nil {
		return err
	}
	return nil
}
