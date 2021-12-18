package vm

import "testing"

func TestBuildBuiltinContract(t *testing.T) {
	tc := new(token)
	//tc.BuiltinContract
	code, err := BuildBuiltinContract(tc)

	if err != nil {
		t.Fatal(err)
	}
	_ = code

}
