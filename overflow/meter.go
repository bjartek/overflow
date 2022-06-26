package overflow

import "github.com/onflow/cadence/runtime/common"

type Meter struct {
	LedgerInteractionUsed  int                           `json:"ledgerInteractionUsed"`
	ComputationUsed        int                           `json:"computationUsed"`
	MemoryUsed             int                           `json:"memoryUsed"`
	ComputationIntensities MeteredComputationIntensities `json:"computationIntensities"`
	MemoryIntensities      MeteredMemoryIntensities      `json:"memoryIntensities"`
}

func (m Meter) FunctionInvocations() int {
	return int(m.ComputationIntensities[common.ComputationKindFunctionInvocation])
}

func (m Meter) Loops() int {
	return int(m.ComputationIntensities[common.ComputationKindLoop])
}

func (m Meter) Statements() int {
	return int(m.ComputationIntensities[common.ComputationKindStatement])
}

type MeteredComputationIntensities map[common.ComputationKind]uint
type MeteredMemoryIntensities map[common.MemoryKind]uint
