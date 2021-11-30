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
)

// GoWithTheFlow Entire configuration to work with Go With the Flow
type GoWithTheFlow struct {
	State                        *flowkit.State
	Services                     *services.Services
	Network                      string
	Logger                       output.Logger
	PrependNetworkToAccountNames bool
}

//NewGoWithTheFlowInMemoryEmulator this method is used to create an in memory emulator, deploy all contracts for the emulator and create all accounts
func NewGoWithTheFlowInMemoryEmulator() *GoWithTheFlow {
	return NewGoWithTheFlow(config.DefaultPaths(), "emulator", true, output.InfoLog).InitializeContracts().CreateAccounts("emulator-account")
}

//NewTEstingEmulator create new emulator that ignore all log messages
func NewTestingEmulator() *GoWithTheFlow {
	return NewGoWithTheFlow(config.DefaultPaths(), "emulator", true, output.NoneLog).InitializeContracts().CreateAccounts("emulator-account")
}

//NewGoWithTheFlowForNetwork creates a new gwtf client for the provided network
func NewGoWithTheFlowForNetwork(network string) *GoWithTheFlow {
	return NewGoWithTheFlow(config.DefaultPaths(), network, false, output.InfoLog)

}

//NewGoWithTheFlowEmulator create a new client
func NewGoWithTheFlowEmulator() *GoWithTheFlow {
	return NewGoWithTheFlow(config.DefaultPaths(), "emulator", false, output.InfoLog)
}

//NewGoWithTheFlowDevNet creates a new gwtf client for devnet/testnet
func NewGoWithTheFlowDevNet() *GoWithTheFlow {
	return NewGoWithTheFlow(config.DefaultPaths(), "testnet", false, output.InfoLog)
}

//NewGoWithTheFlowMainNet creates a new gwft client for mainnet
func NewGoWithTheFlowMainNet() *GoWithTheFlow {
	return NewGoWithTheFlow(config.DefaultPaths(), "mainnet", false, output.InfoLog)
}

// NewGoWithTheFlow with custom file panic on error
func NewGoWithTheFlow(filenames []string, network string, inMemory bool, loglevel int) *GoWithTheFlow {
	gwtf, err := NewGoWithTheFlowError(filenames, network, inMemory, loglevel)
	if err != nil {
		log.Fatalf("%v error %+v", emoji.PileOfPoo, err)
	}
	return gwtf
}

//DoNotPrependNetworkToAccountNames disable the default behavior of prefixing account names with network-
func (f *GoWithTheFlow) DoNotPrependNetworkToAccountNames() *GoWithTheFlow {
	f.PrependNetworkToAccountNames = false
	return f
}

//Account fetch an account from flow.json, prefixing the name with network- as default (can be turned off)
func (f *GoWithTheFlow) Account(key string) *flowkit.Account {
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
func NewGoWithTheFlowError(paths []string, network string, inMemory bool, logLevel int) (*GoWithTheFlow, error) {

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
	return &GoWithTheFlow{
		State:                        state,
		Services:                     service,
		Network:                      network,
		Logger:                       logger,
		PrependNetworkToAccountNames: true,
	}, nil

}
