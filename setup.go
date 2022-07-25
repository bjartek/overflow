// MIT licensed

// Overflow is a DSL to help interact with the flow blockchain using go
//
// By bjartek aka Bjarte Karlsen
package overflow

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	emulator "github.com/onflow/flow-emulator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// OverflowOption and option function that you can send in to configure Overflow
type OverflowOption func(*OverflowBuilder)

// Overflow will start an  Overflow instance that panics if there are initialization errors
//
// Will read the following ENV vars as default:
// OVERFLOW_ENV : set to "mainnet|testnet|emulator|embedded", default embedded
// OVERFLOW_LOGGING: set from 0-3. 0 is silent, 1 is print terse output, 2 is print output from flowkit, 3 is all lots we can
// OVERFLOW_CONTINUE: to continue this overflow on an already running emulator., default false
// OVERFLOW_STOP_ON_ERROR: will the process panic if an erorr is encountered. If set to false the result objects will have the error. default: false
//
// Starting overflow without env vars will make it start in embedded mode deploying all contracts creating accounts
//
// You can then chose to override this setting with the builder methods example
//
//  Overflow(WithNetwork("mainnet"))
func Overflow(opts ...OverflowOption) *OverflowState {
	ob := defaultOverflowBuilder.applyOptions(opts)
	o, err := ob.StartE()

	if err != nil {
		if o.StopOnError {
			panic(err)
		}
		o.Error = err
	}

	return o
}

// The default overflow builder settings
var defaultOverflowBuilder = OverflowBuilder{
	InMemory:                            true,
	DeployContracts:                     true,
	GasLimit:                            9999,
	Path:                                ".",
	TransactionFolderName:               "transactions",
	ScriptFolderName:                    "scripts",
	LogLevel:                            output.NoneLog,
	InitializeAccounts:                  true,
	PrependNetworkName:                  true,
	ServiceSuffix:                       "account",
	ConfigFiles:                         config.DefaultPaths(),
	FilterOutEmptyWithDrawDepositEvents: true,
	FilterOutFeeEvents:                  true,
	GlobalEventFilter:                   OverflowEventFilter{},
	StopOnError:                         true,
	PrintOptions:                        &[]PrinterOption{},
	NewAccountFlowAmount:                10.0,
}

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
	GlobalEventFilter                   OverflowEventFilter
	StopOnError                         bool
	PrintOptions                        *[]PrinterOption
	NewAccountFlowAmount                float64
}

// StartE will start Overflow and return State and error if any
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
		acc, _ := state.EmulatorServiceAccount()

		logrusLogger := &logrus.Logger{
			Formatter: &logrus.JSONFormatter{},
			Level:     logrus.TraceLevel,
			Out:       &memlog,
		}

		writer := io.Writer(&emulatorLog)
		emulatorLogger := zerolog.New(writer).Level(zerolog.DebugLevel)
		gw := gateway.NewEmulatorGatewayWithOpts(acc,
			gateway.WithLogger(logrusLogger),
			gateway.WithEmulatorOptions(
				emulator.WithTransactionFeesEnabled(true),
				emulator.WithLogger(emulatorLogger),
			))

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
		StopOnError:                         o.StopOnError,
		PrintOptions:                        o.PrintOptions,
		NewUserFlowAmount:                   o.NewAccountFlowAmount,
		LogLevel:                            o.LogLevel,
	}

	if o.DeployContracts {
		overflow = overflow.InitializeContracts()
		if overflow.Error != nil {
			return overflow, errors.Wrap(overflow.Error, "could not deploy contracts")
		}
	}

	if o.InitializeAccounts {
		o2, err := overflow.CreateAccountsE()
		if err != nil {
			return o2, errors.Wrap(err, "could not create accounts")
		}
	}

	return overflow, nil
}

//applyOptions will apply all options from the sent in slice to an overflow builder
func (o OverflowBuilder) applyOptions(opts []OverflowOption) *OverflowBuilder {

	network := os.Getenv("OVERFLOW_ENV")
	existing := os.Getenv("OVERFLOW_CONTINUE")
	loglevel := os.Getenv("OVERFLOW_LOGGING")
	stopOnError := os.Getenv("OVERFLOW_STOP_ON_ERROR")

	allOpts := []OverflowOption{}

	if stopOnError == "true" {
		allOpts = append(allOpts, WithPanicOnError())
	}

	if loglevel != "" {
		log, err := strconv.Atoi(loglevel)
		if err != nil {
			panic(err)
		}
		if log == 1 {
			allOpts = append(allOpts, WithLogInfo())
		}
		if log == 2 {
			allOpts = append(allOpts, WithLogFull())
		}
	}

	allOpts = append(allOpts, WithNetwork(network))

	if existing != "" {
		allOpts = append(allOpts, WithExistingEmulator())

	}

	allOpts = append(allOpts, opts...)

	ob := &o
	for _, opt := range allOpts {
		opt(ob)
	}

	return ob
}

