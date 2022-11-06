package overflow

import (
	"encoding/base64"
	"fmt"
)

// Templates
//
// A list of functions you can use to perform some common operations

// UploadFile reads a file, base64 encodes it and chunk upload to /storage/upload
func (o *OverflowState) UploadFile(filename string, accountName string) error {
	content, err := fileAsBase64(filename)
	if err != nil {
		return err
	}

	return o.UploadString(content, accountName)
}

// DownloadAndUploadFile reads a file, base64 encodes it and chunk upload to /storage/upload
func (o *OverflowState) DownloadAndUploadFile(url string, accountName string) error {
	body, err := getUrl(url)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(body)
	return o.UploadString(encoded, accountName)
}

// DownloadImageAndUploadAsDataUrl download an image and upload as data url
func (o *OverflowState) DownloadImageAndUploadAsDataUrl(url, accountName string) error {
	body, err := getUrl(url)
	if err != nil {
		return err
	}
	content := contentAsImageDataUrl(body)

	return o.UploadString(content, accountName)
}

// UploadImageAsDataUrl will upload a image file from the filesystem into /storage/upload of the given account
func (o *OverflowState) UploadImageAsDataUrl(filename string, accountName string) error {
	content, err := fileAsImageData(filename)
	if err != nil {
		return err
	}

	return o.UploadString(content, accountName)
}

// UploadString will upload the given string data in 1mb chunkts to /storage/upload of the given account
func (o *OverflowState) UploadString(content string, accountName string) error {
	//unload previous content if any.
	res := o.Tx(`
	transaction {
		prepare(signer: AuthAccount) {
			let path = /storage/upload
			let existing = signer.load<String>(from: path) ?? ""
			log(existing)
		}
	}
	  `, WithSigner(accountName))
	if res.Err != nil {
		return res.Err
	}

	parts := splitByWidthMake(content, 1_000_000)
	for _, part := range parts {
		res := o.Tx(`
		transaction(part: String) {
			prepare(signer: AuthAccount) {
				let path = /storage/upload
				let existing = signer.load<String>(from: path) ?? ""
				signer.save(existing.concat(part), to: path)
				log(signer.address.toString())
				log(part)
			}
		}
			`, WithSigner(accountName), WithArg("part", part))
		if res.Err != nil {
			return res.Err
		}
	}

	return nil
}

// Get the free capacity in an account
func (o *OverflowState) GetFreeCapacity(accountName string) int {

	result := o.Script(`
pub fun main(user:Address): UInt64{
	let account=getAccount(user)
	return account.storageCapacity - account.storageUsed
}
`, WithArg("user", accountName))

	value, ok := result.Output.(uint64)
	if !ok {
		panic("Type conversion of free capacity failed")
	}
	return int(value)
}

func (o *OverflowState) MintFlowTokens(accountName string, amount float64) *OverflowState {
	if o.Network != "emulator" {
		o.Error = fmt.Errorf("Can only mint new flow on emulator")
		return o
	}
	result := o.Tx(`
import FungibleToken from 0xee82856bf20e2aa6
import FlowToken from 0x0ae53cb6e3f42a79


transaction(recipient: Address, amount: UFix64) {
    let tokenAdmin: &FlowToken.Administrator
    let tokenReceiver: &{FungibleToken.Receiver}

    prepare(signer: AuthAccount) {
        self.tokenAdmin = signer
            .borrow<&FlowToken.Administrator>(from: /storage/flowTokenAdmin)
            ?? panic("Signer is not the token admin")

        self.tokenReceiver = getAccount(recipient)
            .getCapability(/public/flowTokenReceiver)
            .borrow<&{FungibleToken.Receiver}>()
            ?? panic("Unable to borrow receiver reference")
    }

    execute {
        let minter <- self.tokenAdmin.createNewMinter(allowedAmount: amount)
        let mintedVault <- minter.mintTokens(amount: amount)

        self.tokenReceiver.deposit(from: <-mintedVault)

        destroy minter
    }
}
`, WithSignerServiceAccount(),
		WithArg("recipient", accountName),
		WithArg("amount", amount),
		WithName(fmt.Sprintf("Startup Mint tokens for %s", accountName)),
		WithoutLog(),
	)

	if result.Err != nil {
		o.Error = result.Err
	}
	return o
}

// A method to fill up a users storage, useful when testing
// This has some issues with transaction fees
func (o *OverflowState) FillUpStorage(accountName string) *OverflowState {

	capacity := o.GetFreeCapacity(accountName)
	length := capacity - 7400 //we cannot fill up all of storage since we need flow to pay for the transaction that fills it up

	err := o.UploadString(randomString(length), accountName)
	if err != nil {
		o.Error = err
	}
	return o
}
