package wstructs

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	errNotExported = errors.New("structs: field is not exported")
	errNotSettable = errors.New("structs: field is not settable")
)

// Field represents a single struct field that encapsulates high level
// functions around the field.
type Field struct {
	value      reflect.Value
	field      reflect.StructField
	defaultTag string
}

// Value returns the underlying value of the field. It panics if the field
// is not exported.
func (f *Field) Value() interface{} {
	return f.value.Interface()
}

// Tag returns the value associated with key in the tag string. If there is no
// such key in the tag, Tag returns the empty string.
func (f *Field) Tag(key string) string {
	return f.field.Tag.Get(key)
}

// IsAnonymous returns true if the given field is an anonymous field (embedded)
func (f *Field) IsAnonymous() bool {
	return f.field.Anonymous
}

// IsExported returns true if the given field is exported.
func (f *Field) IsExported() bool {
	return f.field.PkgPath == ""
}

// CanInterface reports whether Interface can be used without panicking.
func (f *Field) CanInterface() bool {
	return f.value.CanInterface()
}

// CanSet reports whether the value of v can be changed.
func (f *Field) CanSet() bool {
	return f.value.CanSet()
}

// IsZero returns true if the given field is not initialized (has a zero value).
// It panics if the field is not exported.
func (f *Field) IsZero() bool {
	return isEmptyValue(f.value)
}

// Name returns the name of the given field
func (f *Field) Name() string {
	return f.field.Name
}

// Kind returns the fields kind, such as "string", "map", "bool", etc ..
func (f *Field) Kind() reflect.Kind {
	return f.value.Kind()
}

// Set sets the field to given value v. It returns an error if the field is not
// settable (not addressable or not exported) or if the given value's type
// doesn't match the fields type.
func (f *Field) Set(val interface{}) error {
	// we can't set unexported fields, so be sure this field is exported
	if !f.IsExported() {
		return errNotExported
	}
	// do we get here? not sure...
	if !f.value.CanSet() {
		return errNotSettable
	}

	given := reflect.ValueOf(val)
	if f.value.Kind() != given.Kind() {
		return fmt.Errorf("structs: wrong kind. got: %s want: %s", given.Kind(), f.value.Kind())
	}

	f.value.Set(given)
	return nil
}

// SetZero sets the field to its zero value. It returns an error if the field is not
// settable (not addressable or not exported).
func (f *Field) SetZero() error {
	zero := reflect.Zero(f.value.Type()).Interface()
	return f.Set(zero)
}

// Fields returns a slice of Fields. This is particular handy to get the fields
// of a nested struct . A struct tag with the content of "-" ignores the
// checking of that particular field. Example:
//
//   // Field is ignored by this package.
//   Field *http.Request `structs:"-"`
//
// It panics if field is not exported or if field's kind is not struct
func (f *Field) Fields() []*Field {
	return getFields(f.value, f.defaultTag)
}

// Field returns the field from a nested struct. It panics if the nested struct
// is not exported or if the field was not found.
func (f *Field) MustField(name string) *Field {
	field, ok := f.Field(name)
	if !ok {
		panic("structs: field not found")
	}
	return field
}

// Field returns the field from a nested struct. The boolean returns whether
// the field was found (true) or not (false).
func (f *Field) Field(name string) (*Field, bool) {
	value := &f.value
	// value must be settable so we need to make sure it holds the address of the
	// variable and not a copy, so we can pass the pointer to structVal instead of a
	// copy (which is not assigned to any variable, hence not settable).
	// see "https://blog.golang.org/laws-of-reflection#TOC_8."
	if f.value.Kind() != reflect.Ptr {
		a := f.value.Addr()
		value = &a
	}
	v, err := structVal(value.Interface())
	if err != nil {
		return nil, false
	}
	t := v.Type()

	field, ok := t.FieldByName(name)
	if !ok {
		return nil, false
	}

	return &Field{
		field:      field,
		value:      v.FieldByName(name),
		defaultTag: f.defaultTag,
	}, true
}
