package v3

import (
	"github.com/bjartek/overflow/overflow"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

/*
 Start a new Overflow instance that panics if there are initialization errors

 Will read the following ENV vars as default:
  OVERFLOW_ENV : set to "mainnet|testnet|emulator|embedded"
	OVERFLOW_LOGGING: set from 0-4 to get increasing amount of log output
	OVERFLOW_CONTINUE: to continue this overflow on an already running emulator.

 You can then chose to override this setting with the builder methods example
 `
  Overflow3(WithNetwork("mainnet"))
 `

 Setting the network in this way will reset other builder methods if appropriate so use with care.

*/
func Overflow(opts ...overflow.OverflowOption) *overflow.Overflow {
	o, err := overflow.NewOverflow().ApplyOptions(opts).StartE()
	if err != nil {
		panic(err)
	}
	return o
}

func OverflowE(opts ...overflow.OverflowOption) (*overflow.Overflow, error) {
	return overflow.NewOverflow().ApplyOptions(opts).StartE()
}

/*
	Can be used to start an overflow instance that can be used in tests
*/
func OverflowTesting(opts ...overflow.OverflowOption) (*overflow.Overflow, error) {
	return overflow.NewOverflowBuilder("embedded", true, 0).ApplyOptions(opts).StartE()
}

func WithNetwork(network string) func(o *overflow.OverflowBuilder) {
	return func(o *overflow.OverflowBuilder) {

		o.InMemory = false
		o.DeployContracts = false
		o.InitializeAccounts = false

		if network == "embedded" || network == "" {
			o.Network = "emulator"
			o.DeployContracts = true
			o.InitializeAccounts = true
			o.InMemory = true
			return
		}

		if network == "testing" {
			o.Network = "emulator"
			o.DeployContracts = true
			o.InitializeAccounts = true
			o.LogLevel = output.NoneLog
			o.InMemory = true
			return
		}
		if network == "emulator" {
			o.DeployContracts = true
			o.InitializeAccounts = true
		}
		o.Network = network
	}
}

func WithInMemory() func(o *overflow.OverflowBuilder) {
	return func(o *overflow.OverflowBuilder) {
		o.InMemory = true
		o.DeployContracts = true
		o.InitializeAccounts = true
		o.Network = "emulator"
	}
}

func WithExistingEmulator() func(o *overflow.OverflowBuilder) {
	return func(o *overflow.OverflowBuilder) {
		o.DeployContracts = false
		o.InitializeAccounts = false
	}
}

func DoNotPrependNetworkToAccountNames() func(o *overflow.OverflowBuilder) {
	return func(o *overflow.OverflowBuilder) {
		o.PrependNetworkName = false
	}
}

func WithServiceAccountSuffix(suffix string) func(o *overflow.OverflowBuilder) {
	return func(o *overflow.OverflowBuilder) {
		o.ServiceSuffix = suffix
	}
}

func WithNoLog() func(o *overflow.OverflowBuilder) {
	return func(o *overflow.OverflowBuilder) {
		o.LogLevel = output.NoneLog
	}
}

func WithGas(gas int) func(o *overflow.OverflowBuilder) {
	return func(o *overflow.OverflowBuilder) {
		o.GasLimit = gas
	}
}

func WithBasePath(path string) func(o *overflow.OverflowBuilder) {
	return func(o *overflow.OverflowBuilder) {
		o.Path = path
	}
}

func WithFlowConfig(files ...string) func(o *overflow.OverflowBuilder) {
	return func(o *overflow.OverflowBuilder) {
		o.ConfigFiles = files
	}
}

func WithScriptFolderName(name string) func(o *overflow.OverflowBuilder) {
	return func(o *overflow.OverflowBuilder) {
		o.ScriptFolderName = name
	}
}

func WithTransactionFolderName(name string) func(o *overflow.OverflowBuilder) {
	return func(o *overflow.OverflowBuilder) {
		o.TransactionFolderName = name
	}
}
