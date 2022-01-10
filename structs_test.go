package wstructs

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type Animal struct {
	Name string
	Age  int
}

type Dog struct {
	Animal *Animal `json:"animal,string"` // nolint: staticcheck
}

type Person struct {
	Name string
	Age  int
}

func (p *Person) String() string {
	return fmt.Sprintf("%s(%d)", p.Name, p.Age)
}

func getMapKey(m map[string]interface{}) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

func getMapValue(m map[string]interface{}) []interface{} {
	vs := make([]interface{}, 0, len(m))
	for _, v := range m {
		vs = append(vs, v)
	}
	return vs
}

func TestMap(t *testing.T) {
	t.Run("NonStruct", func(t *testing.T) {
		// this should panic. We are going to recover and and test it
		require.Panics(t, func() {
			_ = Map([]string{"foo"})
		})
	})

	t.Run("mixed indexes", func(t *testing.T) {
		type C struct {
			// something int
			Props map[string]interface{}
		}

		require.NotPanics(t, func() {
			_ = Map(&C{})
			_ = Fields(&C{})
			_ = Values(&C{})
			_ = IsZero(&C{})
			_ = HasZero(&C{})
		})
	})

	t.Run("Normal", func(t *testing.T) {
		var T = struct {
			A string
			B int
			C bool
		}{
			A: "a-value",
			B: 2,
			C: true,
		}

		m := Map(T)
		require.Equal(t, reflect.Map, reflect.TypeOf(m).Kind())
		require.Len(t, m, 3)
		require.ElementsMatch(t, []interface{}{"a-value", 2, true}, getMapValue(m))
	})

	t.Run("Anonymous", func(t *testing.T) {
		type A struct {
			Name string
		}
		type B struct {
			*A
		}

		b := &B{A: &A{Name: "example"}}
		m := Map(b)
		require.Equal(t, reflect.Map, reflect.TypeOf(m).Kind())

		in, ok := m["A"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "example", in["Name"])
	})

	t.Run("TimeField", func(t *testing.T) {
		type A struct {
			CreatedAt time.Time
		}

		a := &A{CreatedAt: time.Now().UTC()}
		m := Map(a)

		_, ok := m["CreatedAt"].(time.Time)
		require.True(t, ok)
	})

	t.Run("tag", func(t *testing.T) {
		var T = struct {
			A string `map:"x"`
			B int    `map:"y"`
			C bool   `map:"z"`
		}{
			A: "a-value",
			B: 2,
			C: true,
		}

		a := Map(T)
		require.ElementsMatch(t, []string{"x", "y", "z"}, getMapKey(a))
	})
	t.Run("CustomTag", func(t *testing.T) {
		type D struct {
			E string `json:"jkl"`
		}
		T := struct {
			A string `json:"x"`
			B int    `json:"y"`
			C bool   `json:"z"`
			D D      `json:"nested"`
		}{
			A: "a-value",
			B: 2,
			C: true,
			D: D{E: "e-value"},
		}

		a := MapWithTag(T, "json")

		require.ElementsMatch(t, []string{"x", "y", "z", "nested"}, getMapKey(a))

		nested, ok := a["nested"].(map[string]interface{})
		require.True(t, ok)

		e, ok := nested["jkl"].(string)
		require.True(t, ok)
		require.Equal(t, "e-value", e)
	})

	t.Run("MultipleCustomTag", func(t *testing.T) {
		var A = struct {
			X string `aa:"ax"`
		}{"a_value"}

		var B = struct {
			X string `bb:"bx"`
		}{"b_value"}

		a, b := MapWithTag(A, "aa"), MapWithTag(B, "bb")
		require.Equal(t, map[string]interface{}{"ax": "a_value"}, a)
		require.Equal(t, map[string]interface{}{"bx": "b_value"}, b)
	})

	t.Run("OmitEmpty", func(t *testing.T) {
		type A struct {
			Name  string
			Value string    `map:",omitempty"`
			Time  time.Time `map:",omitempty"`
		}

		m := Map(A{})

		_, ok := m["Value"].(map[string]interface{})
		require.False(t, ok)
		_, ok = m["Time"].(map[string]interface{})
		require.False(t, ok)
	})

	t.Run("OmitNested", func(t *testing.T) {
		type A struct {
			Name  string
			Value string
			Time  time.Time `map:",omitnested"`
		}
		type B struct {
			Desc string
			A    A
		}

		b := &B{A: A{Time: time.Now()}}
		m := Map(b)

		in, ok := m["A"].(map[string]interface{})
		require.True(t, ok)

		// should not happen
		_, ok = in["Time"].(map[string]interface{})
		require.False(t, ok)

		_, ok = in["Time"].(time.Time)
		require.True(t, ok)
	})

	t.Run("Nested", func(t *testing.T) {
		type A struct {
			Name string
		}
		type B struct {
			A *A
		}

		b := &B{A: &A{Name: "example"}}
		m := Map(b)

		require.Equal(t, reflect.Map, reflect.TypeOf(m).Kind())

		in, ok := m["A"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "example", in["Name"].(string))
	})

	t.Run("NestedMapWithStructValues", func(t *testing.T) {
		type A struct {
			Name string
		}
		type B struct {
			A map[string]*A
		}

		b := &B{
			A: map[string]*A{
				"example_key": {Name: "example"},
			},
		}
		m := Map(b)
		require.Equal(t, reflect.Map, reflect.TypeOf(m).Kind())

		in, ok := m["A"].(map[string]interface{})
		require.True(t, ok)
		example := in["example_key"].(map[string]interface{})
		require.Equal(t, "example", example["Name"].(string))
	})

	t.Run("NestedMapWithStringValues", func(t *testing.T) {
		type B struct {
			Foo map[string]string
		}
		type A struct {
			B *B
		}

		a := &A{
			B: &B{
				Foo: map[string]string{
					"example_key": "example",
				},
			},
		}

		m := Map(a)
		require.Equal(t, reflect.Map, reflect.TypeOf(m).Kind())

		in, ok := m["B"].(map[string]interface{})
		require.True(t, ok)

		foo := in["Foo"].(map[string]string)
		require.Equal(t, "example", foo["example_key"])
	})

	t.Run("NestedMapWithInterfaceValues", func(t *testing.T) {
		type B struct {
			Foo map[string]interface{}
		}
		type A struct {
			B *B
		}

		a := &A{B: &B{
			Foo: map[string]interface{}{
				"example_key": "example",
			},
		}}

		m := Map(a)
		require.Equal(t, reflect.Map, reflect.TypeOf(m).Kind())

		in, ok := m["B"].(map[string]interface{})
		require.True(t, ok)

		foo := in["Foo"].(map[string]interface{})
		require.Equal(t, "example", foo["example_key"])
	})

	t.Run("NestedMapWithSliceIntValues", func(t *testing.T) {
		type B struct {
			Foo map[string][]int
		}
		type A struct {
			B *B
		}

		a := &A{B: &B{
			Foo: map[string][]int{
				"example_key": {80},
			},
		}}
		m := Map(a)
		require.Equal(t, reflect.Map, reflect.TypeOf(m).Kind())

		in, ok := m["B"].(map[string]interface{})
		require.True(t, ok)

		foo := in["Foo"].(map[string][]int)
		require.ElementsMatch(t, []int{80}, foo["example_key"])
	})
	t.Run("NestedMapWithSliceStructValues", func(t *testing.T) {
		type address struct {
			Country string `map:"country"`
		}
		type B struct {
			Foo map[string][]address
		}
		type A struct {
			B *B
		}

		a := &A{B: &B{
			Foo: map[string][]address{
				"example_key": {
					{Country: "Turkey"},
				},
			},
		}}
		m := Map(a)
		require.Equal(t, reflect.Map, reflect.TypeOf(m).Kind())

		in, ok := m["B"].(map[string]interface{})
		require.True(t, ok)

		foo := in["Foo"].(map[string]interface{})
		addresses := foo["example_key"].([]interface{})

		addr, ok := addresses[0].(map[string]interface{})
		require.True(t, ok)
		_, exists := addr["country"]
		require.True(t, exists)
	})

	t.Run("NestedSliceWithStructValues", func(t *testing.T) {
		type address struct {
			Country string `map:"customCountryName"`
		}
		type person struct {
			Name      string    `map:"name"`
			Addresses []address `map:"addresses"`
		}

		p := person{
			Name: "test",
			Addresses: []address{
				{Country: "England"},
				{Country: "Italy"},
			},
		}
		m := Map(p)
		require.Equal(t, reflect.Map, reflect.TypeOf(m).Kind())

		mpAddresses := m["addresses"].([]interface{})

		_, exists := mpAddresses[0].(map[string]interface{})["Country"]
		require.False(t, exists)

		_, exists = mpAddresses[0].(map[string]interface{})["customCountryName"]
		require.True(t, exists)
	})

	t.Run("NestedSliceWithPointerOfStructValues", func(t *testing.T) {
		type address struct {
			Country string `map:"customCountryName"`
		}
		type person struct {
			Name      string     `map:"name"`
			Addresses []*address `map:"addresses"`
		}

		p := person{
			Name: "test",
			Addresses: []*address{
				{Country: "England"},
				{Country: "Italy"},
			},
		}
		m := Map(p)
		require.Equal(t, reflect.Map, reflect.TypeOf(m).Kind())

		mpAddresses := m["addresses"].([]interface{})

		_, exists := mpAddresses[0].(map[string]interface{})["Country"]
		require.False(t, exists)

		_, exists = mpAddresses[0].(map[string]interface{})["customCountryName"]
		require.True(t, exists)
	})

	t.Run("NestedSliceWithIntValues", func(t *testing.T) {
		type person struct {
			Name  string `map:"name"`
			Ports []int  `map:"ports"`
		}

		p := person{
			Name:  "test",
			Ports: []int{80},
		}
		m := Map(p)
		require.Equal(t, reflect.Map, reflect.TypeOf(m).Kind())

		ports, ok := m["ports"].([]int)
		require.True(t, ok)
		require.ElementsMatch(t, []int{80}, ports)
	})
	t.Run("Pointer2Pointer", func(t *testing.T) {
		require.NotPanics(t, func() {
			a := &Animal{
				Name: "Fluff",
				Age:  4,
			}
			_ = Map(&a)

			b := &a
			_ = Map(&b)

			c := &b
			_ = Map(&c)
		})
	})
	t.Run("InterfaceTypeWithMapValue", func(t *testing.T) {
		type A struct {
			Name    string      `map:"name"`
			IP      string      `map:"ip"`
			Query   string      `map:"query"`
			Payload interface{} `map:"payload"`
		}

		a := A{
			Name:    "test",
			IP:      "127.0.0.1",
			Query:   "",
			Payload: map[string]string{"test_param": "test_param"},
		}

		require.NotPanics(t, func() {
			_ = Map(a)
		})
	})

	t.Run("NonStringerTagWithStringOption", func(t *testing.T) {
		a := &Animal{
			Name: "Fluff",
			Age:  4,
		}
		d := &Dog{
			Animal: a,
		}

		m := MapWithTag(d, "json")
		_, exists := m["animal"]
		require.True(t, exists)
	})

	t.Run("InterfaceValue", func(t *testing.T) {
		type TestStruct struct {
			A interface{}
		}

		expected := []byte("test value")
		a := TestStruct{A: expected}
		s := Map(a)

		require.Equal(t, expected, s["A"])
	})

	t.Run("TagWithStringOption", func(t *testing.T) {
		type Address struct {
			Country string  `json:"country"`
			Person  *Person `json:"person,string"` // nolint: staticcheck
		}
		person := &Person{
			Name: "John",
			Age:  23,
		}
		address := &Address{
			Country: "EU",
			Person:  person,
		}
		s := New(address).SetTagName("json")
		m := s.Map()

		require.Equal(t, person.String(), m["person"])

		vs := s.Values()
		require.Contains(t, vs, person.String())
	})

	t.Run("SetValueOnNestedField", func(t *testing.T) {
		type Base struct {
			ID int
		}
		type User struct {
			Base
			Name string
		}

		u := User{}
		f := New(&u).MustField("Base").MustField("ID")
		err := f.Set(10)
		require.NoError(t, err)
		require.Equal(t, 10, f.Value())
	})
	t.Run("NestedNilPointer", func(t *testing.T) {
		type Collar struct {
			Engraving string
		}
		type Dog struct {
			Name   string
			Collar *Collar
		}
		type Person struct {
			Name string
			Dog  *Dog
		}

		person := &Person{
			Name: "John",
		}
		personWithDog := &Person{
			Name: "Ron",
			Dog: &Dog{
				Name: "Rover",
			},
		}
		personWithDogWithCollar := &Person{
			Name: "Kon",
			Dog: &Dog{
				Name: "Ruffles",
				Collar: &Collar{
					Engraving: "If lost, call Kon",
				},
			},
		}

		_ = Map(person)
		_ = Map(personWithDog)
		_ = Map(personWithDogWithCollar)
	})
	t.Run("flattenNested", func(t *testing.T) {
		type A struct {
			Name string
		}
		type B struct {
			A `map:",flatten"`
			C int
		}

		b := &B{
			A: A{Name: "example"},
			C: 123,
		}
		m := Map(b)

		_, ok := m["A"].(map[string]interface{})
		require.False(t, ok)

		require.Equal(t, map[string]interface{}{"Name": "example", "C": 123}, m)
	})

	t.Run("flattenNestedOverwrite", func(t *testing.T) {
		type A struct {
			Name string
		}
		type B struct {
			A    `map:",flatten"`
			Name string
			C    int
		}
		b := &B{
			A:    A{Name: "example"},
			C:    123,
			Name: "bName",
		}
		m := Map(b)
		_, ok := m["A"].(map[string]interface{})
		require.False(t, ok)

		require.Equal(t, map[string]interface{}{"Name": "bName", "C": 123}, m)
	})
}

