package v3

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/bjartek/overflow/overflow"
	"github.com/onflow/cadence"
	"github.com/pkg/errors"
)

type ArgsError func() (cadence.Value, error)

func Args(args ...interface{}) func(ftb *overflow.FlowTransactionBuilder) {

	return func(ftb *overflow.FlowTransactionBuilder) {
		if len(args)%2 != 0 {
			ftb.Error = fmt.Errorf("Please send in an even number of string : interface{} pairs")
			fmt.Println("foo")
			return
		}
		var i = 0
		for i < len(args) {
			key := args[0]
			value, labelOk := key.(string)
			if !labelOk {
				ftb.Error = fmt.Errorf("even parameters in Args needs to be strings")
			}
			ftb.NamedArgs[value] = args[1]
			i = i + 2
		}
	}
}

func ArgsM(args map[string]interface{}) func(ftb *overflow.FlowTransactionBuilder) {
	return func(ftb *overflow.FlowTransactionBuilder) {
		for key, value := range args {
			ftb.NamedArgs[key] = value
		}
	}
}

func Arg(name, value string) func(ftb *overflow.FlowTransactionBuilder) {
	return func(ftb *overflow.FlowTransactionBuilder) {
		ftb.NamedArgs[name] = value
	}
}

func CArg(name string, value cadence.Value) func(ftb *overflow.FlowTransactionBuilder) {
	return func(ftb *overflow.FlowTransactionBuilder) {
		ftb.NamedArgs[name] = value
	}
}

func Addresses(name string, value ...string) func(ftb *overflow.FlowTransactionBuilder) {
	return func(ftb *overflow.FlowTransactionBuilder) {
		array := []cadence.Value{}

		for _, val := range value {
			account, err := ftb.Overflow.AccountE(val)
			if err != nil {
				address, err := HexToAddress(val)
				if err != nil {
					ftb.Error = errors.Wrap(err, fmt.Sprintf("%s is not an valid account name or an address", val))
					return
				}
				cadenceAddress := cadence.BytesToAddress(address.Bytes())
				array = append(array, cadenceAddress)
			} else {
				cadenceAddress := cadence.BytesToAddress(account.Address().Bytes())
				array = append(array, cadenceAddress)
			}
		}
		ftb.NamedArgs[name] = cadence.NewArray(array)
	}
}

// HexToAddress converts a hex string to an Address.
func HexToAddress(h string) (*cadence.Address, error) {
	trimmed := strings.TrimPrefix(h, "0x")
	if len(trimmed)%2 == 1 {
		trimmed = "0" + trimmed
	}
	b, err := hex.DecodeString(trimmed)
	if err != nil {
		return nil, err

	}
	address := cadence.BytesToAddress(b)
	return &address, nil
}

func SignProposeAndPayAs(signer string) func(ftb *overflow.FlowTransactionBuilder) {
	return func(ftb *overflow.FlowTransactionBuilder) {
		account, err := ftb.Overflow.AccountE(signer)
		if err != nil {
			ftb.Error = err
			return
		}
		ftb.MainSigner = account
	}
}

func SignProposeAndPayAsServiceAccount() func(ftb *overflow.FlowTransactionBuilder) {
	return func(ftb *overflow.FlowTransactionBuilder) {
		key := ftb.Overflow.ServiceAccountName()
		account, _ := ftb.Overflow.State.Accounts().ByName(key)
		ftb.MainSigner = account
	}
}
