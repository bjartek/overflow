package overflow

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Event fetching
//
// A function to customize the transaction builder
type OverflowEventFetcherOption func(*OverflowEventFetcherBuilder)

// OverflowEventFetcherBuilder builder to hold info about eventhook context.
type OverflowEventFetcherBuilder struct {
	OverflowState         *OverflowState
	EventsAndIgnoreFields OverflowEventFilter
	FromIndex             int64
	EndAtCurrentHeight    bool
	EndIndex              uint64
	ProgressFile          string
	NumberOfWorkers       int
	EventBatchSize        uint64
}

// Build an event fetcher builder from the sent in options
func (o *OverflowState) buildEventInteraction(opts ...OverflowEventFetcherOption) *OverflowEventFetcherBuilder {
	e := &OverflowEventFetcherBuilder{
		OverflowState:         o,
		EventsAndIgnoreFields: OverflowEventFilter{},
		EndAtCurrentHeight:    true,
		FromIndex:             -10,
		ProgressFile:          "",
		EventBatchSize:        250,
		NumberOfWorkers:       20,
	}

	for _, opt := range opts {
		opt(e)
	}
	return e
}

type EventFetcherResult struct {
	Events []OverflowPastEvent
	Error  error
	State  *OverflowEventFetcherBuilder
	From   int64
	To     uint64
}

func (efr EventFetcherResult) String() string {
	events := []string{}
	for event := range efr.State.EventsAndIgnoreFields {
		events = append(events, event)
	}
	eventString := strings.Join(events, ",")
	return fmt.Sprintf("Fetched number=%d of events within from=%d block to=%d for events=%s\n", len(efr.Events), efr.From, efr.To, eventString)
}

// FetchEvents using the given options
func (o *OverflowState) FetchEventsWithResult(opts ...OverflowEventFetcherOption) EventFetcherResult {

	e := o.buildEventInteraction(opts...)

	res := EventFetcherResult{State: e}
	//if we have a progress file read the value from it and set it as oldHeight
	if e.ProgressFile != "" {

		present, err := exists(e.ProgressFile)
		if err != nil {
			res.Error = err
			return res
		}

		if !present {
			err := writeProgressToFile(e.ProgressFile, 0)
			if err != nil {
				res.Error = fmt.Errorf("could not create initial progress file %v", err)
				return res
			}

			e.FromIndex = 0
		} else {
			oldHeight, err := readProgressFromFile(e.ProgressFile)
			if err != nil {
				res.Error = fmt.Errorf("could not parse progress file as block height %v", err)
				return res
			}
			e.FromIndex = oldHeight
		}
	}

	endIndex := e.EndIndex
	if e.EndAtCurrentHeight {
		blockHeight, err := e.OverflowState.Services.Blocks.GetLatestBlockHeight()
		if err != nil {
			res.Error = err
			return res
		}
		endIndex = blockHeight
	}

	fromIndex := e.FromIndex
	//if we have a negative fromIndex is is relative to endIndex
	if e.FromIndex <= 0 {
		fromIndex = int64(endIndex) + e.FromIndex
	}

	if fromIndex < 0 {
		res.Error = fmt.Errorf("FromIndex is negative")
		return res
	}

	var events []string
	for key := range e.EventsAndIgnoreFields {
		events = append(events, key)
	}

	if uint64(fromIndex) > endIndex {
		return res
	}
	blockEvents, err := e.OverflowState.Services.Events.Get(events, uint64(fromIndex), endIndex, e.EventBatchSize, e.NumberOfWorkers)
	if err != nil {
		res.Error = err
		return res
	}

	formatedEvents := []OverflowPastEvent{}
	for _, blockEvent := range blockEvents {
		events, _ := parseEvents(blockEvent.Events)
		for name, eventList := range events {
			for _, instance := range eventList {
				formatedEvents = append(formatedEvents, OverflowPastEvent{
					Name:        name,
					Time:        blockEvent.BlockTimestamp,
					BlockHeight: blockEvent.Height,
					Event:       instance,
				})
			}
		}
	}
	if e.ProgressFile != "" {
		err := writeProgressToFile(e.ProgressFile, endIndex+1)
		if err != nil {
			res.Error = fmt.Errorf("could not write progress to file %v", err)
			return res
		}
	}
	sort.Slice(formatedEvents, func(i, j int) bool {
		return formatedEvents[i].BlockHeight < formatedEvents[j].BlockHeight
	})

	res.Events = formatedEvents
	res.From = fromIndex
	res.To = endIndex
	return res

}