func TestFillMap(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		var T = struct {
			A string
			B int
			C bool
		}{
			A: "a-value",
			B: 2,
			C: true,
		}

		m := make(map[string]interface{})
		FillMap(T, m)

		require.Len(t, m, 3)
		require.ElementsMatch(t, []interface{}{"a-value", 2, true}, getMapValue(m))
	})
	t.Run("nil", func(t *testing.T) {
		var T = struct {
			A string
			B int
			C bool
		}{
			A: "a-value",
			B: 2,
			C: true,
		}

		// nil should no
		FillMap(T, nil)
	})
}

func TestValues(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		T := struct {
			A string
			B int
			C bool
		}{
			A: "a-value",
			B: 2,
			C: true,
		}
		s := Values(T)

		require.Equal(t, reflect.Slice, reflect.TypeOf(s).Kind())
		require.Equal(t, []interface{}{"a-value", 2, true}, s)
	})

	t.Run("OmitEmpty", func(t *testing.T) {
		a := struct {
			Name  string
			Value int `map:",omitempty"`
		}{
			Name:  "example",
			Value: 0,
		}
		s := Values(a)

		require.Len(t, s, 1)
		require.ElementsMatch(t, []interface{}{"example"}, s)
	})

	type A struct {
		Name string
	}
	t.Run("OmitNested", func(t *testing.T) {
		a := A{
			Name: "example",
		}
		b := struct {
			A A `map:",omitnested"`
			C int
		}{
			A: a,
			C: 123,
		}
		s := Values(b)

		require.Len(t, s, 2)
		require.ElementsMatch(t, []interface{}{a, 123}, s)
	})

	t.Run("Nested", func(t *testing.T) {
		a := A{Name: "example"}
		b := struct {
			A A
			C int
		}{
			A: a,
			C: 123,
		}
		s := Values(b)
		require.Len(t, s, 2)
		require.ElementsMatch(t, []interface{}{"example", 123}, s)
	})

	t.Run("Anonymous", func(t *testing.T) {
		a := A{Name: "example"}
		b := struct {
			A
			C int
		}{
			A: a,
			C: 123,
		}

		s := Values(b)

		require.Len(t, s, 2)
		require.ElementsMatch(t, []interface{}{"example", 123}, s)
	})
}

