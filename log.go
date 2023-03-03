package overflow

// OverflowEmulatorLogMessage a log message from the logrus implementation used in the flow emulator
type OverflowEmulatorLogMessage struct {
	ComputationUsed int    `json:"computationUsed"`
	Level           string `json:"level"`
	Msg             string `json:"message"`
	TxID            string `json:"txID"`
}

//Log  {"level":"info","txID":"7f0f09b89cb64a79a6751f9b9875e0b3738996b4e7c07b0fe27f3d263a492dfb","computationUsed":27,"message":"⭐  Transaction executed"}
