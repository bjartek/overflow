package overflow

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/araddon/dateparse"
	"github.com/enescakir/emoji"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-go-sdk"
)

// TransactionFromFile will start a flow transaction builder
func (f *GoWithTheFlow) TransactionFromFile(filename string) FlowTransactionBuilder {
	return FlowTransactionBuilder{
		GoWithTheFlow:  f,
		FileName:       filename,
		MainSigner:     nil,
		Arguments:      []cadence.Value{},
		PayloadSigners: []*flowkit.Account{},
		GasLimit:       9999,
	}
}

// Transaction will start a flow transaction builder using the inline transaction
func (f *GoWithTheFlow) Transaction(content string) FlowTransactionBuilder {
	return FlowTransactionBuilder{
		GoWithTheFlow:  f,
		FileName:       "inline",
		Content:        content,
		MainSigner:     nil,
		Arguments:      []cadence.Value{},
		PayloadSigners: []*flowkit.Account{},
		GasLimit:       9999,
	}
}

// Gas sets the gas limit for this transaction
func (t FlowTransactionBuilder) Gas(limit uint64) FlowTransactionBuilder {
	t.GasLimit = limit
	return t
}

// SignProposeAndPayAs set the payer, proposer and envelope signer
func (t FlowTransactionBuilder) SignProposeAndPayAs(signer string) FlowTransactionBuilder {
	t.MainSigner = t.GoWithTheFlow.Account(signer)
	return t
}

// SignProposeAndPayAsService set the payer, proposer and envelope signer
func (t FlowTransactionBuilder) SignProposeAndPayAsService() FlowTransactionBuilder {
	key := fmt.Sprintf("%s-account", t.GoWithTheFlow.Network)
	account, err := t.GoWithTheFlow.State.Accounts().ByName(key)
	if err != nil {
		log.Fatal(err)
	}
	t.MainSigner = account
	return t
}

// RawAccountArgument add an account from a string as an argument
func (t FlowTransactionBuilder) RawAccountArgument(key string) FlowTransactionBuilder {
	account := flow.HexToAddress(key)
	accountArg := cadence.BytesToAddress(account.Bytes())
	return t.Argument(accountArg)
}

// AccountArgument add an account as an argument
func (t FlowTransactionBuilder) AccountArgument(key string) FlowTransactionBuilder {
	f := t.GoWithTheFlow

	account := f.Account(key)
	return t.Argument(cadence.BytesToAddress(account.Address().Bytes()))
}

// StringArgument add a String Argument to the transaction
func (t FlowTransactionBuilder) StringArgument(value string) FlowTransactionBuilder {
	return t.Argument(cadence.String(value))
}

// BooleanArgument add a Boolean Argument to the transaction
func (t FlowTransactionBuilder) BooleanArgument(value bool) FlowTransactionBuilder {
	return t.Argument(cadence.NewBool(value))
}

// BytesArgument add a Bytes Argument to the transaction
func (t FlowTransactionBuilder) BytesArgument(value []byte) FlowTransactionBuilder {
	return t.Argument(cadence.NewBytes(value))
}

// IntArgument add an Int Argument to the transaction
func (t FlowTransactionBuilder) IntArgument(value int) FlowTransactionBuilder {
	return t.Argument(cadence.NewInt(value))
}

// Int8Argument add an Int8 Argument to the transaction
func (t FlowTransactionBuilder) Int8Argument(value int8) FlowTransactionBuilder {
	return t.Argument(cadence.NewInt8(value))
}

// Int16Argument add an Int16 Argument to the transaction
func (t FlowTransactionBuilder) Int16Argument(value int16) FlowTransactionBuilder {
	return t.Argument(cadence.NewInt16(value))
}

// Int32Argument add an Int32 Argument to the transaction
func (t FlowTransactionBuilder) Int32Argument(value int32) FlowTransactionBuilder {
	return t.Argument(cadence.NewInt32(value))
}

// Int64Argument add an Int64 Argument to the transaction
func (t FlowTransactionBuilder) Int64Argument(value int64) FlowTransactionBuilder {
	return t.Argument(cadence.NewInt64(value))
}

// Int128Argument add an Int128 Argument to the transaction
func (t FlowTransactionBuilder) Int128Argument(value int) FlowTransactionBuilder {
	return t.Argument(cadence.NewInt128(value))
}

// Int256Argument add an Int256 Argument to the transaction
func (t FlowTransactionBuilder) Int256Argument(value int) FlowTransactionBuilder {
	return t.Argument(cadence.NewInt256(value))
}

// UIntArgument add an UInt Argument to the transaction
func (t FlowTransactionBuilder) UIntArgument(value uint) FlowTransactionBuilder {
	return t.Argument(cadence.NewUInt(value))
}

// UInt8Argument add an UInt8 Argument to the transaction
func (t FlowTransactionBuilder) UInt8Argument(value uint8) FlowTransactionBuilder {
	return t.Argument(cadence.NewUInt8(value))
}

// UInt16Argument add an UInt16 Argument to the transaction
func (t FlowTransactionBuilder) UInt16Argument(value uint16) FlowTransactionBuilder {
	return t.Argument(cadence.NewUInt16(value))
}

// UInt32Argument add an UInt32 Argument to the transaction
func (t FlowTransactionBuilder) UInt32Argument(value uint32) FlowTransactionBuilder {
	return t.Argument(cadence.NewUInt32(value))
}

// UInt64Argument add an UInt64 Argument to the transaction
func (t FlowTransactionBuilder) UInt64Argument(value uint64) FlowTransactionBuilder {
	return t.Argument(cadence.NewUInt64(value))
}

