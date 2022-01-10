// Package wstructs contains various utilities functions to work with structs.
package wstructs

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// DefaultTagName is the default tag name for struct fields which provides
// a more granular to tweak certain structs. Lookup the necessary functions
// for more info.
const DefaultTagName = "map" // struct's field default tag name

// Struct encapsulates a struct type to provide several high level functions
// around the struct.
type Struct struct {
	raw     interface{}
	value   reflect.Value
	tagName string
}

// New returns a new *Struct with the struct s. It panics if the s's kind is
// not struct.
func New(s interface{}) *Struct {
	value, err := structVal(s)
	if err != nil {
		panic(err)
	}
	return &Struct{
		s,
		value,
		DefaultTagName,
	}
}

// SetTagName set struct's field tag name, default is  DefaultTagName.
func (s *Struct) SetTagName(tagName string) *Struct {
	s.tagName = tagName
	return s
}

// Map converts the given struct to a map[string]interface{}, where the keys
// of the map are the field names and the values of the map the associated
// values of the fields. The default key string is the struct field name but
// can be changed in the struct field's tag value. The "structs" key in the
// struct's field tag value is the key name. Example:
//
//   // Field appears in map as key "myName".
//   Name string `map:"myName"`
//
// A tag value with the content of "-" ignores that particular field. Example:
//
//   // Field is ignored by this package.
//   Field bool `map:"-"`
//
// A tag value with the content of "string" uses the stringer to get the value. Example:
//
//   // The value will be output of Animal's String() func.
//   // Map will panic if Animal does not implement String().
//   Field *Animal `map:"field,string"`
//
// A tag value with the option of "flatten" used in a struct field is to flatten its fields
// in the output map. Example:
//
//   // The FieldStruct's fields will be flattened into the output map.
//   FieldStruct time.Time `structs:",flatten"`
//
// A tag value with the option of "omitnested" stops iterating further if the type
// is a struct. Example:
//
//   // Field is not processed further by this package.
//   Field time.Time     `map:"myName,omitnested"`
//   Field *http.Request `map:",omitnested"`
//
// A tag value with the option of "omitempty" ignores that particular field if
// the field value is empty. Example:
//
//   // Field appears in map as key "myName", but the field is
//   // skipped if empty.
//   Field string `map:"myName,omitempty"`
//
//   // Field appears in map as key "MustField" (the default), but
//   // the field is skipped if empty.
//   Field string `map:",omitempty"`
//
// Note that only exported fields of a struct can be accessed, non exported
// fields will be neglected.
func (s *Struct) Map() map[string]interface{} {
	out := make(map[string]interface{})
	s.FillMap(out)
	return out
}

// FillMap is the same as Map. Instead of returning the output, it fills the
// given map.
func (s *Struct) FillMap(out map[string]interface{}) {
	if out == nil {
		return
	}

	fields := getStructFields(s.value, s.tagName)
	for _, field := range fields {
		name := field.Name
		val := s.value.FieldByName(name)

		tagName, tagOpts := parseTag(field.Tag.Get(s.tagName))
		if tagName != "" {
			name = tagName
		}

		// if the value is a zero value and the field is marked as omitempty do
		// not include
		if tagOpts.Contains("omitempty") && isEmptyValue(val) {
			continue
		}
		if tagOpts.Contains("string") {
			if str := toString(val); str != nil {
				out[name] = str
				continue
			}
		}

		var finalVal interface{}
		isSubStruct := false

		if !tagOpts.Contains("omitnested") {
			finalVal = s.nested(val)
			if val.Kind() == reflect.Map || val.Kind() == reflect.Struct {
				isSubStruct = true
			}
		} else {
			finalVal = val.Interface()
		}
		if isSubStruct && tagOpts.Contains("flatten") {
			for k := range finalVal.(map[string]interface{}) {
				out[k] = finalVal.(map[string]interface{})[k]
			}
		} else {
			out[name] = finalVal
		}
	}
}

