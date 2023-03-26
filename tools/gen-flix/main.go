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

	//what if we here add a i18n json file for this same tx.

	data := res.Transactions["mint_tokens"]

	docString := strings.TrimSpace(data.DocString)

	//this should probably read from multiple different files
	fileName := "flix/mint_tokens.json"
	messages, err := overflow.ReadFileArrayIntoStructs[Internationalisation](fileName)
	if err != nil {
		fmt.Printf("%s not found or could not be read err=%s\n", fileName, err.Error())
		messages = []Internationalisation{}
	}

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
	//arg -> language -> message
	argumentsMessage := map[string]map[string]InternationalisationMessage{}
	for _, arg := range data.ParameterOrder {

		argMap := map[string]InternationalisationMessage{}
		for _, msg := range messages {
			argMap[msg.Lang] = msg.Arguments[arg]
		}
		argMap[lang] = InternationalisationMessage{
			Title: params[arg],
		}
		argumentsMessage[arg] = argMap
	}

	mainMessage := map[string]InternationalisationMessage{}

	for _, msg := range messages {
		mainMessage[msg.Lang] = msg.Interaction
	}
	mainMessage[lang] = InternationalisationMessage{
		Title:       name,
		Description: descriptionString,
	}

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
			Messages: createMessages(argumentsMessage[arg]),
			Balance:  balance,
		}

	}
	deps := map[string]map[string]map[string]overflow.Network{}

	latestBlocks := map[string]*flow.Block{}

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
		latestBlocks[network.Name] = latestBlock

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

			if deps[key1] == nil {
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
			Messages:     createMessages(mainMessage),
			Cadence:      data.EnvCode,
			Dependencies: deps,
			Arguments:    flixArguments,
		},
	}

	flix.ID, err = overflow.GenerateFlixID(flix)

	//	out, _ := json.Marshal(flix)
	out2, _ := json.MarshalIndent(flix, "", " ")

	fmt.Println(string(out2))

}

func createMessages(messages map[string]InternationalisationMessage) overflow.Messages {

	titles := map[string]string{}
	descriptions := map[string]string{}
	for lang, message := range messages {

		if message.Title != "" {
			titles[lang] = message.Title
		}

		if message.Description != "" {
			descriptions[lang] = message.Description
		}

	}
	msg := overflow.Messages{}

	if len(titles) == 0 && len(descriptions) == 0 {
		return msg
	}

	title := overflow.Title{
		I18N: titles,
	}
	msg.Title = &title

	desc := overflow.Description{
		I18N: descriptions,
	}
	msg.Description = &desc

	return msg
}

type InternationalisationMessage struct {
	Title       string
	Description string
}

type Internationalisation struct {
	Lang        string
	Interaction InternationalisationMessage
	Arguments   map[string]InternationalisationMessage
}
