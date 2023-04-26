package overflow

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/onflow/cadence/runtime/cmd"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-go-sdk"
	"github.com/samber/lo"
	"github.com/sanity-io/litter"
	"golang.org/x/crypto/sha3"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type FlowInteractionTemplate struct {
	FType    string `json:"f_type"`
	FVersion string `json:"f_version"`
	ID       string `json:"id"`
	Data     Data   `json:"data"`
}

func (flix FlowInteractionTemplate) EncodeRLP(w io.Writer) (err error) {
	input := []interface{}{
		shaHex(flix.FType, "f-type"),
		shaHex(flix.FVersion, "f-version"),
		shaHex(flix.Data.Type, "type"),
		shaHex(flix.Data.Interface, "interface"),
		flix.Data.Messages.ToRLP(),
		shaHex(flix.Data.Cadence, "cadence"),
		flix.Data.Dependencies.ToRLP(),
		flix.Data.Arguments.ToRLP(),
	}

	litter.Dump(input)

	return rlp.Encode(w, input)
}

func (self FlowInteractionTemplate) IsTransaction() bool {
	return self.Data.Type == "transaction"
}

type Message struct {
	Key  string            `json:"key"`
	I18N map[string]string `json:"i18n"`
}

func (this Message) ToRLP(_ int) []interface{} {

	list := ProcessMap(this.I18N, func(lang string, content string) interface{} {
		return []interface{}{
			shaHex(lang, "message title lang"),
			shaHex(content, "message title content"),
		}
	})
	return []interface{}{shaHex(this.Key, "message key"), list}
}

type Messages []Message

func (this Messages) ToRLP() [][]interface{} {
	return lo.Map(this, Message.ToRLP)
}

type Network struct {
	Network        string `json:"network"`
	Address        string `json:"address"`
	FqAddress      string `json:"fq_address"`
	Pin            string `json:"pin"`
	PinBlockHeight uint64 `json:"pin_block_height"`
}

func (this Network) ToRLP(_ int) []interface{} {

	return []interface{}{
		shaHex(this.Address, "dep address"),
		shaHex(this.FqAddress, "dep fqaddress"),
		shaHex(this.Pin, "dep pin"),
		shaHex(this.PinBlockHeight, "dep pin height"),
	}
}

type Dependencies []Dependency

func (this Dependencies) ToRLP() [][]interface{} {
	return lo.Map(this, Dependency.ToRLP)
}

type Dependency struct {
	Address   string     `json:"address"`
	Contracts []Contract `json:"contracts"`
}

func (this Dependency) ToRLP(_ int) []interface{} {
	return []interface{}{
		shaHex(this.Address, "dep address"),
		lo.Map(this.Contracts, Contract.ToRLP),
	}
}

type Contract struct {
	Contract string    `json:"contract"`
	Networks []Network `json:"networks"`
}

func (this Contract) ToRLP(_ int) []interface{} {
	return []interface{}{
		shaHex(this.Contract, "dep contract"),
		lo.Map(this.Networks, Network.ToRLP),
	}
}

type Networks []Network

type Argument struct {
	Key      string   `json:"key"`
	Index    int      `json:"index"`
	Type     string   `json:"type"`
	Messages Messages `json:"messages"`
	Balance  string   `json:"balance"`
}

func (this Argument) ToRLP(_ int) []interface{} {

	list := []interface{}{
		shaHex(this.Key, "argument key"),
		shaHex(this.Index, "argument index"),
		shaHex(this.Type, "argument type"),
		shaHex(this.Balance, "argument balance"),
		this.Messages.ToRLP(),
	}

	return list
}

type Arguments []Argument

func (this Arguments) ToRLP() [][]interface{} {
	return lo.Map(this, Argument.ToRLP)
}

type Data struct {
	Type         string       `json:"type"`
	Interface    string       `json:"interface"`
	Messages     Messages     `json:"messages"`
	Cadence      string       `json:"cadence"`
	Dependencies Dependencies `json:"dependencies"`
	Arguments    Arguments    `json:"arguments"`
}

func (self Data) ResolvedCadence(input string) string {
	code := self.Cadence
	for _, dependency := range self.Dependencies {
		for _, contract := range dependency.Contracts {
			for _, network := range contract.Networks {
				if network.Network == input {
					code = strings.ReplaceAll(code, dependency.Address, network.Address)
				}
			}
		}
	}
	return code
}

