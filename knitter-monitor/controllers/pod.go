package controllers

import (
	"github.com/astaxie/beego"

	"github.com/ZTE/Knitter/knitter-monitor/apps"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/pkg/klog"
)

type PodController struct {
	beego.Controller
}

// Title Get
// Description find pod by pod_name
// Param	pod_name	path 	string	true		"the pod_name you want to get"
// Param	pod_ns	path 	string	true		"the pod_ns you want to get"
// Success 200 {object}
// Failure 404 :
// router /api/v1/pods/:podns/:podname [get]

func (pc *PodController) Get() {

	defer errobj.RecoverPanic()
	klog.Debugf("RECV agent get pod start  ")
	podns := pc.Ctx.Input.Param(":podns")
	podName := pc.Ctx.Input.Param(":podname")

	podForAgent, err := apps.GetPodApp().Get(podns, podName)
	if err != nil {
		if errobj.IsNotFoundError(err) {
			errobj.NotfoundErr404(&pc.Controller, err)
			return
		}
		klog.Debugf("apps.GetPodApp().Get(podns :[%v],podName:[%v]) err, error is [%v]", podns, podName, err)
		errobj.Err500(&pc.Controller, err)
		return
	}
	pc.Data["json"] = podForAgent
	klog.Debugf("Agent get pod end, pod for agent is [%+v]", podForAgent)

	pc.ServeJSON()
}

// Title PatchReportPod
// Description PatchReportPod  port ip
// Param	pod_name	path 	string	true		"the pod_name you want to get"
// Param	pod_ns	path 	string	true		"the pod_ns you want to get"
// Success 200 {object}
// Failure 500 :
// router /api/v1/pods/:podns/:podname [put]

func (pc *PodController) Put() {


	defer errobj.RecoverPanic()
	podns := pc.Ctx.Input.Param(":podns")
	podName := pc.Ctx.Input.Param(":podname")
	klog.Infof("RECV agent put pod start,pod name is [%v]", podName)
	body := pc.Ctx.Input.RequestBody
	err := apps.GetPodApp().PatchReportPod(podns, podName, body)
	if err != nil {
		klog.Errorf("apps.GetPodApp().PatchReportPod(podns :[%v],podName:[%v]) err, error is [%v]", podns, podName, err)
		errobj.Err500(&pc.Controller, err)
		return
	}
	klog.Infof("Agent put pod end ")

	pc.ServeJSON()
}
