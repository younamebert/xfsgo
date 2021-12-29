package vm

import (
	"bytes"
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
		if err = p.call(m, mv, input); err != nil {
			return
		}
		//for i := 0; i < cve.NumField(); i++ {
		//
		//	cvef := cve.Field(i)
		//	cvef.Type()
		//	fmt.Printf("cvf: %s\n", cvef)
		//}
		for i := 0; i < cte.NumField(); i++ {
			ctef := cte.Field(i)
			c := ctef.Tag.Get(contractTag)
			if c != contractStorage {
				continue
				//fvaluebs := fvalue.Bytes()
				//fmt.Printf("name: %s, hash: %x, value: %x\n", ctef.Name, nameHash[:], fvaluebs[:])
				//p.stateTree.SetState(p.address, nameHash[:], )
			}
			nameHash := ahash.SHA256([]byte(ctef.Name))
			_ = cve
			fvalue := cve.FieldByName(ctef.Name)
			switch ctef.Type {
			case reflect.TypeOf(CTypeUint8(0)),
				reflect.TypeOf(CTypeUint16{}),
				reflect.TypeOf(CTypeUint32{}),
				reflect.TypeOf(CTypeUint64{}),
				reflect.TypeOf(CTypeUint256{}),
				reflect.TypeOf(CTypeString{}),
				reflect.TypeOf(CTypeAddress{}):
				fmt.Printf("name: %s, hash: %x, type: %v\n", ctef.Name, nameHash[:], ctef.Type)
			default:
				switch ctef.Type.Kind() {
				case reflect.Map:
					for _, k := range fvalue.MapKeys() {
						tt :=
						k.
							v := fvalue.MapIndex(k)
						fmt.Printf("name: %s, hash: %x, type: %v, val(k): %s, val(v): %x\n", ctef.Name, nameHash[:], ctef.Type, k.Bytes(), v)
						//key := k.Convert(fvalue.Type().Key()) //.Convert(m.Type().Key())
						//value := fvalue.MapIndex(k)
						//value.Type()
						//switch t := value.Interface().(type) {
						//case CTypeAddress:
						//	fmt.Printf("name: %s, hash: %x, type: %v, val: %s\n", ctef.Name, nameHash[:], ctef.Type, t.address())
						//}
					}

				default:
					return errUnsupportedType
				}
			}
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
