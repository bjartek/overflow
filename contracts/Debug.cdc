import "NonFungibleToken"

access(all) contract Debug {

    access(all) struct FooListBar {
        access(all) let foo:[Foo2]
        access(all) let bar:String

        init(foo:[Foo2], bar:String) {
            self.foo=foo
            self.bar=bar
        }
    }
    access(all) struct FooBar {
        access(all) let foo:Foo
        access(all) let bar:String

        init(foo:Foo, bar:String) {
            self.foo=foo
            self.bar=bar
        }
    }


    access(all) struct Foo2{
        access(all) let bar: Address

        init(bar: Address) {
            self.bar=bar
        }
    }

    access(all) struct Foo{
        access(all) let bar: String

        init(bar: String) {
            self.bar=bar
        }
    }

    access(all) event Log(msg: String)
    access(all) event LogNum(id: UInt64)

    access(all) fun id(_ id:UInt64) {
        emit LogNum(id:id)
    }

    access(all) fun log(_ msg: String) : String {
        emit Log(msg: msg)
        return msg
    }

}
