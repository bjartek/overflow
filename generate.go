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
	lines := []string{
		fmt.Sprintf(`  o.%s("%s"`, commandName, interactionName),
	}

	if commandName == "Tx" {
		lines = append(lines, "    WithSigner(\"\")")
	}
	for name, value := range interaction.Parameters {
		lines = append(lines, fmt.Sprintf("    WithArg(\"%s\", \"%s\")", name, value))
	}
	if len(lines) > 1 {
		lines = append(lines, "  )")
	} else {
		return lines[0] + ")", nil
	}

	return strings.Join(lines, ",\n"), nil

}
