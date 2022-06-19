package v3

import "github.com/bjartek/overflow/overflow"

func Arg(name, value string) func(ftb *overflow.FlowTransactionBuilder) {
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
