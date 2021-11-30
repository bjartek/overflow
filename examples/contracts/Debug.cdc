pub contract Debug {


	pub struct Foo{
		pub let bar: String

		init(bar: String) {
			self.bar=bar
		}
	}

	pub event Log(msg: String)

	pub fun log(_ msg: String) : String {
		emit Log(msg: msg)
		return msg
	}

}
