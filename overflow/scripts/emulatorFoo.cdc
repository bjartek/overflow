// test script to ensure code is running
import NonFungibleToken from "../contracts/NonFungibleToken.cdc"

pub fun main(account: Address): String {
    return getAccount(account).address.toString()
}
