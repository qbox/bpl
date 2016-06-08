//
// http://blog.sina.com.cn/s/blog_48f93b530100jz4b.html

box = {
	size  uint32be
	assert size != 1 // largesize: not impl

	typ   [4]char
	_body [size - 8]byte
}

boxtr = {
	let body = _body
}

fixed16be = {
	_v uint16be
	return float64(_v) / 0x100
}

fixed32be = {
	_v uint32be
	return float64(_v) / 0x10000
}

// --------------------------------------------------------------

vmhd = {
	body *byte
}

smhd = {
	body *byte
}

hmhd = {
	body *byte
}

nmhd = {
	body *byte
}

dinf = {
	body *byte
}

stbl = {
	body *byte
}

mibox = box {
	eval _body do case typ {
		"vmhd": vmhd
		"smhd": smhd
		"hmhd": hmhd
		"nmhd": nmhd
		"dinf": dinf
		"stbl": stbl
		default: boxtr
	}
}

// Media Information Box
//
minf = {
	items *mibox
}

// --------------------------------------------------------------

mdhd = {
	version byte
	flags   uint24be

	ctime uint32be  // 创建时间（相对于UTC时间1904-01-01零点的秒数）
	mtime uint32be  // 修改时间

	time_scale uint32be // 文件媒体在1秒时间内的刻度值，可以理解为1秒长度的时间单元数
	duration   uint32be // 该track的时间长度，用duration和time_scale值可以计算track时长

	language   uint16be
	predefined uint16be
}

// Handler Reference Box
//
hdlr = {
	version byte
	flags   uint24be

	predefined uint32be

	// 在media box中，该值为4个字符：
	// “vide”— video track
	// “soun”— audio track
	// “hint”— hint track
	//
	handler_typ [4]char

	reserved [12]byte
	name     cstring // track type name，以‘\0’结尾的字符串
}

mdbox = box {
	eval _body do case typ {
		"mdhd": mdhd
		"hdlr": hdlr
		"minf": minf
		default: boxtr
	}
}

// Media Box
//
mdia = {
	items *mdbox
}

// --------------------------------------------------------------

// Track Header Box
//
tkhd = {
	version byte

	// 按位或操作结果值，预定义如下：
	// 0x000001 track_enabled，否则该track不被播放；
	// 0x000002 track_in_movie，表示该track在播放中被引用；
	// 0x000004 track_in_preview，表示该track在预览时被引用。
	// 一般该值为7，如果一个媒体所有track均未设置track_in_movie和track_in_preview，将被理解为所有track均设置了这两项；
	// 对于hint track，该值为0
	//
	flags uint24be

	ctime uint32be  // 创建时间（相对于UTC时间1904-01-01零点的秒数）
	mtime uint32be  // 修改时间

	track_id uint32be // track id号，不能重复且不能为0

	reserved  uint32be
	duration  uint32be
	reserved2 uint64be

	layer     uint16be  // 视频层，默认为0，值小的在上层
	alt_group uint16be  // alternate group: track分组信息，默认为0表示该track未与其他track有群组关系
	volume    fixed16be // [8.8] 格式，1.0（0x0100）表示最大音量。如果为音频track，该值有效，否则为0
	reserved3 uint16be

	matrix [36]byte  // 视频变换矩阵
	width  fixed32be // 宽
	height fixed32be // 高，均为 [16.16] 格式值，与sample描述中的实际画面大小比值，用于播放时的展示宽高
}

edts = {
	body *byte
}

trkbox = box {
	eval _body do case typ {
		"tkhd": tkhd
		"mdia": mdia
		"edts": edts
		default: boxtr
	}
}

// Track Box
//
trak = {
	items *trkbox
}

// --------------------------------------------------------------

// Movie Header Box
//
mvhd = {
	version    byte
	flags      uint24be
	ctime      uint32be  // 创建时间（相对于UTC时间1904-01-01零点的秒数）
	mtime      uint32be  // 修改时间
	time_scale uint32be  // 文件媒体在1秒时间内的刻度值，可以理解为1秒长度的时间单元数
	duration   uint32be  // 该track的时间长度，用duration和time_scale值可以计算track时长
	rate       fixed32be // 推荐播放速率，高16位和低16位分别为小数点整数部分和小数部分，即[16.16] 格式，该值为1.0（0x00010000）表示正常前向播放
	volume     fixed16be // 与rate类似，[8.8] 格式，1.0（0x0100）表示最大音量
	reserved   [10]byte
	matrix     [36]byte  // 视频变换矩阵
	predefined [24]byte

	// 下一个track使用的id号
	//
	next_track_id uint32be
}

movbox = box {
	eval _body do case typ {
		"mvhd": mvhd
		"trak": trak
		default: boxtr
	}
	dump
}

// Movie Box
//
moov = dump *movbox

// --------------------------------------------------------------

// File Type Box
//
ftyp = {
	major_brand       [4]char
	minor_version     uint32be
	compatible_brands *[4]char
}

free = {
	body *byte
}

mdat = {
	body *byte
}

gblbox = box {
	//
	// 注意：moov 可能太大了，如果等解析完再 dump 不太友好
	//
	eval _body do case typ {
		"ftyp": ftyp dump
		"moov": moov
		"mdat": mdat dump
		default: boxtr dump
	}
}

doc = *gblbox

// --------------------------------------------------------------
