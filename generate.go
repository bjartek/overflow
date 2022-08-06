package overflow

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (o *OverflowState) GenerateStub(network, filePath string) (string, error) {

	solution, err := o.ParseAll()
	if err != nil {
		return "", err
	}

	interactionName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	var interaction *OverflowDeclarationInfo
	var commandName string
	if strings.HasPrefix(filePath, o.TransactionBasePath) || strings.HasPrefix("./"+filePath, o.TransactionBasePath) {
		interaction = solution.Transactions[interactionName]
		commandName = "Tx"
	} else {
		interaction = solution.Scripts[interactionName]
		commandName = "Script"
	}
	if interaction == nil {
		return "", fmt.Errorf("Could not find interaction of type %s with name %s", commandName, interaction)
	}
	stub := fmt.Sprintf(`  o.%s("%s"`, commandName, interactionName)
	if len(interaction.Parameters) > 0 {
		stub = stub + ",\n"
	}
	if commandName == "Tx" {

		if len(interaction.Parameters) == 0 {
			stub = stub + ",\n"
		}
		stub = stub + "    WithSigner(\"\"),\n"
	}
	for name, value := range interaction.Parameters {
		stub = stub + fmt.Sprintf("    WithArg(\"%s\", \"%s\"),\n", name, value)
	}
	stub = stub + `)`
	return stub, nil

}
