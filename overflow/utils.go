package overflow

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-emulator/server"
	"github.com/onflow/flow-go/fvm"
	"github.com/psiemens/graceland"
	"github.com/sirupsen/logrus"
)

func CadenceString(input string) cadence.String {
	value, err := cadence.NewString(input)
	if err != nil {
		panic(err)
	}
	return value
}

// HexToAddress converts a hex string to an Address.
func HexToAddress(h string) (*cadence.Address, error) {
	trimmed := strings.TrimPrefix(h, "0x")
	if len(trimmed)%2 == 1 {
		trimmed = "0" + trimmed
	}
	b, err := hex.DecodeString(trimmed)
	if err != nil {
		return nil, err

	}
	address := cadence.BytesToAddress(b)
	return &address, nil
}

func parseTime(timeString string, location string) (string, error) {
	loc, err := time.LoadLocation(location)
	if err != nil {
		return "", err
	}
	time.Local = loc
	t, err := dateparse.ParseLocal(timeString)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.0", t.Unix()), nil
}

func getAndUnquoteStringAsPointer(value cadence.Value) *string {
	result, err := strconv.Unquote(value.String())
	if err != nil {
		result = value.String()
	}

	if result == "" {
		return nil
	}
	return &result
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func writeProgressToFile(fileName string, blockHeight uint64) error {

	err := ioutil.WriteFile(fileName, []byte(fmt.Sprintf("%d", blockHeight)), 0644)

	if err != nil {
		return fmt.Errorf("could not create initial progress file %v", err)
	}
	return nil
}

func readProgressFromFile(fileName string) (int64, error) {
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		return 0, fmt.Errorf("ProgressFile is not valid %v", err)
	}

	stringValue := strings.TrimSpace(string(dat))

	return strconv.ParseInt(stringValue, 10, 64)
}

func parseCadenceUFix64(value string, valueName string) cadence.UFix64 {
	tokenSupply, err := cadence.NewUFix64(value)
	if err != nil {
		Exit(
			1,
			fmt.Sprintf(
				"Failed to parse %s from value `%s` as an unsigned 64-bit fixed-point number: %s",
				valueName,
				value,
				err.Error()),
		)
	}

	return tokenSupply
}

func newEmulatorServer(state *flowkit.State) *graceland.Group {
	cliLogger := logrus.New()
	cliLogger.Out = os.Stdout

	emulatorGroup := graceland.NewGroup()

	acc, _ := state.EmulatorServiceAccount()

	serverConf := &server.Config{
		GRPCPort:                  state.Config().Emulators.Default().Port,
		GRPCDebug:                 false,
		AdminPort:                 8080,
		RESTPort:                  8888,
		RESTDebug:                 false,
		HTTPHeaders:               nil,
		BlockTime:                 0,
		ServicePublicKey:          acc.Key().ToConfig().PrivateKey.PublicKey(),
		ServicePrivateKey:         acc.Key().ToConfig().PrivateKey,
		ServiceKeySigAlgo:         acc.Key().SigAlgo(),
		ServiceKeyHashAlgo:        acc.Key().HashAlgo(),
		Persist:                   false,
		DBPath:                    "./flowdb",
		GenesisTokenSupply:        parseCadenceUFix64("1000000000.0", "token-supply"),
		TransactionMaxGasLimit:    uint64(9999),
		ScriptGasLimit:            uint64(100000),
		TransactionExpiry:         uint(10),
		StorageLimitEnabled:       true,
		StorageMBPerFLOW:          fvm.DefaultStorageMBPerFLOW,
		MinimumStorageReservation: fvm.DefaultMinimumStorageReservation,
		TransactionFeesEnabled:    false,
		WithContracts:             true,
	}

	emulator := server.NewEmulatorServer(cliLogger, serverConf, emulatorGroup)
	emulator.Start()

	return emulatorGroup
}

func Exit(code int, msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}
