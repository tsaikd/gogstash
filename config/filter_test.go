package config

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsaikd/gogstash/config/logevent"
)

func TestCommonAddTag(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	event := logevent.LogEvent{}

	assert.Equal(0, len(event.Tags))

	filter := FilterConfig{
		AddTags: []string{"add"},
	}
	event = filter.CommonFilter(context.TODO(), event)

	assert.Equal(1, len(event.Tags))
	assert.Equal("add", event.Tags[0])
}

func TestCommonIsConfigured(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	var filter FilterConfig

	assert.False(filter.IsConfigured(), "should be not configured")
	withAddTag := FilterConfig{AddTags: []string{"tag"}}
	assert.True(withAddTag.IsConfigured(), "should be configured")
	withAddFields := FilterConfig{AddFields: []FieldConfig{{Key: "name", Value: "value"}}}
	assert.True(withAddFields.IsConfigured(), "should be configured")
	withRemoveTag := FilterConfig{RemoveTags: []string{"tag"}}
	assert.True(withRemoveTag.IsConfigured(), "should be configured")
	withRemoveFields := FilterConfig{RemoveFields: []string{"field"}}
	assert.True(withRemoveFields.IsConfigured(), "should be configured")
}

func TestCommonRemoveTag(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	event := logevent.LogEvent{Tags: []string{"removeme"}}

	assert.Equal("removeme", event.Tags[0])

	filter := FilterConfig{
		RemoveTags: []string{"removeme"},
	}
	event = filter.CommonFilter(context.TODO(), event)

	assert.Equal(0, len(event.Tags))
}

func TestCommonAddField(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	event := logevent.LogEvent{}

	assert.Equal("", event.GetString("addme"))

	filter := FilterConfig{
		AddFields: []FieldConfig{
			{Key: "addme", Value: "value"},
			{Key: "addme2", Value: "value2"},
		},
	}
	event = filter.CommonFilter(context.TODO(), event)

	assert.Equal("value", event.GetString("addme"))
	assert.Equal("value", event.Get("addme"))

	assert.Equal("value2", event.GetString("addme2"))
	assert.Equal("value2", event.Get("addme2"))
}

func TestCommonRemoveField(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	event := logevent.LogEvent{}
	event.SetValue("removeme", "value")

	assert.Equal("value", event.GetString("removeme"))
	assert.Equal("value", event.Get("removeme"))

	filter := FilterConfig{
		RemoveFields: []string{"removeme"},
	}
	event = filter.CommonFilter(context.TODO(), event)

	assert.Equal("", event.GetString("field"))
	assert.Equal(nil, event.Get("field"))
}

func TestLoadYaml(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)

	conf, err := LoadFromYAML([]byte(strings.TrimSpace(`
filter:
  - type: "whatever"

    # list of tags to add
    add_tag: ["addtag1", "addtag2"]

    # list of tags to remove
    remove_tag: ["removetag1", "removetag2"]

    # list of fields (key/value) to add
    add_field:
      - key: "field1"
        value: "value1"
      - key: "field2"
        value: "value2"
    # list of fields to remove
    remove_field: ["removefield1", "removefield2"]
	`)))
	assert.NoError(err)

	assert.Equal(0, len(conf.InputRaw))
	assert.Equal(1, len(conf.FilterRaw))
	assert.Equal(0, len(conf.OutputRaw))

	// to have config updated
	mapFilterHandler["whatever"] = func(ctx context.Context, raw ConfigRaw, control Control) (TypeFilterConfig, error) {
		conf := WhateverFilterConfig{}
		err := ReflectConfig(raw, &conf)
		if err != nil {
			return nil, err
		}
		return &conf, nil
	}
	filters, err := conf.getFilters()
	assert.NoError(err)
	assert.Equal(1, len(filters))

	tags := []string{"removetag1", "removetag2"}
	event := logevent.LogEvent{Tags: tags}
	event.SetValue("removefield1", "value1")
	event.SetValue("removefield2", "value2")

	assert.Equal(tags, event.Tags)
	assert.Equal("value1", event.GetString("removefield1"))
	assert.Equal("value2", event.GetString("removefield2"))
	assert.Equal("value1", event.Get("removefield1"))
	assert.Equal("value2", event.Get("removefield2"))

	event = filters[0].CommonFilter(context.TODO(), event)

	newTags := []string{"addtag1", "addtag2"}
	assert.Equal(newTags, event.Tags)
	assert.Equal("", event.GetString("removefield1"))
	assert.Equal("", event.GetString("removefield2"))
	assert.Equal(nil, event.Get("removefield1"))
	assert.Equal(nil, event.Get("removefield2"))
	assert.Equal("value1", event.GetString("field1"))
	assert.Equal("value2", event.GetString("field2"))
	assert.Equal("value1", event.Get("field1"))
	assert.Equal("value2", event.Get("field2"))
}

type WhateverFilterConfig struct {
	FilterConfig
}

func (f *WhateverFilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	return event, true
}
