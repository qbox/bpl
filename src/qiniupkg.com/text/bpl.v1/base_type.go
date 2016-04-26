package bpl

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"reflect"
	"unsafe"

	"qiniupkg.com/text/bpl.v1/bufio"
	"qiniupkg.com/x/log.v7"
)

// -----------------------------------------------------------------------------

// A BaseType represents a matching unit of a builtin fixed size type.
//
type BaseType uint

type baseTypeInfo struct {
	read   func(in *bufio.Reader) (v interface{}, err error)
	newn   func(n int) interface{}
	sizeOf int
}

var baseTypes = [...]baseTypeInfo{
	reflect.Int8:    {readInt8, newInt8n, 1},
	reflect.Int16:   {readInt16, newInt16n, 2},
	reflect.Int32:   {readInt32, newInt32n, 4},
	reflect.Int64:   {readInt64, newInt64n, 8},
	reflect.Uint8:   {readUint8, newUint8n, 1},
	reflect.Uint16:  {readUint16, newUint16n, 2},
	reflect.Uint32:  {readUint32, newUint32n, 4},
	reflect.Uint64:  {readUint64, newUint64n, 8},
	reflect.Float32: {readFloat32, newFloat32n, 4},
	reflect.Float64: {readFloat64, newFloat64n, 8},
}

func readInt8(in *bufio.Reader) (v interface{}, err error) {

	t, err := in.ReadByte()
	return int8(t), err
}

func readUint8(in *bufio.Reader) (v interface{}, err error) {

	v, err = in.ReadByte()
	return
}

func readInt16(in *bufio.Reader) (v interface{}, err error) {

	t1, err := in.ReadByte()
	if err != nil {
		return
	}
	t2, err := in.ReadByte()
	return (int16(t2) << 8) | int16(t1), err
}

func readUint16(in *bufio.Reader) (v interface{}, err error) {

	t1, err := in.ReadByte()
	if err != nil {
		return
	}
	t2, err := in.ReadByte()
	return (uint16(t2) << 8) | uint16(t1), err
}

func readInt32(in *bufio.Reader) (v interface{}, err error) {

	t, err := in.Peek(4)
	if err != nil {
		return
	}
	v = int32(binary.LittleEndian.Uint32(t))
	in.Discard(4)
	return
}

func readUint32(in *bufio.Reader) (v interface{}, err error) {

	t, err := in.Peek(4)
	if err != nil {
		return
	}
	v = binary.LittleEndian.Uint32(t)
	in.Discard(4)
	return
}

func readInt64(in *bufio.Reader) (v interface{}, err error) {

	t, err := in.Peek(8)
	if err != nil {
		return
	}
	v = int64(binary.LittleEndian.Uint64(t))
	in.Discard(8)
	return
}

func readUint64(in *bufio.Reader) (v interface{}, err error) {

	t, err := in.Peek(8)
	if err != nil {
		return
	}
	v = binary.LittleEndian.Uint64(t)
	in.Discard(8)
	return
}

func readFloat32(in *bufio.Reader) (v interface{}, err error) {

	t, err := in.Peek(4)
	if err != nil {
		return
	}
	v = *(*float32)(unsafe.Pointer(&t[0]))
	in.Discard(4)
	return
}

func readFloat64(in *bufio.Reader) (v interface{}, err error) {

	t, err := in.Peek(8)
	if err != nil {
		return
	}
	v = *(*float64)(unsafe.Pointer(&t[0]))
	in.Discard(8)
	return
}

func newInt8n(n int) interface{} {

	return make([]int8, n)
}

func newUint8n(n int) interface{} {

	return make([]uint8, n)
}

func newInt16n(n int) interface{} {

	return make([]int16, n)
}

func newUint16n(n int) interface{} {

	return make([]uint16, n)
}

func newInt32n(n int) interface{} {

	return make([]int32, n)
}

func newUint32n(n int) interface{} {

	return make([]uint32, n)
}

func newInt64n(n int) interface{} {

	return make([]int64, n)
}

func newUint64n(n int) interface{} {

	return make([]uint64, n)
}

func newFloat32n(n int) interface{} {

	return make([]float32, n)
}

func newFloat64n(n int) interface{} {

	return make([]float64, n)
}

// Match is required by a matching unit. see Ruler interface.
//
func (p BaseType) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	v, err = baseTypes[p].read(in)
	return
}

// SizeOf is required by a matching unit. see Ruler interface.
//
func (p BaseType) SizeOf() int {

	return baseTypes[p].sizeOf
}

var (
	// Int8 is the matching unit for int8
	Int8 = BaseType(reflect.Int8)

	// Int16 is the matching unit for int16
	Int16 = BaseType(reflect.Int16)

	// Int32 is the matching unit for int32
	Int32 = BaseType(reflect.Int32)

	// Int64 is the matching unit for int64
	Int64 = BaseType(reflect.Int64)

	// Uint8 is the matching unit for uint8
	Uint8 = BaseType(reflect.Uint8)

	// Uint16 is the matching unit for uint16
	Uint16 = BaseType(reflect.Uint16)

	// Uint32 is the matching unit for uint32
	Uint32 = BaseType(reflect.Uint32)

	// Uint64 is the matching unit for uint64
	Uint64 = BaseType(reflect.Uint64)

	// Float32 is the matching unit for float32
	Float32 = BaseType(reflect.Float32)

	// Float64 is the matching unit for float64
	Float64 = BaseType(reflect.Float64)
)

// -----------------------------------------------------------------------------

type nilType int

func (p nilType) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	return nil, nil
}

func (p nilType) SizeOf() int {

	return 0
}

// Nil is a matching unit that matches zero bytes.
//
var Nil Ruler = nilType(0)

// -----------------------------------------------------------------------------

// A String represents result of a string matching unit, such as `CString`.
//
type String struct {
	Data  []byte
	cache string
}

func (p *String) String() string {

	if p.cache == "" {
		p.cache = string(p.Data)
	}
	return p.cache
}

// MarshalJSON is required by json.Marshal
//
func (p *String) MarshalJSON() (b []byte, err error) {

	return json.Marshal(p.String())
}

type cstring int

func (p cstring) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	b, err := in.ReadBytes(0)
	if err != nil {
		return
	}
	return &String{Data: b[:len(b)-1]}, nil
}

func (p cstring) SizeOf() int {

	return -1
}

// CString is a matching unit that matches a C style string.
//
var CString Ruler = cstring(0)

// -----------------------------------------------------------------------------

type fixedType struct {
	typ reflect.Type
}

func (p *fixedType) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	typ := p.typ
	size := typ.Size()
	val := reflect.New(typ)
	b := (*[1 << 30]byte)(unsafe.Pointer(val.UnsafeAddr()))
	_, err = io.ReadFull(in, b[:size])
	if err != nil {
		log.Warn("fixedType.Match: io.ReadFull failed -", err)
		return
	}
	return val.Interface(), nil
}

func (p *fixedType) SizeOf() int {

	return int(p.typ.Size())
}

// FixedType returns a matching unit that matches a C style fixed size struct.
//
func FixedType(t reflect.Type) Ruler {

	return &fixedType{typ: t}
}

// -----------------------------------------------------------------------------
