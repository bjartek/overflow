// MIT licensed

// Overflow is a DSL to help interact with the flow blockchain using go
//
// By bjartek aka Bjarte Karlsen
package overflow

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/onflow/cadence/runtime"
	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/gateway"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-emulator/emulator"
	"github.com/onflow/flow-go/fvm/blueprints"
	fm "github.com/onflow/flow-go/model/flow"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/spf13/afero"
)

const emulatorValue = "emulator"

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
// # Starting overflow without env vars will make it start in embedded mode deploying all contracts creating accounts
//
// You can then chose to override this setting with the builder methods example
//
//	Overflow(WithNetwork("mainnet"))
func Overflow(opts ...OverflowOption) *OverflowState {
	ob := defaultOverflowBuilder.applyOptions(opts)
	o := ob.StartResult()

	if o.Error != nil && o.StopOnError {
		panic(o.Error)
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
	PrintOptions:                        &[]OverflowPrinterOption{},
	NewAccountFlowAmount:                10.0,
	TransactionFees:                     true,
	Coverage:                            nil,
}

// OverflowBuilder is the struct used to gather up configuration when building an overflow instance
type OverflowBuilder struct {
	Ctx                                 context.Context
	TransactionFees                     bool
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
	PrintOptions                        *[]OverflowPrinterOption
	NewAccountFlowAmount                float64
	ReaderWriter                        flowkit.ReaderWriter
	InputResolver                       *InputResolver
	ArchiveNodeUrl                      string
	Coverage                            *runtime.CoverageReport
}

func (o *OverflowBuilder) StartE() (*OverflowState, error) {
	result := o.StartResult()
	if result.Error != nil {
		return nil, result.Error
	}
	return result, nil
}

// StartE will start Overflow and return State and error if any
func (o *OverflowBuilder) StartResult() *OverflowState {
	scriptFolderName := fmt.Sprintf("%s/%s", o.Path, o.ScriptFolderName)
	if o.ScriptFolderName == "" {
		scriptFolderName = o.Path
	} else if o.Path == "" {
		scriptFolderName = o.ScriptFolderName
	}

	txPathName := fmt.Sprintf("%s/%s", o.Path, o.TransactionFolderName)
	if o.TransactionFolderName == "" {
		txPathName = o.Path
	} else if o.Path == "" {
		txPathName = o.TransactionFolderName
	}

	overflow := &OverflowState{
		PrependNetworkToAccountNames:        o.PrependNetworkName,
		ServiceAccountSuffix:                o.ServiceSuffix,
		Gas:                                 o.GasLimit,
		BasePath:                            o.Path,
		TransactionBasePath:                 txPathName,
		ScriptBasePath:                      scriptFolderName,
		FilterOutFeeEvents:                  o.FilterOutFeeEvents,
		FilterOutEmptyWithDrawDepositEvents: o.FilterOutEmptyWithDrawDepositEvents,
		GlobalEventFilter:                   o.GlobalEventFilter,
		StopOnError:                         o.StopOnError,
		PrintOptions:                        o.PrintOptions,
		NewUserFlowAmount:                   o.NewAccountFlowAmount,
		LogLevel:                            o.LogLevel,
		CoverageReport:                      o.Coverage,
	}

	loader := o.ReaderWriter
	if o.ReaderWriter == nil {
		loader = afero.Afero{Fs: afero.NewOsFs()}
	}
	var state *flowkit.State
	var err error
	state, err = flowkit.Load(o.ConfigFiles, loader)
	if err != nil {
		overflow.Error = err
		return overflow
	}
	overflow.State = state

	if o.InputResolver != nil {
		overflow.InputResolver = *o.InputResolver
	} else {
		overflow.InputResolver = func(name string) (string, error) {
			return overflow.QualifiedIdentifierFromSnakeCase(name)
		}
	}

	// This is different for testnet and mainnet
	// TODO: fix this for testnet

	chain := fm.Mainnet.Chain()
	if o.Network == "testnet" {
		chain = fm.Testnet.Chain()
	}

	systemChunkTx, err := blueprints.SystemChunkTransaction(chain)
	if err != nil {
		overflow.Error = err
		return overflow
	}
	systemChunkId := systemChunkTx.ID().String()
	overflow.SystemChunkTransactionId = systemChunkId

	network, err := state.Networks().ByName(o.Network)
	if err != nil {
		overflow.Error = err
		return overflow
	}
	overflow.Network = *network

	logger := output.NewStdoutLogger(o.LogLevel)
	overflow.Logger = logger
	var memlog bytes.Buffer
	overflow.Log = &memlog

	if o.InMemory {
		acc, _ := state.EmulatorServiceAccount()

		// this is the emulator log
		logWriter := io.Writer(&memlog)
		emulatorLogger := zerolog.New(logWriter).Level(zerolog.DebugLevel)

		emulatorOptions := []emulator.Option{
			emulator.WithLogger(emulatorLogger),
		}

		if o.TransactionFees {
			emulatorOptions = append(emulatorOptions, emulator.WithTransactionFeesEnabled(true), emulator.WithCoverageReport(o.Coverage))
		}

		pk, _ := acc.Key.PrivateKey()
		emulatorKey := &gateway.EmulatorKey{
			PublicKey: (*pk).PublicKey(),
			SigAlgo:   acc.Key.SigAlgo(),
			HashAlgo:  acc.Key.HashAlgo(),
		}
		gw := gateway.NewEmulatorGatewayWithOpts(emulatorKey,
			gateway.WithLogger(&emulatorLogger),
			gateway.WithEmulatorOptions(emulatorOptions...),
		)

		overflow.EmulatorGatway = gw
		overflow.Flowkit = flowkit.NewFlowkit(state, *network, gw, logger)
	} else {
		gw, err := gateway.NewGrpcGateway(*network)
		if err != nil {
			overflow.Error = err
			return overflow
		}
		overflow.Flowkit = flowkit.NewFlowkit(state, *network, gw, logger)

		/* TODO: fix archive
		if o.ArchiveNodeUrl != "" {
			gw, err := gateway.NewGrpcGateway(o.ArchiveNodeUrl)
			if err != nil {
				overflow.Error = err
				return overflow
			}
			overflow.ArchiveScripts = services.NewScripts(gw, state, logger)
		}
		*/
	}

	if o.InitializeAccounts {
		_, err := overflow.CreateAccountsE(o.Ctx)
		if err != nil {
			overflow.Error = errors.Wrap(err, "could not create accounts")
			return overflow
		}
	}

	if o.DeployContracts {
		overflow = overflow.InitializeContracts(o.Ctx)
		if overflow.Error != nil {
			overflow.Error = errors.Wrap(overflow.Error, "could not deploy contracts")
			return overflow
		}
	}
	return overflow
}

// applyOptions will apply all options from the sent in slice to an overflow builder
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

		case "testnet", "mainnet", "crescendo":
			o.DeployContracts = false
			o.InitializeAccounts = false
			o.StopOnError = false
			o.InMemory = false
			o.PrintOptions = nil
		case emulatorValue:
			o.InMemory = false
		case "testing":
			o.LogLevel = 0
			o.StopOnError = false
			o.PrintOptions = nil
			o.Network = emulatorValue
		default:
			o.Network = emulatorValue
		}
	}
}

