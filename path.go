package path

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

// Potential errors you could receive from this package. These
// are mostly self explanatory.
var (
	ErrNotFound = errors.New("path: no path could be found with the name provided")
)

// Builder is used to set and retrieve named paths.
type Builder struct {
	// Whether or not to turn additional parameters provided
	// to a path into URL query parameters. Eg if you pass
	// the params:
	//
	//   map[string]interface{}{"name": "jane doe"}
	//
	// into a named path defined as `/blah` and this option
	// was set to true, the `name` parameter would be
	// ignored. On the other hand if this value was false it
	// would return something like `/blah?name=jane%20doe`.
	//
	// The default value is false, meaning that extra params
	// will be turned into URL query params.
	IgnoreExtraParams bool

	// unexported fields
	m     sync.Mutex
	once  sync.Once
	paths map[string]string
}

// Set is used to set a named path.
func (b *Builder) Set(name, format string) {
	b.m.Lock()
	defer b.m.Unlock()
	b.init()
	b.paths[name] = format
}

// Path is used to retrieve a named path or return an empty
// string in no path exists with that name.
func (b *Builder) Path(name string, params map[string]interface{}) string {
	// StrictPath is already thread-safe so no need to lock
	ret, err := b.StrictPath(name, params)
	if err != nil {
		return ""
	}
	return ret
}

// StrictPath is used to retrieve a named path or return an
// error if no path exists with that name.
func (b *Builder) StrictPath(name string, params map[string]interface{}) (string, error) {
	b.m.Lock()
	b.m.Unlock()
	path, ok := b.paths[name]
	if !ok {
		return "", ErrNotFound
	}
	return replace(path, params, !b.IgnoreExtraParams), nil
}

func (b *Builder) init() {
	b.once.Do(func() {
		b.paths = make(map[string]string)
	})
}

func replace(path string, params map[string]interface{}, query bool) string {
	if params == nil {
		return path
	}
	pieces := strings.Split(path, "/")
	fillVals := make(map[string]interface{})
	// Default values are the key - eg :id => :id by default
	// unless we provide a new value for it.
	for _, piece := range pieces {
		k, err := key(piece)
		if err == errInvalidKey {
			continue
		}
		fillVals[k] = piece
	}
	// Overwrite defaults values where the param is provided
	for k, v := range params {
		fillVals[k] = v
	}
	// Build the final path's pieces, deleting keys we use
	// so we can keep track for URL query params
	var ret []string
	for _, piece := range pieces {
		k, err := key(piece)
		if err == errInvalidKey {
			ret = append(ret, piece)
			continue
		}
		ret = append(ret, fmt.Sprintf("%v", fillVals[k]))
		delete(fillVals, k)
	}
	if !query {
		return strings.Join(ret, "/")
	}
	qv := make(url.Values)
	for k, v := range fillVals {
		qv.Set(k, fmt.Sprintf("%v", v))
	}
	if len(qv) > 0 {
		return strings.Join(ret, "/") + "?" + qv.Encode()
	}
	return strings.Join(ret, "/")
}

var (
	errInvalidKey = errors.New("path: invalid key")
)

func key(piece string) (string, error) {
	if len(piece) == 0 {
		return "", errInvalidKey
	}
	if piece[0] != ':' {
		return "", errInvalidKey
	}
	return piece[1:], nil
}
