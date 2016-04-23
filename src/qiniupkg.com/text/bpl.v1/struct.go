package bpl

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"syscall"

	"qiniupkg.com/text/bpl.v1/bufio"
	"qiniupkg.com/x/log.v7"
)

// -----------------------------------------------------------------------------

type memberGroup interface {
	member(name string) (v interface{}, err error)
	all(vars map[string]interface{}) (err error)
}

type memberGroupType interface {
	match(in *bufio.Reader, ctx *Context) (g memberGroup, err error)
}

// -----------------------------------------------------------------------------

type blockMember struct {
	name string
	typ  Ruler
	off  int
}

type block struct {
	data    []byte
	members []blockMember
}

func (p *block) member(name string) (v interface{}, err error) {

	for _, m := range p.members {
		if m.name == name {
			in := bufio.NewReaderBuffer(p.data[m.off:])
			return m.typ.Match(in, nil)
		}
	}
	return nil, syscall.ENOENT
}

func (p *block) all(vars map[string]interface{}) error {

	in := bufio.NewReaderBuffer(p.data)
	for _, m := range p.members {
		v, err := m.typ.Match(in, nil)
		if err != nil {
			log.Warnf("block.member `%s` Match failed: %v\n", m.name, err)
			return err
		}
		vars[m.name] = v
	}
	return nil
}

// -----------------------------------------------------------------------------

type blockType struct {
	members []blockMember
	size    int
}

func (p *blockType) match(in *bufio.Reader, ctx *Context) (g memberGroup, err error) {

	b := make([]byte, p.size)
	_, err = io.ReadFull(in, b)
	if err != nil {
		log.Warn("io.ReadFull failed:", err)
		return
	}
	ctx.CaptureIf(b)
	return &block{data: b, members: p.members}, nil
}

// -----------------------------------------------------------------------------

type namedVar struct {
	name string
	data interface{}
}

type namedVars struct {
	vars []namedVar
}

func (p *namedVars) member(name string) (v interface{}, err error) {

	for _, m := range p.vars {
		if m.name == name {
			return m.data, nil
		}
	}
	return nil, syscall.ENOENT
}

func (p *namedVars) all(vars map[string]interface{}) error {

	for _, m := range p.vars {
		vars[m.name] = m.data
	}
	return nil
}

// -----------------------------------------------------------------------------

type namedType struct {
	name string
	typ  Ruler
}

type namedTypes struct {
	members []namedType
}

func (p *namedTypes) match(in *bufio.Reader, ctx *Context) (g memberGroup, err error) {

	vars := make([]namedVar, len(p.members))
	for i, m := range p.members {
		v, err := m.typ.Match(in, ctx)
		if err != nil {
			return nil, err
		}
		vars[i] = namedVar{name: m.name, data: v}
	}
	return &namedVars{vars: vars}, nil
}

// -----------------------------------------------------------------------------

// A Object represents result of a `Struct` matching unit.
//
type Object struct {
	gs    []memberGroup
	cache map[string]interface{}
}

// Var gets a variable from this object.
//
func (p *Object) Var(name string) (v interface{}, ok bool) {

	for _, g := range p.gs {
		if val, err := g.member(name); err != syscall.ENOENT {
			return val, (err == nil)
		}
	}
	return
}

// Vars returns all variables of this object.
//
func (p *Object) Vars() map[string]interface{} {

	if p.cache == nil {
		cache := make(map[string]interface{})
		for _, g := range p.gs {
			err := g.all(cache)
			if err != nil {
				panic(err)
			}
		}
		p.cache = cache
	}
	return p.cache
}

// MarshalJSON is required by `json.Marshal`.
//
func (p *Object) MarshalJSON() (b []byte, err error) {

	return json.Marshal(p.Vars())
}

// -----------------------------------------------------------------------------

type structType struct {
	gs   []memberGroupType
	size int
}

func (p *structType) Match(in *bufio.Reader, ctx *Context) (interface{}, error) {

	gs := make([]memberGroup, len(p.gs))
	for i, g := range p.gs {
		m, err := g.match(in, ctx)
		if err != nil {
			return nil, err
		}
		gs[i] = m
	}
	return &Object{gs: gs}, nil
}

func (p *structType) SizeOf() int {

	return p.size
}

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

	v, err = p.Type.Match(in, ctx)
	if err != nil {
		return
	}
	if ctx != nil {
		ctx.vars[p.Name] = v
	}
	return
}

// SizeOf is required by a matching unit. see Ruler interface.
//
func (p *Member) SizeOf() int {

	return p.Type.SizeOf()
}

// -----------------------------------------------------------------------------

// Struct returns a compound matching unit.
//
func Struct(members []Member) Ruler {

	n := len(members)
	if n == 0 {
		return Nil
	}

	i := 0
	off := 0
	base := members[0]
	baseSize := base.Type.SizeOf()
	isVariadic := false

	var gs []memberGroupType
	for i < n {
		j := i + 1
		if baseSize < 0 {
			items := []namedType{
				{name: base.Name, typ: base.Type},
			}
			for j < n {
				m := members[j]
				size := m.Type.SizeOf()
				if size >= 0 {
					base, baseSize = m, size
					break
				}
				items = append(items, namedType{name: m.Name, typ: m.Type})
				j++
			}
			gs = append(gs, &namedTypes{members: items})
			isVariadic = true
		} else {
			items := []blockMember{
				{name: base.Name, typ: base.Type, off: 0},
			}
			off = baseSize
			for j < n {
				m := members[j]
				size := m.Type.SizeOf()
				if size < 0 {
					base, baseSize = m, size
					break
				}
				items = append(items, blockMember{name: m.Name, typ: m.Type, off: off})
				off += size
				j++
			}
			gs = append(gs, &blockType{members: items, size: off})
		}
		i = j
	}

	if isVariadic {
		off = -1
	}
	return &structType{gs: gs, size: off}
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
