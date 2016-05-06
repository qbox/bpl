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
        "ts": header.ts,
        "length": header.length,
        "typeid": header.typeid,
        "streamid": header.streamid,
        "remain": header.remain - _length,
        "body": header._body,
    }
    do set(msgs, header.csid, _header)

    do {
        data [_length]byte
    }
    do header._body.write(data)
    if _header.remain == 0 {
        eval header._body.bytes() do {
            msg Msg
        }
    }

    if header.csid == 2 && header.streamid == 0 {
        case header.typeid {
            1: let chunksize = (data[0] << 24) | (data[1] << 16) | (data[2] << 8) | data[3]
            2: {
                let _csid = (data[0] << 24) | (data[1] << 16) | (data[2] << 8) | data[3]
                let _last = msgs[_csid]
                do set(_last, "remain", 0)
            }
        }

    }
}

doc = init Handshake0 dump Handshake1 Handshake2 *(Chunk dump)
