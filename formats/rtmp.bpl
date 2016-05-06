init = {
    global msgs = mkmap("int:var")
    global chunksize = 128
}

Msg = {
}

Handshake0 = {
    h0 byte
}

Handshake1 = {
    h1 [1536]byte
}

Handshake2 = {
    h2 [1536]byte
}

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
            let remain = length
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
    }
}

Chunk = {
    header ChunkHeader

    let _length = chunksize
    if header.remain < _length {
        let _length = header.remain
    }

    if _last == undefined {
        let _msg = bytes.buffer()
    } else {
        let _msg = _last["msg"]
    }

    let _header = {
        "ts": header.ts,
        "length": header.length,
        "typeid": header.typeid,
        "streamid": header.streamid,
        "remain": header.remain - _length,
        "msg": _msg,
    }
    do set(msgs, header.csid, _header)
    do {
        data [_length]byte
    }
    do _msg.write(data)
    if _header.remain == 0 {
        eval _msg.bytes() do {
            msg Msg
        }
    }

    if header.csid == 2 && header.streamid == 0 && header.typeid == 1 {
        let chunksize = (data[0] << 24) | (data[1] << 16) | (data[2] << 8) | data[3]
    }
    if header.csid == 2 && header.streamid == 0 && header.typeid == 2 {
        let _csid = (data[0] << 24) | (data[1] << 16) | (data[2] << 8) | data[3]
        let _last = msgs[_csid]
        let _newLast = {
            "ts": _last["ts"],
            "length": _last["length"],
            "typeid": _last["typeid"],
            "streamid": _last["streamid"],
            "remain": 0,
            "msg": _last["msg"],
        }
        do set(msgs, _csid, _newLast)
    }
}

doc = init Handshake0 dump Handshake1 Handshake2 *(Chunk dump)