// Values converts the given s struct's exported field values to a []interface{}.  A
// struct tag with the content of "-" ignores the that particular field.
// Example:
//
//   // Field is ignored by this package.
//   Field int `map:"-"`
//
// A value with the option of "omitnested" stops iterating further if the type
// is a struct. Example:
//
//   // Fields is not processed further by this package.
//   Field time.Time     `map:",omitnested"`
//   Field *http.Request `map:",omitnested"`
//
// A tag value with the option of "omitempty" ignores that particular field and
// is not added to the values if the field value is empty. Example:
//
//   // Field is skipped if empty
//   Field string `map:",omitempty"`
//
// Note that only exported fields of a struct can be accessed, non exported
// fields  will be neglected.
func (s *Struct) Values() []interface{} {
	fields := getStructFields(s.value, s.tagName)
	t := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		val := s.value.FieldByName(field.Name)

		_, tagOpts := parseTag(field.Tag.Get(s.tagName))

		// if the value is a zero value and the field is marked as omitempty do
		// not include
		if tagOpts.Contains("omitempty") && isEmptyValue(val) {
			continue
		}
		if tagOpts.Contains("string") {
			if str := toString(val); str != nil {
				t = append(t, str)
				continue
			}
		}

		if IsStruct(val.Interface()) && !tagOpts.Contains("omitnested") {
			// look out for embedded structs, and convert them to a
			// []interface{} to be added to the final values slice
			t = append(t, Values(val.Interface())...)
		} else {
			t = append(t, val.Interface())
		}
	}
	return t
}

// Fields returns a slice of Fields. A struct tag with the content of "-"
// ignores the checking of that particular field. Example:
//
//   // Field is ignored by this package.
//   Field bool `map:"-"`
//
// It panics if s's kind is not struct.
func (s *Struct) Fields() []*Field {
	return getFields(s.value, s.tagName)
}

// Names returns a slice of field names. A struct tag with the content of "-"
// ignores the checking of that particular field. Example:
//
//   // Field is ignored by this package.
//   Field bool `map:"-"`
//
// It panics if s's kind is not struct.
func (s *Struct) Names() []string {
	fields := getFields(s.value, s.tagName)

	names := make([]string, 0, len(fields))
	for _, field := range fields {
		names = append(names, field.Name())
	}
	return names
}

// MustField returns a new Field struct that provides several high level functions
// around a single struct field entity. It panics if the field is not found.
func (s *Struct) MustField(name string) *Field {
	f, ok := s.Field(name)
	if !ok {
		panic("structs: field not found")
	}
	return f
}

// Field returns a new Field struct that provides several high level functions
// around a single struct field entity. The boolean returns true if the field
// was found.
func (s *Struct) Field(name string) (*Field, bool) {
	t := s.value.Type()
	field, ok := t.FieldByName(name)
	if !ok {
		return nil, false
	}
	return &Field{
		field:      field,
		value:      s.value.FieldByName(name),
		defaultTag: s.tagName,
	}, true
}

// IsZero returns true if all fields in a struct is a zero value (not
// initialized) A struct tag with the content of "-" ignores the checking of
// that particular field. Example:
//
//   // Field is ignored by this package.
//   Field bool `map:"-"`
//
// A value with the option of "omitnested" stops iterating further if the type
// is a struct. Example:
//
//   // Field is not processed further by this package.
//   Field time.Time     `map:"myName,omitnested"`
//   Field *http.Request `map:",omitnested"`
//
// Note that only exported fields of a struct can be accessed, non exported
// fields  will be neglected. It panics if s's kind is not struct.
func (s *Struct) IsZero() bool {
	fields := getStructFields(s.value, s.tagName)
	for _, field := range fields {
		val := s.value.FieldByName(field.Name)

		_, tagOpts := parseTag(field.Tag.Get(s.tagName))
		if IsStruct(val.Interface()) && !tagOpts.Contains("omitnested") {
			ok := IsZero(val.Interface())
			if !ok {
				return false
			}
			continue
		}
		if !isEmptyWithAll(val) {
			return false
		}
	}
	return true
}

// HasZero returns true if a field in a struct is not initialized (zero value).
// A struct tag with the content of "-" ignores the checking of that particular
// field. Example:
//
//   // Field is ignored by this package.
//   Field bool `map:"-"`
//
// A value with the option of "omitnested" stops iterating further if the type
// is a struct. Example:
//
//   // Field is not processed further by this package.
//   Field time.Time     `map:"myName,omitnested"`
//   Field *http.Request `map:",omitnested"`
//
// Note that only exported fields of a struct can be accessed, non exported
// fields  will be neglected. It panics if s's kind is not struct.
func (s *Struct) HasZero() bool {
	fields := getStructFields(s.value, s.tagName)
	for _, field := range fields {
		val := s.value.FieldByName(field.Name)

		_, tagOpts := parseTag(field.Tag.Get(s.tagName))

		if IsStruct(val.Interface()) && !tagOpts.Contains("omitnested") {
			ok := HasZero(val.Interface())
			if ok {
				return true
			}
			continue
		}

		if isEmptyWithAll(val) {
			return true
		}
	}
	return false
}

