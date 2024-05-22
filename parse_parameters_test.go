package overflow

import (
	"fmt"
	"testing"

	"github.com/hexops/autogold"
	"github.com/onflow/cadence"
	"github.com/stretchr/testify/require"
)

func TestParseArguments(t *testing.T) {
	o, err := OverflowTesting()
	require.NoError(t, err)
	require.NotNil(t, o)

	t.Run("testing user defined struct", func(t *testing.T) {
		code := []byte(`
    import "Debug"
    transaction(data:Debug.Foo2){}

    `)

		data := Debug_Foo2{
			Bar: "foo",
		}
		_, _, err := o.parseArguments("somefile", code, map[string]interface{}{"data": data})
		require.Error(t, err)
	})
	type Args struct {
		inputArgs map[string]interface{}
		code      string
	}

	type TestInput struct {
		want    autogold.Value
		want1   autogold.Value
		wantErr autogold.Value
		name    string
		args    Args
	}

	tests := []TestInput{
		{
			name: "64bit integers",
			args: Args{
				code:      "u:UInt64, i:Int64",
				inputArgs: map[string]interface{}{"u": 42, "i": 42},
			},
			want1:   autogold.Want("map", CadenceArguments{"i": cadence.Int64(42), "u": cadence.UInt64(42)}),
			want:    autogold.Want("slice", []cadence.Value{cadence.UInt64(42), cadence.Int64(42)}),
			wantErr: autogold.Want("err", nil),
		},
		{
			name: "strings",
			args: Args{
				code:      "s:String",
				inputArgs: map[string]interface{}{"s": "foobar"},
			},
			want1:   autogold.Want("map2", CadenceArguments{"s": cadence.String("foobar")}),
			want:    autogold.Want("slice2", []cadence.Value{cadence.String("foobar")}),
			wantErr: autogold.Want("err2", nil),
		},

		{
			name: "address",
			args: Args{
				code:      "adr:Address",
				inputArgs: map[string]interface{}{"adr": "first"},
			},
			want1: autogold.Want("map3", CadenceArguments{"adr": cadence.Address{
				23,
				155,
				107,
				28,
				182,
				117,
				94,
				49,
			}}),
			want: autogold.Want("slice3", []cadence.Value{cadence.Address{
				23,
				155,
				107,
				28,
				182,
				117,
				94,
				49,
			}}),
			wantErr: autogold.Want("err3", nil),
		},
		{
			name: "missing argument",
			args: Args{
				code:      "adr:Address",
				inputArgs: map[string]interface{}{},
			},
			want1:   autogold.Want("map4", CadenceArguments{}),
			want:    autogold.Want("slice4", []cadence.Value{}),
			wantErr: autogold.Want("err4", "extracting arguments: the interaction 'somefile' is missing [adr]"),
		},
		{
			name: "redundant argument",
			args: Args{
				code:      "",
				inputArgs: map[string]interface{}{"foo": "bar"},
			},
			want1:   autogold.Want("map5", CadenceArguments{}),
			want:    autogold.Want("slice5", []cadence.Value{}),
			wantErr: autogold.Want("err5", "extracting arguments: the interaction 'somefile' has the following extra arguments [foo]"),
		},
		{
			name: "array of addresses",
			args: Args{
				code:      "adr:[Address]",
				inputArgs: map[string]interface{}{"adr": []string{"second"}},
			},
			want1: autogold.Want("map6", CadenceArguments{"adr": cadence.Array{
				Values: []cadence.Value{cadence.Address{
					243,
					252,
					210,
					193,
					167,
					143,
					94,
					238,
				}},
			}}),
			want: autogold.Want("slice6", []cadence.Value{cadence.Array{
				Values: []cadence.Value{cadence.Address{
					243,
					252,
					210,
					193,
					167,
					143,
					94,
					238,
				}},
			}}),
			wantErr: autogold.Want("err6", nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("testing")
			code := []byte(fmt.Sprintf("transaction(%s){}", tt.args.code))
			got, got1, err := o.parseArguments("somefile", code, tt.args.inputArgs)

			tt.want.Equal(t, got)
			tt.want1.Equal(t, got1)
			if err != nil {
				tt.wantErr.Equal(t, err.Error())
			} else {
				tt.wantErr.Equal(t, nil)
			}
		})
	}
}
