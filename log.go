package overflow

import "time"

//LogrusMessage a log message from the logrus implementation used in the flow emulator
type LogrusMessage struct {
	ComputationUsed int       `json:"computationUsed"`
	Level           string    `json:"level"`
	Msg             string    `json:"msg"`
	Time            time.Time `json:"time"`
	TxID            string    `json:"txID"`
}
