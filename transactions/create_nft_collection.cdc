import NonFungibleToken from "../contracts/NonFungibleToken.cdc"

// This transaction creates an empty NFT Collection in the signer's account
transaction {
  prepare(acct: AuthAccount) {
    // store an empty NFT Collection in account storage
    acct.save<@NonFungibleToken.Collection>(<-NonFungibleToken.createEmptyCollection(), to: /storage/NFTCollection)

    // publish a capability to the Collection in storage
    acct.link<&{NonFungibleToken.NFTReceiver}>(/public/NFTReceiver, target: /storage/NFTCollection)

    log("Created a new empty collection and published a reference")
  }
}