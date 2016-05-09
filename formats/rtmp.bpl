const (
    VERBOSE = 0
)

init = {
    global msgs = mkmap("int:var")
    global chunksize = 128
    global objectend = errors.new("object end")
    global limitTypes = {
        0: "Hard",
        1: "Soft",
        2: "Dynamic",
    }

    global audioFormats = {
        0:  "Linear PCM",
        1:  "ADPCM",
        2:  "MP3",
        3:  "Linear PCM le",
        4:  "Nellymoser 16kHz",
        5:  "Nellymoser 8kHz",
        6:  "Nellymoser",
        7:  "G711 A-law",
        8:  "G711 mu-law",
        9:  "Reserved",
        10: "AAC",
        11: "Speex",
        12: "MP3 8kHz",
        13: "Device Specific",
    }
    global audioRates = ["5.5 kHz", "11 kHz", "22 kHz", "44 kHz"]
    global audioBits = ["8 bits", "16 bits"]
    global audioChannels = ["Mono", "Stereo"]

    global videoTypes = {
        1: "AVC keyframe",
        2: "AVC inter frame",
        3: "H.263 disposable inter frame",
        4: "generated keyframe",
        5: "video info/command frame",
    }
    global videoCodecs = {
        1: "JPEG",
        2: "H.263",
        3: "screen video",
        4: "On2 VP6",
        5: "On2 VP6 with alpha channel",
        6: "screen video v2",
        7: "AVC",
    }
    global avcTypes = {
        0: "sequence header",
        1: "NALU",
        2: "end of sequence",
    }
}

// --------------------------------------------------------------

AMF0_NUMBER = {
    val float64be
    if VERBOSE == 0 {
        return val
    }
}

AMF0_BOOLEAN = {
    val byte
    if VERBOSE == 0 {
        return byte != 0
    }
}

AMF0_STRING = {
    len uint16be
    val [len]char
    if VERBOSE == 0 {
        return val
    }
}

AMF0_OBJECT_ITEMS = {
    _key AMF0_STRING
    _val AMF0_TYPE
    let items = mkslice("var", 2)
    do set(items, 0, _key, 1, _val)
    if _val != objectend {
        _next AMF0_OBJECT_ITEMS
        let items = append(items, _next.items...)
    }
}

AMF0_OBJECT_NORMAL = {
    val AMF0_OBJECT_ITEMS
    let n = len(val.items)
    return mapFrom(val.items[:n-2]...) // 去掉了最后的 objectend
}

AMF0_OBJECT_VERBOSE = {
    _key AMF0_STRING
    _val AMF0_TYPE
    let items = [{"key": _key, "val": _val}]
    if _val.marker != 0x09 {
        _next AMF0_OBJECT_VERBOSE
        let items = append(items, _next.items...)
    }
}

AMF0_OBJECT = if VERBOSE do AMF0_OBJECT_VERBOSE else AMF0_OBJECT_NORMAL

AMF0_STRICT_ARRAY = {
    len uint32be
    body *byte
    assert false
}

AMF0_MOVIECLIP = {
    body *byte
    assert false
}

AMF0_NULL = {
    if VERBOSE {
        let val = nil
    } else {
        return nil
    }
}

AMF0_UNDEFINED = {
    if VERBOSE {
        let val = undefined
    } else {
        return undefined
    }
}

AMF0_REFERENCE = {
    reference uint16be
}

AMF0_ECMA_ARRAY = {
    len uint32be
    val AMF0_OBJECT
    if VERBOSE == 0 {
        return val
    }
}

AMF0_OBJECT_END = if VERBOSE do nil else {
    return objectend
}

AMF0_DATE = {
    timestamp float64be
    tz        uint16be
}

AMF0_LONG_STRING = {
    len uint32be
    val [len]char
    if VERBOSE == 0 {
        return val
    }
}

AMF0_UNSUPPORTED = {
    body *byte
    assert false
}

AMF0_RECORDSET = {
    body *byte
    assert false
}

AMF0_XML_DOCUMENT = AMF0_LONG_STRING

AMF0_TYPED_OBJECT = {
    type AMF0_STRING
    val  AMF0_OBJECT
}

AMF0_ACMPLUS_OBJECT = { // Switch to AMF3
    body *byte
    assert false
}

