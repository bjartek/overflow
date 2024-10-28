package overflow

import "github.com/onflow/cadence/common"

// a type representing a meter that contains information about the inner workings of an interaction, only available on local emulator
type OverflowMeter struct {
	ComputationIntensities OverflowMeteredComputationIntensities `json:"computationIntensities"`
	MemoryIntensities      OverflowMeteredMemoryIntensities      `json:"memoryIntensities"`
	LedgerInteractionUsed  int                                   `json:"ledgerInteractionUsed"`
	ComputationUsed        int                                   `json:"computationUsed"`
	MemoryUsed             int                                   `json:"memoryUsed"`
}

// get the number of functions invocations
func (m OverflowMeter) FunctionInvocations() int {
	return int(m.ComputationIntensities[common.ComputationKindFunctionInvocation])
}

// get the number of loops
func (m OverflowMeter) Loops() int {
	return int(m.ComputationIntensities[common.ComputationKindLoop])
}

// get the number of statements
func (m OverflowMeter) Statements() int {
	return int(m.ComputationIntensities[common.ComputationKindStatement])
}

// type collecting computatationIntensities
type OverflowMeteredComputationIntensities map[common.ComputationKind]uint

// type collecting memoryIntensities
type OverflowMeteredMemoryIntensities map[common.MemoryKind]uint
