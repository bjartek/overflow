package overflow

import (
	"context"
	"encoding/hex"

	"github.com/onflow/flow-go-sdk"
)

func (f *Overflow) SignUserMessage(account string, message string) (string, err) {

	signer, err := f.Account(account).Key().Signer(context.Background())
	if err != nil {
		return "", err
	}

	signature, err := flow.SignUserMessage(signer, []byte(message))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(signature), nil

}
