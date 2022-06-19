package overflow

import (
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
func (f *Overflow) TransactionFromFile(filename string) FlowTransactionBuilder {
	return FlowTransactionBuilder{
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
func (f *Overflow) Transaction(content string) FlowTransactionBuilder {
	return FlowTransactionBuilder{
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

func (t FlowTransactionBuilder) NamedArguments(args map[string]string) FlowTransactionBuilder {

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
func (t FlowTransactionBuilder) ArgsV(args []cadence.Value) FlowTransactionBuilder {
	t.Arguments = args
	return t
}

// Specify arguments to send to transaction using a builder you send in
func (t FlowTransactionBuilder) Args(args *FlowArgumentsBuilder) FlowTransactionBuilder {
	t.Arguments = args.Build()
	return t
}

// Specify arguments to send to transaction using a function that takes a builder where you call the builder
func (t FlowTransactionBuilder) ArgsFn(fn func(*FlowArgumentsBuilder)) FlowTransactionBuilder {
	args := t.Overflow.Arguments()
	fn(args)
	t.Arguments = args.Build()
	return t
}

func (t FlowTransactionBuilder) TransactionPath(path string) FlowTransactionBuilder {
	t.BasePath = path
	return t
}

// Gas sets the gas limit for this transaction
func (t FlowTransactionBuilder) Gas(limit uint64) FlowTransactionBuilder {
	t.GasLimit = limit
	return t
}

// SignProposeAndPayAs set the payer, proposer and envelope signer
func (t FlowTransactionBuilder) SignProposeAndPayAs(signer string) FlowTransactionBuilder {
	account, err := t.Overflow.AccountE(signer)
	if err != nil {
		t.Error = err
		return t
	}
	t.MainSigner = account
	return t
}

// SignProposeAndPayAsService set the payer, proposer and envelope signer
func (t FlowTransactionBuilder) SignProposeAndPayAsService() FlowTransactionBuilder {
	key := t.Overflow.ServiceAccountName()
	//swallow error as you cannot start a overflow without a valid sa
	account, _ := t.Overflow.State.Accounts().ByName(key)
	t.MainSigner = account
	return t
}

// PayloadSigner set a signer for the payload
func (t FlowTransactionBuilder) PayloadSigner(value string) FlowTransactionBuilder {
	signer := t.Overflow.Account(value)
	t.PayloadSigners = append(t.PayloadSigners, signer)
	return t
}

// RunPrintEventsFull will run a transaction and print all events
func (t FlowTransactionBuilder) RunPrintEventsFull() {
	PrintEvents(t.Run(), map[string][]string{})
}

// RunPrintEvents will run a transaction and print all events ignoring some fields
func (t FlowTransactionBuilder) RunPrintEvents(ignoreFields map[string][]string) {
	PrintEvents(t.Run(), ignoreFields)
}

// Run run the transaction
func (t FlowTransactionBuilder) Run() []flow.Event {
	events, err := t.RunE()
	if err != nil {
		t.Overflow.Logger.Error(fmt.Sprintf("%v Error executing script: %s output %v", emoji.PileOfPoo, t.FileName, err))
		panic(err)
	}
	return events
}

func (t FlowTransactionBuilder) RunGetIdFromEventPrintAll(eventName string, fieldName string) uint64 {
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

func (t FlowTransactionBuilder) RunGetIdFromEvent(eventName string, fieldName string) uint64 {

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

func (t FlowTransactionBuilder) RunGetIds(eventName string, fieldName string) ([]uint64, error) {

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

func (t FlowTransactionBuilder) RunGetEventsWithNameOrError(eventName string) ([]FormatedEvent, error) {

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

func (t FlowTransactionBuilder) RunGetEventsWithName(eventName string) []FormatedEvent {

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
func (t FlowTransactionBuilder) RunE() ([]flow.Event, error) {
	result := t.Send()
	return result.RawEvents, result.Err
}

// The new main way of running an overflow transaction
func (t FlowTransactionBuilder) Send() *OverflowResult {
	result := &OverflowResult{}
	if t.Error != nil {
		result.Err = t.Error
		return result
	}

	if t.MainSigner == nil {
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
	// we append the mainSigners at the end here so that it signs last
	signers := t.PayloadSigners
	signers = append(signers, t.MainSigner)

	signerKeyIndex := t.MainSigner.Key().Index()

	var authorizers []flow.Address
	for _, signer := range signers {
		authorizers = append(authorizers, signer.Address())
	}

	tx, err := t.Overflow.Services.Transactions.Build(
		t.MainSigner.Address(),
		authorizers,
		t.MainSigner.Address(),
		signerKeyIndex,
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

	t.Overflow.Logger.Info(fmt.Sprintf("Transaction ID: %s", txId))
	t.Overflow.Logger.StartProgress("Sending transaction...")
	defer t.Overflow.Logger.StopProgress()
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
	t.Overflow.Logger.Info(fmt.Sprintf("%v Transaction %s successfully applied using gas:%d\n", emoji.OkHand, t.FileName, gas))
	result.RawEvents = res.Events
	return result
}

func (t FlowTransactionBuilder) getContractCode(codeFileName string) ([]byte, error) {
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

// FlowTransactionBuilder used to create a builder pattern for a transaction
type FlowTransactionBuilder struct {
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
}

type OverflowResult struct {
	Err             error
	Id              flow.Identifier
	EmulatorLog     []string
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
type TransactionOption func(*FlowTransactionBuilder)

type TransactionFunction func(filename string, opts ...TransactionOption) *OverflowResult
type TransactionOptsFunction func(opts ...TransactionOption) *OverflowResult

func Arg(name, value string) func(ftb *FlowTransactionBuilder) {
	return func(ftb *FlowTransactionBuilder) {
		ftb.NamedArgs[name] = value
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

func (o *Overflow) TxFileNameFN(filename string, outerOpts ...TransactionOption) TransactionOptsFunction {

	return func(opts ...TransactionOption) *OverflowResult {

		for _, opt := range opts {
			outerOpts = append(outerOpts, opt)
		}
		return o.Tx(filename, outerOpts...)

	}
}

func (o *Overflow) Tx(filename string, opts ...TransactionOption) *OverflowResult {
	return o.Buildv3Transaction(filename, opts...).Send()
}

func (o *Overflow) Buildv3Transaction(filename string, opts ...TransactionOption) *FlowTransactionBuilder {
	ftb := &FlowTransactionBuilder{
		Overflow:       o,
		MainSigner:     nil,
		Arguments:      []cadence.Value{},
		PayloadSigners: []*flowkit.Account{},
		GasLimit:       uint64(o.Gas),
		BasePath:       o.TransactionBasePath,
		NamedArgs:      map[string]interface{}{},
	}

	if strings.Contains(filename, "transaction(") || strings.Contains(filename, "transaction(") {
		ftb.TransactionCode = []byte(filename)
		ftb.FileName = "inline"
	} else {
		filePath := fmt.Sprintf("%s/%s.cdc", o.TransactionBasePath, filename)
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
