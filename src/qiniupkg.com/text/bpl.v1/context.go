package bpl

import (
	"bytes"

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

// A Context represents the matching context of bpl.
//
type Context struct {
	vars map[string]interface{}

	// Capture is optional when matching.
	Capture *bytes.Buffer
}

// NewContext returns a new Context.
//
func NewContext() *Context {

	vars := make(map[string]interface{})
	return &Context{vars: vars}
}

// SetVar sets a new variable to matching context.
//
func (p *Context) SetVar(name string, v interface{}) {

	p.vars[name] = v
}

// Var gets a variable from matching context.
//
func (p *Context) Var(name string) (v interface{}, ok bool) {

	v, ok = p.vars[name]
	return
}

// Vars returns all variables in matching context.
//
func (p *Context) Vars() map[string]interface{} {

	return p.vars
}

// -----------------------------------------------------------------------------

// A Ruler interface is required to a matching unit.
//
type Ruler interface {
	// Match matches input stream `in`, and returns matching result.
	Match(in *bufio.Reader, ctx *Context) (v interface{}, err error)

	// SizeOf returns expected length of result. If length is variadic, it returns -1.
	SizeOf() int
}

// -----------------------------------------------------------------------------