func TestNames(t *testing.T) {
	var T = struct {
		A string
		B int
		C bool
		d struct{}
	}{
		A: "a-value",
		B: 2,
		C: true,
		d: struct{}{},
	}

	s := Names(T)
	require.Len(t, s, 4)
	require.Equal(t, []string{"A", "B", "C", "d"}, s)
}

func TestFields(t *testing.T) {
	getFieldNames := func(fields []*Field) []string {
		names := make([]string, 0, len(fields))
		for _, field := range fields {
			names = append(names, field.Name())
		}
		return names
	}
	t.Run("Normal", func(t *testing.T) {
		var T = struct {
			A string
			B int
			C bool `map:",omitempty"`
			d struct{}
		}{
			A: "a-value",
			B: 2,
			C: true,
			d: struct{}{},
		}

		s := Fields(T)
		require.Len(t, s, 4)
		require.ElementsMatch(t, []string{"A", "B", "C", "d"}, getFieldNames(s))
	})

	t.Run("Ignore", func(t *testing.T) {
		type A struct {
			Name    string
			Enabled bool
		}

		a := A{Name: "example"}
		b := struct {
			A     A
			c     int
			Value string `map:"-"`
		}{
			A: a,
			c: 123,
		}

		s := Fields(b)

		require.Len(t, s, 2)
		require.ElementsMatch(t, []string{"A", "c"}, getFieldNames(s))
	})

	t.Run("Anonymous", func(t *testing.T) {
		type A struct {
			Name string
		}

		a := A{Name: "example"}
		b := struct {
			A
			c int
		}{
			A: a,
			c: 123,
		}
		s := Fields(b)
		names := make([]string, 0, len(s))
		for _, field := range s {
			names = append(names, field.Name())
		}
		require.ElementsMatch(t, []string{"A", "c"}, names)
	})
}

