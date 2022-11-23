package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bjartek/overflow"
	"github.com/sanity-io/litter"
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
	fmt.Println("====")
	litter.Dump(data)

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
			o.Address()
			nw := overflow.Network{
				Address:        "",
				FqAddress:      "",
				Contract:       name,
				Pin:            ""
				PinBlockHeight: latestBlock.Height,
			}
			deps[fmt.Sprintf("0x%s", strings.ToUpper(name))][name][network.Name] = nw
		}
	}
	/*
		"0xFUNGIBLETOKENADDRESS": {
		        "FungibleToken": {
		          "mainnet": {
		            "address": "0xf233dcee88fe0abe",
		            "fq_address": "A.0xf233dcee88fe0abe.FungibleToken",
		            "contract": "FungibleToken",
		            "pin": "83c9e3d61d3b5ebf24356a9f17b5b57b12d6d56547abc73e05f820a0ae7d9cf5",
		            "pin_block_height": 34166296
		          },
		          "testnet": {
		            "address": "0x9a0766d93b6608b7",
		            "fq_address": "A.0x9a0766d93b6608b7.FungibleToken",
		            "contract": "FungibleToken",
		            "pin": "83c9e3d61d3b5ebf24356a9f17b5b57b12d6d56547abc73e05f820a0ae7d9cf5",
		            "pin_block_height": 74776482
		          }
		        }
		      }
	*/
	flix := overflow.FlowInteractionTemplate{
		FType:    "InteractionTemplate",
		FVersion: "1.0",
		ID:       "TBD",
		Data: overflow.Data{
			Type:         "transaction",
			Interface:    flixInterface,
			Messages:     createMessage(lang, name, descriptionString),
			Cadence:      data.EnvCode,
			Dependencies: map[string]map[string]map[string]overflow.Network{},
			Arguments:    flixArguments,
		},
	}

	out, _ := json.MarshalIndent(flix, "", " ")
	fmt.Println(string(out))

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

/*
TODO:
 - pins, contact network, find latest code, SHA3_256 digest hex
*/
