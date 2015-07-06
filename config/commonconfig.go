package config

import (
	"reflect"

	"github.com/codegangsta/inject"
	"github.com/tsaikd/KDGoLib/injectutil"
)

type TypeConfig interface {
	SetInjector(inj inject.Injector)
	GetType() string
	Invoke(f interface{}) (refvs []reflect.Value, err error)
}

type CommonConfig struct {
	inject.Injector `json:"-"`
	Type            string `json:"type"`
}

func (t *CommonConfig) SetInjector(inj inject.Injector) {
	t.Injector = inj
}

func (t *CommonConfig) GetType() string {
	return t.Type
}

func (t *CommonConfig) Invoke(f interface{}) (refvs []reflect.Value, err error) {
	return injectutil.Invoke(t.Injector, f)
}

type ConfigRaw map[string]interface{}
