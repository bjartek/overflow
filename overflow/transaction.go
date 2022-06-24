package overflow

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/enescakir/emoji"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-go-sdk"
)

func (f *Overflow) SimpleTxArgs(filename string, signer string, args *FlowArgumentsBuilder) {
	f.TransactionFromFile(filename).SignProposeAndPayAs(signer).Args(args).RunPrintEventsFull()
}

// TransactionFromFile will start a flow transaction builder
func (f *Overflow) TransactionFromFile(filename string) FlowInteractionBuilder {
	return FlowInteractionBuilder{
		Overflow:       f,
		FileName:       filename,
		MainSigner:     nil,
		Arguments:      []cadence.Value{},
		PayloadSigners: []*flowkit.Account{},
		GasLimit:       uint64(f.Gas),
		BasePath:       fmt.Sprintf("%s/transactions", f.BasePath),
	}
}

// Transaction will start a flow transaction builder using the inline transaction
func (f *Overflow) Transaction(content string) FlowInteractionBuilder {
	return FlowInteractionBuilder{
		Overflow:       f,
		FileName:       "inline",
		Content:        content,
		MainSigner:     nil,
		Arguments:      []cadence.Value{},
		PayloadSigners: []*flowkit.Account{},
		GasLimit:       uint64(f.Gas),
		BasePath:       fmt.Sprintf("%s/transactions", f.BasePath),
	}
}

func (t FlowInteractionBuilder) NamedArguments(args map[string]string) FlowInteractionBuilder {

	codeFileName := fmt.Sprintf("%s/%s.cdc", t.BasePath, t.FileName)
	code, err := t.getContractCode(codeFileName)
	if err != nil {
		fmt.Println(err.Error())
		t.Error = err
	}
	parseArgs, err := t.Overflow.ParseArgumentsWithoutType(t.FileName, code, args)
	if err != nil {
		t.Error = err
	}
	t.Arguments = parseArgs
	return t
}

// Specify arguments to send to transaction using a raw list of values
func (t FlowInteractionBuilder) ArgsV(args []cadence.Value) FlowInteractionBuilder {
	t.Arguments = args
	return t
}

// Specify arguments to send to transaction using a builder you send in
func (t FlowInteractionBuilder) Args(args *FlowArgumentsBuilder) FlowInteractionBuilder {
	t.Arguments = args.Build()
	return t
}

// Specify arguments to send to transaction using a function that takes a builder where you call the builder
func (t FlowInteractionBuilder) ArgsFn(fn func(*FlowArgumentsBuilder)) FlowInteractionBuilder {
	args := t.Overflow.Arguments()
	fn(args)
	t.Arguments = args.Build()
	return t
}

func (t FlowInteractionBuilder) TransactionPath(path string) FlowInteractionBuilder {
	t.BasePath = path
	return t
}

// Gas sets the gas limit for this transaction
func (t FlowInteractionBuilder) Gas(limit uint64) FlowInteractionBuilder {
	t.GasLimit = limit
	return t
}

// SignProposeAndPayAs set the payer, proposer and envelope signer
func (t FlowInteractionBuilder) SignProposeAndPayAs(signer string) FlowInteractionBuilder {
	account, err := t.Overflow.AccountE(signer)
	if err != nil {
		t.Error = err
		return t
	}
	t.Proposer = account
	t.MainSigner = account
	return t
}

// SignProposeAndPayAsService set the payer, proposer and envelope signer
func (t FlowInteractionBuilder) SignProposeAndPayAsService() FlowInteractionBuilder {
	key := t.Overflow.ServiceAccountName()
	//swallow error as you cannot start a overflow without a valid sa
	account, _ := t.Overflow.State.Accounts().ByName(key)
	t.MainSigner = account
	t.Proposer = account
	return t
}

// PayloadSigner set a signer for the payload
func (t FlowInteractionBuilder) PayloadSigner(value string) FlowInteractionBuilder {
	signer := t.Overflow.Account(value)
	t.PayloadSigners = append(t.PayloadSigners, signer)
	return t
}

// RunPrintEventsFull will run a transaction and print all events
func (t FlowInteractionBuilder) RunPrintEventsFull() {
	PrintEvents(t.Run(), map[string][]string{})
}

// RunPrintEvents will run a transaction and print all events ignoring some fields
func (t FlowInteractionBuilder) RunPrintEvents(ignoreFields map[string][]string) {
	PrintEvents(t.Run(), ignoreFields)
}

// Run run the transaction
func (t FlowInteractionBuilder) Run() []flow.Event {
	events, err := t.RunE()
	if err != nil {
		t.Overflow.Logger.Error(fmt.Sprintf("%v Error executing script: %s output %v", emoji.PileOfPoo, t.FileName, err))
		panic(err)
	}
	return events
}

