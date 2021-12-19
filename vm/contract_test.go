package vm

import (
	"testing"
)

func TestBuildBuiltinContract(t *testing.T) {
	tc := new(token)
	//tc.BuiltinContract
	code, err := BuildBuiltinContract(tc)

	if err != nil {
		t.Fatal(err)
	}
	_ = code

}
func createToken(
	name CTypeString,
	symbol CTypeString,
	decimals CTypeUINT8,
	totalSupply CTypeUINT256) {

}
func TestBuildBuiltinContract2(t *testing.T) {
	//t.Logf("%s", n.Name())

}
