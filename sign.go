package overflow

import (
	"context"
	"encoding/hex"

	"github.com/onflow/flow-go-sdk"
)

// Sign a user message
func (o *OverflowState) SignUserMessage(account string, message string) (string, error) {

	a, err := o.AccountE(account)
	if err != nil {
		return "", err
	}

	signer, err := a.Key().Signer(context.Background())
	if err != nil {
		return "", err
	}

	signature, err := flow.SignUserMessage(signer, []byte(message))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(signature), nil

}
