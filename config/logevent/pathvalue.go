package logevent

import (
	"reflect"
	"strconv"
	"strings"

	lru "github.com/hashicorp/golang-lru"
)

type pathtoken struct {
	isSlice bool
	index   int
	key     string
}

const cacheSize = 200

var cache *lru.Cache

func init() {
	var err error
	cache, err = lru.New(cacheSize)
	if err != nil {
		panic(err)
	}
}

func getPathArrayToken(path string) []pathtoken {
	var tokens []pathtoken
	for len(path) > 0 && path[len(path)-1] == ']' {
		pos := strings.LastIndexByte(path, '[')
		if pos < 0 {
			break
		}
		indexStr := path[pos+1 : len(path)-1]
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			panic(err)
		}
		t := pathtoken{
			isSlice: true,
			index:   index,
		}
		tokens = append([]pathtoken{t}, tokens...)
		path = path[:pos]
	}
	if len(tokens) > 0 && len(path) > 0 {
		// object key
		tokens = append([]pathtoken{pathtoken{
			isSlice: false,
			key:     path,
		}}, tokens...)
	}
	return tokens
}

func compilePath(path string) []pathtoken {
	fields := strings.Split(path, ".")
	tokens := make([]pathtoken, 0, len(fields)+2)
	for _, field := range fields {
		atokens := getPathArrayToken(field)
		if len(atokens) > 0 {
			tokens = append(tokens, atokens...)
		} else {
			tokens = append(tokens, pathtoken{
				isSlice: false,
				key:     field,
			})
		}
	}
	return tokens
}

func compilePathWithCache(path string) []pathtoken {
	var tokens []pathtoken
	cachedTokens, ok := cache.Get(path)
	if ok {
		tokens = cachedTokens.([]pathtoken)
	} else {
		tokens = compilePath(path)
		cache.Add(path, tokens)
	}
	return tokens
}

func getPathValue(obj interface{}, path string) (interface{}, bool) {
	tokens := compilePathWithCache(path)
	return getPathValueFromTokens(obj, tokens)
}

func getPathValueFromTokens(obj interface{}, tokens []pathtoken) (interface{}, bool) {
	for _, t := range tokens {
		switch v := obj.(type) {
		case map[string]interface{}:
			if t.isSlice {
				// invalid path
				return nil, false
			}
			var ok bool
			obj, ok = v[t.key]
			if !ok {
				return nil, false
			}
		case []interface{}:
			if !t.isSlice {
				// invalid path
				return nil, false
			}
			if t.index < 0 {
				if -t.index > len(v) {
					// array index out of range
					return nil, false
				}
				obj = v[len(v)+t.index]
			} else if t.index >= len(v) {
				// array index out of range
				return nil, false
			} else {
				obj = v[t.index]
			}
		default:
			s := reflect.ValueOf(obj)
			if s.Kind() == reflect.Slice {
				if !t.isSlice {
					// invalid path
					return nil, false
				}
				if t.index < 0 {
					if -t.index > s.Len() {
						// array index out of range
						return nil, false
					}
					obj = s.Index(s.Len() + t.index).Interface()
				} else if t.index >= s.Len() {
					// array index out of range
					return nil, false
				} else {
					obj = s.Index(t.index).Interface()
				}
			} else {
				// TODO: reflect struct
				return nil, false
			}
		}
	}
	return obj, true
}

func setPathValue(obj map[string]interface{}, path string, v interface{}) bool {
	fieldSplits := strings.Split(path, ".")
	for i, key := range fieldSplits {
		if i >= len(fieldSplits)-1 {
			obj[key] = v
			return true
		} else if node, ok := obj[key]; ok {
			switch v := node.(type) {
			case map[string]interface{}:
				obj = v
			case nil:
				obj[key] = map[string]interface{}{}
				obj = obj[key].(map[string]interface{})
			default:
				return false
			}
		} else {
			obj[key] = map[string]interface{}{}
			obj = obj[key].(map[string]interface{})
		}
	}
	return false
}

func removePathValue(obj map[string]interface{}, field string) bool {
	fieldSplits := strings.Split(field, ".")
	for i, key := range fieldSplits {
		if i >= len(fieldSplits)-1 {
			delete(obj, key)
			return true
		} else if node, ok := obj[key]; ok {
			switch v := node.(type) {
			case map[string]interface{}:
				obj = v
			default:
				return false
			}
		} else {
			break
		}
	}
	return false
}
