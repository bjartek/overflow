package overflow

import (
	"fmt"
	"log"

	"github.com/enescakir/emoji"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// Overflow Entire configuration to work with Go With the Flow
type Overflow struct {
	State                        *flowkit.State
	Services                     *services.Services
	Network                      string
	Logger                       output.Logger
	PrependNetworkToAccountNames bool
}

/* Type
  - mainnet
	- devnet
	- emulator
	- embedded

  Logs:

	Contracts:true/false
	Accounts:true/false
	BasePath: this will change default path for scripts and transactons
	GasLimit: standard gas limit
	SericeAccountSuffix
	PrependNetworkToAccountNames
*/

type OverflowConfig struct {
	Network         string
	DeployContracts *bool
	GasLimit        int
	BasePath        string
	LogLevel        int
	Accounts        struct {
		Initialize         bool
		PrependNetworkName bool
		ServiceSuffix      string
	}
}

func LoadConfig(file string) (config OverflowConfig, err error) {

	viper.SetDefault("network", "emulator")
	viper.SetDefault("gasLimit", 9999)
	viper.SetDefault("basePath", ".")
	viper.SetDefault("accounts.initialize", true)
	viper.SetDefault("accounts.prependNetworkName", true)
	viper.SetDefault("accounts.serviceSuffix", "account")

	viper.AddConfigPath("$HOME/.overflow")
	viper.AddConfigPath("$XDB_CONFIG_HOME/.overflow")
	viper.AddConfigPath(".")

	viper.SetConfigName("overflow")
	viper.SetConfigType("yaml")

	if file != "" {
		viper.SetConfigFile(file)
	}

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

//NewOverflowInMemoryEmulator this method is used to create an in memory emulator, deploy all contracts for the emulator and create all accounts
func NewOverflowInMemoryEmulator() *Overflow {
	return NewOverflow(config.DefaultPaths(), "emulator", true, output.InfoLog).InitializeContracts().CreateAccounts("emulator-account")
}

//NewTestingEmulator create new emulator that ignore all log messages
func NewTestingEmulator() *Overflow {
	return NewOverflow(config.DefaultPaths(), "emulator", true, output.NoneLog).InitializeContracts().CreateAccounts("emulator-account")
}

//NewOverflowForNetwork creates a new overflow client for the provided network
func NewOverflowForNetwork(network string) *Overflow {
	return NewOverflow(config.DefaultPaths(), network, false, output.InfoLog)

}

//NewOverflowEmulator create a new client
func NewOverflowEmulator() *Overflow {
	return NewOverflow(config.DefaultPaths(), "emulator", false, output.InfoLog)
}

//NewOverflowTestnet creates a new overflow client for devnet/testnet
func NewOverflowTestnet() *Overflow {
	return NewOverflow(config.DefaultPaths(), "testnet", false, output.InfoLog)
}

//NewOverflowMainnet creates a new gwft client for mainnet
func NewOverflowMainnet() *Overflow {
	return NewOverflow(config.DefaultPaths(), "mainnet", false, output.InfoLog)
}

// NewOverflow with custom file panic on error
func NewOverflow(filenames []string, network string, inMemory bool, loglevel int) *Overflow {
	overflow, err := NewGoWithTheFlowError(filenames, network, inMemory, loglevel)
	if err != nil {
		log.Fatalf("%v error %+v", emoji.PileOfPoo, err)
	}
	return overflow
}

//DoNotPrependNetworkToAccountNames disable the default behavior of prefixing account names with network-
func (f *Overflow) DoNotPrependNetworkToAccountNames() *Overflow {
	f.PrependNetworkToAccountNames = false
	return f
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

// NewGoWithTheFlowError creates a new local go with the flow client
func NewGoWithTheFlowError(paths []string, network string, inMemory bool, logLevel int) (*Overflow, error) {

	loader := &afero.Afero{Fs: afero.NewOsFs()}
	state, err := flowkit.Load(paths, loader)
	if err != nil {
		return nil, err
	}

	logger := output.NewStdoutLogger(logLevel)
	var service *services.Services
	if inMemory {
		//YAY we can run it inline in memory!
		acc, _ := state.EmulatorServiceAccount()
		gw := gateway.NewEmulatorGateway(acc)
		service = services.NewServices(gw, state, logger)
	} else {
		network, err := state.Networks().ByName(network)
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
	return &Overflow{
		State:                        state,
		Services:                     service,
		Network:                      network,
		Logger:                       logger,
		PrependNetworkToAccountNames: true,
	}, nil

}
