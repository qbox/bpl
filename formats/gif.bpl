const (
    fColorTable         = 0x80
    fColorTableBitsMask = 7
)

ColorTable = {
    colortable [(1 << (1 + (fields&fColorTableBitsMask))) * 3]byte
}

Header = {
    tag             [6]char
    width           int16
    height          int16
    fields          uint8
    backgroundIndex uint8
    tmp             uint8
    assert tag == "GIF87a" || tag == "GIF89a"
    case fields & fColorTable {
        0: nil
        default: ColorTable
    }
}

sTrailer = nil

ImageHeader = {
    left   int16
    top    int16
    width  int16
    height int16
    fields byte
    case fields & fColorTable {
        0: nil
        default: ColorTable
    }
}

sImage = {
    h        ImageHeader
    litWidth byte
    assert litWidth >= 2 && litWidth <= 8
    lzw 0, litWidth do {
        _ [h.width*h.height]byte
        _ eof
    }
}

eText = {
    text [13]char
}

eGraphicControl = {
    _                byte
    flags            byte
    delayTime        int16
    transparentIndex byte
    _                byte
}

eComment = nil

eApplication = {
    len  byte
    name [len]char
}

ExtHeader = {
    etag byte
    case etag {
        0x01: eText
        0xF9: eGraphicControl
        0xFE: eComment
        0xFF: eApplication
    }
}

ExtBlocks = {
    len byte
    case len {
        0: nil
        default: {
            data [len]byte
            next ExtBlocks
        }
    }
}

sExtension = {
    h      ExtHeader
    blocks ExtBlocks
}

Record = {
    tag byte
    case tag {
        0x21: sExtension
        0x2C: sImage
        0x3B: sTrailer
    }
}

doc = *(Record dump)