// Name returns the map's type name within its package. For more info refer
// to Name() function.
func (s *Struct) Name() string {
	return s.value.Type().Name()
}

// Map converts the given struct to a map[string]interface{}. For more info
// refer to Struct types Map() method. It panics if s's kind is not struct.
func Map(s interface{}) map[string]interface{} {
	return MapWithTag(s, DefaultTagName)
}

// MapWithTag same as Map() but with tagName. It panics if s's kind is not struct.
func MapWithTag(s interface{}, tagName string) map[string]interface{} {
	return New(s).SetTagName(tagName).Map()
}

// FillMap is the same as Map. Instead of returning the output, it fills the
// given map.
func FillMap(s interface{}, out map[string]interface{}) {
	FillMapWithTag(s, out, DefaultTagName)
}

// FillMapWithTag is the same as MapTag(). Instead of returning the output, it fills the
// given map.
func FillMapWithTag(s interface{}, out map[string]interface{}, tagName string) {
	New(s).SetTagName(tagName).FillMap(out)
}

// Values converts the given s struct's exported field values to a []interface{}.
// For more info refer to Struct types Values() method.  It panics if s's kind is
// not struct.
func Values(s interface{}) []interface{} {
	return ValuesWithTag(s, DefaultTagName)
}

// ValuesWithTag is the same as Values() but with tagName.
func ValuesWithTag(s interface{}, tagName string) []interface{} {
	return New(s).SetTagName(tagName).Values()
}

// Names returns a slice of field names. For more info refer to Struct types
// Names() method.  It panics if s's kind is not struct.
func Names(s interface{}) []string {
	return NamesWithTag(s, DefaultTagName)
}

// NamesWithTag is the same as Names() but with tagName.
func NamesWithTag(s interface{}, tagName string) []string {
	return New(s).SetTagName(tagName).Names()
}

// Fields returns a slice of *Field. For more info refer to Struct types
// Fields() method.  It panics if s's kind is not struct.
func Fields(s interface{}) []*Field {
	return FieldsWithTag(s, DefaultTagName)
}

// FieldsWithTag is the same as Fields() but with tagName.
func FieldsWithTag(s interface{}, tagName string) []*Field {
	return New(s).SetTagName(tagName).Fields()
}

// IsZero returns true if all fields is equal to a zero value. For more info
// refer to Struct types IsZero() method.  It panics if s's kind is not struct.
func IsZero(s interface{}) bool {
	return IsZeroWithTag(s, DefaultTagName)
}

// IsZeroWithTag is the same as IsZero() but with tagName.
func IsZeroWithTag(s interface{}, tagName string) bool {
	return New(s).SetTagName(tagName).IsZero()
}

// HasZero returns true if any field is equal to a zero value. For more info
// refer to Struct types HasZero() method.  It panics if s's kind is not struct.
func HasZero(s interface{}) bool {
	return HasZeroWithTag(s, DefaultTagName)
}

// HasZeroWithTag is the same as HasZero() but with tagName.
func HasZeroWithTag(s interface{}, tagName string) bool {
	return New(s).SetTagName(tagName).HasZero()
}

// IsStruct returns true if the given variable is a struct or a pointer to
// struct.
func IsStruct(s interface{}) bool {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	// uninitialized zero value of a struct
	if v.Kind() == reflect.Invalid {
		return false
	}
	return v.Kind() == reflect.Struct
}

// Name returns the map's type name within its package. It returns an
// empty string for unnamed types. It panics if s's kind is not struct.
func Name(s interface{}) string {
	return New(s).Name()
}

// MapSlice converts the given struct slice to a []map[string]interface{}.
// For more info refer to MapSliceWithTag() method
func MapSlice(s interface{}) []map[string]interface{} {
	return MapSliceWithTag(s, DefaultTagName)
}

// MapSliceWithTag converts the given struct slice to a []map[string]interface{} with tagName.
// It returns empty []map[string]interface{} if s is not a slice struct.
func MapSliceWithTag(s interface{}, tagName string) []map[string]interface{} {
	if s == nil {
		return make([]map[string]interface{}, 0)
	}
	v := reflect.Indirect(reflect.ValueOf(s))
	if (v.Type().Kind() == reflect.Slice && !v.IsNil() || v.Type().Kind() == reflect.Array) &&
		v.IsValid() && v.Len() != 0 && reflect.Indirect(v.Index(0)).Kind() == reflect.Struct {
		length := v.Len()
		result := make([]map[string]interface{}, length)
		for i := 0; i < length; i++ {
			result[i] = MapWithTag(v.Index(i).Interface(), tagName)
		}
		return result
	}
	return make([]map[string]interface{}, 0)
}

