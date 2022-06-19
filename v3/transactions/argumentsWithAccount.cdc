
// This transaction creates an empty NFT Collection in the signer's account
transaction(test:Address) {
  prepare(acct: AuthAccount) {
    log("signer")
    log(acct)
    log("argument")
    log(test)
 }
}