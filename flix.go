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

	/*
		template-encoded              = RLP([
			    sha3_256(template-f-type),
			    sha3_256(template-f-version),
			    sha3_256(template-type),
			    sha3_256(template-interface),
			    template-messages,
			    sha3_256(template-cadence),
			    template-dependencies,
			    template-arguments
			])

	*/

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

type Title struct {
	I18N map[string]string `json:"i18n"`
}

//function som tar inn en map av K,V og returnerer interface{}

func (this Title) ToRLP() []interface{} {

	list := ProcessMap(this.I18N, func(lang string, content string) interface{} {
		return []interface{}{
			shaHex(lang, "message title lang"),
			shaHex(content, "message title content"),
		}
	})
	return []interface{}{shaHex("title", "message title"), list}
}

type Description struct {
	I18N map[string]string `json:"i18n"`
}

func (this Description) ToRLP() []interface{} {
	list := ProcessMap(this.I18N, func(lang string, content string) interface{} {
		return []interface{}{
			shaHex(lang, "message description lang"),
			shaHex(content, "message description content"),
		}
	})

	return []interface{}{shaHex("description", "message description"), list}
}

type Messages struct {
	Title       *Title       `json:"title,omitempty"`
	Description *Description `json:"description,omitempty"`
}

/*
template-argument-content-message-key-content   = UTF-8 string content of the message
template-argument-content-message-key-bcp47-tag = BCP-47 language tag
template-argument-content-message-translation   = [

	sha3_256(template-argument-content-message-key-bcp47-tag),
	sha3_256(template-argument-content-message-key-content)

]
template-argument-content-message-key           = Key for a template message (eg: "title", "description" etc)
template-argument-content-message = [

	sha3_256(template-argument-content-message-key),
	[ ...template-argument-content-message-translation ]

]
*/
func (this Messages) ToRLP() []interface{} {

	parts := []interface{}{this.Title.ToRLP()}
	if this.Description != nil {
		parts = append(parts, this.Description.ToRLP())
	}

	return parts
}

type Network struct {
	Address        string `json:"address"`
	FqAddress      string `json:"fq_address"`
	Contract       string `json:"contract"`
	Pin            string `json:"pin"`
	PinBlockHeight uint64 `json:"pin_block_height"`
}

/*
template-dependency-contract-pin-block-height = Network block height the pin was generated against.
template-dependency-contract-pin              = Pin of contract
template-dependency-contract-fq-addr          = Fully qualified contract identifier
template-dependency-network-address           = Address of an account
template-dependency-network                   = "mainnet" | "testnet" | "emulator" | Custom Network Tag
template-dependency-contract-network          = [

*/

func (this Network) ToRLP() []interface{} {

	return []interface{}{
		shaHex(this.Address, "dep address"),
		shaHex(this.Contract, "dep contract"),
		shaHex(this.FqAddress, "dep fqaddress"),
		shaHex(this.Pin, "dep pin"),
		shaHex(this.PinBlockHeight, "dep pin height"),
	}
}

type Dependencies map[string]Contracts
type Contracts map[string]Networks
type Networks map[string]Network

func (this Dependencies) ToRLP() []interface{} {
	return ProcessMap(this, func(placeholder string, contracts Contracts) interface{} {
		contractRLP := ProcessMap(contracts, func(name string, networks Networks) interface{} {
			networkRLP := ProcessMap(networks, func(networkName string, network Network) interface{} {
				return []interface{}{shaHex(networkName, "network"), network.ToRLP()}
			})
			return []interface{}{shaHex(name, "contract"), networkRLP}
		})
		return []interface{}{shaHex(placeholder, "placeholder"), contractRLP}
	})
}

type Argument struct {
	Index    int      `json:"index"`
	Type     string   `json:"type"`
	Messages Messages `json:"messages"`
	Balance  string   `json:"balance"`
}

/*
	template-argument-content-index   = Cadence type of argument
	template-argument-content-index   = Index of argument in cadence transaction or script
	template-argument-content-balance = Fully qualified contract identifier of a token this argument acts upon | ""
	template-argument-content         = [
	    sha3_256(template-argument-content-index),
	    sha3_256(template-argument-content-type),
	    sha3_256(template-argument-content-balance),
	    [ ...template-argument-content-message ]
	]

*/

func (this Argument) ToRLP() []interface{} {

	list := []interface{}{
		shaHex(this.Index, "argument index"),
		shaHex(this.Type, "argument type"),
		shaHex(this.Balance, "argument balance"),
		this.Messages.ToRLP(),
	}

	return list
}

type Arguments map[string]Argument

/*
	template-argument-label         = Label for an argument

template-argument               = [ sha3_256(template-argument-label), [ ...template-argument-content ]]
template-arguments            = [ ...template-argument ] | []
*/
func (this Arguments) ToRLP() []interface{} {
	return ProcessMap(this, func(label string, arg Argument) interface{} {
		return []interface{}{
			shaHex(label, "argument label"),
			arg.ToRLP(),
		}

	})
}

type Data struct {
	Type         string       `json:"type"`
	Interface    string       `json:"interface"`
	Messages     Messages     `json:"messages"`
	Cadence      string       `json:"cadence"`
	Dependencies Dependencies `json:"dependencies"`
	Arguments    Arguments    `json:"arguments"`
}

func (self Data) ResolvedCadence(network string) string {
	code := self.Cadence
	for placeholder, dependency := range self.Dependencies {
		for _, networks := range dependency {
			address := networks[network].Address
			code = strings.ReplaceAll(code, placeholder, address)
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

	codes := map[common.Location][]byte{}
	location := common.StringLocation(name)
	program, _ := cmd.PrepareProgram(code, location, codes)

	hashes := []string{shaHex(code, "pin code")}
	for _, imp := range program.ImportDeclarations() {
		address, isAddressImport := imp.Location.(common.AddressLocation)
		if isAddressImport {
			adr := address.Address.Hex()
			impName := imp.Identifiers[0].Identifier
			dep, err := o.GenerateDependentPin(adr, impName, cache)
			if err != nil {
				//TODO: gather up errors?
				return nil, err
			}
			hashes = append(hashes, dep...)
		}
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
