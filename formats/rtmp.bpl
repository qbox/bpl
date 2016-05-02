uint24be = {
    b3 uint8
    b2 uint8
    b1 uint8
    return (b3 << 16) | (b2 << 8) | b1
}

uint32be = {
    b4 uint8
    b3 uint8
    b2 uint8
    b1 uint8
    return (b4 << 24) | (b3 << 16) | (b2 << 8) | b1
}


Chunk = {
    tag byte
    let cfmt = (tag >> 6) & 3
    let csid = tag & 0x3f
    if csid == 0 {
        _v byte
        let csid = _v + 0x40
    } elif csid == 1 {
        _v uint16
        let csid = _v + 0x40
    }
    if cfmt < 3 {
        ts uint24be
        if cfmt == 0 {
            mlen   uint24be
            typeid byte
            strmid uint32
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
    do _msg.write(data)
}

init = {
    global msgs = mkmap("int:var")
}

doc = init *(Chunk dump)
