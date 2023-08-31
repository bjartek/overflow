// test script to ensure code is running
import FlowToken from "../contracts/FlowToken.cdc"

access(all) fun main(): Type {
    return Type<@FlowToken.Vault>()
}
