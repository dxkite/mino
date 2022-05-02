package handler

import (
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/server/comm"
	"dxkite.cn/mino/server/context"
	"dxkite.cn/mino/util"
	"net/http"
	"reflect"
	"strings"
)

type ConfigProperty struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	ReadOnly    bool   `json:"readOnly"`
	Description string `json:"description"`
}

type ConfigSchema struct {
	Properties map[string]*ConfigProperty `json:"properties"`
}

func NewConfigHandler(ctx *context.Context) http.Handler {
	sm := http.NewServeMux()

	schema := BuildSchemaFromConfig(ctx)
	sm.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		comm.WriteResp(w, nil, ctx.Cfg)
	})

	sm.HandleFunc("/schema", func(w http.ResponseWriter, r *http.Request) {
		comm.WriteResp(w, nil, schema)
	})

	sm.Handle("/set", NewCallbackHandler(func(from map[string]interface{}, success *[]string) (err error) {
		*success, err = config.CopyObjectMap(ctx.Cfg, from)
		ctx.Cfg.Notify()
		return
	}))
	return sm
}

func BuildSchemaFromConfig(ctx *context.Context) *ConfigSchema {
	s := &ConfigSchema{map[string]*ConfigProperty{}}
	v := reflect.ValueOf(ctx.Cfg).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		tag := t.Field(i).Tag.Get("json")
		prop := strings.ToLower(t.Field(i).Tag.Get("prop"))

		name := util.TagName(tag)
		if name == "-" || len(name) == 0 {
			continue
		}

		title := t.Field(i).Tag.Get("title")
		if len(title) == 0 {
			title = name
		}

		desc := t.Field(i).Tag.Get("desc")

		typ := "string"
		switch f.Kind() {
		case reflect.Bool:
			typ = "boolean"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			typ = "integer"
		}
		s.Properties[name] = &ConfigProperty{
			Type:        typ,
			Title:       title,
			ReadOnly:    strings.Index(prop, "readonly") >= 0,
			Description: desc,
		}
	}
	return s
}
