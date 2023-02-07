package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorsInAccountCreation(t *testing.T) {

	t.Run("Should deploy contracts to multiple accounts", func(t *testing.T) {
		_, err := OverflowTesting(WithFlowConfig("testdata/flow-with-multiple-deployments.json"), WithLogFull(), WithFlowForNewUsers(100.0))
		assert.NoError(t, err)
	})

	t.Run("Should give error on wrong contract name", func(t *testing.T) {
		_, err := OverflowTesting(WithFlowConfig("testdata/non_existing_contract.json"))
		assert.ErrorContains(t, err, "deployment contains nonexisting contract Debug2")
	})

	t.Run("Should give error on invalid env var in flow.json", func(t *testing.T) {
		_, err := OverflowTesting(WithFlowConfig("testdata/invalid_env_flow.json"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid private key for account: emulator-5")
	})

	t.Run("Should give error on wrong account name", func(t *testing.T) {
		_, err := OverflowTesting(WithFlowConfig("testdata/invalid_account_in_deployment.json"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deployment contains nonexisting account emulator-firs")
	})

}

func TestGetAccount(t *testing.T) {

	t.Run("Should return the account", func(t *testing.T) {
		g, err := OverflowTesting()
		require.NotNil(t, g)
		require.NoError(t, err)
		assert.NotNil(t, g)
		account, err := g.GetAccount("account")
		require.NoError(t, err)
		assert.Equal(t, "f8d6e0586b0a20c7", account.Address.String())
	})

	t.Run("Should return an error if account doesn't exist", func(t *testing.T) {
		g, err := OverflowTesting()
		require.NotNil(t, g)
		require.NoError(t, err)
		_, err = g.GetAccount("doesnotexist")
		assert.ErrorContains(t, err, "could not find account with name emulator-doesnotexist in the configuration")

	})

	t.Run("Should return an error if sa does not exist", func(t *testing.T) {
		_, err := OverflowTesting(WithServiceAccountSuffix("dummy"))

		assert.ErrorContains(t, err, "could not find account with name emulator-dummy in the configuration")

	})
}

/*
func TestCheckContractUpdate(t *testing.T) {

	t.Run("Should return the updatable contracts", func(t *testing.T) {

		g, err := OverflowTesting()
		require.NoError(t, err)
		res, err := g.CheckContractUpdates()

		assert.Nil(t, err)
		autogold.Equal(t, res)
	})

	t.Run("Should return the updatable contracts (updatable)", func(t *testing.T) {
		g, err := OverflowTesting()
		require.NoError(t, err)

		code := []byte(`pub contract Debug{

	pub struct FooListBar {
		pub let foo:[Foo2]
		pub let bar:String

		init(foo:[Foo2], bar:String) {
			self.foo=foo
			self.bar=bar
		}
	}

	pub struct Foo2{
		pub let bar: Address

		init(bar: Address) {
			self.bar=bar
		}
	}
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

			pub fun haha(){}
		}`)

		contract := flowkit.NewScript(code, []cadence.Value{}, "Debug")

		err = g.AddContract("account", contract, true)
		assert.Nil(t, err)
		res, err := g.CheckContractUpdates()

		assert.Nil(t, err)
		autogold.Equal(t, res)
	})

}
*/
