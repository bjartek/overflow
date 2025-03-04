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
	"io/fs"
	"os"
	"strconv"

	"github.com/bjartek/underflow"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/flixkit-go/v2/flixkit"
	"github.com/onflow/flow-emulator/emulator"
	grpcAccess "github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/gateway"
	"github.com/onflow/flowkit/v2/output"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"google.golang.org/grpc"
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
	UnderflowOptions:                    underflow.Options{},
}

// OverflowBuilder is the struct used to gather up configuration when building an overflow instance
type OverflowBuilder struct {
	Ctx                                 context.Context
	ReaderWriter                        flowkit.ReaderWriter
	Coverage                            *runtime.CoverageReport
	InputResolver                       *underflow.InputResolver
	PrintOptions                        *[]OverflowPrinterOption
	GlobalEventFilter                   OverflowEventFilter
	Path                                string
	NetworkHost                         string
	Network                             string
	ScriptFolderName                    string
	ServiceSuffix                       string
	TransactionFolderName               string
	EmulatorOptions                     []emulator.Option
	GrpcDialOptions                     []grpc.DialOption
	ConfigFiles                         []string
	NewAccountFlowAmount                float64
	GasLimit                            int
	LogLevel                            int
	UnderflowOptions                    underflow.Options
	DeployContracts                     bool
	InMemory                            bool
	InitializeAccounts                  bool
	StopOnError                         bool
	TransactionFees                     bool
	FilterOutEmptyWithDrawDepositEvents bool
	FilterOutFeeEvents                  bool
	PrependNetworkName                  bool
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
		UnderflowOptions:                    o.UnderflowOptions,
	}

	loader := o.ReaderWriter
	if o.ReaderWriter == nil {
		loader = afero.Afero{Fs: afero.NewOsFs()}
	}
	var state *flowkit.State
	var err error
	state, err = flowkit.Load(o.ConfigFiles, loader)
	if err != nil {
		overflow.Error = errors.Wrapf(err, "could not find flow configuration")
		return overflow
	}
	overflow.State = state

	overflow.Flixkit = flixkit.NewFlixService(&flixkit.FlixServiceConfig{
		FileReader: state,
	})

	if o.InputResolver != nil {
		overflow.InputResolver = *o.InputResolver
	} else {
		overflow.InputResolver = func(name string, resolveType underflow.ResolveType) (string, error) {
			if resolveType == underflow.Identifier {
				return overflow.QualifiedIdentifierFromSnakeCase(name)
			}

			adr, err2 := hexToAddress(name)
			if err2 == nil {
				return adr.String(), nil
			}

			address, err2 := overflow.FlowAddressE(name)
			if err2 != nil {
				return "", errors.Wrapf(err2, "could not parse %s into an address", name)
			}
			return address.HexWithPrefix(), nil
		}
	}
	network, err := state.Networks().ByName(o.Network)
	if err != nil {
		overflow.Error = err
		return overflow
	}
	if o.NetworkHost != "" {
		network.Host = o.NetworkHost
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

		emulatorOptions = append(emulatorOptions, o.EmulatorOptions...)

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
		clientOpts := grpcAccess.WithGRPCDialOptions(o.GrpcDialOptions...)
		gw, err := gateway.NewGrpcGateway(*network, clientOpts)
		if err != nil {
			overflow.Error = err
			return overflow
		}
		overflow.Flowkit = flowkit.NewFlowkit(state, *network, gw, logger)
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

		case "testnet", "mainnet", "crescendo", "previewnet", "migrationnet":
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

func WithInputResolver(ir underflow.InputResolver) OverflowOption {
	return func(o *OverflowBuilder) {
		o.InputResolver = &ir
	}
}

func WithGrpcDialOption(opt ...grpc.DialOption) OverflowOption {
	return func(o *OverflowBuilder) {
		o.GrpcDialOptions = append(o.GrpcDialOptions, opt...)
	}
}

func WithCoverageReport() OverflowOption {
	return func(o *OverflowBuilder) {
		o.Coverage = runtime.NewCoverageReport()
	}
}

func WithEmulatorOption(opt ...emulator.Option) OverflowOption {
	return func(o *OverflowBuilder) {
		o.EmulatorOptions = append(o.EmulatorOptions, opt...)
	}
}

// Set custom network host if different from the one in flow.json since we cannot env substs there
func WithNetworkHost(host string) OverflowOption {
	return func(o *OverflowBuilder) {
		o.NetworkHost = host
	}
}

func WithUnderflowOptions(opt underflow.Options) OverflowOption {
	return func(o *OverflowBuilder) {
		o.UnderflowOptions = opt
	}
}

type EmbedWrapper struct {
	Embed embed.FS
}

func (ew *EmbedWrapper) ReadFile(source string) ([]byte, error) {
	return ew.Embed.ReadFile(source)
}

func (ew *EmbedWrapper) MkdirAll(path string, perm os.FileMode) error {
	fmt.Printf("Writing dirs %s is not supported by embed.FS", path)
	return nil
}

func (ew *EmbedWrapper) WriteFile(filename string, data []byte, perm os.FileMode) error {
	fmt.Printf("Writing file %s is not supported by embed.FS", filename)
	return nil
}

func (ew *EmbedWrapper) Stat(string) (fs.FileInfo, error) {
	return nil, nil
}
