package overflow

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

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

/// The main overflow struct that we add methods to to interact with overflow
type OverflowState struct {
	State                        *flowkit.State
	Services                     *services.Services
	Network                      string
	Logger                       output.Logger
	PrependNetworkToAccountNames bool
	ServiceAccountSuffix         string
	Gas                          int
	BasePath                     string
	Log                          *bytes.Buffer
	Error                        error
	TransactionBasePath          string
	ScriptBasePath               string
	EmulatorLog                  *bytes.Buffer

	//TODO: add config on what events to skip, like skip fees or empty deposit/withdraw
}

func (o *OverflowState) ServiceAccountName() string {
	if o.PrependNetworkToAccountNames {
		return fmt.Sprintf("%s-%s", o.Network, o.ServiceAccountSuffix)
	}
	return o.ServiceAccountSuffix
}

//Account fetch an account from flow.json, prefixing the name with network- as default (can be turned off)
func (f *OverflowState) AccountE(key string) (*flowkit.Account, error) {
	if f.PrependNetworkToAccountNames {
		key = fmt.Sprintf("%s-%s", f.Network, key)
	}

	account, err := f.State.Accounts().ByName(key)
	if err != nil {
		return nil, err
	}

	return account, nil

}

type OverflowBuilder struct {
	Network               string
	InMemory              bool
	DeployContracts       bool
	GasLimit              int
	Path                  string
	LogLevel              int
	InitializeAccounts    bool
	PrependNetworkName    bool
	ServiceSuffix         string
	ConfigFiles           []string
	TransactionFolderName string
	ScriptFolderName      string
}

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
		Network:               network,
		InMemory:              inMemory,
		DeployContracts:       deployContracts,
		GasLimit:              9999,
		Path:                  ".",
		TransactionFolderName: "transactions",
		ScriptFolderName:      "scripts",
		LogLevel:              logLevel,
		InitializeAccounts:    initializeAccounts,
		PrependNetworkName:    true,
		ServiceSuffix:         "account",
		ConfigFiles:           config.DefaultPaths(),
	}
}

//Set this if you are using an existing emulator and you do not want to create contracts or initializeAccounts
func (o *OverflowBuilder) ExistingEmulator() *OverflowBuilder {
	o.DeployContracts = false
	o.InitializeAccounts = false
	return o
}

func (o *OverflowBuilder) DoNotPrependNetworkToAccountNames() *OverflowBuilder {
	o.PrependNetworkName = false
	return o
}

func (o *OverflowBuilder) SetServiceSuffix(suffix string) *OverflowBuilder {
	o.ServiceSuffix = suffix
	return o
}

func (o *OverflowBuilder) NoneLog() *OverflowBuilder {
	o.LogLevel = output.NoneLog
	return o
}

func (o *OverflowBuilder) DefaultGas(gas int) *OverflowBuilder {
	o.GasLimit = gas
	return o
}

func (o *OverflowBuilder) BasePath(path string) *OverflowBuilder {
	o.Path = path
	return o
}

func (o *OverflowBuilder) Config(files ...string) *OverflowBuilder {
	o.ConfigFiles = files
	return o
}

// NewOverflow with custom file panic on error
func (ob *OverflowBuilder) Start() *OverflowState {
	o, err := ob.StartE()
	if err != nil {
		panic(fmt.Sprintf("%v error %+v", emoji.PileOfPoo, err))
	}
	return o
}

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
		State:                        state,
		Services:                     service,
		Network:                      o.Network,
		Logger:                       logger,
		PrependNetworkToAccountNames: o.PrependNetworkName,
		ServiceAccountSuffix:         o.ServiceSuffix,
		Gas:                          o.GasLimit,
		BasePath:                     o.Path,
		TransactionBasePath:          fmt.Sprintf("%s/%s", o.Path, o.TransactionFolderName),
		ScriptBasePath:               fmt.Sprintf("%s/%s", o.Path, o.ScriptFolderName),
		Log:                          &memlog,
		EmulatorLog:                  &emulatorLog,
		//TODO; what events do you want to skip by default
		//TODO: remove fees
		//TODO: remove empty deposit/withdraw events
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
func NewOverflowInMemoryEmulator() *OverflowBuilder {
	return NewOverflowBuilder("embedded", true, output.InfoLog)
}

//NewOverflowForNetwork creates a new overflow client for the provided network
func NewOverflowForNetwork(network string) *OverflowBuilder {
	return NewOverflowBuilder(network, false, output.InfoLog)
}

//NewOverflowEmulator create a new client
func NewOverflowEmulator() *OverflowBuilder {
	return NewOverflowBuilder("emulator", false, output.InfoLog)
}

func NewTestingEmulator() *OverflowBuilder {
	return NewOverflowBuilder("embedded", true, 0)
}

//NewOverflowTestnet creates a new overflow client for devnet/testnet
func NewOverflowTestnet() *OverflowBuilder {
	return NewOverflowBuilder("testnet", false, output.InfoLog)
}

//NewOverflowMainnet creates a new gwft client for mainnet
func NewOverflowMainnet() *OverflowBuilder {
	return NewOverflowBuilder("mainnet", false, output.InfoLog)
}

type LogrusMessage struct {
	ComputationUsed int       `json:"computationUsed"`
	Level           string    `json:"level"`
	Msg             string    `json:"msg"`
	Time            time.Time `json:"time"`
	TxID            string    `json:"txID"`
}

type OverflowOption func(*OverflowBuilder)

func (o *OverflowBuilder) ApplyOptions(opts []OverflowOption) *OverflowBuilder {
	for _, opt := range opts {
		opt(o)
	}

	return o
}

/*
 Start a new Overflow instance that panics if there are initialization errors

 Will read the following ENV vars as default:
  OVERFLOW_ENV : set to "mainnet|testnet|emulator|embedded"
	OVERFLOW_LOGGING: set from 0-4 to get increasing amount of log output
	OVERFLOW_CONTINUE: to continue this overflow on an already running emulator.

 You can then chose to override this setting with the builder methods example
 `
  Overflow(WithNetwork("mainnet"))
 `

 Setting the network in this way will reset other builder methods if appropriate so use with care.

*/
func Overflow(opts ...OverflowOption) *OverflowState {
	o, err := NewOverflow().ApplyOptions(opts).StartE()
	if err != nil {
		panic(err)
	}
	return o
}

func OverflowE(opts ...OverflowOption) (*OverflowState, error) {
	return NewOverflow().ApplyOptions(opts).StartE()
}

/*
	Can be used to start an overflow instance that can be used in tests
*/
func OverflowTesting(opts ...OverflowOption) (*OverflowState, error) {
	return NewOverflowBuilder("embedded", true, 0).ApplyOptions(opts).StartE()
}

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

func WithInMemory() func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.InMemory = true
		o.DeployContracts = true
		o.InitializeAccounts = true
		o.Network = "emulator"
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

func WithScriptFolderName(name string) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.ScriptFolderName = name
	}
}

func WithTransactionFolderName(name string) func(o *OverflowBuilder) {
	return func(o *OverflowBuilder) {
		o.TransactionFolderName = name
	}
}
