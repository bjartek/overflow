package overflow

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/enescakir/emoji"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/rs/zerolog"
	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

//OverflowBuilder is the struct used to gather up configuration when building an overflow instance
type OverflowBuilder struct {
	Network                             string
	InMemory                            bool
	DeployContracts                     bool
	GasLimit                            int
	Path                                string
	LogLevel                            int
	InitializeAccounts                  bool
	PrependNetworkName                  bool
	ServiceSuffix                       string
	ConfigFiles                         []string
	TransactionFolderName               string
	ScriptFolderName                    string
	FilterOutFeeEvents                  bool
	FilterOutEmptyWithDrawDepositEvents bool
	GlobalEventFilter                   OverFlowEventFilter
}

//NewOverflow creates a new OverflowBuilder reading some confiuration from ENV var (
// - OVERFLOW_ENV : sets the environment to use, valid values here are emulator|testnet|mainnet|embedded
// - OVERFLOW_CONTINUE : if set to `true` will not create accounts and deploy contracts even if on embeded/emulator
// - OVERFLOW_LOGGING : set the logging level of flowkit and overflow itself, 0 = No Log, 1 = Errors only, 2 = Debug, 3(default) = Info

//Deprecated use Overflow function with builder
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

//Deprecated use Overflow function with builder
func NewOverflowBuilder(network string, newEmulator bool, logLevel int) *OverflowBuilder {
	inMemory := false
	deployContracts := newEmulator
	initializeAccounts := newEmulator

	if network == "embedded" || network == "" {
		inMemory = true
		network = "emulator"
	}

	if network == "emulator" {
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
		GlobalEventFilter:                   OverFlowEventFilter{},
	}
}

//ExistingEmulator this if you are using an existing emulator and you do not want to create contracts or initializeAccounts
//Deprecated use Overflow function with builder
func (o *OverflowBuilder) ExistingEmulator() *OverflowBuilder {
	o.DeployContracts = false
	o.InitializeAccounts = false
	return o
}

//DoNotPrependNetworkToAccountNames sets that network names will not be prepends to account names
//Deprecated use Overflow function with builder
func (o *OverflowBuilder) DoNotPrependNetworkToAccountNames() *OverflowBuilder {
	o.PrependNetworkName = false
	return o
}

//SetServiceSuffix will set the suffix to use for the service account. The default is `account`
//Deprecated use Overflow function with builder
func (o *OverflowBuilder) SetServiceSuffix(suffix string) *OverflowBuilder {
	o.ServiceSuffix = suffix
	return o
}

//NoneLog will turn of logging, making the script work well in batch jobs
//Deprecated use Overflow function with builder
func (o *OverflowBuilder) NoneLog() *OverflowBuilder {
	o.LogLevel = output.NoneLog
	return o
}

//DefaultGas sets the default gas limit to use
//Deprecated use Overflow function with builder
func (o *OverflowBuilder) DefaultGas(gas int) *OverflowBuilder {
	o.GasLimit = gas
	return o
}

//BasePath set the base path for transactions/scripts/contracts
//Deprecated use Overflow function with builder
func (o *OverflowBuilder) BasePath(path string) *OverflowBuilder {
	o.Path = path
	return o
}

//Config sets the file path to the flow.json config files to use
//Deprecated use Overflow function with builder
func (o *OverflowBuilder) Config(files ...string) *OverflowBuilder {
	o.ConfigFiles = files
	return o
}

//Start will start the overflow builder and return OverflowState, will panic if there are errors
//Deprecated use Overflow function with builder
func (ob *OverflowBuilder) Start() *OverflowState {
	o, err := ob.StartE()
	if err != nil {
		panic(fmt.Sprintf("%v error %+v", emoji.PileOfPoo, err))
	}
	return o
}

