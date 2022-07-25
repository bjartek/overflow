package overflow

import "github.com/onflow/cadence/runtime/common"

// a type representing a meter that contains information about the inner workings of an interaction, only available on local emulator
type OverflowMeter struct {
	LedgerInteractionUsed  int                           `json:"ledgerInteractionUsed"`
	ComputationUsed        int                           `json:"computationUsed"`
	MemoryUsed             int                           `json:"memoryUsed"`
	ComputationIntensities MeteredComputationIntensities `json:"computationIntensities"`
	MemoryIntensities      MeteredMemoryIntensities      `json:"memoryIntensities"`
}

//get the number of functions invocations
func (m OverflowMeter) FunctionInvocations() int {
	return int(m.ComputationIntensities[common.ComputationKindFunctionInvocation])
}

// get the number of loops
func (m OverflowMeter) Loops() int {
	return int(m.ComputationIntensities[common.ComputationKindLoop])
}

//get the number of statements
func (m OverflowMeter) Statements() int {
	return int(m.ComputationIntensities[common.ComputationKindStatement])
}

// type collecting computatationIntensities
type MeteredComputationIntensities map[common.ComputationKind]uint

// type collecting memoryIntensities
type MeteredMemoryIntensities map[common.MemoryKind]uint
