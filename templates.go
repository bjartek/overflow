package overflow

import (
	"encoding/base64"
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
	if _, err := o.Transaction(`
	transaction {
		prepare(signer: AuthAccount) {
			let path = /storage/upload
			let existing = signer.load<String>(from: path) ?? ""
			log(existing)
		}
	}
	  `).SignProposeAndPayAs(accountName).RunE(); err != nil {
		return err
	}

	parts := splitByWidthMake(content, 1_000_000)
	for _, part := range parts {
		if _, err := o.Transaction(`
		transaction(part: String) {
			prepare(signer: AuthAccount) {
				let path = /storage/upload
				let existing = signer.load<String>(from: path) ?? ""
				signer.save(existing.concat(part), to: path)
				log(signer.address.toString())
				log(part)
			}
		}
			`).SignProposeAndPayAs(accountName).Args(o.Arguments().String(part)).RunE(); err != nil {
			return err
		}
	}

	return nil
}

// Get the free capacity in an account
func (o *OverflowState) GetFreeCapacity(accountName string) int {

	result := o.InlineScript(`
pub fun main(user:Address): UInt64{
	let account=getAccount(user)
	return account.storageCapacity - account.storageUsed
}
`).Args(o.Arguments().Account(accountName)).RunReturnsInterface().(uint64)

	return int(result)

}

// A method to fill up a users storage, useful when testing
func (o *OverflowState) FillUpStorage(accountName string) *OverflowState {

	length := o.GetFreeCapacity(accountName) - 67 //some storage is made outside of the string so need to adjust

	err := o.UploadString(randomString(length), accountName)
	if err != nil {
		o.Error = err
	}
	return o
}
