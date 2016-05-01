package bpl

import (
	"qlang.io/qlang.spec.v1"
	"qlang.io/qlang/bytes"

	// import qlang builtin
	_ "qlang.io/qlang/builtin"
)

// -----------------------------------------------------------------------------

func init() {

	qlang.Import("bytes", bytes.Exports)
}

// -----------------------------------------------------------------------------
