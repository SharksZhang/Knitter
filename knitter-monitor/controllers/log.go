package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/syndtr/goleveldb/leveldb/errors"

	"github.com/ZTE/Knitter/pkg/klog"
)

type LogController struct {
	beego.Controller
}

// @Title modify log level
// @Description modify log level for klog
// @Failure 400 params is invalid
func (l *LogController) Put() {
	logLevel := l.GetString(":log_level")
	level, err := strconv.Atoi(logLevel)
	if err != nil {
		klog.Errorf("input log level is: %s, can not transform to integer", logLevel)
		Err400(&l.Controller, err)
		return
	}

	levelNum := klog.Level(level)

	if levelNum < klog.TraceLevel || levelNum >= klog.NumLevel {
		klog.Errorf("input new log level is: %s, invalid level", logLevel)
		Err400(&l.Controller, errors.New("invalid log level number, should between 0 to 5"))
		return
	}

	klog.SetLogLevel(levelNum)
	klog.Errorf("error level")
	klog.Warningf("warning level")
	klog.Infof("info level")
	klog.Debugf("debug level")
	klog.Tracef("trace level")
	l.Data["json"] = fmt.Sprintf("modify log level to %d success", levelNum)
	l.ServeJSON()
}

func HandleErr(o *beego.Controller, err error) {
	klog.Info("HandleErr:", err)

	parts := strings.Split(err.Error(), "::")
	var i int
	var msg string

	if len(parts) < 2 {
		i = http.StatusInternalServerError
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
	o.Redirect(o.Ctx.Request.URL.RequestURI(), i)
	o.ServeJSON()
}

func BuildErrWithCode(code int, err error) error {
	status := http.StatusText(code)
	if status == "" {
		return fmt.Errorf("%v::%v", http.StatusInternalServerError, err)
	}
	return fmt.Errorf("%v::%v", code, err)
}

func Err400(o *beego.Controller, err error) {
	HandleErr(o, BuildErrWithCode(http.StatusBadRequest, err))
}