func OverflowTesting(opts ...OverflowOption) (*OverflowState, error) {
	allOpts := []OverflowOption{WithNetwork("testing")}
	allOpts = append(allOpts, opts...)
	o := Overflow(allOpts...)
	if o.Error != nil {
		return nil, o.Error
	}
	return o, nil
}

// WithNetwork will start overflow with the given network.
// This function will also set up other options that are common for a given Network.
// embedded: starts in memory, will deploy contracts and create accounts, will panic on errors and show terse output
// testing: as embedeed, but will not stop on errors and turn off all logs
// emulator: will connect to running local emulator, deploy contracts and create account
// testnet|mainnet: will only set network, not deploy contracts or create accounts
func WithNetwork(network string) OverflowOption {
	return func(o *OverflowBuilder) {
		o.Network = network
		switch network {

		case "testnet", "mainnet":
			o.DeployContracts = false
			o.InitializeAccounts = false
			o.InMemory = false
		case "emulator":
			o.InMemory = false
		case "testing":
			o.LogLevel = 0
			o.StopOnError = false
			o.PrintOptions = nil
			o.Network = "emulator"
		default:
			o.Network = "emulator"
		}
	}
}

//WithExistingEmulator will attach to an existing emulator, not deploying contracts and creating accounts
func WithExistingEmulator() OverflowOption {
	return func(o *OverflowBuilder) {
		o.DeployContracts = false
		o.InitializeAccounts = false
		o.Network = "emulator"
		o.InMemory = false
	}
}

//DoNotPrependNetworkToAccountNames will not prepend the name of the network to account names
func WithNoPrefixToAccountNames() OverflowOption {
	return func(o *OverflowBuilder) {
		o.PrependNetworkName = false
	}
}

//WithServiceAccountSuffix will set the suffix of the service account
func WithServiceAccountSuffix(suffix string) OverflowOption {
	return func(o *OverflowBuilder) {
		o.ServiceSuffix = suffix
	}
}

func WithLogInfo() OverflowOption {
	return func(ob *OverflowBuilder) {
		ob.LogLevel = output.InfoLog
		ob.PrintOptions = &[]PrinterOption{}
	}
}

func WithLogFull() OverflowOption {
	return func(ob *OverflowBuilder) {
		ob.LogLevel = output.InfoLog
		ob.PrintOptions = &[]PrinterOption{WithFullMeter(), WithEmulatorLog()}
	}
}

//WithNoLog will not log anything from results or flowkit logger
func WithLogNone() OverflowOption {
	return func(o *OverflowBuilder) {
		o.LogLevel = output.NoneLog
		o.PrintOptions = nil
	}
}

//WithGas set the default gas limit, standard is 9999 (max)
func WithGas(gas int) OverflowOption {
	return func(o *OverflowBuilder) {
		o.GasLimit = gas
	}
}

//WithBasePath will change the standard basepath `.` to another folder
func WithBasePath(path string) OverflowOption {
	return func(o *OverflowBuilder) {
		o.Path = path
	}
}

//WithFlowConfig will set the path to one or more flow.json config files
//The default is ~/flow.json and ./flow.json.
func WithFlowConfig(files ...string) OverflowOption {
	return func(o *OverflowBuilder) {
		o.ConfigFiles = files
	}
}

//WithScriptFolderName will overwite the default script subdir for scripts `scripts`
func WithScriptFolderName(name string) OverflowOption {
	return func(o *OverflowBuilder) {
		o.ScriptFolderName = name
	}
}

//WithTransactionFolderName will overwite the default script subdir for transactions `transactions`
func WithTransactionFolderName(name string) OverflowOption {
	return func(o *OverflowBuilder) {
		o.TransactionFolderName = name
	}
}

//WithTransactionFolderName will overwite the default script subdir for transactions `transactions`
func WithFeesEvents() OverflowOption {
	return func(o *OverflowBuilder) {
		o.FilterOutFeeEvents = false
	}
}

// filter out empty deposit and withdraw events
func WithEmptyDepositWithdrawEvents() OverflowOption {
	return func(o *OverflowBuilder) {
		o.FilterOutEmptyWithDrawDepositEvents = false
	}
}

// set global filters to events
func WithGlobalEventFilter(filter OverflowEventFilter) OverflowOption {
	return func(o *OverflowBuilder) {
		o.GlobalEventFilter = filter
	}
}

//If this option is used a panic will be called if an error occurs after an interaction is run
func WithPanicOnError() OverflowOption {
	return func(o *OverflowBuilder) {
		o.StopOnError = true
	}
}

// automatically print interactions using the following options
func WithGlobalPrintOptions(opts ...PrinterOption) OverflowOption {
	return func(o *OverflowBuilder) {
		o.PrintOptions = &opts
	}
}

// Set the amount of flow for new account, default is 0.001
func WithFlowForNewUsers(amount float64) OverflowOption {
	return func(o *OverflowBuilder) {
		o.NewAccountFlowAmount = amount
	}
}
