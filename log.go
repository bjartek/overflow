package overflow

import (
	"fmt"
)

// OverflowEmulatorLogMessage a log message from the logrus implementation used in the flow emulator
type OverflowEmulatorLogMessage struct {
	Fields          map[string]interface{}
	Level           string
	Msg             string
	ComputationUsed int
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