/*

	template-f-version            = Version of the InteractionTemplate data structure being serialized.
	template-f-type               = "InteractionTemplate"
	template-type                 = "transaction" | "script"
	template-interface            = ID of the InteractionTemplateInterface this template implements | ""
	template-messages             = [ ...template-message ] | []
	template-cadence              = Cadence content of the template
	template-dependencies         = [ ...template-dependency ] | []


*/

func (o *OverflowState) GeneratePin(address string, name string) (string, error) {

	identifier := fmt.Sprintf("%s.%s", address, name)

	horizen := []string{identifier}

	importHash := []string{}
	for _, contract := range horizen {

		split := strings.Split(contract, ".")
		address, name := split[0], split[1]
		account, err := o.Services.Accounts.Get(flow.HexToAddress(address))
		if err != nil {
			return "", err
		}
		code := account.Contracts[name]
		importHash = append(importHash, shaHex(code, ""))
		deps := GetAddressImports(code, name)
		horizen = append(horizen, deps...)
	}
	return shaHex(strings.Join(importHash, ""), ""), nil
}

func GetAddressImports(code []byte, name string) []string {

	deps := []string{}
	codes := map[common.Location][]byte{}
	location := common.StringLocation(name)
	program, _ := cmd.PrepareProgram(code, location, codes)
	for _, imp := range program.ImportDeclarations() {
		address, isAddressImport := imp.Location.(common.AddressLocation)
		if isAddressImport {
			adr := address.Address.Hex()
			impName := imp.Identifiers[0].Identifier
			deps = append(deps, fmt.Sprintf("%s.%s", adr, impName))
		}
	}
	return deps
}

// TODO: send in the cache to make it spawn multiple generators
func (o *OverflowState) GeneratePinDebthFirst(address string, name string) (string, error) {

	memoize := map[string][]string{}
	pin, err := o.GenerateDependentPin(address, name, memoize)

	if err != nil {
		return "", err
	}
	return shaHex(strings.Join(pin, ""), ""), nil
}

// https://github.com/onflow/fcl-js/blob/master/packages/fcl/src/interaction-template-utils/generate-dependency-pin.js
func (o *OverflowState) GenerateDependentPin(address string, name string, cache map[string][]string) ([]string, error) {

	identifier := fmt.Sprintf("A.%s.%s", strings.ReplaceAll(address, "0x", ""), name)
	existingHash, ok := cache[identifier]
	if ok {
		return existingHash, nil
	}
	account, err := o.Services.Accounts.Get(flow.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	code := account.Contracts[name]

	imports := GetAddressImports(code, name)
	hashes := []string{shaHex(code, "")}
	for _, imp := range imports {
		split := strings.Split(imp, ".")
		address, name := split[0], split[1]
		dep, err := o.GenerateDependentPin(address, name, cache)
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, dep...)
	}
	cache[identifier] = hashes
	return hashes, nil
}

func GenerateFlixID(flix FlowInteractionTemplate) (string, error) {
	rlpOutput, err := rlp.EncodeToBytes(flix)
	if err != nil {
		return "", err
	}

	//template-encoded-hex          = hex( template-encoded )
	dst := make([]byte, hex.EncodedLen(len(rlpOutput)))
	hex.Encode(dst, rlpOutput)

	//template-id                   = sha3_256( template-encoded-hex )
	shaOutput := sha256.Sum256(dst)
	return hex.EncodeToString(shaOutput[:]), nil
}

func ProcessMap[M ~map[K]V, K string, V any](m M, fn func(key K, value V) interface{}) []interface{} {
	keys := maps.Keys(m)
	slices.Sort(keys)

	list := []interface{}{}
	for _, key := range keys {
		value := m[key]
		list = append(list, fn(key, value))
	}
	return list
}

func shaHex(value interface{}, debugKey string) string {

	// Convert the value to a byte array
	data, err := convertToBytes(value)
	if err != nil {
		if debugKey != "" {
			fmt.Printf("%30s value=%v hex=%x\n", debugKey, value, err.Error())
		}
		return ""
	}

	// Calculate the SHA-3 hash
	hash := sha3.Sum256(data)

	// Convert the hash to a hexadecimal string
	hashHex := hex.EncodeToString(hash[:])

	if debugKey != "" {
		fmt.Printf("%30s hex=%v value=%v \n", debugKey, hashHex, value)
	}
	return hashHex
}

func convertToBytes(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case int:
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(v))
		return buf, nil
	case uint64:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, v)
		return buf, nil
	default:
		return nil, fmt.Errorf("unsupported type %T", v)
	}
}
