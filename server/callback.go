package server

import (
	"dxkite.cn/log"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
)

type HttpContext struct {
	Request  *http.Request
	Response http.ResponseWriter
}

type callback struct {
	f interface{}
}

func NewCallback(fun interface{}) http.Handler {
	return &callback{fun}
}

var errType = reflect.TypeOf(errors.New("error type"))

func (fh *callback) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	rf := reflect.ValueOf(fh.f)

	isStruct := rf.Kind() == reflect.Ptr && rf.Elem().Kind() == reflect.Struct || rf.Kind() == reflect.Struct

	if rf.Kind() != reflect.Func && !isStruct {
		writeMsg(w, "handler must be function, current is "+rf.String(), nil)
		return
	}

	if isStruct {
		method := req.URL.Query().Get("method")
		st := rf

		if st.NumMethod() < 1 {
			writeMsg(w, "no method: "+st.String(), nil)
			return
		}

		rf = st.MethodByName(method)

		if !rf.IsValid() {
			method = "Call"
			rf = st.MethodByName(method)
		}

		if !rf.IsValid() {
			writeMsg(w, "call not found: "+method, nil)
			return
		}
	}

	if rf.Type().NumOut() < 1 {
		writeMsg(w, "handler must return error", nil)
		return
	}

	if rf.Type().Out(0).ConvertibleTo(reflect.TypeOf(errType)) {
		writeMsg(w, "handler must return error", nil)
		return
	}

	if rf.Type().NumIn() < 1 {
		writeMsg(w, "handler input must be (in, *out [, http.ResponseWriter ])", nil)
		return
	}

	params := []reflect.Value{}

	reqValue := reflect.New(rf.Type().In(0))
	respValue := reflect.New(rf.Type().In(1).Elem())

	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		writeMsg(w, err.Error(), nil)
		return
	}

	if err := json.Unmarshal(body, reqValue.Interface()); err != nil {
		writeMsg(w, "json decode error:"+err.Error(), map[string]interface{}{
			"request":  reqValue.Interface(),
			"response": respValue.Interface(),
		})
		return
	}

	params = append(params, reqValue.Elem())
	params = append(params, respValue)

	numIn := rf.Type().NumIn()

	if numIn > 2 && rf.Type().In(2).ConvertibleTo(reflect.TypeOf(w)) {
		params = append(params, reflect.ValueOf(w))
	}

	if numIn > 2 && rf.Type().In(2).ConvertibleTo(reflect.TypeOf(&HttpContext{})) {
		params = append(params, reflect.ValueOf(&HttpContext{req, w}))
	}

	log.Println(params[1].IsZero())
	ret := rf.Call(params)
	writeMsg(w, ret[0].Interface(), params[1].Interface())
}
