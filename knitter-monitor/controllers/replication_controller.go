package controllers

import (
	"github.com/astaxie/beego"
	"github.com/ZTE/Knitter/knitter-monitor/domain/services"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/pkg/klog"
)

type ReplicationController struct {
	beego.Controller
}

//	beego.Router("/test/v1/namespaces/:namespace/rcs/:rcname", &controllers.ReplicationController{})

func (pc *ReplicationController) Get() {
	namespace := pc.GetString(":namespace")
	rcName := pc.GetString(":rcname")
	rc, err := services.GetMonitorK8sClient().GetReplicationController(namespace, rcName)
	if err != nil {
		klog.Errorf("IsReplicationControllerExist(namespace:[%v], rcName:[%v])", namespace, rcName, err)
		errobj.Err500(&pc.Controller, err)
		return
	}
	pc.Data["json"] = rc
	klog.Infof("Agent get pod end, pod for agent is [%+v]", rc)

	pc.ServeJSON()
}
