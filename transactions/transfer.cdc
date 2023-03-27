import FungibleToken from "../contracts/FungibleToken.cdc"

/**
Transfer Tokens

Transfer tokens from one account to another

@lang en-US
@param to: The Flow account the tokens will go to
@param amount: The amount of FLOW tokens to send
@balance amount: FlowToken
*/
transaction(amount: UFix64, to: Address) {
    let vault: @FungibleToken.Vault
    prepare(signer: AuthAccount) {
        self.vault <- signer
        .borrow<&{FungibleToken.Provider}>(from: /storage/flowTokenVault)!
        .withdraw(amount: amount)

    }
    execute {
        getAccount(to)
        .getCapability(/public/flowTokenReceiver)
        .borrow<&{FungibleToken.Receiver}>()!
        .deposit(from: <-self.vault)
    }
}
