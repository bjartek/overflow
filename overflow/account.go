package overflow

import (
	"fmt"
	"sort"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

// CreateAccountsE ensures that all accounts present in the deployment block for the given network is present
func (f *OverflowState) CreateAccountsE() (*OverflowState, error) {
	p := f.State
	signerAccount, err := p.Accounts().ByName(f.ServiceAccountName())
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
			[]crypto.SignatureAlgorithm{account.Key().SigAlgo()},
			[]crypto.HashAlgorithm{account.Key().HashAlgo()},
			[]string{})
		if err != nil {
			return nil, err
		}
		f.Logger.Info("Account created " + a.Address.String())
	}
	return f, nil
}

// InitializeContracts installs all contracts in the deployment block for the configured network
func (f *OverflowState) InitializeContracts() *OverflowState {
	f.Logger.Info("Deploying contracts")
	if _, err := f.Services.Project.Deploy(f.Network, false); err != nil {
		panic(err)
	}
	return f
}

// GetAccount takes the account name  and returns the state of that account on the given network.
//TODO: consider renamting this method as this is getting a remove account a flow account not a flowkit account
func (f *OverflowState) GetAccount(key string) (*flow.Account, error) {
	account, err := f.AccountE(key)
	if err != nil {
		return nil, err
	}
	rawAddress := account.Address()
	return f.Services.Accounts.Get(rawAddress)
}
