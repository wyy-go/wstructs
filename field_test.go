package wstructs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// A test struct that defines all cases
type Foo struct {
	A string
	B int    `map:"y"`
	C bool   `json:"c"`
	d string // not exported
	E *Baz
	// not exported, with tag
	x    string `xml:"x"` // nolint: govet
	Y    []string
	Z    map[string]interface{}
	*Bar // embedded
}

type Baz struct {
	A string
	B int
}

type Bar struct {
	E string
	F int
	g []string
}

func newStruct() *Struct {
	// B and x is not initialized for testing
	return New(&Foo{
		A: "gopher",
		C: true,
		d: "small",
		E: nil,
		x: "",
		Y: []string{"example"},
		Z: nil,
		Bar: &Bar{
			E: "example",
			F: 2,
			g: []string{"zeynep", "fatih"},
		},
	})
}

func TestField_Set(t *testing.T) {
	s := newStruct()

	f := s.MustField("A")
	require.NoError(t, f.Set("fatih"))
	require.Equal(t, "fatih", f.Value())
	// let's pass a different type
	require.Error(t, f.Set(123))         // Field A is of type string, but we are going to pass an integer
	require.Equal(t, "fatih", f.Value()) // old value should be still there :)

	f = s.MustField("Y")
	require.NoError(t, f.Set([]string{"override", "with", "this"}))
	require.Len(t, f.Value(), 3)

	f = s.MustField("C")
	require.NoError(t, f.Set(false))
	require.Equal(t, false, f.Value())

	// let's access an unexported field, which should give an error
	f = s.MustField("d")
	require.ErrorIs(t, errNotExported, f.Set("large"))

	// let's set a pointer to struct
	bar := &Bar{
		E: "gopher",
		F: 2,
	}
	f = s.MustField("Bar")
	require.NoError(t, f.Set(bar))

	baz := &Baz{
		A: "helloWorld",
		B: 42,
	}
	f = s.MustField("E")
	require.NoError(t, f.Set(baz))

	ba := s.MustField("E").Value().(*Baz)
	require.Equal(t, "helloWorld", ba.A)
}

func TestField_CanInterface(t *testing.T) {
	s := newStruct()

	f := s.MustField("A")
	require.True(t, f.CanInterface())
}

func TestField_NotSettable(t *testing.T) {
	a := map[int]Baz{
		4: {A: "value"},
	}

	s := New(a[4])
	require.False(t, s.MustField("A").CanSet())
	require.ErrorIs(t, errNotSettable, s.MustField("A").Set("newValue"))
}

func TestField_Zero(t *testing.T) {
	s := newStruct()

	f := s.MustField("A")
	require.NoError(t, f.SetZero())
	require.Empty(t, f.Value())

	f = s.MustField("Y")
	require.NoError(t, f.SetZero())
	require.Empty(t, f.Value())

	f = s.MustField("C")
	require.NoError(t, f.SetZero())
	require.Empty(t, f.Value())

	// let's access an unexported field, which should give an error
	f = s.MustField("d")
	require.ErrorIs(t, errNotExported, f.SetZero())

	f = s.MustField("Bar")
	require.NoError(t, f.SetZero())
	require.Empty(t, f.Value())

	f = s.MustField("E")
	require.NoError(t, f.SetZero())
	require.Empty(t, f.Value())

	v := s.MustField("E").Value()
	require.Empty(t, v)
}

func TestField(t *testing.T) {
	s := newStruct()

	require.Panics(t, func() {
		_ = s.MustField("no-field")
	})
}

func TestField_Kind(t *testing.T) {
	s := newStruct()

	f := s.MustField("A")
	require.Equal(t, reflect.String, f.Kind())

	f = s.MustField("B")
	require.Equal(t, reflect.Int, f.Kind())

	// unexported
	f = s.MustField("d")
	require.Equal(t, reflect.String, f.Kind())
}

func TestField_Tag(t *testing.T) {
	s := newStruct()

	require.Empty(t, s.MustField("B").Tag("json"))
	require.Equal(t, "c", s.MustField("C").Tag("json"))
	require.Empty(t, s.MustField("d").Tag("json"))
	require.Equal(t, "x", s.MustField("x").Tag("xml"))
	require.Empty(t, s.MustField("A").Tag("json"))
}

func TestField_Value(t *testing.T) {
	s := newStruct()

	v := s.MustField("A").Value()
	val, ok := v.(string)
	require.True(t, ok)
	require.Equal(t, "gopher", val)

	require.Panics(t, func() {
		// should panic
		_ = s.MustField("d").Value()
	})
}

func TestField_IsAnonymous(t *testing.T) {
	s := newStruct()

	require.True(t, s.MustField("Bar").IsAnonymous())
	require.False(t, s.MustField("d").IsAnonymous())
}

func TestField_IsExported(t *testing.T) {
	s := newStruct()

	require.True(t, s.MustField("Bar").IsExported())
	require.True(t, s.MustField("A").IsExported())
	require.False(t, s.MustField("d").IsExported())
}

func TestField_IsZero(t *testing.T) {
	s := newStruct()

	require.False(t, s.MustField("A").IsZero())
	require.True(t, s.MustField("B").IsZero())
}

func TestField_Name(t *testing.T) {
	s := newStruct()

	require.Equal(t, "A", s.MustField("A").Name())
}

func TestField_MustField(t *testing.T) {
	s := newStruct()

	e := s.MustField("Bar").MustField("E")

	val, ok := e.Value().(string)
	require.True(t, ok)
	require.Equal(t, "example", val)

	require.Panics(t, func() {
		_ = s.MustField("Bar").MustField("e")
	})
}

func TestField_Fields(t *testing.T) {
	s := newStruct()

	require.Len(t, s.MustField("Bar").Fields(), 3)
}

func TestField_Field(t *testing.T) {
	s := newStruct()

	b, ok := s.Field("Bar")
	require.True(t, ok)

	// field A not exist
	a, ok := b.Field("A")
	require.False(t, ok)
	require.Nil(t, a)

	// field E exist
	e, ok := b.Field("E")
	require.True(t, ok)
	val, ok := e.Value().(string)
	require.True(t, ok)
	require.Equal(t, "example", val)

	// field e not a struct
	x, ok := e.Field("X")
	require.False(t, ok)
	require.Nil(t, x)
}
