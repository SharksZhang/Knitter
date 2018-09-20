package routers

import (
	"github.com/astaxie/beego"
	"github.com/ZTE/Knitter/knitter-monitor/controllers"
)

func init() {
	beego.Router("/api/v1/pods/:podns/:podname", &controllers.PodController{})
	beego.Router("/api/v1/loglevel/:log_level", &controllers.LogController{})
	beego.Router("/test/v1/namespaces/:namespace/rcs/:rcname", &controllers.ReplicationController{})

}