func TestIsZero(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		require.True(t,
			IsZero(struct {
				A string
				B int
				C bool `map:"-"`
				D []string
			}{}),
		)
		require.False(t,
			IsZero(struct {
				A string
				F *bool
			}{
				A: "a-value",
			}),
		)
		require.False(t,
			IsZero(struct {
				A string
				B int
			}{
				A: "a-value",
				B: 123,
			}),
		)
	})

	type A struct {
		Name string
		D    string
	}

	t.Run("OmitNested", func(t *testing.T) {
		type B struct {
			A A `map:",omitnested"`
			C int
		}

		require.True(t, IsZero(&B{A: A{}}))
		require.False(t, IsZero(&B{A: A{Name: "example"}, C: 123}))
	})

	t.Run("Nested", func(t *testing.T) {
		type B struct {
			A A
			C int
		}

		require.False(t, IsZero(&B{A: A{Name: "example"}, C: 123}))
		require.True(t, IsZero(&B{A: A{}}))
	})

	t.Run("Anonymous", func(t *testing.T) {
		type B struct {
			A
			C int
		}

		require.False(t, IsZero(&B{A: A{Name: "example"}, C: 123}))
		require.True(t, IsZero(&B{A: A{}, C: 0}))
	})
}

func TestHasZero(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		require.True(t,
			HasZero(
				struct {
					A string
					B int
					C bool `map:"-"`
					D []string
				}{
					A: "a-value",
					B: 2,
				},
			),
		)
		require.True(t,
			HasZero(
				struct {
					A string
					F *bool
				}{
					A: "a-value",
				},
			),
		)
		require.False(t,
			HasZero(
				struct {
					A string
					B int
				}{
					A: "a-value",
					B: 123,
				},
			),
		)
	})

	type A struct {
		Name string
		D    string
	}

	t.Run("OmitNested", func(t *testing.T) {
		// Because the MustField A inside B is omitted  HasZero should return false
		// because it will stop iterating deeper andnot going to lookup for D
		require.False(t,
			HasZero(
				struct {
					A A `map:",omitnested"`
					C int
				}{
					A: A{Name: "example"},
					C: 123,
				},
			),
		)
	})

	t.Run("Nested", func(t *testing.T) {
		require.True(t,
			HasZero(
				struct {
					A A
					C int
				}{
					A: A{Name: "example"},
					C: 123,
				},
			),
		)
	})

	t.Run("Anonymous", func(t *testing.T) {
		require.True(t,
			HasZero(
				struct {
					A
					C int
				}{
					A: A{Name: "example"},
					C: 123,
				},
			),
		)
	})
}

