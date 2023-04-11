import FungibleToken from "../contracts/FungibleToken.cdc"
// This is a generic script

pub fun main(account: Address): String {
    return getAccount(account).address.toString()
}
