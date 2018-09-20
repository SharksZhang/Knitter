package errobj

import (
	"github.com/astaxie/beego/context"
	"github.com/bouk/monkey"
	"html/template"
	"io"
	"net/http"
	"reflect"
	"testing"
)

func TestPageNotFound(t *testing.T) {
	var tm *template.Template
	monkey.PatchInstanceMethod(reflect.TypeOf(tm), "Execute", func(_ *template.Template, wr io.Writer, data interface{}) error {
		return nil
	})
	rw := &context.Response{}
	r := &http.Request{}
	PageNotFound(rw, r)
}