func TestStruct(t *testing.T) {
	v := struct{}{}

	require.True(t, IsStruct(v))
	require.True(t, IsStruct(&v))
	require.False(t, IsStruct(""))
	require.False(t, IsStruct((*struct{})(nil)))
}

func TestName(t *testing.T) {
	type Foo struct {
		A string
		B bool
	}

	f := &Foo{}
	require.Equal(t, "Foo", Name(f))

	unnamed := struct{ Name string }{Name: "Cihangir"}
	require.Empty(t, Name(unnamed))

	require.Panics(t, func() {
		Name([]string{})
	})
}

func TestMapSlice(t *testing.T) {
	type Foo struct {
		A string
		B bool
	}

	t.Run("Normal struct slice", func(t *testing.T) {
		want := []map[string]interface{}{
			{"A": "a1", "B": false},
			{"A": "a2", "B": true},
		}
		got := MapSlice([]Foo{{"a1", false}, {"a2", true}})

		require.ElementsMatch(t, want, got)
	})
	t.Run("Normal empty struct slice", func(t *testing.T) {
		want := []map[string]interface{}{}
		got := MapSlice([]Foo{})

		require.ElementsMatch(t, want, got)

		var a []Foo
		got = MapSlice(a)

		require.ElementsMatch(t, want, got)
	})
	t.Run("Normal array struct", func(t *testing.T) {
		want := []map[string]interface{}{
			{"A": "a1", "B": false},
			{"A": "a2", "B": true},
		}
		got := MapSlice([2]Foo{{"a1", false}, {"a2", true}})

		require.ElementsMatch(t, want, got)
	})
	t.Run("Normal empty array struct", func(t *testing.T) {
		want := []map[string]interface{}{}
		got := MapSlice([0]Foo{})

		require.ElementsMatch(t, want, got)
	})

	t.Run("Not a slice struct", func(t *testing.T) {
		want := []map[string]interface{}{}
		got := MapSlice("")

		require.ElementsMatch(t, want, got)

		got = MapSlice([]int{1, 2})
		require.ElementsMatch(t, want, got)
	})
	t.Run("nil", func(t *testing.T) {
		want := []map[string]interface{}{}
		got := MapSlice(nil)

		require.ElementsMatch(t, want, got)
	})
}