//StartE will start Overflow and return State and error if any
//Deprecated use Overflow function with builder
func (o *OverflowBuilder) StartE() (*OverflowState, error) {

	loader := &afero.Afero{Fs: afero.NewOsFs()}
	state, err := flowkit.Load(o.ConfigFiles, loader)
	if err != nil {
		return nil, err
	}

	logger := output.NewStdoutLogger(o.LogLevel)
	var service *services.Services
	var memlog bytes.Buffer
	var emulatorLog bytes.Buffer

	if o.InMemory {
		//YAY we can run it inline in memory!
		acc, _ := state.EmulatorServiceAccount()

		logrusLogger := &logrus.Logger{
			Formatter: &logrus.JSONFormatter{},
			Level:     logrus.TraceLevel,
			Out:       &memlog,
		}

		writer := io.Writer(&emulatorLog)
		emulatorLogger := zerolog.New(writer).Level(zerolog.DebugLevel)

		//YAY we can now get out embedded logs!
		gw := gateway.NewEmulatorGatewayWithOpts(acc, gateway.WithLogger(logrusLogger), gateway.WithEmulatorLogger(&emulatorLogger))
		service = services.NewServices(gw, state, logger)
	} else {
		network, err := state.Networks().ByName(o.Network)
		if err != nil {
			return nil, err
		}
		host := network.Host
		gw, err := gateway.NewGrpcGateway(host)
		if err != nil {
			return nil, err
		}
		service = services.NewServices(gw, state, logger)
	}
	overflow := &OverflowState{
		State:                               state,
		Services:                            service,
		Network:                             o.Network,
		Logger:                              logger,
		PrependNetworkToAccountNames:        o.PrependNetworkName,
		ServiceAccountSuffix:                o.ServiceSuffix,
		Gas:                                 o.GasLimit,
		BasePath:                            o.Path,
		TransactionBasePath:                 fmt.Sprintf("%s/%s", o.Path, o.TransactionFolderName),
		ScriptBasePath:                      fmt.Sprintf("%s/%s", o.Path, o.ScriptFolderName),
		Log:                                 &memlog,
		EmulatorLog:                         &emulatorLog,
		FilterOutFeeEvents:                  o.FilterOutFeeEvents,
		FilterOutEmptyWithDrawDepositEvents: o.FilterOutEmptyWithDrawDepositEvents,
		GlobalEventFilter:                   o.GlobalEventFilter,
	}

	if o.DeployContracts {
		overflow = overflow.InitializeContracts()
	}

	if o.InitializeAccounts {
		o2, err := overflow.CreateAccountsE()
		return o2, err
	}
	return overflow, nil
}

//NewOverflowInMemoryEmulator this method is used to create an in memory emulator, deploy all contracts for the emulator and create all accounts
//Deprecated use Overflow function with builder
func NewOverflowInMemoryEmulator() *OverflowBuilder {
	return NewOverflowBuilder("embedded", true, output.InfoLog)
}

//NewOverflowForNetwork creates a new overflow client for the provided network
//Deprecated use Overflow function with builder
func NewOverflowForNetwork(network string) *OverflowBuilder {
	return NewOverflowBuilder(network, false, output.InfoLog)
}

//NewOverflowEmulator create a new client
//Deprecated use Overflow function with builder
func NewOverflowEmulator() *OverflowBuilder {
	return NewOverflowBuilder("emulator", false, output.InfoLog)
}

//NewTestingEmulator starts an embeded emulator with no log to be used most often in tests
//Deprecated use Overflow function with builder
func NewTestingEmulator() *OverflowBuilder {
	return NewOverflowBuilder("embedded", true, 0)
}

//NewOverflowTestnet creates a new overflow client for devnet/testnet
//Deprecated use Overflow function with builder
func NewOverflowTestnet() *OverflowBuilder {
	return NewOverflowBuilder("testnet", false, output.InfoLog)
}

//NewOverflowMainnet creates a new gwft client for mainnet
//Deprecated use Overflow function with builder
func NewOverflowMainnet() *OverflowBuilder {
	return NewOverflowBuilder("mainnet", false, output.InfoLog)
}

//OverflowOption and option function that you can send in to configure Overflow
type OverflowOption func(*OverflowBuilder)

