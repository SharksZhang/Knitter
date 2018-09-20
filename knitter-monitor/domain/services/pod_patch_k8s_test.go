package services

import (
	"encoding/json"
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/bouk/monkey"
	"github.com/pkg/errors"
	"github.com/smartystreets/goconvey/convey"

	"github.com/ZTE/Knitter/knitter-monitor/infra/clients"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
)

func TestReportPodToK8s(t *testing.T) {
	convey.Convey("TestReportPodToK8s succ\n", t, func() {
		reportPod := &clients.ReportPod{
			PodNs:   "pod_ns",
			PodName: "pod_name",
		}
		monkey.Patch(makePortsForK8sPodsPatch, func(ports []clients.PortInfo) []portForPatch {
			return []portForPatch{}
		})
		defer monkey.UnpatchAll()
		monkey.Patch(makeArgsForK8sPodsPatch, func(portsForPatch []portForPatch) ([]ArgsForPodsPatch, error) {
			return []ArgsForPodsPatch{}, nil
		})
		monkey.Patch(json.Marshal, func(v interface{}) ([]byte, error) {
			return []byte{}, nil
		})
		client := &MonitorK8sClient{}
		monkey.Patch(GetMonitorK8sClient, func() *MonitorK8sClient {
			return client
		})
		monkey.PatchInstanceMethod(reflect.TypeOf(client), "GetPodsPatchController", func(mc *MonitorK8sClient, nameSpace string, name string, pt types.PatchType, data []byte, subresources ...string) (*v1.Pod, error) {
			return &v1.Pod{}, nil
		})
		err := reportPodToK8s(reportPod)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestReportPodToK8s makeArgsForK8sPodsPatch fail\n", t, func() {
		reportPod := &clients.ReportPod{
			PodNs:   "pod_ns",
			PodName: "pod_name",
		}
		monkey.Patch(makePortsForK8sPodsPatch, func(ports []clients.PortInfo) []portForPatch {
			return []portForPatch{}
		})
		defer monkey.UnpatchAll()
		monkey.Patch(makeArgsForK8sPodsPatch, func(portsForPatch []portForPatch) ([]ArgsForPodsPatch, error) {
			return nil, errors.New("makeArgsForK8sPodsPatch fail")
		})
		err := reportPodToK8s(reportPod)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("TestReportPodToK8s Marshal fail\n", t, func() {
		reportPod := &clients.ReportPod{
			PodNs:   "pod_ns",
			PodName: "pod_name",
		}
		monkey.Patch(makePortsForK8sPodsPatch, func(ports []clients.PortInfo) []portForPatch {
			return []portForPatch{}
		})
		defer monkey.UnpatchAll()
		monkey.Patch(makeArgsForK8sPodsPatch, func(portsForPatch []portForPatch) ([]ArgsForPodsPatch, error) {
			return []ArgsForPodsPatch{}, nil
		})
		monkey.Patch(json.Marshal, func(v interface{}) ([]byte, error) {
			return nil, errors.New("Marshal fail")
		})
		err := reportPodToK8s(reportPod)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("TestReportPodToK8s GetPodsPatchController fail\n", t, func() {
		reportPod := &clients.ReportPod{
			PodNs:   "pod_ns",
			PodName: "pod_name",
		}
		monkey.Patch(makePortsForK8sPodsPatch, func(ports []clients.PortInfo) []portForPatch {
			return []portForPatch{}
		})
		defer monkey.UnpatchAll()
		monkey.Patch(makeArgsForK8sPodsPatch, func(portsForPatch []portForPatch) ([]ArgsForPodsPatch, error) {
			return []ArgsForPodsPatch{}, nil
		})
		monkey.Patch(json.Marshal, func(v interface{}) ([]byte, error) {
			return []byte{}, nil
		})
		client := &MonitorK8sClient{}
		monkey.Patch(GetMonitorK8sClient, func() *MonitorK8sClient {
			return client
		})
		monkey.PatchInstanceMethod(reflect.TypeOf(client), "GetPodsPatchController", func(mc *MonitorK8sClient, nameSpace string, name string, pt types.PatchType, data []byte, subresources ...string) (*v1.Pod, error) {
			return nil, errors.New("GetPodsPatchController fail")
		})
		err := reportPodToK8s(reportPod)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestMakePortsForK8sPodsPatch(t *testing.T) {
	convey.Convey("TestMakePortsForK8sPodsPatch succ\n", t, func() {
		ports := []clients.PortInfo{
			{
				PortID:        "port_id1",
				NetworkName:   "network_name1",
				NetworkPlanes: "network_planes1",
				FixedIPInfos: []iaasaccessor.FixedIPItem{
					{
						IPVersion: uint8(4),
						IPAddress: "ipv4_1",
					},
					{
						IPVersion: uint8(6),
						IPAddress: "ipv6_1",
					},
				},
			},
			{
				PortID:        "port_id2",
				NetworkName:   "network_name2",
				NetworkPlanes: "network_planes2",
				FixedIPInfos: []iaasaccessor.FixedIPItem{
					{
						IPVersion: uint8(4),
						IPAddress: "ipv4_2",
					},
					{
						IPVersion: uint8(6),
						IPAddress: "ipv6_2",
					},
				},
			},
		}
		portsForK8s := []portForPatch{
			{
				Function:    "network_planes1",
				NetworkName: "network_name1",
				IPAddress:   "ipv4_1",
				IPv6Address: "ipv6_1",
			},
			{
				Function:    "network_planes2",
				NetworkName: "network_name2",
				IPAddress:   "ipv4_2",
				IPv6Address: "ipv6_2",
			},
		}
		portsForPatch := makePortsForK8sPodsPatch(ports)
		convey.So(portsForPatch, convey.ShouldResemble, portsForK8s)
	})
}

func TestMakeArgsForK8sPodsPatch(t *testing.T) {
	convey.Convey("TestMakeArgsForK8sPodsPatch succ\n", t, func() {
		portsForPatch := []portForPatch{
			{
				Function:    "network_planes1",
				NetworkName: "network_name1",
				IPAddress:   "ipv4_1",
				IPv6Address: "ipv6_1",
			},
			{
				Function:    "network_planes2",
				NetworkName: "network_name2",
				IPAddress:   "ipv4_2",
				IPv6Address: "ipv6_2",
			},
		}
		args, err := makeArgsForK8sPodsPatch(portsForPatch)
		convey.So(err, convey.ShouldBeNil)
		convey.So(args[0].Op, convey.ShouldEqual, "add")
		convey.So(args[0].Path, convey.ShouldEqual, "/metadata/annotations/network.knitter.io~1configuration-result")
	})
}
