package v3

import (
	"fmt"

	"github.com/bjartek/overflow/overflow"
	"github.com/onflow/cadence"
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

func ArgE(name string, fn ArgsError) func(ftb *overflow.FlowTransactionBuilder) {
	return func(ftb *overflow.FlowTransactionBuilder) {
		value, err := fn()
		if err != nil {
			ftb.Error = err
			return
		}
		ftb.NamedArgs[name] = value
	}
}
func CArg(name string, value cadence.Value) func(ftb *overflow.FlowTransactionBuilder) {
	return func(ftb *overflow.FlowTransactionBuilder) {
		ftb.NamedArgs[name] = value
	}
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
