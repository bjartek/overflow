// test script to ensure code is running
access(all) fun main(): UInt64 {
    let height = getCurrentBlock().height
    log(height)
    return height
}