func (t FlowInteractionBuilder) RunGetIdFromEventPrintAll(eventName string, fieldName string) uint64 {
	result, err := t.RunE()
	if err != nil {
		panic(err)
	}
	PrintEvents(result, map[string][]string{})

	number, err := getUInt64FieldFromEvent(result, eventName, fieldName)
	if err != nil {
		panic(err)
	}
	return number
}

func getUInt64FieldFromEvent(result []flow.Event, eventName string, fieldName string) (uint64, error) {
	for _, event := range result {
		ev := ParseEvent(event, uint64(0), time.Unix(0, 0), []string{})
		if ev.Name == eventName {
			return ev.GetFieldAsUInt64(fieldName), nil
		}
	}
	return 0, fmt.Errorf("did not find field %s", fieldName)
}

func (t FlowInteractionBuilder) RunGetIdFromEvent(eventName string, fieldName string) uint64 {

	result, err := t.RunE()
	if err != nil {
		panic(err)
	}

	value, err := getUInt64FieldFromEvent(result, eventName, fieldName)
	if err != nil {
		panic(err)
	}
	return value
}

func (t FlowInteractionBuilder) RunGetIds(eventName string, fieldName string) ([]uint64, error) {

	result, err := t.RunE()
	if err != nil {
		return nil, err
	}
	var ids []uint64
	for _, event := range result {
		ev := ParseEvent(event, uint64(0), time.Unix(0, 0), []string{})
		if ev.Name == eventName {
			ids = append(ids, ev.GetFieldAsUInt64(fieldName))
		}
	}
	return ids, nil
}

func (t FlowInteractionBuilder) RunGetEventsWithNameOrError(eventName string) ([]FormatedEvent, error) {

	result, err := t.RunE()
	if err != nil {
		return nil, err
	}
	var events []FormatedEvent
	for _, event := range result {
		ev := ParseEvent(event, uint64(0), time.Unix(0, 0), []string{})
		if ev.Name == eventName {
			events = append(events, *ev)
		}
	}
	return events, nil
}

func (t FlowInteractionBuilder) RunGetEventsWithName(eventName string) []FormatedEvent {

	result, err := t.RunE()
	if err != nil {
		panic(err)
	}
	var events []FormatedEvent
	for _, event := range result {
		ev := ParseEvent(event, uint64(0), time.Unix(0, 0), []string{})
		if ev.Name == eventName {
			events = append(events, *ev)
		}
	}
	return events
}

// RunE runs returns events and error
func (t FlowInteractionBuilder) RunE() ([]flow.Event, error) {
	result := t.Send()
	return result.RawEvents, result.Err
}

// The new main way of running an overflow transaction
func (t FlowInteractionBuilder) Send() *OverflowResult {
	result := &OverflowResult{}
	if t.Error != nil {
		result.Err = t.Error
		return result
	}

	if t.Proposer == nil {
		fmt.Println("err")
		result.Err = fmt.Errorf("%v You need to set the main signer", emoji.PileOfPoo)
		return result
	}

	codeFileName := fmt.Sprintf("%s/%s.cdc", t.BasePath, t.FileName)

	if len(t.TransactionCode) == 0 {
		code, err := t.getContractCode(codeFileName)
		if err != nil {
			result.Err = err
			return result
		}
		t.TransactionCode = code
	}

	t.Overflow.Log.Reset()
	t.Overflow.EmulatorLog.Reset()
	// we append the mainSigners at the end here so that it signs last
	signers := t.PayloadSigners
	if t.MainSigner != nil {
		signers = append(signers, t.MainSigner)
	}

	var authorizers []flow.Address
	for _, signer := range signers {
		authorizers = append(authorizers, signer.Address())
	}
	if t.MainSigner == nil {
		signers = append(signers, t.Proposer)
	}

	tx, err := t.Overflow.Services.Transactions.Build(
		t.Proposer.Address(),
		authorizers,
		t.Proposer.Address(),
		t.Proposer.Key().Index(),
		t.TransactionCode,
		codeFileName,
		t.GasLimit,
		t.Arguments,
		t.Overflow.Network,
		true,
	)
	if err != nil {
		result.Err = err
		return result
	}

	for _, signer := range signers {
		err = tx.SetSigner(signer)
		if err != nil {
			result.Err = err
			return result
		}

		tx, err = tx.Sign()
		if err != nil {
			result.Err = err
			return result
		}
	}
	txId := tx.FlowTransaction().ID()
	result.Id = txId

	txBytes := []byte(fmt.Sprintf("%x", tx.FlowTransaction().Encode()))
	ftx, res, err := t.Overflow.Services.Transactions.SendSigned(txBytes, true)
	result.Transaction = ftx

	if err != nil {
		result.Err = err
		return result
	}

	var logMessage []LogrusMessage
	dec := json.NewDecoder(t.Overflow.Log)
	for {
		var doc LogrusMessage

		err := dec.Decode(&doc)
		if err == io.EOF {
			// all done
			break
		}
		if err != nil {
			result.Err = err
			return result
		}

		logMessage = append(logMessage, doc)
	}

	result.Meter = &Meter{}
	var meter Meter
	scanner := bufio.NewScanner(t.Overflow.EmulatorLog)
	for scanner.Scan() {
		txt := scanner.Text()
		if strings.Contains(txt, "transaction execution data") {
			err = json.Unmarshal([]byte(txt), &meter)
			if err == nil {
				result.Meter = &meter
			}
		}
	}
	var gas int
	messages := []string{}
	for _, msg := range logMessage {
		if msg.ComputationUsed != 0 {
			result.ComputationUsed = msg.ComputationUsed
			gas = msg.ComputationUsed
		}
		messages = append(messages, msg.Msg)
	}
	result.RawLog = logMessage
	result.EmulatorLog = messages

	if res.Error != nil {
		result.Err = res.Error
		return result
	}

	t.Overflow.Log.Reset()
	t.Overflow.EmulatorLog.Reset()
	t.Overflow.Logger.Info(fmt.Sprintf("%v Transaction %s successfully applied using gas:%d\n", emoji.OkHand, t.FileName, gas))
	result.RawEvents = res.Events
	return result
}

