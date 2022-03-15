package overflow

import (
	"fmt"
	"log"
	"os"
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
	t.MainSigner = t.Overflow.Account(signer)
	return t
}

// SignProposeAndPayAsService set the payer, proposer and envelope signer
func (t FlowTransactionBuilder) SignProposeAndPayAsService() FlowTransactionBuilder {
	key := t.Overflow.ServiceAccountName()
	account, err := t.Overflow.State.Accounts().ByName(key)
	if err != nil {
		log.Fatal(err)
	}
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
		os.Exit(1)
	}
	return events
}

func (t FlowTransactionBuilder) RunGetIdFromEventPrintAll(eventName string, fieldName string) uint64 {
	result, err := t.RunE()
	if err != nil {
		panic(err)
	}
	PrintEvents(result, map[string][]string{})

	return getUInt64FieldFromEvent(result, eventName, fieldName)
}

func getUInt64FieldFromEvent(result []flow.Event, eventName string, fieldName string) uint64 {
	for _, event := range result {
		ev := ParseEvent(event, uint64(0), time.Unix(0, 0), []string{})
		if ev.Name == eventName {
			return ev.GetFieldAsUInt64(fieldName)
		}
	}
	panic("did not find field")
}

func (t FlowTransactionBuilder) RunGetIdFromEvent(eventName string, fieldName string) uint64 {

	result, err := t.RunE()
	if err != nil {
		panic(err)
	}

	return getUInt64FieldFromEvent(result, eventName, fieldName)
}

// RunE runs returns error
func (t FlowTransactionBuilder) RunE() ([]flow.Event, error) {
	if t.MainSigner == nil {
		return nil, fmt.Errorf("%v You need to set the main signer", emoji.PileOfPoo)
	}

	codeFileName := fmt.Sprintf("%s/%s.cdc", t.BasePath, t.FileName)
	code, err := t.getContractCode(codeFileName)
	if err != nil {
		return nil, err
	}
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
		code,
		codeFileName,
		t.GasLimit,
		t.Arguments,
		t.Overflow.Network,
		true,
	)
	if err != nil {
		return nil, err
	}

	for _, signer := range signers {
		err = tx.SetSigner(signer)
		if err != nil {
			return nil, err
		}

		tx, err = tx.Sign()
		if err != nil {
			return nil, err
		}
	}

	t.Overflow.Logger.Info(fmt.Sprintf("Transaction ID: %s", tx.FlowTransaction().ID()))
	t.Overflow.Logger.StartProgress("Sending transaction...")
	defer t.Overflow.Logger.StopProgress()
	txBytes := []byte(fmt.Sprintf("%x", tx.FlowTransaction().Encode()))
	_, res, err := t.Overflow.Services.Transactions.SendSigned(txBytes, true)

	if err != nil {
		return nil, err
	}

	if res.Error != nil {
		return nil, res.Error
	}

	t.Overflow.Logger.Info(fmt.Sprintf("%v Transaction %s successfully applied\n", emoji.OkHand, t.FileName))
	return res.Events, nil
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
}
