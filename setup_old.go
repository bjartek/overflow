package overflow

import (
	"fmt"
	"os"
	"strconv"

	"github.com/enescakir/emoji"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

// NewOverflow creates a new OverflowBuilder reading some confiuration from ENV var (
// OVERFLOW_ENV : sets the environment to use, valid values here are emulator|testnet|mainnet|embedded
// OVERFLOW_CONTINUE : if set to `true` will not create accounts and deploy contracts even if on embedded/emulator
// OVERFLOW_LOGGING : set the logging level of flowkit and overflow itself, 0 = No Log, 1 = Errors only, 2 = Debug, 3(default) = Info
//
// Deprecated: use Overflow function with builder
func NewOverflow() *OverflowBuilder {
	network := os.Getenv("OVERFLOW_ENV")
	existing := os.Getenv("OVERFLOW_CONTINUE")
	loglevel := os.Getenv("OVERFLOW_LOGGING")
	var log int
	var err error
	if loglevel != "" {
		log, err = strconv.Atoi(loglevel)
		if err != nil {
			panic(err)
		}
	} else {
		log = output.InfoLog
	}
	return NewOverflowBuilder(network, existing != "true", log)

}

// Deprecated: use Overflow function with builder
func NewOverflowBuilder(network string, newEmulator bool, logLevel int) *OverflowBuilder {
	inMemory := false
	deployContracts := newEmulator
	initializeAccounts := newEmulator

	if network == "embedded" || network == "" {
		inMemory = true
		network = emulatorValue
	}

	if network == emulatorValue {
		deployContracts = true
		initializeAccounts = true
	}

	return &OverflowBuilder{
		Network:                             network,
		InMemory:                            inMemory,
		DeployContracts:                     deployContracts,
		GasLimit:                            9999,
		Path:                                ".",
		TransactionFolderName:               "transactions",
		ScriptFolderName:                    "scripts",
		LogLevel:                            logLevel,
		InitializeAccounts:                  initializeAccounts,
		PrependNetworkName:                  true,
		ServiceSuffix:                       "account",
		ConfigFiles:                         config.DefaultPaths(),
		FilterOutEmptyWithDrawDepositEvents: true,
		FilterOutFeeEvents:                  true,
		GlobalEventFilter:                   OverflowEventFilter{},
		StopOnError:                         false,
		PrintOptions:                        nil,
		NewAccountFlowAmount:                0.0,
	}
}

// ExistingEmulator this if you are using an existing emulator and you do not want to create contracts or initializeAccounts
//
// Deprecated: use Overflow function with builder
func (o *OverflowBuilder) ExistingEmulator() *OverflowBuilder {
	o.DeployContracts = false
	o.InitializeAccounts = false
	return o
}

// DoNotPrependNetworkToAccountNames sets that network names will not be prepends to account names
//
// Deprecated: use Overflow function with builder
func (o *OverflowBuilder) DoNotPrependNetworkToAccountNames() *OverflowBuilder {
	o.PrependNetworkName = false
	return o
}

// SetServiceSuffix will set the suffix to use for the service account. The default is `account`
//
// Deprecated: use Overflow function with builder
func (o *OverflowBuilder) SetServiceSuffix(suffix string) *OverflowBuilder {
	o.ServiceSuffix = suffix
	return o
}

// NoneLog will turn of logging, making the script work well in batch jobs
//
// Deprecated: use Overflow function with builder
func (o *OverflowBuilder) NoneLog() *OverflowBuilder {
	o.LogLevel = output.NoneLog
	return o
}

// DefaultGas sets the default gas limit to use
//
// Deprecated: use Overflow function with builder
func (o *OverflowBuilder) DefaultGas(gas int) *OverflowBuilder {
	o.GasLimit = gas
	return o
}

// BasePath set the base path for transactions/scripts/contracts
//
// Deprecated: use Overflow function with builder
func (o *OverflowBuilder) BasePath(path string) *OverflowBuilder {
	o.Path = path
	return o
}

// Config sets the file path to the flow.json config files to use
//
// Deprecated: use Overflow function with builder
func (o *OverflowBuilder) Config(files ...string) *OverflowBuilder {
	o.ConfigFiles = files
	return o
}

// Start will start the overflow builder and return OverflowState, will panic if there are errors
//
// Deprecated: use Overflow function with builder
func (ob *OverflowBuilder) Start() *OverflowState {
	o, err := ob.StartE()
	if err != nil {
		panic(fmt.Sprintf("%v error %+v", emoji.PileOfPoo, err))
	}
	return o
}

// NewOverflowInMemoryEmulator this method is used to create an in memory emulator, deploy all contracts for the emulator and create all accounts
// Deprecated: use Overflow function with builder
func NewOverflowInMemoryEmulator() *OverflowBuilder {
	return NewOverflowBuilder("embedded", true, output.InfoLog)
}

// NewOverflowForNetwork creates a new overflow client for the provided network
//
// Deprecated: use Overflow function with builder
func NewOverflowForNetwork(network string) *OverflowBuilder {
	return NewOverflowBuilder(network, false, output.InfoLog)
}

// NewOverflowEmulator create a new client
//
// Deprecated: use Overflow function with builder
func NewOverflowEmulator() *OverflowBuilder {
	return NewOverflowBuilder(emulatorValue, false, output.InfoLog)
}

// NewTestingEmulator starts an embedded emulator with no log to be used most often in tests
//
// Deprecated: use Overflow function with builder
func NewTestingEmulator() *OverflowBuilder {
	return NewOverflowBuilder("embedded", true, 0)
}

// NewOverflowTestnet creates a new overflow client for devnet/testnet
//
// Deprecated: use Overflow function with builder
func NewOverflowTestnet() *OverflowBuilder {
	return NewOverflowBuilder("testnet", false, output.InfoLog)
}

// NewOverflowMainnet creates a new gwft client for mainnet
//
// Deprecated: use Overflow function with builder
func NewOverflowMainnet() *OverflowBuilder {
	return NewOverflowBuilder("mainnet", false, output.InfoLog)
}
