// This transaction creates an empty NFT Collection in the signer's account
transaction(test:String) {
    prepare(acct: &Account, account2: &Account) {
        log(acct)
        log(account2)
    }
}
