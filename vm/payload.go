package vm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"xfsgo/common"
	"xfsgo/common/ahash"
	"xfsgo/core"
)

type payload interface {
	Create([]byte) error
	Call([]byte) error
}

var (
	errNotfoundCreateFn = errors.New("notfound create function")
	errNotfoundMethod   = errors.New("notfound method")
	errUnsupportedType  = errors.New("unsupported type")
)

const contractTag = "contract"
const contractStorage = "storage"

type xvmPayload struct {
	code      []byte
	stateTree core.StateTree
	address   common.Address
	contract  BuiltinContract
	resultBuf Buffer
}

func (p *xvmPayload) goReturn(vs []reflect.Value) error {
	for i := 0; i < len(vs); i++ {
		if vs[i].IsNil() {
			continue
		}
		if err, ok := vs[i].Interface().(error); ok {
			return err
		}
		_, _ = p.resultBuf.Write(vs[i].Bytes())
	}
	return nil
}
func (p *xvmPayload) call(fn reflect.Method, fnv reflect.Value, input []byte) error {
	buf := NewBuffer(input)
	mType := fn.Type
	n := mType.NumIn()

	var args = make([]reflect.Value, 0)
	for i := 1; i < n; i++ {
		parameterType := mType.In(i)
		switch parameterType.Name() {
		case "CTypeString":
			ssize, err := buf.ReadUint32()
			if err != nil {
				return err
			}
			s, err := buf.ReadString(int(ssize.uint32()))
			if err != nil {
				return err
			}
			args = append(args, reflect.ValueOf(s))
		case "CTypeUint8":
			m, err := buf.ReadUint8()
			if err != nil {
				return err
			}
			args = append(args, reflect.ValueOf(m))
		case "CTypeUint256":
			m, err := buf.ReadUint256()
			if err != nil {
				return err
			}
			args = append(args, reflect.ValueOf(m))
		}
	}
	r := fnv.Call(args)
	return p.goReturn(r)
}

type stv struct {
	reflect.StructField
	nameHash [32]byte
	val      reflect.Value
}

func findContractStorageValue(cte reflect.Type, cve reflect.Value) []*stv {
	stvs := make([]*stv, 0)
	for i := 0; i < cte.NumField(); i++ {
		ctef := cte.Field(i)
		c := ctef.Tag.Get(contractTag)
		if c != contractStorage {
			continue
		}
		nameHash := ahash.SHA256Array([]byte(ctef.Name))
		fvalue := cve.FieldByName(ctef.Name)
		if !fvalue.CanInterface() {
			continue
		}
		stvs = append(stvs, &stv{
			StructField: ctef,
			nameHash:    nameHash,
			val:         fvalue,
		})
	}
	return stvs
}
func (p *xvmPayload) setupContract(stvs []*stv) (err error) {
	for i := 0; i < len(stvs); i++ {
		st := stvs[i]
		data := p.stateTree.GetStateValue(p.address, st.nameHash)
		if data == nil {
			continue
		}
		switch st.Type {
		case reflect.TypeOf(CTypeString{}):
			cs := CTypeString{}
			if err = json.Unmarshal(data, &cs); err != nil {
				return
			}
			st.val.Set(reflect.ValueOf(cs))
		case reflect.TypeOf(CTypeUint8(0)):
			cs := CTypeUint8(0)
			if err = json.Unmarshal(data, &cs); err != nil {
				return
			}
			st.val.Set(reflect.ValueOf(cs))
		case reflect.TypeOf(CTypeUint256{}):
			cs := CTypeUint256{}
			if err = json.Unmarshal(data, &cs); err != nil {
				return
			}
			st.val.Set(reflect.ValueOf(cs))
		}
		fmt.Printf("name: %s, hash: %x, type: %v, val: %s\n", st.Name, st.nameHash[:], st.Type, "nil")
	}
	fmt.Println()
	//fmt.Printf("name: %s, hash: %x, type: %v, val: %s\n", st.Name, st.nameHash[:], st.Type, "nil")
	return
}

func (p *xvmPayload) updateContractState(stvs []*stv) (err error) {
	for i := 0; i < len(stvs); i++ {
		st := stvs[i]
		fvalue := st.val
		if !fvalue.CanInterface() {
			continue
		}
		jb, err := json.Marshal(fvalue.Interface())
		if err != nil {
			return err
		}
		p.stateTree.SetState(p.address, st.nameHash, jb)
		fmt.Printf("name: %s, hash: %x, type: %v, val: %s\n", st.Name, st.nameHash[:], st.Type, string(jb))
	}
	return
}
func (p *xvmPayload) callFn(fn common.Hash, input []byte) (err error) {
	ct := reflect.TypeOf(p.contract)
	cv := reflect.ValueOf(p.contract)
	cte := ct.Elem()
	cve := cv.Elem()
	findMethod := func(hash common.Hash) (reflect.Method, reflect.Value, bool) {
		for i := 0; i < ct.NumMethod(); i++ {
			sf := ct.Method(i)
			aname := sf.Name

			namehash := ahash.SHA256([]byte(aname))
			if sf.Type.Kind() == reflect.Func && bytes.Equal(hash[:], namehash) {
				mv := cv.MethodByName(aname)
				return sf, mv, true
			}
		}
		return reflect.Method{}, reflect.Value{}, false
	}
	if m, mv, ok := findMethod(fn); ok {
		stvs := findContractStorageValue(cte, cve)
		if err = p.setupContract(stvs); err != nil {
			return
		}
		if err = p.call(m, mv, input); err != nil {
			return
		}
		if err = p.updateContractState(stvs); err != nil {
			return
		}
		return
	}
	return errNotfoundMethod
}

func readCallMethod(r io.Reader) (m common.Hash, e error) {
	var hashdata [32]byte
	n, e := r.Read(hashdata[:])
	if e != nil {
		return common.Hash{}, e
	}
	if n != len(hashdata) {
		return common.Hash{}, errors.New("eof")
	}
	copy(m[:], hashdata[:])
	return
}

func (p *xvmPayload) exec(input []byte) error {
	buf := bytes.NewBuffer(input)
	fn, err := readCallMethod(buf)
	if err != nil {
		return err
	}
	return p.callFn(fn, buf.Bytes())
}
func (p *xvmPayload) Call(input []byte) error {
	return p.exec(input)
}

func (p *xvmPayload) Create(input []byte) error {
	return p.exec(input)
}
func makeKey(k []byte) (key [32]byte) {
	keyhash := ahash.SHA256(k)
	copy(key[:], keyhash)
	return
}