func (t FlowInteractionBuilder) getContractCode(codeFileName string) ([]byte, error) {
	code := []byte(t.Content)
	var err error
	if t.Content == "" {
		code, err = t.Overflow.State.ReaderWriter().ReadFile(codeFileName)
		if err != nil {
			return nil, fmt.Errorf("%v Could not read transaction file from path=%s", emoji.PileOfPoo, codeFileName)
		}
	}
	return code, nil
}

// FlowInteractionBuilder used to create a builder pattern for a transaction
type FlowInteractionBuilder struct {
	Overflow       *Overflow
	FileName       string
	Content        string
	Arguments      []cadence.Value
	MainSigner     *flowkit.Account
	PayloadSigners []*flowkit.Account
	GasLimit       uint64
	BasePath       string
	Error          error

	//these are used for v3, but can still be here for v2
	TransactionCode []byte
	NamedArgs       map[string]interface{}
	Proposer        *flowkit.Account
}

type OverflowScriptResult struct {
	Err    error
	Result cadence.Value
	Input  *FlowInteractionBuilder
}

func (osr *OverflowScriptResult) GetAsJson() string {
	if osr.Err != nil {
		panic(fmt.Sprintf("%v Error executing script: %s output %v", emoji.PileOfPoo, osr.Input.FileName, osr.Err))
	}
	return CadenceValueToJsonStringCompact(osr.Result)
}

func (osr *OverflowScriptResult) GetAsInterface() interface{} {
	if osr.Err != nil {
		panic(fmt.Sprintf("%v Error executing script: %s output %v", emoji.PileOfPoo, osr.Input.FileName, osr.Err))
	}
	return CadenceValueToInterfaceCompact(osr.Result)
}

func (osr *OverflowScriptResult) Print() {
	json := osr.GetAsJson()
	osr.Input.Overflow.Logger.Info(fmt.Sprintf("%v Script %s run from result: %v\n", emoji.Star, osr.Input.FileName, json))
}

func (osr *OverflowScriptResult) MarhalAs(value interface{}) error {
	if osr.Err != nil {
		return osr.Err
	}
	jsonResult := CadenceValueToJsonStringCompact(osr.Result)
	err := json.Unmarshal([]byte(jsonResult), &value)
	return err
}

type OverflowResult struct {
	Err             error
	Id              flow.Identifier
	EmulatorLog     []string
	Meter           *Meter
	ComputationUsed int
	RawEvents       []flow.Event
	RawLog          []LogrusMessage
	Transaction     *flow.Transaction
}

func (o OverflowResult) GetIdFromEvent(eventName string, fieldName string) uint64 {
	number, err := getUInt64FieldFromEvent(o.RawEvents, eventName, fieldName)
	if err != nil {
		panic(err)
	}
	return number
}

func (o OverflowResult) GetIdsFromEvent(eventName string, fieldName string) []uint64 {
	var ids []uint64
	for _, event := range o.RawEvents {
		ev := ParseEvent(event, uint64(0), time.Unix(0, 0), []string{})
		if ev.Name == eventName {
			ids = append(ids, ev.GetFieldAsUInt64(fieldName))
		}
	}
	return ids
}

