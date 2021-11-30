// This transaction creates an empty NFT Collection in the signer's account
transaction(test:String) {
  prepare(acct: AuthAccount, account2: AuthAccount) {
    log(acct)
    log(account2)
 }
}