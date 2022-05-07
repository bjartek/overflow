package overflow

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/enescakir/emoji"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Overflow Entire configuration to work with Go With the Flow
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
}

func (o *Overflow) ServiceAccountName() string {
	if o.PrependNetworkToAccountNames {
		return fmt.Sprintf("%s-%s", o.Network, o.ServiceAccountSuffix)
	}
	return o.ServiceAccountSuffix
}

//Account fetch an account from flow.json, prefixing the name with network- as default (can be turned off)
func (f *Overflow) Account(key string) *flowkit.Account {
	if f.PrependNetworkToAccountNames {
		key = fmt.Sprintf("%s-%s", f.Network, key)
	}

	account, err := f.State.Accounts().ByName(key)
	if err != nil {
		log.Fatal(err)
	}

	return account

}

type OverflowBuilder struct {
	Network            string
	InMemory           bool
	DeployContracts    bool
	GasLimit           int
	Path               string
	LogLevel           int
	InitializeAccounts bool
	PrependNetworkName bool
	ServiceSuffix      string
	ConfigFiles        []string
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
	if network == "" {
		network = "embedded"
	}

	inMemory := false
	deployContracts := newEmulator
	initializeAccounts := newEmulator

	if network == "embedded" {
		inMemory = true
		network = "emulator"
	}

	if network == "emulator" {
		deployContracts = true
		initializeAccounts = true
	}

	return &OverflowBuilder{
		Network:            network,
		InMemory:           inMemory,
		DeployContracts:    deployContracts,
		GasLimit:           9999,
		Path:               ".",
		LogLevel:           logLevel,
		InitializeAccounts: initializeAccounts,
		PrependNetworkName: true,
		ServiceSuffix:      "account",
		ConfigFiles:        config.DefaultPaths(),
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
		log.Fatalf("%v error %+v", emoji.PileOfPoo, err)
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
	if o.InMemory {
		//YAY we can run it inline in memory!
		acc, _ := state.EmulatorServiceAccount()

		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetLevel(logrus.TraceLevel)
		logrus.SetOutput(&memlog)
		gw := gateway.NewEmulatorGatewayWithLogger(logrus.StandardLogger(), acc)
		//		gw := gateway.NewEmulatorGateway(acc)
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
		Log:                          &memlog,
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
