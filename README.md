# wstructs

> origin: github.com/things-go/structs

Go library for encoding native Go structures into generic map values.

![GitHub Repo stars](https://img.shields.io/github/stars/wyy-go/wstructs?style=social)
![GitHub](https://img.shields.io/github/license/wyy-go/wstructs)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/wyy-go/wstructs)
![GitHub CI Status](https://img.shields.io/github/workflow/status/wyy-go/wstructs/ci?label=CI)
[![Go Report Card](https://goreportcard.com/badge/github.com/wyy-go/wstructs)](https://goreportcard.com/report/github.com/wyy-go/wstructs)
[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wyy-go/wstructs?tab=doc)
[![codecov](https://codecov.io/gh/wyy-go/wstructs/branch/main/graph/badge.svg)](https://codecov.io/gh/wyy-go/wstructs)
### Installation

Use go get.

```bash
    go get -u github.com/wyy-go/wstructs
```

Then import the structs package into your own code.

```go
    import "github.com/wyy-go/wstructs"
```

### Usage && Example

#### API

Just like the standard lib strings, bytes and co packages, structs has many global functions to manipulate or organize your struct data. Lets define and declare a struct:

```go
type Server struct {
    Name        string `json:"name,omitempty"`
    ID          int
    Enabled     bool
    users       []string // not exported
    http.Server          // embedded
}

server := &Server{
    Name:    "gopher",
    ID:      123456,
    Enabled: true,
}
```

Here is an example:

```go
// Convert a struct to a map[string]interface{}
// => {"Name":"gopher", "ID":123456, "Enabled":true}
m := structs.Map(server)

// Convert the values of a struct to a []interface{}
// => ["gopher", 123456, true]
v := structs.Values(server)

// Convert the names of a struct to a []string
// (see "Names methods" for more info about fields)
n := structs.Names(server)

// Convert the values of a struct to a []*Field
// (see "Field methods" for more info about fields)
f := structs.Fields(server)

// Return the struct name => "Server"
n := structs.Name(server)

// Check if any field of a struct is initialized or not.
h := structs.HasZero(server)

// Check if all fields of a struct is initialized or not.
z := structs.IsZero(server)

// Check if server is a struct or a pointer to struct
i := structs.IsStruct(server)
```

Only [public fields](https://golang.org/doc/effective_go.html#names) will be processed. So **fields
starting with lowercase will be ignored**.

#### Name Tags

```go
type AA struct {
    Id        int64    `map:"id"`
    Name      string   `map:"name"`
}
```
We can give the field a tag to specify another name to be used as the key.

#### Ignore Field

```go
type AA struct {
    Ignore string `map:"-"`
}
```
If we give the special tag "-" to a field, it will be ignored.

#### Omit Empty

```go
type AA struct {
    Desc        string    `map:"desc,omitempty"`
}
```
If tag option is "omitempty", this field will not appear in the map if the value is empty.
Empty values are 0, false, "", nil, empty array and empty map.

#### Omit Nested

```go
type AA struct {
    // Field is not processed further by this package.
    Field *http.Request `map:",omitnested"`
}
```
A value with the option of "omitnested" stops iterating further if the type
is a struct. it do not converted value if a value are no exported fields, ie: time.Time

#### To String

```go
type AA struct {
    Id        int64    `map:"id,string"`
    Price     float32  `map:"price,string"`
}
```
If tag option is "string", this field will be converted to string type. Encode will put the
original value to the map if the conversion is failed.