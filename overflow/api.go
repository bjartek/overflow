package overflow

import (
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

func main() {
	fmt.Println("vim-go")
}

type OverflowOption func(*OverflowBuilder)

func (o *OverflowBuilder) StartWithOpts(opts []OverflowOption) (*Overflow, error) {
	for _, opt := range opts {
		opt(o)
	}

	return o.StartE()
}

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
func Overflow3(opts ...OverflowOption) *Overflow {
	o, err := NewOverflow().StartWithOpts(opts)
	if err != nil {
		panic(err)
	}
	return o
}

func OverflowE(opts ...OverflowOption) (*Overflow, error) {
	return NewOverflow().StartWithOpts(opts)
}

/*
	Can be used to start an overflow instance that can be used in tests
*/
func OverflowTesting(opts ...OverflowOption) (*Overflow, error) {
	return NewOverflowBuilder("embedded", true, 0).StartWithOpts(opts)
}

func WithNetwork(network string) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {

		if network == "embedded" || network == "" {
			o.Network = "emulator"
			o.DeployContracts = true
			o.InitializeAccounts = true
			o.InMemory = true
			o.LogLevel = output.InfoLog
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

func WithInMemory() func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.InMemory = true
	}
}

func WithExistingEmulator() func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.DeployContracts = false
		o.InitializeAccounts = false
	}
}

func DoNotPrependNetworkToAccountNames() func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.PrependNetworkName = false
	}
}

func WithServiceAccountSuffix(suffix string) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.ServiceSuffix = suffix
	}
}

func WithNoLog() func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.LogLevel = output.NoneLog
	}
}

func WithGas(gas int) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.GasLimit = gas
	}
}

func WithBasePath(path string) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.Path = path
	}
}

func WithFlowConfig(files ...string) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.ConfigFiles = files
	}
}