// WithExistingEmulator will attach to an existing emulator, not deploying contracts and creating accounts
func WithExistingEmulator() OverflowOption {
	return func(o *OverflowBuilder) {
		o.DeployContracts = false
		o.InitializeAccounts = false
		o.Network = emulatorValue
		o.InMemory = false
	}
}

// DoNotPrependNetworkToAccountNames will not prepend the name of the network to account names
func WithNoPrefixToAccountNames() OverflowOption {
	return func(o *OverflowBuilder) {
		o.PrependNetworkName = false
	}
}

// WithServiceAccountSuffix will set the suffix of the service account
func WithServiceAccountSuffix(suffix string) OverflowOption {
	return func(o *OverflowBuilder) {
		o.ServiceSuffix = suffix
	}
}

func WithLogInfo() OverflowOption {
	return func(ob *OverflowBuilder) {
		ob.LogLevel = output.InfoLog
		ob.PrintOptions = &[]OverflowPrinterOption{}
	}
}

func WithLogFull() OverflowOption {
	return func(ob *OverflowBuilder) {
		ob.LogLevel = output.InfoLog
		ob.PrintOptions = &[]OverflowPrinterOption{WithFullMeter(), WithEmulatorLog()}
	}
}

// WithNoLog will not log anything from results or flowkit logger
func WithLogNone() OverflowOption {
	return func(o *OverflowBuilder) {
		o.LogLevel = output.NoneLog
		o.PrintOptions = nil
	}
}

