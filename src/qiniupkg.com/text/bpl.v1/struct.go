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

	v, err = p.Type.Match(in, NewSubContext(ctx))
	if err != nil {
		return
	}
	if p.Name != "_" {
		if ctx != nil {
			ctx.SetVar(p.Name, v)
		}
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
	members []Member
	size    int
}

func (p *structType) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	for _, m := range p.members {
		_, err = m.Match(in, ctx)
		if err != nil {
			return
		}
	}
	return domOf(ctx), nil
}

func (p *structType) SizeOf() int {

	return p.size
}

// Struct returns a compound matching unit.
//
func Struct(members []Member) Ruler {

	n := len(members)
	if n == 0 {
		return Nil
	}

	size := 0
	for _, m := range members {
		if n := m.Type.SizeOf(); n < 0 {
			size = -1
			break
		} else {
			size += n
		}
	}
	return &structType{members: members, size: size}
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
