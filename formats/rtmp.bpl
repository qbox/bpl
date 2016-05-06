init = {
    global msgs = mkmap("int:var")
    global chunksize = 128
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

ChunkHeaderF0 = {
    _v [11]byte

    let timestamp = (_v[0] << 16) | (_v[1] << 8) | (_v[2])
    let length = (_v[3] << 16) | (_v[4] << 8) | (_v[5])
    let typeid = (_v[6])
    let streamid = (_v[7] << 24) | (_v[8] << 16) | (_v[9] << 8) | (_v[10])
}

ChunkHeaderF1 = {
    _v [7]byte

    let timedelta = (_v[0] << 16) | (_v[1] << 8) | (_v[2])
    let length = (_v[3] << 16) | (_v[4] << 8) | (_v[5])
    let typeid = (_v[6])
}

ChunkHeaderF2 = {
    _v [3]byte

    let timedelta = (_v[0] << 16) | (_v[1] << 8) | (_v[2])
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
                streamid uint23be
            } else {
                let ts = ts + _last["ts"]
                let streamid = _last["streamid"]
            }
        } else {
            let ts = ts + _last["ts"]
            let length = _last["length"]
            let typeid = _last["typeid"]
            let streamid = _last["streamid"]
        }
    } else {
        let ts = _last["ts"]
        let length = _last["length"]
        let typeid = _last["typeid"]
        let streamid = _last["streamid"]
    }
}

Chunk = {
    header ChunkHeader

    let _length = chunksize
    if header.length < _length {
        let _length = header.length
    }

    let _header = {
        "ts": header.ts,
        "length": header.length - _length,
        "typeid": header.typeid,
        "streamid": header.streamid,
    }
    do set(msgs, header.csid, _header)

    read _length do {
        data []byte
    }

    if header.csid == 2 && header.streamid == 0 && header.typeid == 1 {
        let chunksize = (data[0] << 24) | (data[1] << 16) | (data[2] << 8) | data[3]
    }
}

doc = init (Handshake0 dump) Handshake1 Handshake2 *(Chunk dump)
