package overflow

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/enescakir/emoji"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/rs/zerolog"
	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Overflow Entire configuration to work with Go With the Flow
//TODO rename this to OverflowState
type Overflow struct {
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

func (o *Overflow) ServiceAccountName() string {
	if o.PrependNetworkToAccountNames {
		return fmt.Sprintf("%s-%s", o.Network, o.ServiceAccountSuffix)
	}
	return o.ServiceAccountSuffix
}

//Account fetch an account from flow.json, prefixing the name with network- as default (can be turned off)
func (f *Overflow) AccountE(key string) (*flowkit.Account, error) {
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
func (ob *OverflowBuilder) Start() *Overflow {
	o, err := ob.StartE()
	if err != nil {
		panic(fmt.Sprintf("%v error %+v", emoji.PileOfPoo, err))
	}
	return o
}

func (o *OverflowBuilder) StartE() (*Overflow, error) {

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
	overflow := &Overflow{
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

type Meter struct {
	LedgerInteractionUsed  int                           `json:"ledgerInteractionUsed"`
	ComputationUsed        int                           `json:"computationUsed"`
	MemoryUsed             int                           `json:"memoryUsed"`
	ComputationIntensities MeteredComputationIntensities `json:"computationIntensities"`
	MemoryIntensities      MeteredMemoryIntensities      `json:"memoryIntensities"`
}

func (m Meter) FunctionInvocations() int {
	return int(m.ComputationIntensities[common.ComputationKindFunctionInvocation])
}

func (m Meter) Loops() int {
	return int(m.ComputationIntensities[common.ComputationKindLoop])
}

func (m Meter) Statements() int {
	return int(m.ComputationIntensities[common.ComputationKindStatement])
}

type MeteredComputationIntensities map[common.ComputationKind]uint
type MeteredMemoryIntensities map[common.MemoryKind]uint
