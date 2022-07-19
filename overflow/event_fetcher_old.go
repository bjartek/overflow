package overflow

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/onflow/flow-go-sdk"
)

// EventFetcher create an event fetcher builder.
func (o *OverflowState) EventFetcher() EventFetcherBuilder {
	return EventFetcherBuilder{
		OverflowState:         o,
		EventsAndIgnoreFields: map[string][]string{},
		EndAtCurrentHeight:    true,
		FromIndex:             -10,
		ProgressFile:          "",
		EventBatchSize:        250,
		NumberOfWorkers:       20,
	}
}

// Workers sets the number of workers.
func (e EventFetcherBuilder) Workers(workers int) EventFetcherBuilder {
	e.NumberOfWorkers = workers
	return e
}

// BatchSize sets the size of a batch
func (e EventFetcherBuilder) BatchSize(batchSize uint64) EventFetcherBuilder {
	e.EventBatchSize = batchSize
	return e
}

// Event fetches and Events and all its fields
func (e EventFetcherBuilder) Event(eventName string) EventFetcherBuilder {
	e.EventsAndIgnoreFields[eventName] = []string{}
	return e
}

//EventIgnoringFields fetch event and ignore the specified fields
func (e EventFetcherBuilder) EventIgnoringFields(eventName string, ignoreFields []string) EventFetcherBuilder {
	e.EventsAndIgnoreFields[eventName] = ignoreFields
	return e
}

//Start specify what blockHeight to fetch starting atm. This can be negative related to end/until
func (e EventFetcherBuilder) Start(blockHeight int64) EventFetcherBuilder {
	e.FromIndex = blockHeight
	return e
}

//From specify what blockHeight to fetch from. This can be negative related to end.
func (e EventFetcherBuilder) From(blockHeight int64) EventFetcherBuilder {
	e.FromIndex = blockHeight
	return e
}

//End specify what index to end at
func (e EventFetcherBuilder) End(blockHeight uint64) EventFetcherBuilder {
	e.EndIndex = blockHeight
	e.EndAtCurrentHeight = false
	return e
}

//Last fetch events from the number last blocks
func (e EventFetcherBuilder) Last(number uint64) EventFetcherBuilder {
	e.EndAtCurrentHeight = true
	e.FromIndex = -int64(number)
	return e
}

//Until specify what index to end at
func (e EventFetcherBuilder) Until(blockHeight uint64) EventFetcherBuilder {
	e.EndIndex = blockHeight
	e.EndAtCurrentHeight = false
	return e
}

//UntilCurrent Specify to fetch events until the current Block
func (e EventFetcherBuilder) UntilCurrent() EventFetcherBuilder {
	e.EndAtCurrentHeight = true
	e.EndIndex = 0
	return e
}

//TrackProgressIn Specify a file to store progress in
func (e EventFetcherBuilder) TrackProgressIn(fileName string) EventFetcherBuilder {
	e.ProgressFile = fileName
	e.EndIndex = 0
	e.FromIndex = 0
	e.EndAtCurrentHeight = true
	return e
}

//Run runs the eventfetcher returning events or an error
func (e EventFetcherBuilder) Run() ([]*FormatedEvent, error) {

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

	e.OverflowState.Logger.Info(fmt.Sprintf("Fetching events from %d to %d", fromIndex, endIndex))

	var events []string
	for key := range e.EventsAndIgnoreFields {
		events = append(events, key)
	}

	blockEvents, err := e.OverflowState.Services.Events.Get(events, uint64(fromIndex), endIndex, e.EventBatchSize, e.NumberOfWorkers)
	if err != nil {
		return nil, err
	}

	formatedEvents := FormatEvents(blockEvents, e.EventsAndIgnoreFields)

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

//PrintEvents prints th events, ignoring fields specified for the given event typeID
//Deprecated use PrintEvent on overflowResult
func PrintEvents(events []flow.Event, ignoreFields map[string][]string) {
	if len(events) > 0 {
		log.Println("EVENTS")
		log.Println("======")
	}

	for _, event := range events {
		ignoreFieldsForType := ignoreFields[event.Type]
		ev := ParseEvent(event, uint64(0), time.Now(), ignoreFieldsForType)
		prettyJSON, err := json.MarshalIndent(ev, "", "    ")

		if err != nil {
			panic(err)
		}

		log.Printf("%s\n", string(prettyJSON))
	}
	if len(events) > 0 {
		log.Println("======")
	}
}

//FormatEvents
func FormatEvents(blockEvents []flow.BlockEvents, ignoreFields map[string][]string) []*FormatedEvent {
	var events []*FormatedEvent

	for _, blockEvent := range blockEvents {
		for _, event := range blockEvent.Events {
			ev := ParseEvent(event, blockEvent.Height, blockEvent.BlockTimestamp, ignoreFields[event.Type])
			events = append(events, ev)
		}
	}
	return events
}

//ParseEvent parses a flow event into a more terse representation
func ParseEvent(event flow.Event, blockHeight uint64, time time.Time, ignoreFields []string) *FormatedEvent {

	var fieldNames []string

	for _, eventTypeFields := range event.Value.EventType.Fields {
		fieldNames = append(fieldNames, eventTypeFields.Identifier)
	}

	finalFields := map[string]interface{}{}

	for id, field := range event.Value.Fields {

		skip := false
		name := fieldNames[id]

		for _, ignoreField := range ignoreFields {
			if ignoreField == name {
				skip = true
			}
		}
		if skip {
			continue
		}
		value := CadenceValueToInterface(field)
		if value != nil {
			finalFields[name] = value
		}
	}
	return &FormatedEvent{
		Name:        event.Type,
		Fields:      finalFields,
		BlockHeight: blockHeight,
		Time:        time,
	}
}

// FormatedEvent event in a more condensed formated form
type FormatedEvent struct {
	Name        string                 `json:"name"`
	BlockHeight uint64                 `json:"blockHeight,omitempty"`
	Time        time.Time              `json:"time,omitempty"`
	Fields      map[string]interface{} `json:"fields"`
}

func (o FormatedEvent) ExistIn(events []*FormatedEvent) bool {
	//	litter.Dump(o)
	for _, ev := range events {
		//;		litter.Dump(*ev)
		result := reflect.DeepEqual(o, *ev)
		if result {
			return true
		}
	}
	return false
}

func (fe FormatedEvent) ShortName() string {
	return fe.Name[19:]
}

func NewTestEvent(name string, fields map[string]interface{}) *FormatedEvent {
	loc, _ := time.LoadLocation("UTC")
	// handle err
	time.Local = loc // -> this is setting the global timezone
	return &FormatedEvent{
		Name:        name,
		BlockHeight: 0,
		Time:        time.Unix(0, 0),
		Fields:      fields,
	}
}

//String pretty print an event as a String
func (e FormatedEvent) String() string {
	j, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(j)
}

func (e FormatedEvent) GetFieldAsUInt64(field string) uint64 {
	id := e.Fields[field]
	fieldAsString := fmt.Sprintf("%v", id)
	if fieldAsString == "" {
		panic("field is empty")
	}
	n, err := strconv.ParseUint(fieldAsString, 10, 64)
	if err != nil {
		panic(err)
	}
	return n
}
