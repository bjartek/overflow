//Example startup using WithFlowEmulator to start the Emulator with HTTP Servers
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bjartek/overflow/overflow"
)

type FclAccount struct {
	Type    string    `json:"type"`
	Address string    `json:"address"`
	KeyId   int       `json:"keyId"`
	Label   string    `json:"label"`
	Scopes  *[]string `json:"scopes"`
}

type fclAccounts []FclAccount

func main() {

	o, err := overflow.OverflowE(overflow.WithEmulatorServer())
	if err != nil {
		log.Fatal(err)
	}

	fclAccountList := []FclAccount{}

	for _, account := range *o.State.Accounts() {
		fclAccount := FclAccount{
			Type:    "ACCOUNT",
			Address: account.Address().String(),
			KeyId:   0,
			Label:   account.Name(),
			Scopes:  new([]string),
		}

		fclAccountList = append(fclAccountList, fclAccount)
	}

	acctJSON, err := json.MarshalIndent(fclAccountList, "", " ")
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println(string(acctJSON))

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("press ctrl+c to stop emulator...")
	<-done // Will block here until user hits ctrl+c
}
