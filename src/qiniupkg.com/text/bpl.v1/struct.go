package bpl

import (
	"fmt"
	"reflect"
	"strings"

	"qiniupkg.com/text/bpl.v1/bufio"
	"qiniupkg.com/x/log.v7"
)

// -----------------------------------------------------------------------------

// A Member is typeinfo of a `Struct` member.
//
type Member struct {
	Name string
	Type Ruler
}

// Match is required by a matching unit. see Ruler interface.
//
func (p *Member) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	v, err = p.Type.Match(in, ctx.NewSub())
	if err != nil {
		return
	}
	if p.Name != "_" {
		ctx.SetVar(p.Name, v)
	}
	return
}

// SizeOf is required by a matching unit. see Ruler interface.
//
func (p *Member) SizeOf() int {

	return p.Type.SizeOf()
}

// -----------------------------------------------------------------------------

type structType struct {
	rulers []Ruler
	retFn  func(ctx *Context) (v interface{}, err error)
	size   int
}

func (p *structType) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	for _, r := range p.rulers {
		_, err = r.Match(in, ctx)
		if err != nil {
			return
		}
	}
	if p.retFn != nil {
		return p.retFn(ctx)
	}
	return ctx.Dom(), nil
}

func (p *structType) SizeOf() int {

	if p.size == -2 {
		p.size = p.sizeof()
	}
	return p.size
}

func (p *structType) sizeof() int {

	if p.retFn != nil {
		return -1
	}

	size := 0
	for _, r := range p.rulers {
		if n := r.SizeOf(); n < 0 {
			size = -1
			break
		} else {
			size += n
		}
	}
	return size
}

// Struct returns a compound matching unit.
//
func Struct(members []Member) Ruler {

	n := len(members)
	if n == 0 {
		return Nil
	}

	rulers := make([]Ruler, len(members))
	for i := range members {
		rulers[i] = &members[i]
	}
	return &structType{rulers: rulers, size: -2}
}

// StructEx returns a compound matching unit.
//
func StructEx(rulers []Ruler, retFn func(ctx *Context) (v interface{}, err error)) Ruler {

	n := len(rulers)
	if n == 0 && retFn == nil {
		return Nil
	}

	return &structType{rulers: rulers, size: -2, retFn: retFn}
}

// -----------------------------------------------------------------------------

func structFrom(t reflect.Type) (r Ruler, err error) {

	n := t.NumField()
	members := make([]Member, n)
	for i := 0; i < n; i++ {
		sf := t.Field(i)
		r, err = TypeFrom(sf.Type)
		if err != nil {
			log.Warn("bpl.TypeFrom failed:", err)
			return
		}
		members[i] = Member{Name: strings.ToLower(sf.Name), Type: r}
	}
	return Struct(members), nil
}

// TypeFrom creates a matching unit from a Go type.
//
func TypeFrom(t reflect.Type) (r Ruler, err error) {

retry:
	kind := t.Kind()
	switch {
	case kind == reflect.Struct:
		return structFrom(t)
	case kind >= reflect.Int8 && kind <= reflect.Float64:
		return BaseType(kind), nil
	case kind == reflect.String:
		return CString, nil
	case kind == reflect.Ptr:
		t = t.Elem()
		goto retry
	}
	return nil, fmt.Errorf("bpl.TypeFrom: unsupported type - %v", t)
}

// -----------------------------------------------------------------------------
