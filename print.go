package overflow

import (
	"fmt"
	"strings"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
)

// a type represneting seting an option in the printer builder
type PrinterOption func(*PrinterBuilder)

// a type representing the accuumlated state in the builder
//
// the default setting is to print one line for each transaction with meter and all events
type PrinterBuilder struct {
	Events      bool
	EventFilter OverflowEventFilter
	Meter       int
	EmulatorLog bool
}

// print full meter verbose mode
func WithFullMeter() PrinterOption {
	return func(opt *PrinterBuilder) {
		opt.Meter = 2
	}
}

// print meters as part of the transaction output line
func WithMeter() PrinterOption {
	return func(opt *PrinterBuilder) {
		opt.Meter = 1
	}
}

// do not print meter
func WithoutMeter(value int) PrinterOption {
	return func(opt *PrinterBuilder) {
		opt.Meter = 0
	}
}

// print the emulator log. NB! Verbose
func WithEmulatorLog() PrinterOption {
	return func(opt *PrinterBuilder) {
		opt.EmulatorLog = true
	}
}

// filter out events that are printed
func WithEventFilter(filter OverflowEventFilter) PrinterOption {
	return func(opt *PrinterBuilder) {
		opt.EventFilter = filter
	}
}

// do not print events
func WithoutEvents() PrinterOption {
	return func(opt *PrinterBuilder) {
		opt.Events = false
	}
}

// print out an result
func (o OverflowResult) Print(opts ...PrinterOption) OverflowResult {

	printOpts := &PrinterBuilder{
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

	nameMessage := fmt.Sprintf("Tx:%s", o.Name)
	if o.Name == "inline" {
		nameMessage = "Inline TX"
	}
	messages = append(messages, nameMessage)

	/*
		if printOpts.Meter == 1 && o.Meter != nil {
			messages = append(messages, fmt.Sprintf("loops:%d", o.Meter.Loops()))
			messages = append(messages, fmt.Sprintf("statements:%d", o.Meter.Statements()))
			messages = append(messages, fmt.Sprintf("invocations:%d", o.Meter.FunctionInvocations()))
		}
	*/

	if len(o.Fee) != 0 {
		messages = append(messages, fmt.Sprintf("fee:%.8f gas:%d", o.Fee["amount"], o.FeeGas))
	} else {
		if o.ComputationUsed != 0 {
			messages = append(messages, fmt.Sprintf("gas:%d", o.ComputationUsed))
		}
	}

	messages = append(messages, fmt.Sprintf("id:%s", o.Id.String()))

	fmt.Println()
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

					format := fmt.Sprintf("%%%ds -> %%v\n", length+2)
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
		if printOpts.Meter == 2 {
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