// UInt128Argument add an UInt128 Argument to the transaction
func (t FlowTransactionBuilder) UInt128Argument(value uint) FlowTransactionBuilder {
	return t.Argument(cadence.NewUInt128(value))
}

// UInt256Argument add an UInt256 Argument to the transaction
func (t FlowTransactionBuilder) UInt256Argument(value uint) FlowTransactionBuilder {
	return t.Argument(cadence.NewUInt256(value))
}

// Word8Argument add a Word8 Argument to the transaction
func (t FlowTransactionBuilder) Word8Argument(value uint8) FlowTransactionBuilder {
	return t.Argument(cadence.NewWord8(value))
}

// Word16Argument add a Word16 Argument to the transaction
func (t FlowTransactionBuilder) Word16Argument(value uint16) FlowTransactionBuilder {
	return t.Argument(cadence.NewWord16(value))
}

// Word32Argument add a Word32 Argument to the transaction
func (t FlowTransactionBuilder) Word32Argument(value uint32) FlowTransactionBuilder {
	return t.Argument(cadence.NewWord32(value))
}

// Word64Argument add a Word64 Argument to the transaction
func (t FlowTransactionBuilder) Word64Argument(value uint64) FlowTransactionBuilder {
	return t.Argument(cadence.NewWord64(value))
}

// Fix64Argument add a Fix64 Argument to the transaction
func (t FlowTransactionBuilder) Fix64Argument(value string) FlowTransactionBuilder {
	amount, err := cadence.NewFix64(value)
	if err != nil {
		panic(err)
	}
	return t.Argument(amount)
}

// DateStringAsUnixTimestamp sends a dateString parsed in the timezone as a unix timeszone ufix
func (t FlowTransactionBuilder) DateStringAsUnixTimestamp(dateString string, timezone string) FlowTransactionBuilder {
	return t.UFix64Argument(parseTime(dateString, timezone))
}

func parseTime(timeString string, location string) string {
	loc, err := time.LoadLocation(location)
	if err != nil {
		panic(err)
	}

	time.Local = loc
	t, err := dateparse.ParseLocal(timeString)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%d.0", t.Unix())
}

// UFix64Argument add a UFix64 Argument to the transaction
func (t FlowTransactionBuilder) UFix64Argument(value string) FlowTransactionBuilder {
	amount, err := cadence.NewUFix64(value)
	if err != nil {
		panic(err)
	}
	return t.Argument(amount)
}

// Argument add an argument to the transaction
func (t FlowTransactionBuilder) Argument(value cadence.Value) FlowTransactionBuilder {
	t.Arguments = append(t.Arguments, value)
	return t
}

// Argument add an argument to the transaction
func (t FlowTransactionBuilder) StringArrayArgument(value ...string) FlowTransactionBuilder {
	array := []cadence.Value{}
	for _, val := range value {
		stringVal, err := cadence.NewString(val)
		if err != nil {
			//TODO: what to do with errors here? Accumulate in builder?
			panic(err)
		}
		array = append(array, stringVal)
	}
	t.Arguments = append(t.Arguments, cadence.NewArray(array))
	return t
}

// PayloadSigner set a signer for the payload
func (t FlowTransactionBuilder) PayloadSigner(value string) FlowTransactionBuilder {
	signer := t.GoWithTheFlow.Account(value)
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
		t.GoWithTheFlow.Logger.Error(fmt.Sprintf("%v Error executing script: %s output %v", emoji.PileOfPoo, t.FileName, err))
		os.Exit(1)
	}
	return events
}

// RunE runs returns error
func (t FlowTransactionBuilder) RunE() ([]flow.Event, error) {
	if t.MainSigner == nil {
		return nil, fmt.Errorf("%v You need to set the main signer", emoji.PileOfPoo)
	}

	codeFileName := fmt.Sprintf("./transactions/%s.cdc", t.FileName)
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

	tx, err := t.GoWithTheFlow.Services.Transactions.Build(
		t.MainSigner.Address(),
		authorizers,
		t.MainSigner.Address(),
		signerKeyIndex,
		code,
		codeFileName,
		t.GasLimit,
		t.Arguments,
		t.GoWithTheFlow.Network,
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

	t.GoWithTheFlow.Logger.Info(fmt.Sprintf("Transaction ID: %s", tx.FlowTransaction().ID()))
	t.GoWithTheFlow.Logger.StartProgress("Sending transaction...")
	defer t.GoWithTheFlow.Logger.StopProgress()
	txBytes := []byte(fmt.Sprintf("%x", tx.FlowTransaction().Encode()))
	_, res, err := t.GoWithTheFlow.Services.Transactions.SendSigned(txBytes)

	if err != nil {
		return nil, err
	}

	if res.Error != nil {
		return nil, res.Error
	}

	t.GoWithTheFlow.Logger.Info(fmt.Sprintf("%v Transaction %s successfully applied\n", emoji.OkHand, t.FileName))
	return res.Events, nil
}

func (t FlowTransactionBuilder) getContractCode(codeFileName string) ([]byte, error) {
	code := []byte(t.Content)
	var err error
	if t.Content == "" {
		code, err = t.GoWithTheFlow.State.ReaderWriter().ReadFile(codeFileName)
		if err != nil {
			return nil, fmt.Errorf("%v Could not read transaction file from path=%s", emoji.PileOfPoo, codeFileName)
		}
	}
	return code, nil
}

// FlowTransactionBuilder used to create a builder pattern for a transaction
type FlowTransactionBuilder struct {
	GoWithTheFlow  *GoWithTheFlow
	FileName       string
	Content        string
	Arguments      []cadence.Value
	MainSigner     *flowkit.Account
	PayloadSigners []*flowkit.Account
	GasLimit       uint64
}
