package apps

import (
	"encoding/json"
	"github.com/ZTE/Knitter/knitter-monitor/domain/services"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/klog"
)

func GetPodApp() PodAppInterface {
	return &podApp{}
}

type PodAppInterface interface {
	Get(podNs, podName string) (*PodForAgent, error)
	PatchReportPod(podNs, podName string, body []byte) error
}

type podApp struct {
}

func (pa *podApp) PatchReportPod(podNs, podName string, body []byte) error {
	//todo add validate
	podReq := &agtmgr.AgentPodReq{}
	err := json.Unmarshal(body, podReq)
	if err != nil {
		klog.Errorf("podApp.PatchReportPod: json.Unmarshal(body:[%v], podReq:[%v]) error, err is [%v]", string(body), podReq, err)
		return err
	}
	err = services.GetPodService().PatchReportPod(podNs, podName, podReq)
	if err != nil {
		klog.Errorf("podApp.PatchReportPod: services.GetPodService().PatchReportPod"+
			"(podNs:[%v], podName:[%v], podReq:[%v]) error, err is [%v]", podNs, podName, podReq, err)
		return err
	}
	return nil

}

func (pa *podApp) Get(podNs, podName string) (*PodForAgent, error) {
	if podName == "" || podNs == "" {
		klog.Errorf("podApp.Get: check podName or PodNs err, err is [%v]", errobj.ErrPodNSOrPodNameIsNil)
		return nil, errobj.ErrPodNSOrPodNameIsNil
	}
	key := podNs + podName
	if services.GetCreatePorts4PodController().PodEventMap.Get(key) != nil {
		return nil, errobj.ErrPodIsOperating
	}
	pod, err := services.GetPodService().Get(podNs, podName)
	klog.Debugf("services.GetPodService().Get(podNs:[%v], podName:[%v]) pod is [%v]", podNs, podName, pod)
	if errobj.IsNotFoundError(err) {
		return nil, err
	}
	if err != nil {
		klog.Errorf("PodApp.Get err, error is [%v]", err)
		return nil, err
	}
	podForAgent := newPodForAgent(pod)
	return podForAgent, nil

}

type PodForAgent struct {
	TenantId     string          `json:"tenant_id"`
	PodID        string          `json:"pod_id"`
	PodName      string          `json:"pod_name"`
	PodNs        string          `json:"pod_ns"`
	PodType      string          `json:"pod_type"`
	IsSuccessful bool            `json:"is_successful"`
	ErrorMsg     string          `json:"error_msg"`
	Ports        []*PortForAgent `json:"ports"`
}

func newPodForAgent(pod *services.Pod) *PodForAgent {
	podForAgent := &PodForAgent{}
	aPorts := make([]*PortForAgent, 0)
	for _, port := range pod.Ports {
		portForAgent := newPortForAgent(port)
		aPorts = append(aPorts, portForAgent)
	}

	podForAgent.TenantId = pod.TenantId
	podForAgent.PodID = pod.PodID
	podForAgent.PodName = pod.PodName
	podForAgent.PodNs = pod.PodNs
	podForAgent.PodType = pod.PodType
	podForAgent.IsSuccessful = pod.IsSuccessful
	podForAgent.ErrorMsg = pod.ErrorMsg
	podForAgent.Ports = aPorts
	return podForAgent

}
