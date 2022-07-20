package overflow

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

//A function to customize the transaction builder
type EventFetcherOption func(*EventFetcherBuilder)

// EventFetcherBuilder builder to hold info about eventhook context.
type EventFetcherBuilder struct {
	OverflowState         *OverflowState
	EventsAndIgnoreFields OverflowEventFilter
	FromIndex             int64
	EndAtCurrentHeight    bool
	EndIndex              uint64
	ProgressFile          string
	NumberOfWorkers       int
	EventBatchSize        uint64
}

type EventFetcherResult struct {
	Err    error
	Events *OverflowEvents
}

func (o *OverflowState) buildEventInteraction(opts ...EventFetcherOption) *EventFetcherBuilder {
	e := &EventFetcherBuilder{
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

// EventFetcher create an event fetcher builder.
func (o *OverflowState) FetchEvents(opts ...EventFetcherOption) ([]OverflowPastEvent, error) {

	e := o.buildEventInteraction(opts...)
	//if we have a progress file read the value from it and set it as oldHeight
	if e.ProgressFile != "" {

		present, err := exists(e.ProgressFile)
		if err != nil {
			return nil, err
		}

		if !present {
			err := writeProgressToFile(e.ProgressFile, 0)
			if err != nil {
				return nil, fmt.Errorf("could not create initial progress file %v", err)
			}

			e.FromIndex = 0
		} else {
			oldHeight, err := readProgressFromFile(e.ProgressFile)
			if err != nil {
				return nil, fmt.Errorf("could not parse progress file as block height %v", err)
			}
			e.FromIndex = oldHeight
		}
	}

	endIndex := e.EndIndex
	if e.EndAtCurrentHeight {
		blockHeight, err := e.OverflowState.Services.Blocks.GetLatestBlockHeight()
		if err != nil {
			return nil, err
		}
		endIndex = blockHeight
	}

	fromIndex := e.FromIndex
	//if we have a negative fromIndex is is relative to endIndex
	if e.FromIndex <= 0 {
		fromIndex = int64(endIndex) + e.FromIndex
	}

	if fromIndex < 0 {
		return nil, fmt.Errorf("FromIndex is negative")
	}

	var events []string
	for key := range e.EventsAndIgnoreFields {
		events = append(events, key)
	}

	blockEvents, err := e.OverflowState.Services.Events.Get(events, uint64(fromIndex), endIndex, e.EventBatchSize, e.NumberOfWorkers)
	if err != nil {
		return nil, err
	}

	formatedEvents := []OverflowPastEvent{}
	for _, blockEvent := range blockEvents {
		events, _ := ParseEvents(blockEvent.Events)
		for name, eventList := range events {
			for _, instance := range eventList {
				formatedEvents = append(formatedEvents, OverflowPastEvent{
					Name:        name,
					Time:        blockEvent.BlockTimestamp,
					BlockHeight: blockEvent.Height,
					Fields:      instance,
				})
			}
		}
	}
	if e.ProgressFile != "" {
		err := writeProgressToFile(e.ProgressFile, endIndex+1)
		if err != nil {
			return nil, fmt.Errorf("could not write progress to file %v", err)
		}
	}
	sort.Slice(formatedEvents, func(i, j int) bool {
		return formatedEvents[i].BlockHeight < formatedEvents[j].BlockHeight
	})

	return formatedEvents, nil

}

func WithWorkers(workers int) EventFetcherOption {
	return func(e *EventFetcherBuilder) {
		e.NumberOfWorkers = workers
	}
}
func WithBatchSize(size uint64) EventFetcherOption {
	return func(e *EventFetcherBuilder) {
		e.EventBatchSize = size
	}
}
func WithEvent(eventName string) EventFetcherOption {
	return func(e *EventFetcherBuilder) {
		e.EventsAndIgnoreFields[eventName] = []string{}
	}
}
func WithEventIgnoringField(eventName string, ignoreFields []string) EventFetcherOption {
	return func(e *EventFetcherBuilder) {
		e.EventsAndIgnoreFields[eventName] = ignoreFields
	}
}

func WithStartHeight(blockHeight int64) EventFetcherOption {
	return func(e *EventFetcherBuilder) {
		e.FromIndex = blockHeight
	}
}

func WithFromIndex(blockHeight int64) EventFetcherOption {
	return func(e *EventFetcherBuilder) {
		e.FromIndex = blockHeight
	}
}
func WithEndIndex(blockHeight uint64) EventFetcherOption {
	return func(e *EventFetcherBuilder) {
		e.EndIndex = blockHeight
		e.EndAtCurrentHeight = false
	}
}
func WithLastBlocks(number uint64) EventFetcherOption {
	return func(e *EventFetcherBuilder) {
		e.EndAtCurrentHeight = true
		e.FromIndex = -int64(number)
	}
}
func UntilBlock(blockHeight uint64) EventFetcherOption {
	return func(e *EventFetcherBuilder) {
		e.EndIndex = blockHeight
		e.EndAtCurrentHeight = false
	}
}
func UntilCurrentBlock() EventFetcherOption {
	return func(e *EventFetcherBuilder) {
		e.EndAtCurrentHeight = true
		e.EndIndex = 0
	}
}
func TrackProgressIn(fileName string) EventFetcherOption {
	return func(e *EventFetcherBuilder) {
		e.ProgressFile = fileName
		e.EndIndex = 0
		e.FromIndex = 0
		e.EndAtCurrentHeight = true
	}
}

type OverflowPastEvent struct {
	Name        string        `json:"name"`
	BlockHeight uint64        `json:"blockHeight,omitempty"`
	Time        time.Time     `json:"time,omitempty"`
	Fields      OverflowEvent `json:"fields"`
}

func NewTestPastEvent(name string, fields map[string]interface{}) *OverflowPastEvent {
	loc, _ := time.LoadLocation("UTC")
	// handle err
	time.Local = loc // -> this is setting the global timezone
	return &OverflowPastEvent{
		Name:        name,
		BlockHeight: 0,
		Time:        time.Unix(0, 0),
		Fields:      fields,
	}
}

//String pretty print an event as a String
func (e OverflowPastEvent) String() string {
	j, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(j)
}

func (e OverflowPastEvent) GetFieldAsUInt64(field string) uint64 {
	return e.Fields[field].(uint64)
}
