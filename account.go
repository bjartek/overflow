package overflow

import (
	"sort"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/pkg/errors"
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

	for _, accountName := range accounts {

		// this error can never happen here, there is a test for it.
		account, _ := p.Accounts().ByName(accountName)

		if _, err := f.Services.Accounts.Get(account.Address()); err == nil {
			continue
		}

		_, err := f.Services.Accounts.Create(
			signerAccount,
			[]crypto.PublicKey{account.Key().ToConfig().PrivateKey.PublicKey()},
			[]int{1000},
			[]crypto.SignatureAlgorithm{account.Key().SigAlgo()},
			[]crypto.HashAlgorithm{account.Key().HashAlgo()},
			[]string{})
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}

// InitializeContracts installs all contracts in the deployment block for the configured network
func (o *OverflowState) InitializeContracts() *OverflowState {
	o.Log.Reset()
	if _, err := o.Services.Project.Deploy(o.Network, false); err != nil {
		log, _ := o.readLog()
		if len(log) != 0 {
			messages := []string{}
			for _, msg := range log {
				if msg.Level == "warning" || msg.Level == "error" {
					messages = append(messages, msg.Msg)
				}
			}
			o.Error = errors.Wrapf(err, "errors : %v", messages)
		} else {
			o.Error = err
		}
	}
	o.Log.Reset()
	return o
}

// GetAccount takes the account name  and returns the state of that account on the given network.
//TODO: consider renaming this method as this is getting a remove account a flow account not a flowkit account
func (f *OverflowState) GetAccount(key string) (*flow.Account, error) {
	account, err := f.AccountE(key)
	if err != nil {
		return nil, err
	}
	rawAddress := account.Address()
	return f.Services.Accounts.Get(rawAddress)
}
