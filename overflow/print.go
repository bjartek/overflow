package overflow

import (
	"fmt"
	"strings"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
)

type PrinterOption func(*PrintOptions)

type PrintOptions struct {
	Events      bool
	EventFilter OverflowEventFilter
	Meter       int
	EmulatorLog bool
}

func WithFullMeter() func(opt *PrintOptions) {
	return func(opt *PrintOptions) {
		opt.Meter = 2
	}
}

func WithMeter(value int) func(opt *PrintOptions) {
	return func(opt *PrintOptions) {
		opt.Meter = 1
	}
}

func WithoutMeter(value int) func(opt *PrintOptions) {
	return func(opt *PrintOptions) {
		opt.Meter = 0
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
		Meter:       1,
		EmulatorLog: false,
	}

	for _, opt := range opts {
		opt(printOpts)
	}

	if o.Err != nil {
		color.Red("%v Error executing transaction: %s error:%v", emoji.PileOfPoo, o.Name, o.Err)
		return o //is it best to return here or not?
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

	fmt.Printf("%v %s\n", emoji.OkHand, strings.Join(messages, " "))

	if printOpts.Events {
		events := o.Events
		if len(printOpts.EventFilter) != 0 {
			events = events.FilterEvents(printOpts.EventFilter)
		}
		if len(events) != 0 {
			fmt.Println("=== Events ===")
			for name, eventList := range events {
				for _, event := range eventList {
					fmt.Println(name)
					length := 0
					for key, _ := range event {
						keyLength := len(key)
						if keyLength > length {
							length = keyLength
						}
					}

					format := fmt.Sprintf("%%%ds:%%v\n", length+2)
					for key, value := range event {
						fmt.Printf(format, key, value)
					}
				}
			}
		}
	}

	if printOpts.EmulatorLog && len(o.RawLog) > 0 {
		fmt.Println("=== LOG ===")
		for _, msg := range o.RawLog {
			fmt.Println(msg.Msg)
		}
	}

	if printOpts.Meter != 0 && o.Meter != nil {
		if printOpts.Meter == 1 {
			fmt.Println("=== METER ===")
			fmt.Println(fmt.Sprintf("Computation: %d", o.Meter.ComputationUsed))
			fmt.Println(fmt.Sprintf("      loops: %d", o.Meter.Loops()))
			fmt.Println(fmt.Sprintf(" statements: %d", o.Meter.Statements()))
			fmt.Println(fmt.Sprintf("invocations: %d", o.Meter.FunctionInvocations()))
		} else {
			fmt.Println("=== METER ===")
			fmt.Println(fmt.Sprintf("LedgerInteractionUsed: %d", o.Meter.LedgerInteractionUsed))
			if o.Meter.MemoryUsed != 0 {
				fmt.Println(fmt.Sprintf("Memory: %d", o.Meter.MemoryUsed))
				memories := strings.ReplaceAll(strings.Trim(fmt.Sprintf("%+v", o.Meter.MemoryIntensities), "map[]"), " ", "\n  ")

				fmt.Println("Memory Intensities")
				fmt.Println(fmt.Sprintf(" %s", memories))
			}
			fmt.Println(fmt.Sprintf("Computation: %d", o.Meter.ComputationUsed))
			intensities := strings.ReplaceAll(strings.Trim(fmt.Sprintf("%+v", o.Meter.ComputationIntensities), "map[]"), " ", "\n  ")

			fmt.Println("Computation Intensities:")
			fmt.Println(fmt.Sprintf(" %s", intensities))
		}
	}
	return o
}
