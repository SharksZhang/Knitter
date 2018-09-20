package services

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/types"
	"github.com/ZTE/Knitter/knitter-monitor/infra/clients"
	"github.com/ZTE/Knitter/pkg/klog"
)

type ArgsForPodsPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

type ValueOfPodsPatch struct {
	Version string         `json:"version"`
	Ports   []portForPatch `json:"ports"`
}

type portForPatch struct {
	Function    string `json:"function"`
	NetworkName string `json:"network_name"`
	IPAddress   string `json:"ip_address"`
	IPv6Address string `json:"ipv6_address"`
}

func reportPodToK8s(reportPod *clients.ReportPod) error {
	portsForPatch := makePortsForK8sPodsPatch(reportPod.Ports)
	args, err := makeArgsForK8sPodsPatch(portsForPatch)
	if err != nil {
		return err
	}
	patchBytes, err := json.Marshal(args)
	if err != nil {
		klog.Errorf("reportPodToK8s: Marshal(args: %v), error: %v", args, err)
		return err
	}
	_, err = GetMonitorK8sClient().GetPodsPatchController(reportPod.PodNs, reportPod.PodName, types.JSONPatchType, patchBytes)
	if err != nil {
		return err
	}
	return nil
}

func makePortsForK8sPodsPatch(ports []clients.PortInfo) []portForPatch {
	portsForK8s := make([]portForPatch, 0)
	for _, port := range ports {
		var ipv4, ipv6 string
		for _, ipInfo := range port.FixedIPInfos {
			if ipInfo.IPVersion == uint8(4) {
				ipv4 = ipInfo.IPAddress
			}
			if ipInfo.IPVersion == uint8(6) {
				ipv6 = ipInfo.IPAddress
			}
		}

		tmpPort := portForPatch{
			Function:    port.NetworkPlanes,
			NetworkName: port.NetworkName,
			IPAddress:   ipv4,
			IPv6Address: ipv6,
		}
		portsForK8s = append(portsForK8s, tmpPort)
	}
	return portsForK8s
}

func makeArgsForK8sPodsPatch(portsForPatch []portForPatch) ([]ArgsForPodsPatch, error) {
	const (
		addOptionOfPodsPatch     string = "add"
		cfgResultPathOfPodsPatch string = "/metadata/annotations/network.knitter.io~1configuration-result"
		versionOfPodsPatch       string = "v1"
	)

	args := make([]ArgsForPodsPatch, 1)
	args[0].Op = addOptionOfPodsPatch
	args[0].Path = cfgResultPathOfPodsPatch
	value := ValueOfPodsPatch{
		Version: versionOfPodsPatch,
		Ports:   portsForPatch,
	}
	valueBytes, err := json.Marshal(value)
	if err != nil {
		klog.Errorf("makeArgsForK8sPodsPatch: Marshal(value: %v), error: %v", value, err)
		return nil, err
	}
	args[0].Value = string(valueBytes)

	klog.Infof("makeArgsForK8sPodsPatch: args for k8s patch: %+v", args)
	return args, nil
}
