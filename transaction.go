package overflow

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/bjartek/underflow"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser"
	"github.com/onflow/flow-go-sdk"
)

type FilterFunction func(OverflowTransaction) bool

type Argument struct {
	Value interface{}
	Key   string
}

type OverflowTransaction struct {
	Error            error
	AuthorizerTypes  map[string][]string
	Stakeholders     map[string][]string
	Payer            string
	Id               string
	Status           string
	BlockId          string
	Authorizers      []string
	Arguments        []Argument
	Events           []OverflowEvent
	Imports          []Import
	Script           []byte
	ProposalKey      flow.ProposalKey
	Fee              float64
	TransactionIndex int
	GasLimit         uint64
	GasUsed          uint64
	ExecutionEffort  float64
}

func (o *OverflowState) CreateOverflowTransaction(blockId string, transactionResult flow.TransactionResult, transaction flow.Transaction, txIndex int) (*OverflowTransaction, error) {
	feeAmount := 0.0
	events, fee := o.ParseEvents(transactionResult.Events, "")
	feeRaw, ok := fee.Fields["amount"]
	if ok {
		feeAmount, ok = feeRaw.(float64)
		if !ok {
			return nil, fmt.Errorf("failed casting fee amount to float64")
		}
	}

	executionEffort, ok := fee.Fields["executionEffort"].(float64)
	gas := 0
	if ok {
		factor := 100000000
		gas = int(math.Round(executionEffort * float64(factor)))
	}

	status := transactionResult.Status.String()

	args := []Argument{}
	argInfo := declarationInfo(transaction.Script)
	for i := range transaction.Arguments {
		arg, err := transaction.Argument(i)
		if err != nil {
			status = fmt.Sprintf("%s failed getting argument at index %d", status, i)
		}
		var key string
		if len(argInfo.ParameterOrder) <= i {
			key = "invalid"
		} else {
			key = argInfo.ParameterOrder[i]
		}
		argStruct := Argument{
			Key:   key,
			Value: underflow.CadenceValueToInterfaceWithOption(arg, o.UnderflowOptions),
		}
		args = append(args, argStruct)
	}

	standardStakeholders := map[string][]string{}
	imports, err := GetAddressImports(transaction.Script)
	if err != nil {
		status = fmt.Sprintf("%s failed getting imports", status)
	}

	authorizerTypes := map[string][]string{}

	authorizers := []string{}
	for i, authorizer := range transaction.Authorizers {
		auth := fmt.Sprintf("0x%s", authorizer.Hex())
		authorizers = append(authorizers, auth)
		standardStakeholders[auth] = []string{"authorizer"}
		authorizerTypes[auth] = argInfo.Authorizers[i]
	}

	payerRoles, ok := standardStakeholders[fmt.Sprintf("0x%s", transaction.Payer.Hex())]
	if !ok {
		standardStakeholders[fmt.Sprintf("0x%s", transaction.Payer.Hex())] = []string{"payer"}
	} else {
		payerRoles = append(payerRoles, "payer")
		standardStakeholders[fmt.Sprintf("0x%s", transaction.Payer.Hex())] = payerRoles
	}

	proposer, ok := standardStakeholders[fmt.Sprintf("0x%s", transaction.ProposalKey.Address.Hex())]
	if !ok {
		standardStakeholders[fmt.Sprintf("0x%s", transaction.ProposalKey.Address.Hex())] = []string{"proposer"}
	} else {
		proposer = append(proposer, "proposer")
		standardStakeholders[fmt.Sprintf("0x%s", transaction.ProposalKey.Address.Hex())] = proposer
	}

	eventsWithoutFees := events.FilterFees(feeAmount, fmt.Sprintf("0x%s", transaction.Payer.Hex()))

	eventList := []OverflowEvent{}
	for _, evList := range eventsWithoutFees {
		eventList = append(eventList, evList...)
	}

	return &OverflowTransaction{
		Id:               transactionResult.TransactionID.String(),
		TransactionIndex: txIndex,
		BlockId:          blockId,
		Status:           status,
		Events:           eventList,
		Stakeholders:     eventsWithoutFees.GetStakeholders(standardStakeholders),
		Imports:          imports,
		Error:            transactionResult.Error,
		Arguments:        args,
		Fee:              feeAmount,
		Script:           transaction.Script,
		Payer:            fmt.Sprintf("0x%s", transaction.Payer.String()),
		ProposalKey:      transaction.ProposalKey,
		GasLimit:         transaction.GasLimit,
		GasUsed:          uint64(gas),
		ExecutionEffort:  executionEffort,
		Authorizers:      authorizers,
		AuthorizerTypes:  authorizerTypes,
	}, nil
}

func (o *OverflowState) GetOverflowTransactionById(ctx context.Context, id flow.Identifier) (*OverflowTransaction, error) {
	tx, txr, err := o.Flowkit.GetTransactionByID(ctx, id, false)
	if err != nil {
		return nil, err
	}
	txIndex := 0
	if len(txr.Events) > 0 {
		txIndex = txr.Events[0].TransactionIndex
	}
	return o.CreateOverflowTransaction(txr.BlockID.String(), *txr, *tx, txIndex)
}

func (o *OverflowState) GetTransactionById(ctx context.Context, id flow.Identifier) (*flow.Transaction, error) {
	tx, _, err := o.Flowkit.GetTransactionByID(ctx, id, false)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (o *OverflowState) GetTransactionByBlockId(ctx context.Context, id flow.Identifier) ([]*flow.Transaction, []*flow.TransactionResult, error) {
	tx, txr, err := o.Flowkit.GetTransactionsByBlockID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return tx, txr, nil
}

func GetAddressImports(code []byte) ([]Import, error) {
	deps := []Import{}
	program, err := parser.ParseProgram(nil, code, parser.Config{})
	if err != nil {
		return deps, err
	}

	for _, imp := range program.ImportDeclarations() {
		address, isAddressImport := imp.Location.(common.AddressLocation)
		if isAddressImport {
			for _, id := range imp.Identifiers {
				deps = append(deps, Import{
					Address: fmt.Sprintf("0x%s", address.Address.Hex()),
					Name:    id.Identifier,
				})
			}
		}
	}
	return deps, nil
}

type Import struct {
	Address string
	Name    string
}

func (i Import) Identifier() string {
	return fmt.Sprintf("A.%s.%s", strings.TrimPrefix(i.Address, "0x"), i.Name)
}
