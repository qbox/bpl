BPL - Binary Processing Language
============

## 快速入门

了解 BPL 最快的方式是学习 qbpl 和 qbplproxy 两个实用程序：

### qbpl

qbpl 可用来分析任意的文件格式。使用方法如下：

```
qbpl [-p <protocol>.bpl -o <output>.log] <file>
```

多数情况下，你不需要自己指定格式，我们根据文件后缀来确定应该使用何种 protocol 来解析这个文件。例如：

```
qbpl 1.gif
```

不过为了让 qbpl 能够找到所有的 protocols，我们需要先安装：

```
make install # 这将将所有的bpl文件拷贝到 ~/.qbpl/formats/
```

### qbplproxy

qbplproxy 可用来分析服务器和客户端之间的网络包。它通过代理要分析的服务，让客户端请求自己来分析请求包和返回包。使用方式如下：

```
qbplproxy -h <listenIp:port> -b <backendIp:port> -p <protocol>.bpl [-o <output>.log]
```

其中，`<listenIp:port>` 是 qbplproxy 自身监听的IP和端口，`<backendIp:port>` 是原始的服务。


## 基础规则

其实也就是内建类型 (builtin types)，如下：

* int8, char, uint8(byte), int16, uint16
* int32, uint32, int64, uint64
* float32, float64
* cstring, `[n]char`
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
doc = R1 $R2 R3 ... Rn
```

但是如果你感兴趣多个元素，你不能多处使用 `$` 来捕获。下面样例不能正常工作：

```
doc = R1 $R2 R3 ... $Rn
```

因为这段规则的匹配过程是这样的：

* 在匹配 `$R2` 成功后，我们把 ctx.dom = `<R2的匹配结果>`。
* 在匹配 `$Rn` 成功后，我们试图把 ctx.dom 修改为 `<Rn的匹配结果>` 但是失败，因为 ctx.dom 已经赋值。

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


## dump

dump 规则不匹配任何内容，但是会打印当前 Context 中已经捕获的所有变量值。如：

doc = *(record dump)

这样每个 record 匹配成功后会 dump 匹配结果。如果希望某个变量不进行 dump，则该变量需要以 _ 开头。

## case

```
case <expr> {
	<val1>: R1
	<val2>: R2
	...
	<valn>: Rn
	default: Rdefault // 如果没有 default 并且前面各个分支都没有匹配成功，那么整个规则匹配会失败
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

另外条件规则也可以出现在结构体中（下文大部分规则除非特殊说明，一般都可以同时出现在规则列表和结构体）。如：

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

## if..elif..else

```
if <condition1> do R1 elif <condition2> do R2 ... else Rn
```

对 `<condition1>` 进行求值，如果结果为 true 或非零整数则执行 R1 规则，以此类推。如果 R1 是结构体 { ... }，则 do 可以忽略。例如：

```
record = {
	len uint32
	if len {
		data [len]byte
		next record
	}
}
```

## assert

```
assert <condition>
```

对 `<condition>` 进行求值，如果结果为 true 或非零整数表示成功，其他情况均失败。

## read..do

```
read <nbytes> do R
```

这里 `<nbytes>` 是一个 qlang 表达式。对 `<nbytes>` 求值，读如相应字节数的内容后，再用 R 匹配这段内容。如：

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

## eval..do

```
eval <expr> do R
```

对 `<expr>` 进行求值（要求求值结果为[]byte类型）后，再用 R 匹配它。如：

```
record = {
	h    header
	body [h.len - sizeof(header)]byte
	eval body do case h.type {
		type1: body1
		type2: body2
		...
	}
}
```

## let

```
let <var> = <expr>
```

这将给当前 Context 添加一个名为 `<var>` 的变量，其值为 `<expr>` 的求值结果。如：

```
record = {
	let a = 1
	let _b = 2  // 变量 _b 不会被 dump
}
```

## return

return 语句只能出现在结构体中，用来改写结构体的匹配结果。如：

```
uint24be = {
    b3 uint8
    b2 uint8
    b1 uint8
    return (b3 << 16) | (b2 << 8) | b1
}
```

在正常情况下，以上 uint24be 的匹配结果应该是 `{"b1": <val1>, "b2": <val2>, "b3": <val3>}`，但是由于 return 语句的存在，其匹配结果变成返回一个整数。类似地我们可以有：

```
uint32be = {
    b4 uint8
    b3 uint8
    b2 uint8
    b1 uint8
    return (b4 << 24) | (b3 << 16) | (b2 << 8) | b1
}
```

## global

```
global <var> = <expr>
```

global 语句用于定义全局变量，这些变量并不出现在 Context 的捕获结果中。如：

```
init = {
	global msgs = mkmap("int:var")
}
```

## do

```
do <expr>
```

do 语句用来执行一个表达式。如：

```
init = {
	global msgs = {"a": 12, "b": 32}
}

record = {
	do set(msgs, "c", 56, "d", 78) // 现在 msgs = {"a": 12, "b": 32, "c": 56, "d": 78}
	let a = msgs["a"]
}

doc = init record
```

## 常量

```
const (
	<ident> = <constvalue>
	...
)
```

语法和 Go 语言类似。例如：

```
const (
	N = 6
)

record = {
	tag [N]char
    assert tag == "GIF87a" || tag == "GIF89a"
}
```

## qlang 表达式

bpl 集成了 qlang 表达式（不包含赋值）。以上所有 `<expr>`、`<condition>`、`<nbytes>` 这些地方，都是 bpl 引用 qlang 表达式的地方。

bpl 中的 qlang 表达式支持如下这些特性：

* 所有 qlang 操作符；
* string、slice、map 等内置类型；
* 变量、成员变量引用；
* 函数、成员函数调用；
* 模块（但是我们很克制地支持了非常有限的几个模块，如：builtin、bytes 等）；


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
	cstring   fullCollectionName;     // "dbname.collectionname"
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
	body   [header.messageLength - sizeof(MsgHeader)]byte
	eval body do case header.opCode {
		1:    OP_REPLY    // Reply to a client request. responseTo is set.
		1000: OP_MSG      // Generic msg command followed by a string.
		2001: OP_UPDATE
		2002: OP_INSERT
		2004: OP_QUERY
		2005: OP_GET_MORE // Get more data from a query. See Cursors.
		2006: OP_DELETE
		2007: OP_KILL_CURSORS // Notify database that the client has finished with the cursor.
		default: nil
	}
}

doc = *Message
```
