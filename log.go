package overflow

import "time"

//OverflowEmulatorLogMessage a log message from the logrus implementation used in the flow emulator
type OverflowEmulatorLogMessage struct {
	ComputationUsed int       `json:"computationUsed"`
	Level           string    `json:"level"`
	Msg             string    `json:"msg"`
	Time            time.Time `json:"time"`
	TxID            string    `json:"txID"`
}