AMF0_TYPE = {
    marker byte
    case marker {
        0x00: AMF0_NUMBER
        0x01: AMF0_BOOLEAN
        0x02: AMF0_STRING
        0x03: AMF0_OBJECT
        0x04: AMF0_MOVIECLIP
        0x05: AMF0_NULL
        0x06: AMF0_UNDEFINED
        0x07: AMF0_REFERENCE
        0x08: AMF0_ECMA_ARRAY
        0x09: AMF0_OBJECT_END
        0x0a: AMF0_STRICT_ARRAY
        0x0b: AMF0_DATE
        0x0c: AMF0_LONG_STRING
        0x0d: AMF0_UNSUPPORTED
        0x0e: AMF0_RECORDSET
        0x0f: AMF0_XML_DOCUMENT
        0x10: AMF0_TYPED_OBJECT
        0x11: AMF0_ACMPLUS_OBJECT
    }
}

AMF0_CMDDATA = {
    cmd           AMF0_TYPE
    transactionId AMF0_TYPE
    value         *AMF0_TYPE
}

AMF0 = {
    msg *AMF0_TYPE
}

AMF0_CMD = {
    msg AMF0_CMDDATA
}

// --------------------------------------------------------------

AMF3_UNDEFINED = AMF0_UNDEFINED

AMF3_NULL = AMF0_NULL

AMF3_FALSE = {
    if VERBOSE {
        let val = false
    } else {
        return false
    }
}

AMF3_TRUE = {
    if VERBOSE {
        let val = true
    } else {
        return true
    }
}

AMF3_INT = {
    b1 byte
    if b1 & 0x80 {
        let b1 = b1 & 0x7f
        b2 byte
        if b2 & 0x80 {
            let b2 = b2 & 0x7f
            b3 byte
            if b3 & 0x80 {
                let b3 = b3 & 0x7f
                b4 byte
                return (b1 << 22) | (b2 << 15) | (b3 << 8) | b4
            } else {
                return (b1 << 14) | (b2 << 7) | b3
            }
        } else {
            return (b1 << 7) | b2
        }
    } else {
        return b1
    }
}

AMF3_INTEGER_VERBOSE = {
    val AMF3_INT
}

AMF3_INTEGER = if VERBOSE do AMF3_INTEGER_VERBOSE else AMF3_INT

AMF3_DOUBLE = {
    val float64be
    if VERBOSE == 0 {
        return val
    }
}

AMF3_STRING = {
    tag AMF3_INT
    assert (tag & 1) != 0 // reference unsupported
    if tag & 1 {
        val [tag >> 1]char
    }
    if VERBOSE == 0 {
        return val
    }
}

AMF3_XMLDOC = {
    body *byte
}

AMF3_DATE = {
    tag AMF3_INT
    assert (tag & 1) != 0 // reference unsupported
    timestamp float64be
    let tz = tag >> 1
    if VERBOSE == 0 {
        do unset("tag")
    }
}

AMF3_ARRAY = {
    tag AMF3_INT
    assert (tag & 1) != 0 // reference unsupported
    let len = tag >> 1
    body *byte
    assert false
}

AMF3_OBJECT = {
    body *byte
    assert false
}

AMF3_XML = {
    body *byte
    assert false
}

AMF3_BYTE_ARRAY = {
    body *byte
    assert false
}

AMF3_VECTOR_INT = {
    body *byte
    assert false
}

AMF3_VECTOR_UINT = {
    body *byte
    assert false
}

AMF3_VECTOR_DOUBLE = {
    body *byte
    assert false
}

AMF3_VECTOR_OBJECT = {
    body *byte
    assert false
}

AMF3_DICTIONARY = {
    body *byte
    assert false
}

AMF3_TYPE = {
    marker byte
    case marker {
        0x00: AMF3_UNDEFINED
        0x01: AMF3_NULL
        0x02: AMF3_FALSE
        0x03: AMF3_TRUE
        0x04: AMF3_INTEGER
        0x05: AMF3_DOUBLE
        0x06: AMF3_STRING
        0x07: AMF3_XMLDOC
        0x08: AMF3_DATE
        0x09: AMF3_ARRAY
        0x0a: AMF3_OBJECT
        0x0b: AMF3_XML
        0x0c: AMF3_BYTE_ARRAY
        0x0d: AMF3_VECTOR_INT
        0x0e: AMF3_VECTOR_UINT
        0x0f: AMF3_VECTOR_DOUBLE
        0x10: AMF3_VECTOR_OBJECT
        0x11: AMF3_DICTIONARY
    }
}

