// test script to ensure code is running
pub fun main(): UInt64 {
    let height = getCurrentBlock().height
    log(height)
    return height
}