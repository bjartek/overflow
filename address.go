package overflow

import (
	"math/big"
	"strings"
)

var (
	linearCodeN = 64

	parityCheckMatrixColumns = []*big.Int{
		big.NewInt(0x00001), big.NewInt(0x00002), big.NewInt(0x00004), big.NewInt(0x00008),
		big.NewInt(0x00010), big.NewInt(0x00020), big.NewInt(0x00040), big.NewInt(0x00080),
		big.NewInt(0x00100), big.NewInt(0x00200), big.NewInt(0x00400), big.NewInt(0x00800),
		big.NewInt(0x01000), big.NewInt(0x02000), big.NewInt(0x04000), big.NewInt(0x08000),
		big.NewInt(0x10000), big.NewInt(0x20000), big.NewInt(0x40000), big.NewInt(0x7328d),
		big.NewInt(0x6689a), big.NewInt(0x6112f), big.NewInt(0x6084b), big.NewInt(0x433fd),
		big.NewInt(0x42aab), big.NewInt(0x41951), big.NewInt(0x233ce), big.NewInt(0x22a81),
		big.NewInt(0x21948), big.NewInt(0x1ef60), big.NewInt(0x1deca), big.NewInt(0x1c639),
		big.NewInt(0x1bdd8), big.NewInt(0x1a535), big.NewInt(0x194ac), big.NewInt(0x18c46),
		big.NewInt(0x1632b), big.NewInt(0x1529b), big.NewInt(0x14a43), big.NewInt(0x13184),
		big.NewInt(0x12942), big.NewInt(0x118c1), big.NewInt(0x0f812), big.NewInt(0x0e027),
		big.NewInt(0x0d00e), big.NewInt(0x0c83c), big.NewInt(0x0b01d), big.NewInt(0x0a831),
		big.NewInt(0x982b), big.NewInt(0x07034), big.NewInt(0x0682a), big.NewInt(0x05819),
		big.NewInt(0x03807), big.NewInt(0x007d2), big.NewInt(0x00727), big.NewInt(0x0068e),
		big.NewInt(0x0067c), big.NewInt(0x0059d), big.NewInt(0x004eb), big.NewInt(0x003b4),
		big.NewInt(0x0036a), big.NewInt(0x002d9), big.NewInt(0x001c7), big.NewInt(0x0003f),
	}
)

func GetNetworkFromAddress(input string) string {
	address := strings.TrimPrefix(input, "0x")
	testnet, _ := new(big.Int).SetString("6834ba37b3980209", 16)
	emulator, _ := new(big.Int).SetString("1cb159857af02018", 16)

	networkCodewords := map[string]*big.Int{
		"mainnet":  big.NewInt(0),
		"testnet":  testnet,
		"emulator": emulator,
	}

	for network, codeWord := range networkCodewords {
		if IsValidAddressForNetwork(address, network, codeWord) {
			return network
		}
	}
	return ""
}

func IsValidAddressForNetwork(address, network string, codeWord *big.Int) bool {
	flowAddress, ok := new(big.Int).SetString(address, 16)
	if !ok {
		panic("not valid address")
	}
	codeWord.Xor(codeWord, flowAddress)

	if codeWord.Cmp(big.NewInt(0)) == 0 {
		return false
	}

	parity := big.NewInt(0)
	for i := 0; i < linearCodeN; i++ {
		if codeWord.Bit(0) == 1 {
			parity.Xor(parity, parityCheckMatrixColumns[i])
		}
		codeWord.Rsh(codeWord, 1)
	}

	return parity.Cmp(big.NewInt(0)) == 0 && codeWord.Cmp(big.NewInt(0)) == 0
}