AMF3_CMDDATA = {
    cmd           AMF3_TYPE
    transactionId AMF3_TYPE
    value         *AMF3_TYPE
}

AMF3 = {
    msg AMF3_CMDDATA
}

AMF3_CMD = {
    msg AMF3_CMDDATA
}

// --------------------------------------------------------------

AudioData = {
    tag byte
    let format = tag >> 4
    let formatKind = audioFormats[format]
    let rate = (tag>>2) & 3
    let rateKind = audioRates[rate]
    let bits = (tag>>1) & 1
    let bitsKind = audioBits[bits]
    let channel = tag & 1
    let channelKind = audioChannels[channel]

    if formatKind == "AAC" {
        aacPacketType byte
        if aacPacketType == 0 {
            let aacPacketTypeKind = "sequence header"
        } else {
            let aacPacketTypeKind = "raw data"
        }
    }
    raw *byte
}

Audio = {
    audio AudioData
}

// --------------------------------------------------------------

VideoData = {
    tag byte
    let type = tag >> 4
    let typeKind = videoTypes[type]
    let codec = tag & 0xf
    let codecKind = videoCodecs[codec]

    if codecKind == "AVC" {
        avctype         byte
        compositionTime uint24be
        let avctypeKind = avcTypes[int(avctype)]
    }
    raw *byte
}

Video = {
    video VideoData
}

// --------------------------------------------------------------

SetChunkSize = {
    size uint32be
    let chunksize = size
}

Abort = {
    csid uint32be
    let _last = msgs[csid]
    do set(_last, "remain", 0)
}

UserControl = {
    evType uint16be
    evData *byte
}

AckWinsize = {
    winsize uint32be
}

SetPeerBandwidth = {
    winsize   uint32be
    limitType byte
    let limitTypeKind = limitTypes[int(limitType)]
}

// --------------------------------------------------------------

Handshake0 = {
    h0 byte
}

Handshake1 = {
    h1 [1536]byte
}

Handshake2 = {
    h2 [1536]byte
}

// --------------------------------------------------------------

ChunkHeader = {
    _tag byte

    let format = (_tag >> 6) & 3
    assert format <= 3

    let csid = _tag & 0x3f
    if csid == 0 {
        _v byte
        let csid = _v + 0x40
    } elif csid == 1 {
        _v uint16le
        let csid = _v + 0x40
    }

    let _last = msgs[csid]

    if format < 3 {
        ts uint24be
        if format < 2 {
            length uint24be
            typeid byte
            if format < 1 {
                streamid uint32le
            } else {
                let ts = ts + _last["ts"]
                let streamid = _last["streamid"]
            }
            let remain = 0
        } else {
            let ts = ts + _last["ts"]
            let length = _last["length"]
            let typeid = _last["typeid"]
            let streamid = _last["streamid"]
            let remain = _last["remain"]
        }
    } else {
        let ts = _last["ts"]
        let length = _last["length"]
        let typeid = _last["typeid"]
        let streamid = _last["streamid"]
        let remain = _last["remain"]
    }

    if remain == 0 {
        let remain = length
        let _body = bytes.buffer()
    } else {
        let _body = _last["body"]
    }
}

Chunk = {
    header ChunkHeader

    let _length = chunksize
    if header.remain < _length {
        let _length = header.remain
    }

    let _header = {
        "ts":       header.ts,
        "length":   header.length,
        "typeid":   header.typeid,
        "streamid": header.streamid,
        "remain":   header.remain - _length,
        "body":     header._body,
    }
    do set(msgs, header.csid, _header)

    data [_length]byte
    do header._body.write(data)
    if _length > 16 {
        let data = data[:16]
    }

    if _header.remain == 0 {
        let _body = header._body.bytes()
        if header.csid == 2 && header.streamid == 0 {
            eval _body do case header.typeid {
                1: SetChunkSize
                2: Abort
                4: UserControl
                5: AckWinsize
                6: SetPeerBandwidth
                default: let body = _body
            }
        } else {
            eval _body do case header.typeid {
                18: AMF0
                20: AMF0_CMD
                15: AMF3
                17: AMF3_CMD
                8:  Audio
                9:  Video
                default: let body = _body
            }
        }
    }
}

doc = init Handshake0 Handshake1 Handshake2 *(Chunk dump)

// --------------------------------------------------------------