func Test_isEmptyValue(t *testing.T) {
	type A struct{}
	var a = &A{}

	tests := []struct {
		name string
		args reflect.Value
		want bool
	}{
		{
			"bool",
			reflect.ValueOf(false),
			true,
		},
		{
			"int",
			reflect.ValueOf(0),
			true,
		},
		{
			"uint",
			reflect.ValueOf(uint(0)),
			true,
		},
		{
			"float",
			reflect.ValueOf(float32(0)),
			true,
		},
		{
			"slice",
			reflect.ValueOf([]string{}),
			true,
		},
		{
			"ptr",
			reflect.ValueOf((*string)(nil)),
			true,
		},
		{
			"struct{}",
			reflect.ValueOf(struct{}{}),
			false,
		},
		{
			"ptr2ptr",
			reflect.ValueOf(&a),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEmptyValue(tt.args); got != tt.want {
				t.Errorf("isEmptyValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isEmpty(t *testing.T) {
	type A struct{}
	var a = &A{}

	tests := []struct {
		name string
		args reflect.Value
		want bool
	}{
		{
			"bool",
			reflect.ValueOf(false),
			true,
		},
		{
			"int",
			reflect.ValueOf(0),
			true,
		},
		{
			"uint",
			reflect.ValueOf(uint(0)),
			true,
		},
		{
			"float",
			reflect.ValueOf(float32(0)),
			true,
		},
		{
			"slice",
			reflect.ValueOf([]string{}),
			true,
		},
		{
			"ptr",
			reflect.ValueOf((*string)(nil)),
			true,
		},
		{
			"struct{}",
			reflect.ValueOf(struct{}{}),
			true,
		},
		{
			"ptr2ptr",
			reflect.ValueOf(&a),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEmptyWithAll(tt.args); got != tt.want {
				t.Errorf("isEmptyValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toString(t *testing.T) {
	tests := []struct {
		name string
		args reflect.Value
		want interface{}
	}{
		{
			"bool",
			reflect.ValueOf(false),
			"false",
		},
		{
			"int",
			reflect.ValueOf(0),
			"0",
		},
		{
			"uint",
			reflect.ValueOf(uint(0)),
			"0",
		},
		{
			"float",
			reflect.ValueOf(float32(0)),
			"0",
		},
		{
			"slice",
			reflect.ValueOf([]string{}),
			nil,
		},
		{
			"ptr",
			reflect.ValueOf((*string)(nil)),
			nil,
		},
		{
			"struct{}",
			reflect.ValueOf(struct{}{}),
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toString(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toString() = %v, want %v", got, tt.want)
			}
		})
	}
}
