package bpl

import (
	"encoding/binary"
	"reflect"
	"unsafe"

	"qiniupkg.com/text/bpl.v1/bufio"
)

/* -----------------------------------------------------------------------------

builtin types:

* int8, uint8(byte), int16, uint16, int32, uint32, int64, uint64
* float32, float64, cstring, bson

document = bson

MsgHeader = {/C
    int32   messageLength; // total message size, including this
    int32   requestID;     // identifier for this message
    int32   responseTo;    // requestID from the original request (used in responses from db)
    int32   opCode;        // request type - see table below
}

OP_UPDATE = {/C
	MsgHeader header;             // standard message header
	int32     ZERO;               // 0 - reserved for future use
	cstring   fullCollectionName; // "dbname.collectionname"
	int32     flags;              // bit vector. see below
	document  selector;           // the query to select the document
	document  update;             // specification of the update to perform
}

OP_INSERT = {/C
	MsgHeader  header;             // standard message header
	int32      flags;              // bit vector - see below
	cstring    fullCollectionName; // "dbname.collectionname"
	document*  documents;          // one or more documents to insert into the collection
}

OP_QUERY = {/C
	MsgHeader header;                 // standard message header
	int32     flags;                  // bit vector of query options.  See below for details.
	cstring   fullCollectionName ;    // "dbname.collectionname"
	int32     numberToSkip;           // number of documents to skip
	int32     numberToReturn;         // number of documents to return
		                              //  in the first OP_REPLY batch
	document  query;                  // query object.  See below for details.
	document? returnFieldsSelector;   // Optional. Selector indicating the fields
		                              //  to return.  See below for details.
}

OP_GET_MORE = {/C
	MsgHeader header;             // standard message header
	int32     ZERO;               // 0 - reserved for future use
	cstring   fullCollectionName; // "dbname.collectionname"
	int32     numberToReturn;     // number of documents to return
	int64     cursorID;           // cursorID from the OP_REPLY
}

OP_DELETE = {/C
	MsgHeader header;             // standard message header
	int32     ZERO;               // 0 - reserved for future use
	cstring   fullCollectionName; // "dbname.collectionname"
	int32     flags;              // bit vector - see below for details.
	document  selector;           // query object.  See below for details.
}

OP_KILL_CURSORS = {/C
	MsgHeader header;            // standard message header
	int32     ZERO;              // 0 - reserved for future use
	int32     numberOfCursorIDs; // number of cursorIDs in message
	int64*    cursorIDs;         // sequence of cursorIDs to close
}

OP_MSG = {/C
	MsgHeader header;  // standard message header
	cstring   message; // message for the database
}

OP_REPLY = {/C
	MsgHeader header;         // standard message header
	int32     responseFlags;  // bit vector - see details below
	int64     cursorID;       // cursor id if client needs to do get more's
	int32     startingFrom;   // where in the cursor this reply is starting
	int32     numberReturned; // number of documents in the reply
	document* documents;      // documents
}

Message = {
	header MsgHeader   // standard message header
	data   [header.messageLength - sizeof(header)]byte
}/case header.opCode {
	1:    OP_REPLY,    // Reply to a client request. responseTo is set.
	1000: OP_MSG,      // Generic msg command followed by a string.
	2001: OP_UPDATE,
	2002: OP_INSERT,
	2003: RESERVED,
	2004: OP_QUERY,
	2005: OP_GET_MORE, // Get more data from a query. See Cursors.
	2006: OP_DELETE,
	2007: OP_KILL_CURSORS, // Notify database that the client has finished with the cursor.
}

doc = *Message

// ---------------------------------------------------------------------------*/

type Context struct {
	vars map[string]interface{}
}

func NewContext() *Context {

	vars := make(map[string]interface{})
	return &Context{vars: vars}
}

func (p *Context) SetVar(name string, v interface{}) {

	p.vars[name] = v
}

func (p *Context) Var(name string) (v interface{}, ok bool) {

	v, ok = p.vars[name]
	return
}

func (p *Context) Vars() map[string]interface{} {

	return p.vars
}

type Ruler interface {
	Match(in *bufio.Reader, ctx *Context) (v interface{}, err error)
	SizeOf() int
}

// -----------------------------------------------------------------------------

type BaseType uint

type baseTypeInfo struct {
	read   func(in *bufio.Reader) (v interface{}, err error)
	sizeOf int
}

var baseTypes = [...]baseTypeInfo{
	reflect.Int8:    {readInt8, 1},
	reflect.Int16:   {readInt16, 2},
	reflect.Int32:   {readInt32, 2},
	reflect.Int64:   {readInt64, 2},
	reflect.Uint8:   {readUint8, 2},
	reflect.Uint16:  {readUint16, 2},
	reflect.Uint32:  {readUint32, 2},
	reflect.Uint64:  {readUint64, 2},
	reflect.Float32: {readFloat32, 2},
	reflect.Float64: {readFloat64, 2},
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

func (p BaseType) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	v, err = baseTypes[p].read(in)
	return
}

func (p BaseType) SizeOf() int {

	return baseTypes[p].sizeOf
}

var (
	Int8    = BaseType(reflect.Int8)
	Int16   = BaseType(reflect.Int16)
	Int32   = BaseType(reflect.Int32)
	Int64   = BaseType(reflect.Int64)
	Uint8   = BaseType(reflect.Uint8)
	Uint16  = BaseType(reflect.Uint16)
	Uint32  = BaseType(reflect.Uint32)
	Uint64  = BaseType(reflect.Uint64)
	Float32 = BaseType(reflect.Float32)
	Float64 = BaseType(reflect.Float64)
)

// -----------------------------------------------------------------------------

type namedBaseType struct {
	Name string
	Type BaseType
}

func NamedBaseType(name string, typ BaseType) *namedBaseType {

	return &namedBaseType{Name: name, Type: typ}
}

func (p *namedBaseType) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	v, err = baseTypes[p.Type].read(in)
	if ctx.vars != nil {
		ctx.vars[p.Name] = v
	}
	return
}

func (p *namedBaseType) SizeOf() int {

	return baseTypes[p.Type].sizeOf
}

// -----------------------------------------------------------------------------

