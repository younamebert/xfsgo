package vm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"xfsgo/common"
	"xfsgo/common/ahash"
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
	//a := reflect.TypeOf(c).Elem()
	//_ = a
	//n := a.NumField()
	//_ = n
	//pkgpa := a.PkgPath()
	//fmt.Printf("pkgpa: %s\n", pkgpa)
	//for i := 0; i < a.NumField(); i++ {
	//	sf := a.Field(i)
	//	aname := sf.Name
	//	atype := sf.Type
	//	fmt.Printf("filed: name=%s, type=%s\n", aname, atype.Kind().String())
	//}
	//fmt.Println("---------------")
	b := reflect.TypeOf(c)

	pkgpb := b.PkgPath()
	fmt.Printf("pkgpb: %s\n", pkgpb)
	id := c.BuiltinId()
	buf := bytes.NewBuffer(nil)
	var mar [4]byte
	binary.LittleEndian.PutUint32(mar[:], magicNumberXVM)
	buf.Write(mar[:])
	buf.Write([]byte{id})
	for i := 0; i < b.NumMethod(); i++ {
		sf := b.Method(i)
		aname := sf.Name
		atype := sf.Type
		if sf.Type.Kind() == reflect.Func && aname == "Create" {
			anamehash := ahash.SHA256([]byte(aname))
			buf.Write(anamehash[:])
			fmt.Printf("method: name=%s, hash=%x\n", aname, anamehash)
			//at := reflect.ValueOf(atype)
			ni := atype.NumIn()
			//fmt.Printf("method: paramsLen=%v\n", ni)
			for j := 1; j < ni; j++ {
				ft := atype.In(j)
				ftname := ft.Name()
				ftkind := ft.Kind()

				//switch ftkind {
				//
				//}
				fmt.Printf("method(%s): ftname=%s, type=%s, as=%s\n", aname, ftname, ftkind, "---")
			}
		}
	}
	fmt.Printf("ali: %x\n", buf.Bytes())
	return nil, nil
}
