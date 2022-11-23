package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bjartek/overflow"
	"github.com/onflow/flow-go-sdk"
)

func main() {

	ot := overflow.Overflow(overflow.WithNetwork("testnet"))
	if ot.Error != nil {
		panic(ot.Error)
	}
	o := overflow.Overflow(overflow.WithNetwork("mainnet"))

	res, err := o.ParseAllWithConfig(true, []string{
		"^admin*",
		"^setup_*",
	}, []string{})
	if err != nil {
		panic(err)
	}

	data := res.Transactions["mint_tokens"]

	docString := strings.TrimSpace(data.DocString)

	lines := strings.Split(strings.ReplaceAll(docString, "\r\n", "\n"), "\n")

	name := ""
	lang := "en-US"
	description := []string{}
	params := map[string]string{}
	balance := map[string]string{}

	paramKeyword := "@param"
	langKeyword := "@lang"
	balanceKeyword := "@balance"
	interfaceKeyword := "@flixInterface"
	flixInterface := ""

	for i, line := range lines {
		if i == 0 {
			name = line
			continue
		}
		if strings.HasPrefix(line, interfaceKeyword) {
			flixInterface = strings.TrimSpace(strings.TrimPrefix(line, interfaceKeyword))
			continue
		}

		if strings.HasPrefix(line, langKeyword) {
			lang = strings.TrimSpace(strings.TrimPrefix(line, langKeyword))
			continue
		}

		if strings.HasPrefix(line, balanceKeyword) {
			parts := strings.Split(strings.TrimPrefix(line, balanceKeyword), ":")
			balance[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			continue
		}

		if strings.HasPrefix(line, paramKeyword) {
			parts := strings.Split(strings.TrimPrefix(line, paramKeyword), ":")
			params[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			continue
		}

		description = append(description, line)
	}
	descriptionString := strings.TrimSpace(strings.Join(description, "\n"))

	flixArguments := map[string]overflow.Argument{}
	for i, arg := range data.ParameterOrder {

		value, ok := balance[arg]
		balance := &value
		if !ok {
			balance = nil
		}

		flixArguments[arg] = overflow.Argument{
			Index:    i,
			Type:     data.Parameters[arg],
			Messages: createMessage(lang, params[arg], ""),
			Balance:  balance,
		}

	}
	deps := map[string]map[string]map[string]overflow.Network{}

	for _, network := range *o.State.Networks() {
		if network.Name == "emulator" {
			continue
		}
		ovf := o
		if network.Name == "testnet" {
			ovf = ot
		}

		latestBlock, err := ovf.GetLatestBlock()
		if err != nil {
			panic(err)
		}

		for name := range data.Imports {
			address := ovf.Address(name)

			account, err := ovf.Services.Accounts.Get(flow.HexToAddress(address))
			if err != nil {
				panic(err)
			}
			contractBytes := account.Contracts[name]
			hash := sha256.Sum256(contractBytes)

			nw := overflow.Network{
				Address:        address,
				FqAddress:      fmt.Sprintf("A.%s.%s", strings.TrimPrefix(address, "0x"), name),
				Contract:       name,
				Pin:            hex.EncodeToString(hash[:]),
				PinBlockHeight: latestBlock.Height,
			}

			key1 := fmt.Sprintf("0x%s", strings.ToUpper(name))
			key2 := name
			key3 := network.Name

			fmt.Println(key1, " ", key2, " ", key3)
			if deps[key1] == nil {
				fmt.Println("cannot find key1")
				deps[key1] = map[string]map[string]overflow.Network{
					key2: {
						key3: nw,
					},
				}
			}
			deps[key1][key2][key3] = nw
		}
	}
	flix := overflow.FlowInteractionTemplate{
		FType:    "InteractionTemplate",
		FVersion: "1.0",
		Data: overflow.Data{
			Type:         "transaction",
			Interface:    flixInterface,
			Messages:     createMessage(lang, name, descriptionString),
			Cadence:      data.EnvCode,
			Dependencies: deps,
			Arguments:    flixArguments,
		},
	}

	out, _ := json.Marshal(flix)
	idHash := sha256.Sum256(out)
	flix.ID = hex.EncodeToString(idHash[:])

	out2, _ := json.MarshalIndent(flix, "", " ")

	fmt.Println(string(out2))

}

func createMessage(lang, title, description string) overflow.Messages {
	msg := overflow.Messages{}
	if title == "" && description == "" {
		return msg
	}

	if title != "" {
		title := overflow.Title{
			I18N: map[string]string{lang: title},
		}
		msg.Title = &title
	}

	if description != "" {
		desc := overflow.Description{
			I18N: map[string]string{lang: description},
		}
		msg.Description = &desc
	}

	return msg
}
