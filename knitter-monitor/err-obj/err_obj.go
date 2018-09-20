package errobj

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/astaxie/beego"

	"html/template"
	"runtime/debug"
	"github.com/ZTE/Knitter/pkg/klog"
)

func IsEqual(errLeft, errRight error) bool {
	return errLeft.Error() == errRight.Error()
}

var (
	NoNetworkMessageInBluePrint = errors.New("no network message in bluePrint")
	ErrPortNotFound             = errors.New("port not found")
	ErrJasonNewObjectFailed     = errors.New("jason new object failed")
	ErrJasonGetStringFailed     = errors.New("jason get string failed")
	ErrGetPortConfigError       = errors.New("get port config error")
	ErrPodNSOrPodNameIsNil      = errors.New("podNs is nil or podName is nil")
	ErrDeleteBulkPortsError     = errors.New("delete ports error")
	ErrPodIsOperating           = errors.New("pod is operating")
	ErrResourceInUse            = errors.New("resource is using, wait delete")
	ErrK8sMasterUrlIsNil        = errors.New("k8s master url is nil")
)

func GetErrMsg(respData []byte) string {
	var respMap map[string]string

	if respData == nil {
		return ""
	}
	err := json.Unmarshal(respData, &respMap)
	if err != nil {
		return ""
	}
	return respMap["message"]
}

var EtcdKeyNotFound = "Key not found"
var RecordNotExist = "record already not exsit"
var k8sResourceNotFound = "not found"

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	if strings.Contains(errStr, EtcdKeyNotFound) || strings.Contains(errStr, RecordNotExist) ||
		strings.Contains(errStr, k8sResourceNotFound) {
		return true
	}
	return false
}

func Err500(o *beego.Controller, err error) {
	HandleErr(o, BuildErrWithCode(http.StatusInternalServerError, err))
}

func BuildErrWithCode(code int, err error) error {
	status := http.StatusText(code)
	if status == "" {
		return fmt.Errorf("%v::%v", http.StatusInternalServerError, err)
	}
	return fmt.Errorf("%v::%v", code, err)
}

func HandleErr(o *beego.Controller, err error) {
	klog.Info("HandleErr:", err)
	parts := strings.Split(err.Error(), "::")
	var i int
	var msg string
	if len(parts) < 2 {
		i = http.StatusInternalServerError
		msg = http.StatusText(i)
	} else {
		i, _ = strconv.Atoi(parts[0])
		if i == 0 {
			i = http.StatusInternalServerError
		}

		msg = http.StatusText(i)
		if msg == "" {
			i = http.StatusInternalServerError
			msg = http.StatusText(i)
		}
	}

	o.Data["json"] = map[string]string{"ERROR": msg,
		"message": parts[len(parts)-1]}
	o.Ctx.Output.SetStatus(i)
	o.ServeJSON()
}

func NotfoundErr404(o *beego.Controller, err error) {
	HandleErr(o, BuildErrWithCode(http.StatusNotFound, err))
}

func RecoverPanic() {
	defer klog.Flush()
	if p := recover(); p != nil {
		klog.Errorf("panic: [%v] ,Stack:%v", p, string(debug.Stack()))
	}
}

func PageNotFound(rw http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"StatusCode": "404",
		"Error":      "Bad URI",
		"Message":    "",
	}
	bytes, _ := json.MarshalIndent(data, "", "  ")
	t, _ := template.New("404.html").Parse(string(bytes))
	t.Execute(rw, "")
}

var (
	ErrObjectPointerIsNil = errors.New("object pointer is nil")
	ErrRecordNotExist     = errors.New(RecordNotExist)
	ErrObjectTypeMismatch = errors.New("object type mismatch")
	ErrArgTypeMismatch    = errors.New("argument type mismatch")
	ErrUnmarshalFailed    = errors.New("json.Unmarshal Err")
)
