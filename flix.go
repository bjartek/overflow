package overflow

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io"
	"strings"

	"github.com/ethereum/go-ethereum/rlp"
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
		Sha256String(flix.FType),
		Sha256String(flix.FVersion),
		Sha256String(flix.Data.Type),
		Sha256String(flix.Data.Interface),
		flix.Data.Messages,
		Sha256String(flix.Data.Cadence),
		flix.Data.Dependencies,
		flix.Data.Arguments,
	}

	return rlp.Encode(w, input)
}

func (self FlowInteractionTemplate) IsTransaction() bool {
	return self.Data.Type == "transaction"
}

type Title struct {
	I18N map[string]string `json:"i18n"`
}

func (this Title) EncodeRLP(w io.Writer) (err error) {
	list := []interface{}{}
	for lang, content := range this.I18N {
		list = append(list, []interface{}{
			Sha256String(lang),
			Sha256String(content),
		})
	}
	return rlp.Encode(w, []interface{}{Sha256String("title"), list})
}

type Description struct {
	I18N map[string]string `json:"i18n"`
}

func (this Description) EncodeRLP(w io.Writer) (err error) {
	list := []interface{}{}
	for lang, content := range this.I18N {
		list = append(list, []interface{}{
			Sha256String(lang),
			Sha256String(content),
		})
	}
	return rlp.Encode(w, []interface{}{Sha256String("description"), list})
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
func (this Messages) EncodeRLP(w io.Writer) (err error) {
	return rlp.Encode(w, []interface{}{this.Title, this.Description})
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

func (this Network) EncodeRLP(w io.Writer) (err error) {

	return rlp.Encode(w, []interface{}{
		Sha256String(this.Address),
		Sha256String(this.Contract),
		Sha256String(this.FqAddress),
		Sha256String(this.Pin),
	})
}

type Dependencies map[string]map[string]map[string]Network

func (this Dependencies) EncodeRLP(w io.Writer) (err error) {
	list := []interface{}{}
	for placeholder, contracts := range this {

		contractRLP := []interface{}{}
		for name, networks := range contracts {

			networkRLP := []interface{}{}
			for networkName, network := range networks {
				/*
						sha3_256(template-dependency-network),
						[
						    sha3_256(template-dependency-network-address),
						    sha3_256(template-dependency-contract-name),
						    sha3_256(template-dependency-contract-fq-address),
						    sha3_256(template-dependency-contract-pin),
						    sha3_256(template-dependency-contract-pin-block-height)
						]

					]
				*/
				networkRLP = append(networkRLP, []interface{}{Sha256String(networkName), network})
			}
			/*
				template-dependency-contract-name    = Name of a contract
				template-dependency-contract         = [
						sha3_256(template-dependency-contract-name),
						[ ...template-dependency-contract-network ]
				]

			*/
			contractRLP = append(contractRLP, []interface{}{Sha256String(name), networkRLP})
		}
		/*
			template-dependency-addr-placeholder = Placeholder address
			template-dependency                  = [
			    sha3_256(template-dependency-addr-placeholder),
			    [ ...template-dependency-contract ]
			]
		*/
		list = append(list, []interface{}{Sha256String(placeholder), contractRLP})
	}
	return rlp.Encode(w, list)
}

type Argument struct {
	Index    int      `json:"index"`
	Type     string   `json:"type"`
	Messages Messages `json:"messages"`
	Balance  *string  `json:"balance"`
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

func (this Argument) EncodeRLP(w io.Writer) (err error) {

	balance := []byte{}
	if this.Balance != nil {
		balance = Sha256String(*this.Balance)
	}
	list := []interface{}{
		Sha256Int(this.Index),
		Sha256String(this.Type),
		balance,
		this.Messages,
	}

	return rlp.Encode(w, list)
}

type Arguments map[string]Argument

/*
	template-argument-label         = Label for an argument

template-argument               = [ sha3_256(template-argument-label), [ ...template-argument-content ]]
template-arguments            = [ ...template-argument ] | []
*/
func (this Arguments) EncodeRLP(w io.Writer) (err error) {

	arguments := []interface{}{}
	for label, arg := range this {

		argRlp := []interface{}{
			Sha256String(label),
			arg,
		}
		arguments = append(arguments, argRlp)
	}
	return rlp.Encode(w, arguments)

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

func Sha256String(value string) []byte {
	return sha256Sum([]byte(value))
}

func Sha256Int(value int) []byte {
	return sha256Sum(intToBytes(value))
}

func Sha256Uint64(value uint64) []byte {
	return sha256Sum(uint64ToBytes(value))
}

func sha256Sum(b []byte) []byte {
	h := sha256.Sum256(b)
	return h[:]
}

func intToBytes(num int) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(num))
	return buf
}

func uint64ToBytes(num uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, num)
	return b
}
