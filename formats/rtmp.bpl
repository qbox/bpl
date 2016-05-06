init = {
    global msgs = mkmap("int:var")
}

Chunk = {
    tag byte
    let cfmt = (tag >> 6) & 3
    let csid = tag & 0x3f
    if csid == 0 {
        _v byte
        let csid = _v + 0x40
    } elif csid == 1 {
        _v uint16le
        let csid = _v + 0x40
    }
    if cfmt < 3 {
        ts uint24be
        if cfmt == 0 {
            mlen   uint24be
            typeid byte
            strmid uint32le
            let left = mlen
            let _msg = {mlen: mlen, data: bytes.buffer()}
            do set(msgs, csid, _msg)
        } elif cfmt == 1 {
            mlen   uint24be
            typeid byte
            let left = mlen
            let _msg = {mlen: mlen, data: bytes.buffer()}
            do set(msgs, csid, _msg)
        } else {
            let _msg = msgs[csid]
            let left = _msg.mlen - _msg.data.len()
        }
        if ts == 0xffffff {
            _dw uint32be
            let ts = _dw
        }
    } else {
        let _msg = msgs[csid]
        let left = _msg.mlen - _msg.data.len()
    }
    if left <= 128 {
        data [left]byte
    } else {
        data [128]byte
    }
    do _msg.data.write(data)
    let left = left - len(data)
}

Msg = {
}

MsgChunk = Chunk {
    if left == 0 {
        eval _msg.data.bytes() do {
            msg Msg
        }
    }
}

HandshakeMsg = {
}

HandshakeChunks = Chunk {
    if left {
        next HandshakeChunks
    } else {
        eval _msg.data.bytes() do {
            handshake HandshakeMsg
        }
    }
}

doc = init HandshakeChunks *(MsgChunk dump)
