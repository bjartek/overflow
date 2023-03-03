package overflow

import (
	"fmt"
)

// OverflowEmulatorLogMessage a log message from the logrus implementation used in the flow emulator
type OverflowEmulatorLogMessage struct {
	ComputationUsed int
	Level           string
	Msg             string
	Fields          map[string]interface{}
}

func (me OverflowEmulatorLogMessage) String() string {

	fields := ""
	if len(me.Fields) > 0 {
		for key, value := range me.Fields {
			fields = fmt.Sprintf("%s %s=%v", fields, key, value)
		}
	}

	return fmt.Sprintf("%s - %s%s", me.Level, me.Msg, fields)
}

//Log  {"level":"info","txID":"7f0f09b89cb64a79a6751f9b9875e0b3738996b4e7c07b0fe27f3d263a492dfb","computationUsed":27,"message":"⭐  Transaction executed"}