// nested retrieves recursively all types for the given value and returns the
// nested value.
func (s *Struct) nested(val reflect.Value) interface{} {
	var finalVal interface{}

	v := reflect.ValueOf(val.Interface())
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		m := New(val.Interface()).SetTagName(s.tagName).Map()

		// do not add the converted value if there are no exported fields, ie:
		// time.Time
		if len(m) == 0 {
			finalVal = val.Interface()
		} else {
			finalVal = m
		}
	case reflect.Map:
		// get the element type of the map
		mapElem := val.Type()
		switch val.Type().Kind() {
		case reflect.Ptr, reflect.Array, reflect.Map,
			reflect.Slice, reflect.Chan:
			mapElem = val.Type().Elem()
			if mapElem.Kind() == reflect.Ptr {
				mapElem = mapElem.Elem()
			}
		}

		// only iterate over struct types, ie: map[string]StructType,
		// map[string][]StructType,
		if mapElem.Kind() == reflect.Struct ||
			(mapElem.Kind() == reflect.Slice &&
				mapElem.Elem().Kind() == reflect.Struct) {
			m := make(map[string]interface{}, val.Len())
			for _, k := range val.MapKeys() {
				m[k.String()] = s.nested(val.MapIndex(k))
			}
			finalVal = m
			break
		}

		// TODO(arslan): should this be optional?
		finalVal = val.Interface()
	case reflect.Slice, reflect.Array:
		if val.Type().Kind() == reflect.Interface {
			finalVal = val.Interface()
			break
		}

		// TODO(arslan): should this be optional?
		// do not iterate of non struct types, just pass the value. Ie: []int,
		// []string, co... We only iterate further if it's a struct.
		// i.e []foo or []*foo
		if val.Type().Elem().Kind() != reflect.Struct &&
			!(val.Type().Elem().Kind() == reflect.Ptr &&
				val.Type().Elem().Elem().Kind() == reflect.Struct) {
			finalVal = val.Interface()
			break
		}

		slices := make([]interface{}, val.Len())
		for x := 0; x < val.Len(); x++ {
			slices[x] = s.nested(val.Index(x))
		}
		finalVal = slices
	default:
		finalVal = val.Interface()
	}

	return finalVal
}

func structVal(s interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(s)
	// if pointer get the underlying element
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return v, errors.New("structs: not struct")
	}
	return v, nil
}

func getFields(v reflect.Value, tagName string) []*Field {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	var fields []*Field

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get(tagName); tag == "-" {
			continue
		}
		fields = append(fields, &Field{
			v.Field(i),
			field,
			tagName,
		})
	}
	return fields
}

// getStructFields returns the exported struct fields for a given s struct. This
// is a convenient helper method to avoid duplicate code in some functions.
func getStructFields(v reflect.Value, tagName string) []reflect.StructField {
	t := v.Type()
	numField := t.NumField()

	f := make([]reflect.StructField, 0, numField)
	for i := 0; i < numField; i++ {
		field := t.Field(i)
		fv := v.Field(i)
		// we can't access the value of unexported fields
		if !fv.CanInterface() || field.PkgPath != "" {
			continue
		}
		// don't check if it's omitted
		if tag := field.Tag.Get(tagName); tag == "-" {
			continue
		}
		f = append(f, field)
	}
	return f
}

func isEmptyValue(v reflect.Value) bool {
	switch k := v.Kind(); k {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String, reflect.Chan:
		return v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func isEmptyWithAll(v reflect.Value) bool {
	switch k := v.Kind(); k {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String, reflect.Chan:
		return v.Len() == 0
	case reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return isEmptyWithAll(v.Elem())
	default:
		zero := reflect.Zero(v.Type()).Interface()
		current := v.Interface()
		return reflect.DeepEqual(current, zero)
	}
}

func toString(fv reflect.Value) interface{} {
	vv := reflect.Indirect(fv)
	switch k := vv.Kind(); k {
	case reflect.Bool:
		return strconv.FormatBool(vv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(vv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(vv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(vv.Float(), 'f', -1, 64)
	default:
		s, ok := fv.Interface().(fmt.Stringer)
		if ok {
			return s.String()
		}
	}
	return nil
}
