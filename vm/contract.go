package vm

import (
	"fmt"
	"reflect"
	"xfsgo/common"
	"xfsgo/core"
)

type ContractPayload interface {
	Create()
}
type BuiltinContract interface {
	GetStateTree() core.StateTree
	GetAddress() common.Address
	Storage([32]byte, []byte)
	StorageFiled(string, []byte)
	BuiltinId() uint8
}

//type BuiltinContract interface {
//	GetStateTree() core.StateTree
//	GetAddress() common.Address
//	Storage([32]byte, []byte)
//	StorageFiled(string, []byte)
//}

//
func BuildBuiltinContract(c BuiltinContract) ([]byte, error) {
	a := reflect.TypeOf(c).Elem()
	_ = a
	n := a.NumField()
	_ = n
	pkgpa := a.PkgPath()
	fmt.Printf("pkgpa: %s\n", pkgpa)
	for i := 0; i < a.NumField(); i++ {
		sf := a.Field(i)
		aname := sf.Name
		atype := sf.Type
		fmt.Printf("filed: name=%s, type=%s\n", aname, atype.Kind().String())
	}
	fmt.Println("---------------")
	b := reflect.TypeOf(c)

	pkgpb := b.PkgPath()
	fmt.Printf("pkgpb: %s\n", pkgpb)
	for i := 0; i < b.NumMethod(); i++ {
		sf := b.Method(i)
		aname := sf.Name
		atype := sf.Type
		if sf.Type.Kind() == reflect.Func {
			fmt.Printf("method: name=%s, type=%s\n", aname, atype.Kind().String())
			//at := reflect.ValueOf(atype)
			ni := atype.NumIn()
			fmt.Printf("method: paramsLen=%v\n", ni)
		}
	}
	return nil, nil
}
