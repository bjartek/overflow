package overflow

import "strings"

type FlowInteractionTemplate struct {
	FType    string `json:"f_type"`
	FVersion string `json:"f_version"`
	ID       string `json:"id"`
	Data     Data   `json:"data"`
}

func (self FlowInteractionTemplate) IsTransaction() bool {
	return self.Data.Type == "transaction"
}

type Title struct {
	I18N map[string]string `json:"i18n"`
}
type Description struct {
	I18N map[string]string `json:"i18n"`
}
type Messages struct {
	Title       Title       `json:"title"`
	Description Description `json:"description"`
}
type Network struct {
	Address        string `json:"address"`
	FqAddress      string `json:"fq_address"`
	Contract       string `json:"contract"`
	Pin            string `json:"pin"`
	PinBlockHeight int    `json:"pin_block_height"`
}

type Dependencies = map[string]map[string]map[string]Network

type Argument struct {
	Index    int      `json:"index"`
	Type     string   `json:"type"`
	Messages Messages `json:"messages"`
	Balance  *string  `json:"balance"`
}

type Arguments map[string]Argument
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
