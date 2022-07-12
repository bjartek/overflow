package overflow

import (
	"fmt"
	"strings"

	"github.com/enescakir/emoji"
)

type PrinterOption func(*PrintOptions)
type PrintOptions struct {
	Events      bool
	EventFilter OverflowEventFilter
	Meter       bool
	EmulatorLog bool
}

func WithMeter() func(opt *PrintOptions) {
	return func(opt *PrintOptions) {
		opt.Meter = true
	}
}

func WithEmulatorLog() func(opt *PrintOptions) {
	return func(opt *PrintOptions) {
		opt.EmulatorLog = true
	}
}

func WithEventFilter(filter OverflowEventFilter) func(opt *PrintOptions) {
	return func(opt *PrintOptions) {
		opt.EventFilter = filter
	}
}

func WithoutEvents() func(opt *PrintOptions) {
	return func(opt *PrintOptions) {
		opt.Events = false
	}
}

func (o OverflowResult) Print(opts ...PrinterOption) OverflowResult {

	printOpts := &PrintOptions{
		Events:      true,
		EventFilter: OverflowEventFilter{},
		Meter:       false,
		EmulatorLog: false,
	}

	for _, opt := range opts {
		opt(printOpts)
	}

	if o.Err != nil {
		panic(o.Err)
		o.Logger.Error(fmt.Sprintf("%v Error executing transaction: %s ", emoji.PileOfPoo, o.Name))
	}

	messages := []string{}

	if o.ComputationUsed != 0 {
		messages = append(messages, fmt.Sprintf("%d%v", o.ComputationUsed, emoji.HighVoltage))
	}
	nameMessage := fmt.Sprintf("Tx %s", o.Name)
	if o.Name == "inline" {
		nameMessage = "Inline TX"
	}
	messages = append(messages, nameMessage)

	if len(o.Fee) != 0 {
		messages = append(messages, fmt.Sprintf("%v:%f (%f/%f)", emoji.MoneyBag, o.Fee["amount"].(float64), o.Fee["inclusionEffort"].(float64), o.Fee["exclusionEffort"].(float64)))
	}
	messages = append(messages, fmt.Sprintf("id:%s", o.Id.String()))

	o.Logger.Info(fmt.Sprintf("%v %s", emoji.OkHand, strings.Join(messages, " ")))

	if printOpts.Events {
		events := o.Events
		if len(printOpts.EventFilter) != 0 {
			events = events.FilterEvents(printOpts.EventFilter)
		}
		if len(events) != 0 {
			o.Logger.Info("=== Events ===")
			for name, eventList := range events {
				for _, event := range eventList {
					o.Logger.Info(name)
					length := 0
					for key, _ := range event {
						keyLength := len(key)
						if keyLength > length {
							length = keyLength
						}
					}

					format := fmt.Sprintf("%%%ds:%%v", length+2)
					for key, value := range event {
						o.Logger.Info(fmt.Sprintf(format, key, value))
					}
				}
			}
		}
	}

	if printOpts.EmulatorLog && len(o.RawLog) > 0 {
		o.Logger.Info("=== LOG ===")
		for _, msg := range o.RawLog {
			o.Logger.Info(msg.Msg)
		}
	}

	if printOpts.Meter && o.Meter != nil {
		fmt.Println("=== METER ===")
		fmt.Println("LedgerInteractionUsed:", o.Meter.LedgerInteractionUsed)
		if o.Meter.MemoryUsed != 0 {
			fmt.Println("Memory:", o.Meter.MemoryUsed)
			memories := strings.ReplaceAll(strings.Trim(fmt.Sprintf("%+v", o.Meter.MemoryIntensities), "map[]"), " ", "\n  ")

			fmt.Println("Memory Intensities")
			fmt.Println(" ", memories)
		}
		fmt.Println("Computation:", o.Meter.ComputationUsed)
		intensities := strings.ReplaceAll(strings.Trim(fmt.Sprintf("%+v", o.Meter.ComputationIntensities), "map[]"), " ", "\n  ")

		fmt.Println("Computation Intensities:")
		fmt.Println(" ", intensities)
	}
	return o
}