//applyOptions will apply all options from the sent in slice to an overflow builder
func (o *OverflowBuilder) applyOptions(opts []OverflowOption) *OverflowBuilder {
	for _, opt := range opts {
		opt(o)
	}

	return o
}

/*
 Overflow will start an  Overflow instance that panics if there are initialization errors

 Will read the following ENV vars as default:
  OVERFLOW_ENV : set to "mainnet|testnet|emulator|embedded", default embedded
	OVERFLOW_LOGGING: set from 0-3 to get increasing amount of log output, default 3
	OVERFLOW_CONTINUE: to continue this overflow on an already running emulator., default false

 Starting overflow without env vars will make it start in embedded mode deploying all contracts creating accounts

 You can then chose to override this setting with the builder methods example
 `
  Overflow(WithNetwork("mainnet"))
 `

 Setting the network in this way will reset other builder methods if appropriate so use with care.

*/
func Overflow(opts ...OverflowOption) *OverflowState {
	o, err := NewOverflow().applyOptions(opts).StartE()
	if err != nil {
		panic(err)
	}
	return o
}

// OverfloewE will start overflow and return state or an error if there is one
// See Overflow doc comment for an better docs
func OverflowE(opts ...OverflowOption) (*OverflowState, error) {
	return NewOverflow().applyOptions(opts).StartE()
}

//OverflowTesting starts an overflow emulator that is suitable for testing that will print no logs to stdout
func OverflowTesting(opts ...OverflowOption) (*OverflowState, error) {
	return NewOverflowBuilder("embedded", true, 0).applyOptions(opts).StartE()
}

//WithNetwork will start overflow with the given network.
//This function will also set up other options that are common for a given Network.
// - embedded: starts in memory, will deploy contracts and create accounts, info log
// - testing: starts in memory, will deploy contracts and create accounts, no log
// - testnet|mainnet: will only set network, not deploy contracts or create accounts
func WithNetwork(network string) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {

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

//WithInMemory will set that this instance is an in memoy instance createing accounts/deploying contracts
func WithInMemory() func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.InMemory = true
		o.DeployContracts = true
		o.InitializeAccounts = true
		o.Network = "emulator"
	}
}

//WithExistingEmulator will attach to an existing emulator, not deploying contracts and creating accounts
func WithExistingEmulator() func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.DeployContracts = false
		o.InitializeAccounts = false
		o.InMemory = false
		o.Network = "emulator"
	}
}

//DoNotPrependNetworkToAccountNames will not prepend the name of the network to account names
func DoNotPrependNetworkToAccountNames() func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.PrependNetworkName = false
	}
}

//WithServiceAccountSuffix will set the suffix of the service account
func WithServiceAccountSuffix(suffix string) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.ServiceSuffix = suffix
	}
}

//WithNoLog will start emulator with no output
func WithNoLog() func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.LogLevel = output.NoneLog
	}
}

//WithGas set the default gas limit, standard is 9999 (max)
func WithGas(gas int) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.GasLimit = gas
	}
}

//WithBasePath will change the standard basepath `.` to another folder
func WithBasePath(path string) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.Path = path
	}
}

//WithFlowConfig will set the path to one or more flow.json config files
//The default is ~/flow.json and ./flow.json.
func WithFlowConfig(files ...string) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.ConfigFiles = files
	}
}

//WithScriptFolderName will overwite the default script subdir for scripts `scripts`
func WithScriptFolderName(name string) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.ScriptFolderName = name
	}
}

//WithTransactionFolderName will overwite the default script subdir for transactions `transactions`
func WithTransactionFolderName(name string) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.TransactionFolderName = name
	}
}

//WithTransactionFolderName will overwite the default script subdir for transactions `transactions`
func WithFeesEvents() func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.FilterOutFeeEvents = false
	}
}

func WithEmptyDepoitWithdrawEvents() func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.FilterOutEmptyWithDrawDepositEvents = false
	}
}

func WithGlobalEventTilter(filter OverFlowEventFilter) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.GlobalEventFilter = filter
	}
}
