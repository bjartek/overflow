package overflow

import (
	"fmt"
	"log"
	"sort"

	"github.com/onflow/flow-go-sdk/crypto"
)

func (f *GoWithTheFlow) CreateAccounts(saAccountName string) *GoWithTheFlow {
	gwtf, err := f.CreateAccountsE(saAccountName)
	if err != nil {
		log.Fatal(err)
	}

	return gwtf

}

// CreateAccountsE ensures that all accounts present in the deployment block for the given network is present
func (f *GoWithTheFlow) CreateAccountsE(saAccountName string) (*GoWithTheFlow, error) {
	p := f.State
	signerAccount, err := p.Accounts().ByName(saAccountName)
	if err != nil {
		return nil, err
	}

	accounts := p.AccountNamesForNetwork(f.Network)
	sort.Strings(accounts)

	f.Logger.Info(fmt.Sprintf("%v\n", accounts))

	for _, accountName := range accounts {
		f.Logger.Info(fmt.Sprintf("Ensuring account with name '%s' is present", accountName))

		// this error can never happen here, there is a test for it.
		account, _ := p.Accounts().ByName(accountName)

		if _, err := f.Services.Accounts.Get(account.Address()); err == nil {
			f.Logger.Info("Account is present")
			continue
		}

		a, err := f.Services.Accounts.Create(
			signerAccount,
			[]crypto.PublicKey{account.Key().ToConfig().PrivateKey.PublicKey()},
			[]int{1000},
			account.Key().SigAlgo(),
			account.Key().HashAlgo(),
			[]string{})
		if err != nil {
			return nil, err
		}
		f.Logger.Info("Account created " + a.Address.String())
	}
	return f, nil
}

// InitializeContracts installs all contracts in the deployment block for the configured network
func (f *GoWithTheFlow) InitializeContracts() *GoWithTheFlow {
	f.Logger.Info("Deploying contracts")
	if _, err := f.Services.Project.Deploy(f.Network, false); err != nil {
		log.Fatal(err)
	}

	return f
}
