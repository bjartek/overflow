import FungibleToken from "../contracts/FungibleToken.cdc"
// This is a mainnet specific script

pub fun main(account: Address): String {
    return getAccount(account).address.toString()
}