// FetchEvents using the given options
func (o *OverflowState) FetchEvents(opts ...OverflowEventFetcherOption) ([]OverflowPastEvent, error) {
	res := o.FetchEventsWithResult(opts...)
	return res.Events, res.Error
}

// Set the Workers size for FetchEvents
func WithWorkers(workers int) OverflowEventFetcherOption {
	return func(e *OverflowEventFetcherBuilder) {
		e.NumberOfWorkers = workers
	}
}

// Set the batch sice for FetchEvents
func WithBatchSize(size uint64) OverflowEventFetcherOption {
	return func(e *OverflowEventFetcherBuilder) {
		e.EventBatchSize = size
	}
}

// set that we want to fetch an event and all its fields
func WithEvent(eventName string) OverflowEventFetcherOption {
	return func(e *OverflowEventFetcherBuilder) {
		e.EventsAndIgnoreFields[eventName] = []string{}
	}
}

// set that we want the following events and ignoring the fields mentioned
func WithEventIgnoringField(eventName string, ignoreFields []string) OverflowEventFetcherOption {
	return func(e *OverflowEventFetcherBuilder) {
		e.EventsAndIgnoreFields[eventName] = ignoreFields
	}
}

// set the start height to use
func WithStartHeight(blockHeight int64) OverflowEventFetcherOption {
	return func(e *OverflowEventFetcherBuilder) {
		e.FromIndex = blockHeight
	}
}

// set the from index to use alias to WithStartHeight
func WithFromIndex(blockHeight int64) OverflowEventFetcherOption {
	return func(e *OverflowEventFetcherBuilder) {
		e.FromIndex = blockHeight
	}
}

// set the end index to use
func WithEndIndex(blockHeight uint64) OverflowEventFetcherOption {
	return func(e *OverflowEventFetcherBuilder) {
		e.EndIndex = blockHeight
		e.EndAtCurrentHeight = false
	}
}

// set the relative list of blocks to fetch events from
func WithLastBlocks(number uint64) OverflowEventFetcherOption {
	return func(e *OverflowEventFetcherBuilder) {
		e.EndAtCurrentHeight = true
		e.FromIndex = -int64(number)
	}
}

// fetch events until theg given height alias to WithEndHeight
func WithUntilBlock(blockHeight uint64) OverflowEventFetcherOption {
	return func(e *OverflowEventFetcherBuilder) {
		e.EndIndex = blockHeight
		e.EndAtCurrentHeight = false
	}
}

// set the end index to the current height
func WithUntilCurrentBlock() OverflowEventFetcherOption {
	return func(e *OverflowEventFetcherBuilder) {
		e.EndAtCurrentHeight = true
		e.EndIndex = 0
	}
}

// track what block we have read since last run in a file
func WithTrackProgressIn(fileName string) OverflowEventFetcherOption {
	return func(e *OverflowEventFetcherBuilder) {
		e.ProgressFile = fileName
		e.EndIndex = 0
		e.FromIndex = 0
		e.EndAtCurrentHeight = true
	}
}

// a type to represent an event that we get from FetchEvents
type OverflowPastEvent struct {
	Name        string        `json:"name"`
	BlockHeight uint64        `json:"blockHeight,omitempty"`
	Time        time.Time     `json:"time,omitempty"`
	Event       OverflowEvent `json:"event"`
}

// String pretty print an event as a String
func (e OverflowPastEvent) String() string {
	j, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(j)
}

// get the given field as an uint64
func (e OverflowPastEvent) GetFieldAsUInt64(field string) uint64 {
	return e.Event.Fields[field].(uint64)
}