// WithGas set the default gas limit, standard is 9999 (max)
func WithGas(gas int) OverflowOption {
	return func(o *OverflowBuilder) {
		o.GasLimit = gas
	}
}

// WithBasePath will change the standard basepath `.` to another folder
func WithBasePath(path string) OverflowOption {
	return func(o *OverflowBuilder) {
		o.Path = path
	}
}

// WithFlowConfig will set the path to one or more flow.json config files
// The default is ~/flow.json and ./flow.json.
func WithFlowConfig(files ...string) OverflowOption {
	return func(o *OverflowBuilder) {
		o.ConfigFiles = files
	}
}

// WithScriptFolderName will overwite the default script subdir for scripts `scripts`
func WithScriptFolderName(name string) OverflowOption {
	return func(o *OverflowBuilder) {
		o.ScriptFolderName = name
	}
}

// WithTransactionFolderName will overwite the default script subdir for transactions `transactions`
func WithTransactionFolderName(name string) OverflowOption {
	return func(o *OverflowBuilder) {
		o.TransactionFolderName = name
	}
}

// WithTransactionFolderName will overwite the default script subdir for transactions `transactions`
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

// If this option is used a panic will be called if an error occurs after an interaction is run
func WithPanicOnError() OverflowOption {
	return func(o *OverflowBuilder) {
		o.StopOnError = true
	}
}

// If this option is used a panic will be called if an error occurs after an interaction is run
func WithReturnErrors() OverflowOption {
	return func(o *OverflowBuilder) {
		o.StopOnError = false
	}
}

// automatically print interactions using the following options
func WithGlobalPrintOptions(opts ...OverflowPrinterOption) OverflowOption {
	return func(o *OverflowBuilder) {
		o.PrintOptions = &opts
	}
}

// alias for WithGLobalPrintOptions
func WithPrintResults(opts ...OverflowPrinterOption) OverflowOption {
	return WithGlobalPrintOptions(opts...)
}

// Set the amount of flow for new account, default is 0.001
func WithFlowForNewUsers(amount float64) OverflowOption {
	return func(o *OverflowBuilder) {
		o.NewAccountFlowAmount = amount
	}
}

// Turn off storage fees
func WithoutTransactionFees() OverflowOption {
	return func(o *OverflowBuilder) {
		o.TransactionFees = false
	}
}

func WithEmbedFS(fs embed.FS) OverflowOption {
	return func(o *OverflowBuilder) {
		wrapper := EmbedWrapper{Embed: fs}
		o.ReaderWriter = &wrapper
	}
}

func WithInputResolver(ir InputResolver) OverflowOption {
	return func(o *OverflowBuilder) {
		o.InputResolver = &ir
	}
}

func WithArchiveNodeUrl(url string) OverflowOption {
	return func(o *OverflowBuilder) {
		o.ArchiveNodeUrl = url
	}
}

func WithCoverageReport() OverflowOption {
	return func(o *OverflowBuilder) {
		o.Coverage = runtime.NewCoverageReport()
	}
}

type EmbedWrapper struct {
	Embed embed.FS
}

func (ew *EmbedWrapper) ReadFile(source string) ([]byte, error) {
	return ew.Embed.ReadFile(source)
}

func (ew *EmbedWrapper) WriteFile(filename string, data []byte, perm os.FileMode) error {
	fmt.Printf("Writing file %s is not supported by embed.FS", filename)
	return nil
}

func (ew *EmbedWrapper) MkdirAll(path string, perm os.FileMode) error {
	fmt.Printf("Creating dir is not %s is not supported by embed.FS", path)
	return nil
}
