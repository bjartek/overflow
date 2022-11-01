import NonFungibleToken from "./NonFungibleToken.cdc"

pub contract Debug {

	pub struct FooBar {
		pub let foo:Foo
		pub let bar:String

		init(foo:Foo, bar:String) {
			self.foo=foo
			self.bar=bar
		}
	}

	pub struct Foo{
		pub let bar: String

		init(bar: String) {
			self.bar=bar
		}
	}

	pub event Log(msg: String)
	pub event LogNum(id: UInt64)

	pub fun id(_ id:UInt64) {
		emit LogNum(id:id)
	}

	pub fun log(_ msg: String) : String {
		emit Log(msg: msg)
		return msg
	}

}