func (o OverflowResult) GetEventsWithName(eventName string) []FormatedEvent {

	var events []FormatedEvent
	for _, event := range o.RawEvents {
		ev := ParseEvent(event, uint64(0), time.Unix(0, 0), []string{})
		if ev.Name == eventName {
			events = append(events, *ev)
		}
	}
	return events
}

// v3

//A function to customize the transaction builder
type TransactionOption func(*FlowInteractionBuilder)

type TransactionFunction func(filename string, opts ...TransactionOption) *OverflowResult
type TransactionOptsFunction func(opts ...TransactionOption) *OverflowResult

type ScriptFunction func(filename string, opts ...TransactionOption) *OverflowScriptResult
type ScriptOptsFunction func(opts ...TransactionOption) *OverflowScriptResult

func Arg(name, value string) func(ftb *FlowInteractionBuilder) {
	return func(ftb *FlowInteractionBuilder) {
		ftb.NamedArgs[name] = value
	}
}

func (o *Overflow) ScriptFN(outerOpts ...TransactionOption) ScriptFunction {

	return func(filename string, opts ...TransactionOption) *OverflowScriptResult {

		for _, opt := range opts {
			outerOpts = append(outerOpts, opt)
		}
		return o.Script(filename, outerOpts...)
	}
}

func (o *Overflow) TxFN(outerOpts ...TransactionOption) TransactionFunction {

	return func(filename string, opts ...TransactionOption) *OverflowResult {

		for _, opt := range opts {
			outerOpts = append(outerOpts, opt)
		}
		return o.Tx(filename, outerOpts...)

	}
}

func (o *Overflow) ScriptFileNameFN(filename string, outerOpts ...TransactionOption) ScriptOptsFunction {

	return func(opts ...TransactionOption) *OverflowScriptResult {

		for _, opt := range opts {
			outerOpts = append(outerOpts, opt)
		}
		return o.Script(filename, outerOpts...)
	}
}

func (o *Overflow) TxFileNameFN(filename string, outerOpts ...TransactionOption) TransactionOptsFunction {

	return func(opts ...TransactionOption) *OverflowResult {

		for _, opt := range opts {
			outerOpts = append(outerOpts, opt)
		}
		return o.Tx(filename, outerOpts...)

	}
}

func (o *Overflow) Tx(filename string, opts ...TransactionOption) *OverflowResult {
	return o.BuildInteraction(filename, "transaction", opts...).Send()
}

func (o *Overflow) Script(filename string, opts ...TransactionOption) *OverflowScriptResult {
	interaction := o.BuildInteraction(filename, "script", opts...)

	if interaction.Error != nil {
		return &OverflowScriptResult{Err: interaction.Error, Input: interaction}
	}

	filePath := fmt.Sprintf("%s/%s.cdc", interaction.BasePath, interaction.FileName)

	result, err := o.Services.Scripts.Execute(
		interaction.TransactionCode,
		interaction.Arguments,
		filePath,
		o.Network)
	if err != nil {
		return &OverflowScriptResult{Result: result, Err: interaction.Error, Input: interaction}
	}

	o.Logger.Info(fmt.Sprintf("%v Script run from path %s\n", emoji.Star, filePath))
	return &OverflowScriptResult{Result: result, Input: interaction}
}

//shouls this be private?
func (o *Overflow) BuildInteraction(filename string, interactionType string, opts ...TransactionOption) *FlowInteractionBuilder {

	path := o.TransactionBasePath
	if interactionType == "script" {
		path = o.ScriptBasePath
	}
	ftb := &FlowInteractionBuilder{
		Overflow:       o,
		MainSigner:     nil,
		Arguments:      []cadence.Value{},
		PayloadSigners: []*flowkit.Account{},
		GasLimit:       uint64(o.Gas),
		BasePath:       path,
		NamedArgs:      map[string]interface{}{},
	}

	if strings.Contains(filename, "transaction (") ||
		strings.Contains(filename, "transaction {") ||
		strings.Contains(filename, "transaction{") ||
		strings.Contains(filename, "transaction(") ||
		strings.Contains(filename, "transaction ") ||
		strings.Contains(filename, "pub fun main(") {
		ftb.TransactionCode = []byte(filename)
		ftb.FileName = "inline"
	} else {
		filePath := fmt.Sprintf("%s/%s.cdc", ftb.BasePath, filename)
		code, err := ftb.getContractCode(filePath)
		ftb.TransactionCode = code
		ftb.FileName = filename
		if err != nil {
			ftb.Error = err
			return ftb
		}
	}
	for _, opt := range opts {
		opt(ftb)
	}
	if ftb.Error != nil {
		return ftb
	}

	parseArgs, err := o.ParseArguments(ftb.FileName, ftb.TransactionCode, ftb.NamedArgs)
	if err != nil {
		ftb.Error = err
		return ftb
	}
	ftb.Arguments = parseArgs
	return ftb
}
