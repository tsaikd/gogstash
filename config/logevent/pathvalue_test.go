package logevent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkCompilePath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compilePath("a.b.c")
		compilePath("a.b[1].c")
		compilePath("a.b[0][2].c")
		compilePath("[0]")
		compilePath("nginx.access.url")
	}
}

func BenchmarkCompilePathWithCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compilePathWithCache("a.b.c")
		compilePathWithCache("a.b[1].c")
		compilePathWithCache("a.b[0][2].c")
		compilePathWithCache("[0]")
		compilePathWithCache("nginx.access.url")
	}
}

func TestCompilePath(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	tokens := compilePath("a.b.c")
	assert.Equal([]pathtoken{
		pathtoken{isSlice: false, key: "a"},
		pathtoken{isSlice: false, key: "b"},
		pathtoken{isSlice: false, key: "c"},
	}, tokens)

	tokens = compilePath("a.b[1].c")
	assert.Equal([]pathtoken{
		pathtoken{isSlice: false, key: "a"},
		pathtoken{isSlice: false, key: "b"},
		pathtoken{isSlice: true, index: 1},
		pathtoken{isSlice: false, key: "c"},
	}, tokens)

	tokens = compilePath("a.b[0][2].c")
	assert.Equal([]pathtoken{
		pathtoken{isSlice: false, key: "a"},
		pathtoken{isSlice: false, key: "b"},
		pathtoken{isSlice: true, index: 0},
		pathtoken{isSlice: true, index: 2},
		pathtoken{isSlice: false, key: "c"},
	}, tokens)

	tokens = compilePath("[0]")
	assert.Equal([]pathtoken{
		pathtoken{isSlice: true, index: 0},
	}, tokens)
}

func TestGetPathValue(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	d := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "foo",
			},
		},
	}

	r, ok := getPathValue(d, "a.b.c")
	assert.True(ok)
	assert.Equal("foo", r)

	// string will not return the first byte
	r, ok = getPathValue(d, "a.b.c[0]")
	assert.False(ok)
	assert.Nil(r)

	r, ok = getPathValue(d, "a.b.c.d")
	assert.False(ok)
	assert.Nil(r)

	d = map[string]interface{}{
		"a": map[string]interface{}{
			"b": []interface{}{
				map[string]interface{}{"c": "foo"},
			},
		},
	}

	r, ok = getPathValue(d, "a.b[0].c")
	assert.True(ok)
	assert.Equal("foo", r)

	r, ok = getPathValue(d, "a.b.c")
	assert.False(ok)
	assert.Nil(r)

	r, ok = getPathValue(d, "a.b.c.d")
	assert.False(ok)
	assert.Nil(r)

	r, ok = getPathValue(d, "a[0].c")
	assert.False(ok)
	assert.Nil(r)

	d = map[string]interface{}{
		"a": map[string]interface{}{
			"b": []string{"c", "foo"},
		},
	}

	r, ok = getPathValue(d, "a.b[1]")
	assert.True(ok)
	assert.Equal("foo", r)

	r, ok = getPathValue([]string{"a", "b", "c", "foo"}, "[3]")
	assert.True(ok)
	assert.Equal("foo", r)

	d = map[string]interface{}{
		"a": []interface{}{"b", "c", "foo"},
	}
	r, ok = getPathValue(d, "a[-1]")
	assert.True(ok)
	assert.Equal("foo", r)

	r, ok = getPathValue(d, "a[-3]")
	assert.True(ok)
	assert.Equal("b", r)

	r, ok = getPathValue(d, "a[-10]")
	assert.False(ok)
	assert.Nil(r)

	r, ok = getPathValue(d, "a[3]")
	assert.False(ok)
	assert.Nil(r)

	d = map[string]interface{}{
		"a": []string{"b", "c", "foo"},
	}
	r, ok = getPathValue(d, "a[-1]")
	assert.True(ok)
	assert.Equal("foo", r)

	r, ok = getPathValue(d, "a[-10]")
	assert.False(ok)
	assert.Nil(r)

	r, ok = getPathValue(d, "a[3]")
	assert.False(ok)
	assert.Nil(r)

	r, ok = getPathValue(d, "a.b.c")
	assert.False(ok)
	assert.Nil(r)
}
