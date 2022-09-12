package overflow

import (
	"fmt"
	"strings"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
)

// a type represneting seting an obtion in the printer builder
type OverflowPrinterOption func(*OverflowPrinterBuilder)

// a type representing the accuumlated state in the builder
//
// the default setting is to print one line for each transaction with meter and all events
type OverflowPrinterBuilder struct {

	//set to false to disable all events
	Events bool

	//filter out some events
	EventFilter OverflowEventFilter

	//0 to print no meter, 1 to print some, 2 to pritn all NB verbose
	Meter int

	//print the emulator log, NB! Verbose
	EmulatorLog bool

	//print transaction id, useful to disable in tests
	Id bool

	Arguments bool
}

// print full meter verbose mode
func WithFullMeter() OverflowPrinterOption {
	return func(opb *OverflowPrinterBuilder) {
		opb.Meter = 2
	}
}

// print meters as part of the transaction output line
func WithMeter() OverflowPrinterOption {
	return func(opb *OverflowPrinterBuilder) {
		opb.Meter = 1
	}
}

// do not print meter
func WithoutMeter(value int) OverflowPrinterOption {
	return func(opb *OverflowPrinterBuilder) {
		opb.Meter = 0
	}
}

// print the emulator log. NB! Verbose
func WithEmulatorLog() OverflowPrinterOption {
	return func(opb *OverflowPrinterBuilder) {
		opb.EmulatorLog = true
	}
}

// filter out events that are printed
func WithEventFilter(filter OverflowEventFilter) OverflowPrinterOption {
	return func(opb *OverflowPrinterBuilder) {
		opb.EventFilter = filter
	}
}

// do not print events
func WithoutEvents() OverflowPrinterOption {
	return func(opb *OverflowPrinterBuilder) {
		opb.Events = false
	}
}

func WithoutId() OverflowPrinterOption {
	return func(opb *OverflowPrinterBuilder) {
		opb.Id = false
	}
}

func WithArguments() OverflowPrinterOption {
	return func(opb *OverflowPrinterBuilder) {
		opb.Arguments = true
	}
}

// print out an result
func (o OverflowResult) Print(opbs ...OverflowPrinterOption) OverflowResult {

	printOpts := &OverflowPrinterBuilder{
		Events:      true,
		EventFilter: OverflowEventFilter{},
		Meter:       1,
		EmulatorLog: false,
		Id:          true,
		Arguments:   false,
	}

	for _, opb := range opbs {
		opb(printOpts)
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

	if len(o.Fee) != 0 {
		messages = append(messages, fmt.Sprintf("fee:%.8f gas:%d", o.Fee["amount"], o.FeeGas))
	} else {
		if o.ComputationUsed != 0 {
			messages = append(messages, fmt.Sprintf("gas:%d", o.ComputationUsed))
		}
	}

	if printOpts.Id {
		messages = append(messages, fmt.Sprintf("id:%s", o.Id.String()))
	}

	fmt.Printf("%v %s\n", emoji.OkHand, strings.Join(messages, " "))

	if printOpts.Arguments {
		o.PrintArguments(nil)
	}

	if printOpts.Events {
		events := o.Events
		if len(printOpts.EventFilter) != 0 {
			events = events.FilterEvents(printOpts.EventFilter)
		}
		if len(events) != 0 {
			events.Print(nil)
		}
	}

	if printOpts.EmulatorLog && len(o.RawLog) > 0 {
		fmt.Println("=== LOG ===")
		for _, msg := range o.RawLog {
			fmt.Println(msg.Msg)
		}
	}
	/*
		//TODO: print how a meter is computed
			if printOpts.Meter == 1 && o.Meter != nil {
				messages = append(messages, fmt.Sprintf("loops:%d", o.Meter.Loops()))
				messages = append(messages, fmt.Sprintf("statements:%d", o.Meter.Statements()))
				messages = append(messages, fmt.Sprintf("invocations:%d", o.Meter.FunctionInvocations()))
			}
	*/

	if printOpts.Meter != 0 && o.Meter != nil {
		if printOpts.Meter == 2 {
			fmt.Println("=== METER ===")
			fmt.Printf("LedgerInteractionUsed: %d\n", o.Meter.LedgerInteractionUsed)
			if o.Meter.MemoryUsed != 0 {
				fmt.Printf("Memory: %d\n", o.Meter.MemoryUsed)
				memories := strings.ReplaceAll(strings.Trim(fmt.Sprintf("%+v", o.Meter.MemoryIntensities), "map[]"), " ", "\n  ")

				fmt.Println("Memory Intensities")
				fmt.Printf(" %s\n", memories)
			}
			fmt.Printf("Computation: %d\n", o.Meter.ComputationUsed)
			intensities := strings.ReplaceAll(strings.Trim(fmt.Sprintf("%+v", o.Meter.ComputationIntensities), "map[]"), " ", "\n  ")

			fmt.Println("Computation Intensities:")
			fmt.Printf(" %s\n", intensities)
		}
	}
	return o
}
