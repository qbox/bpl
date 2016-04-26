BPL - Binary Processing Language
============

## 基础规则

其实也就是内建类型 (builtin types)，如下：

* int8, uint8(byte), int16, uint16
* int32, uint32, int64, uint64
* float32, float64
* cstring
* bson
* nil


## 复合规则

* `*R`: 反复匹配规则 R，直到无法成功匹配为止。
* `+R`: 反复匹配规则 R，直到无法成功匹配为止。要求至少匹配成功1次。
* `?R`: 匹配规则 R 1次或0次。
* `R1 R2 ... Rn`: 要求要匹配的文本满足规则序列 R1 R2 ... Rn。


## 别名

如果规则太复杂并且要出现在多个地方，那么我们就可以定义下别名。例如：

```
R = R1 R2 ... Rn
```

这里 R 就是规则序列 `R1 R2 ... Rn` 的别名。


## 结构体

结构体本质上和 `R1 R2 ... Rn` 有些类似，属于规则序列，但是它给每个子规则都有命名，如下：

```
{
	Var1 R1
	Var2 R2
	...
	Varn Rn
}
```

以上是 Go 风格的结构体。我们也可以改成 C 风格的：

```
{/C
	R1 Var1
	R2 Var2
	...
	Rn Varn
}
```

在结构体里面出现的规则 R，有一些特殊性，如下：

1) 规则不能太复杂，如果复杂应该定义别名。例如以下是非法的：

```
{
	Var (R1 R2 ... Rn)
}
```

我们不建议写这样不清晰的东西，而是先给 `R1 R2 ... Rn` 定义别名：

```
R = R1 R2 ... Rn
```

然后再引用 R：

```
{
	Var R
}
```

2) 以尽可能符合常理为原则，我们设置了规则可以是 `R`, `?R`, `*R`, `+R`, `[len]R` 这样一些情况。当然在 C 风格中应该是 `R`, `R?`, `R*`, `R+`, `R[len]` 这种结构。

3) 需要注意的一个细节是，`[len]R` 这样的规则当前只能在结构体里面出现。


## 捕获

如果我们希望生成 DOM，那么我们就需要去捕获感兴趣的数据。例如：

```
doc = R1 R2 R3 ... Rn
```

对于这样一段数据，如果我们感兴趣 R2，那么可以：

```
doc = R1 #R2 R3 ... Rn
```

但是如果你感兴趣多个元素，你不能多处使用 `#` 来捕获。下面样例不能正常工作：

```
doc = R1 #R2 R3 ... #Rn
```

因为这段规则的匹配过程是这样的：

* 在匹配 `#R2` 成功后，我们把 ctx.dom = `<R2的匹配结果>`。
* 在匹配 `#Rn` 成功后，我们试图把 ctx.dom 修改为 `<Rn的匹配结果>` 但是失败，因为 ctx.dom 已经赋值。

如果我们对多个匹配的结果感兴趣，那么我们需要写成：

```
doc = R1 [R2] R3 ... [Rn]
```

得到的结果是 `[<R2的匹配结果>, <Rn的匹配结果>]`。

如果我们希望结果不是数组而是对象(object)。那么应该用结构体。例如：

```
doc = R1 {var2 R2; var3 R3} ... {varn Rn}
```

那么结果就是 `{"var2": <R2的匹配结果>, "var3": <R3的匹配结果>, "varn": <Rn的匹配结果>}`。

如果我们这样去捕获：

```
Rx = [R2 R3]

doc = R1 {varx Rx} ... {varn Rn}
```

那么结果就是 `{"varx": [<R2的匹配结果>, <R3的匹配结果>], "varn": <Rn的匹配结果>}`。


## 条件规则

```
case <expr> {
<val1>: R1
<val2>: R2
...
<valn>: Rn
default: Rdefault // 如果没有 default 并且前面各个分支都没有匹配成功，那么整个条件规则匹配会失败
}
```

典型例子：

```
header = {type uint32; ...}

body1 = {...}

body2 = {...}

// 每个记录有个 header，header里面有个记录type，不同记录有不同的body
//
record = {h header} case h.type {type1: body1; type2: body2; ...}

doc = *record
```

另外条件规则也可以出现在结构体中。如：

```
record = {
	h header
	case h.type {
	type1: body1
	type2: body2
	...
	}
}
```

## read..do

```
read <nbytes> do R
```

读 `<nbytes>` 字节的内容后，再用 R 匹配这段内容。如：

```
record = {
	h header
	read h.len - sizeof(header) do case h.type {
		type1: body1
		type2: body2
		...
	}
}
```

## 样例：MongoDB 网络协议

```
document = bson

MsgHeader = {/C
    int32   messageLength; // total message size, including this
    int32   requestID;     // identifier for this message
    int32   responseTo;    // requestID from the original request (used in responses from db)
    int32   opCode;        // request type - see table below
}

OP_UPDATE = {/C
	int32     ZERO;               // 0 - reserved for future use
	cstring   fullCollectionName; // "dbname.collectionname"
	int32     flags;              // bit vector. see below
	document  selector;           // the query to select the document
	document  update;             // specification of the update to perform
}

OP_INSERT = {/C
	int32      flags;              // bit vector - see below
	cstring    fullCollectionName; // "dbname.collectionname"
	document*  documents;          // one or more documents to insert into the collection
}

OP_QUERY = {/C
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
	int32     ZERO;               // 0 - reserved for future use
	cstring   fullCollectionName; // "dbname.collectionname"
	int32     numberToReturn;     // number of documents to return
	int64     cursorID;           // cursorID from the OP_REPLY
}

OP_DELETE = {/C
	int32     ZERO;               // 0 - reserved for future use
	cstring   fullCollectionName; // "dbname.collectionname"
	int32     flags;              // bit vector - see below for details.
	document  selector;           // query object.  See below for details.
}

OP_KILL_CURSORS = {/C
	int32     ZERO;              // 0 - reserved for future use
	int32     numberOfCursorIDs; // number of cursorIDs in message
	int64*    cursorIDs;         // sequence of cursorIDs to close
}

OP_MSG = {/C
	cstring   message; // message for the database
}

OP_REPLY = {/C
	int32     responseFlags;  // bit vector - see details below
	int64     cursorID;       // cursor id if client needs to do get more's
	int32     startingFrom;   // where in the cursor this reply is starting
	int32     numberReturned; // number of documents in the reply
	document* documents;      // documents
}

Message = {
	header MsgHeader   // standard message header
	read header.messageLength - sizeof(MsgHeader) do case header.opCode {
		1:    OP_REPLY    // Reply to a client request. responseTo is set.
		1000: OP_MSG      // Generic msg command followed by a string.
		2001: OP_UPDATE
		2002: OP_INSERT
		2003: RESERVED
		2004: OP_QUERY
		2005: OP_GET_MORE // Get more data from a query. See Cursors.
		2006: OP_DELETE
		2007: OP_KILL_CURSORS // Notify database that the client has finished with the cursor.
	}
}

doc = *Message
```
