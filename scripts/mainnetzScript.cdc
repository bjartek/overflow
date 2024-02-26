import FungibleToken from "../contracts/FungibleToken.cdc"
// This is a mainnet specific script

access(all) fun main(account: Address): String {
    return getAccount(account).address.toString()
}
