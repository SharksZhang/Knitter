package services

import (
	"errors"
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bouk/monkey"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"

	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/daos"
	"github.com/ZTE/Knitter/knitter-monitor/infra/clients"
	"github.com/ZTE/Knitter/knitter-monitor/tests/mocks/daos"
)

func TestPodService_NewPodFromK8sPodSucc(t *testing.T) {
	podNS, PodName, k8sPod := createK8sPod()

	portsForService, podForServiceExpect := createPod(PodName, podNS)

	monkey.Patch(createPortByK8sAnnotations,
		func(annotations map[string]string, pod4CreatePort *PodForCreatPort) ([]*Port, error) {
			return portsForService, nil
		})
	defer monkey.UnpatchAll()
	convey.Convey("TestNewPodFromK8sPodSucc", t, func() {
		ps := &podService{}
		pod, err := ps.NewPodFromK8sPod(k8sPod)
		convey.So(pod, convey.ShouldResemble, podForServiceExpect)
		convey.So(err, convey.ShouldBeNil)
	})

}

func TestPodService_NewPodFromK8sPodFail(t *testing.T) {
	_, _, k8sPod := createK8sPod()

	monkey.Patch(createPortByK8sAnnotations,
		func(annotations map[string]string, pod4CreatePort *PodForCreatPort) ([]*Port, error) {
			return nil, errors.New("create err")
		})
	defer monkey.UnpatchAll()
	convey.Convey("TestNewPodFromK8sPodSucc", t, func() {
		ps := &podService{}
		_, err := ps.NewPodFromK8sPod(k8sPod)
		convey.So(err.Error(), convey.ShouldEqual, "create err")
	})

}

func createPod(PodName string, podNS string) ([]*Port, *Pod) {
	portEagerAttrApi := PortEagerAttr{
		NetworkName:  "net_api",
		NetworkPlane: "net_api",
		PortName:     "eth0",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "192.168.1.0",
		IPGroupName:  "ig0",
		Metadata:     "",
		Combinable:   "false",
		Roles:        []string{"control"},
	}

	portLazyAttrApi := PortLazyAttr{
		ID:         "net_api",
		Name:       "eth0",
		TenantID:   podNS,
		MacAddress: "",
	}

	portEagerAttrControl := PortEagerAttr{
		NetworkName:  "control",
		NetworkPlane: "control",
		PortName:     "nwrfcontrol",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "192.168.1.1",
		IPGroupName:  "ig1",
		Metadata:     "",
		Combinable:   "false",
		Roles:        []string{"control"},
	}
	portLazyAttrControl := PortLazyAttr{
		ID:         "11111-control",
		Name:       "nwrfcontrol",
		TenantID:   podNS,
		MacAddress: "",
	}

	portEagerAttrMedia := PortEagerAttr{
		NetworkName:  "media",
		NetworkPlane: "media",
		PortName:     "eth2",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "192.168.1.2",
		IPGroupName:  "ig2",
		Metadata:     "",
		Combinable:   "false",
		Roles:        []string{"media"},
	}
	portLazyAttrMedia := PortLazyAttr{
		ID:         "11111-media",
		Name:       "eth1",
		TenantID:   podNS,
		MacAddress: "",
	}
	portsForService := []*Port{
		{
			EagerAttr: portEagerAttrApi,
			LazyAttr:  portLazyAttrApi,
		},
		{
			EagerAttr: portEagerAttrControl,
			LazyAttr:  portLazyAttrControl,
		},
		{
			EagerAttr: portEagerAttrMedia,
			LazyAttr:  portLazyAttrMedia,
		},
	}
	podForServiceExpect := &Pod{
		TenantId:     podNS,
		PodID:        "testpod-1111",
		PodName:      PodName,
		PodNs:        podNS,
		PodType:      "",
		IsSuccessful: true,
		ErrorMsg:     "",
		Ports:        portsForService,
	}
	return portsForService, podForServiceExpect
}

func createK8sPod() (string, string, *v1.Pod) {
	podNS := "admin"
	PodName := "pod1"
	networks := "{\"ports\": " +
		"[{\"attach_to_network\": \"net_api\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"std\", \"nic_name\": \"eth0\", \"nic_type\": \"normal\"}}," +
		" {\"attach_to_network\": \"control\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"control\", \"nic_name\": \"nwrfcontrol\", \"nic_type\": \"normal\"}}, " +
		"{\"attach_to_network\": \"media\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"media\", \"nic_name\": \"nwrfmedia\", \"nic_type\": \"normal\"}}]}"
	annotations := map[string]string{
		"networks": networks,
	}
	objectMeta := metav1.ObjectMeta{
		Name:        PodName,
		UID:         "testpod-1111",
		Namespace:   podNS,
		Annotations: annotations,
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}
	return podNS, PodName, k8sPod
}

func TestIsNetworkNotConfigExist(t *testing.T) {

	convey.Convey("TestIsNetworkNotConfigExist", t, func() {
		nwStr := ""
		ok := true
		expect := isNotNetworkConfigExist(nwStr, ok)
		convey.So(expect, convey.ShouldBeTrue)

		nwStr = "\"\""
		expect = isNotNetworkConfigExist(nwStr, ok)
		convey.So(expect, convey.ShouldBeTrue)

		ok = false
		nwStr = ""
		expect = isNotNetworkConfigExist(nwStr, ok)
		convey.So(expect, convey.ShouldBeTrue)

		ok = true
		nwStr = "nwstr"
		expect = isNotNetworkConfigExist(nwStr, ok)
		convey.So(expect, convey.ShouldBeFalse)

	})
}

func createDefaultNetworkMessage(defaultNetName string) *BluePrintNetworkMessage {
	bpnmExpect := &BluePrintNetworkMessage{}
	ports := make([]BluePrintPort, 0)
	port := BluePrintPort{
		AttachToNetwork: defaultNetName,
		Attributes: BluePrintAttributes{
			Accelerate: constvalue.DefaultIsAccelerate,
			Function:   constvalue.DefaultNetworkPlane,
			NicName:    constvalue.DefaultPortName,
			NicType:    constvalue.DefaultVnicType,
		},
	}
	ports = append(ports, port)
	bpnmExpect.Ports = ports
	return bpnmExpect
}

func TestPodService_SaveSucc(t *testing.T) {
	podNs := "admin"
	podName := "test1"

	pod := &Pod{
		TenantId: podNs,
		PodName:  podName,
		PodNs:    podNs,
	}

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Save(gomock.Any()).Return(nil)

	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodService_GetSucc", t, func() {
		err := GetPodService().Save(pod)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPodService_SaveFail(t *testing.T) {
	podNs := "admin"
	podName := "test1"
	pod := &Pod{
		TenantId: podNs,
		PodName:  podName,
		PodNs:    podNs,
	}

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Save(gomock.Any()).Return(errors.New("save err"))

	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodService_SaveFail", t, func() {
		err := GetPodService().Save(pod)
		convey.So(err.Error(), convey.ShouldEqual, "save err")
	})
}

func TestPodService_Get(t *testing.T) {
	podNs := "admin"
	podName := "test1"

	portsForDb := []*daos.PortForDB{
		{
			NetworkName:  "net_api",
			NetworkPlane: "net_api",
			PortName:     "eth0",
			VnicType:     "normal",
			Accelerate:   "false",
			PodName:      podName,
			PodNs:        podNs,
			FixIP:        "192.168.1.0",
			IPGroupName:  "ig0",
			Metadata:     "",
			Combinable:   "false",
			Roles:        []string{"control"},

			ID:         "net_api",
			LazyName:   "eth0",
			TenantID:   podNs,
			MacAddress: "",
		},
	}

	podForDb := &daos.PodForDB{
		TenantId: podNs,
		PodNs:    podNs,
		PodName:  podName,
		Ports:    portsForDb,
	}

	portEagerAttrApi := PortEagerAttr{
		NetworkName:  "net_api",
		NetworkPlane: "net_api",
		PortName:     "eth0",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      podName,
		PodNs:        podNs,
		FixIP:        "192.168.1.0",
		IPGroupName:  "ig0",
		Metadata:     "",
		Combinable:   "false",
		Roles:        []string{"control"},
	}

	portLazyAttrApi := PortLazyAttr{
		ID:         "net_api",
		Name:       "eth0",
		TenantID:   podNs,
		MacAddress: "",
	}

	portsForService := []*Port{
		&Port{
			EagerAttr: portEagerAttrApi,
			LazyAttr:  portLazyAttrApi,
		},
	}

	podExpect := &Pod{
		TenantId: podNs,
		PodName:  podName,
		PodNs:    podNs,
		Ports:    portsForService,
	}

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Get(podNs, podName).Return(podForDb, nil)

	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodService_GetSucc", t, func() {
		pod, err := GetPodService().Get(podNs, podName)
		convey.So(pod, convey.ShouldResemble, podExpect)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPodService_GetFail(t *testing.T) {
	podNs := "admin"
	podName := "test1"

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Get(podNs, podName).Return(nil, errors.New("get err"))

	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodService_GetFail", t, func() {
		pod, err := GetPodService().Get(podNs, podName)
		convey.So(pod, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "get err")
	})
}

func TestPodService_DeleteSucc(t *testing.T) {
	podNs := "admin"
	podName := "test1"
	podForDb := &daos.PodForDB{
		TenantId: podNs,
		PodNs:    podNs,
		PodName:  podName,
	}

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Move(podNs, podName).Return(podForDb, nil)
	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	var pool *PortRecyclePool
	monkey.PatchInstanceMethod(reflect.TypeOf(pool), "AddAndRecycle", func(_ *PortRecyclePool, recyclePod *RecyclePod) {

	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodService_GetSucc", t, func() {
		err := GetPodService().DeletePodAndPorts(podNs, podName)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPodService_DeleteGetFail(t *testing.T) {
	podNs := "admin"
	podName := "test1"

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Move(podNs, podName).Return(nil, errors.New("move err"))
	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	var pool *PortRecyclePool
	monkey.PatchInstanceMethod(reflect.TypeOf(pool), "AddAndRecycle", func(_ *PortRecyclePool, recyclePod *RecyclePod) {

	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodService_DeleteGetFail", t, func() {
		err := GetPodService().DeletePodAndPorts(podNs, podName)
		convey.So(err.Error(), convey.ShouldEqual, "move err")
	})
}

func TestTransferToPodForDaoSucc(t *testing.T) {

	podNs := "admin"
	podName := "test1"
	portsForDb := []*daos.PortForDB{
		{
			NetworkName:  "net_api",
			NetworkPlane: "net_api",
			PortName:     "eth0",
			VnicType:     "normal",
			Accelerate:   "false",
			PodName:      podName,
			PodNs:        podNs,
			FixIP:        "192.168.1.0",
			IPGroupName:  "ig0",
			Metadata:     "",
			Combinable:   "false",
			Roles:        []string{"control"},

			ID:         "net_api",
			LazyName:   "eth0",
			TenantID:   podNs,
			MacAddress: "",
		},
	}
	podForDb := &daos.PodForDB{
		TenantId:     podNs,
		PodID:        "testpod-1111",
		PodNs:        podNs,
		PodName:      podName,
		PodType:      "",
		ErrorMsg:     "",
		IsSuccessful: true,
		Ports:        portsForDb,
	}

	portEagerAttrApi := PortEagerAttr{
		NetworkName:  "net_api",
		NetworkPlane: "net_api",
		PortName:     "eth0",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      podName,
		PodNs:        podNs,
		FixIP:        "192.168.1.0",
		IPGroupName:  "ig0",
		Metadata:     "",
		Combinable:   "false",
		Roles:        []string{"control"},
	}

	portLazyAttrApi := PortLazyAttr{
		ID:         "net_api",
		Name:       "eth0",
		TenantID:   podNs,
		MacAddress: "",
	}

	portsForService := []*Port{
		&Port{
			EagerAttr: portEagerAttrApi,
			LazyAttr:  portLazyAttrApi,
		},
	}
	pod := &Pod{
		TenantId:     podNs,
		PodID:        "testpod-1111",
		PodName:      podName,
		PodNs:        podNs,
		PodType:      "",
		IsSuccessful: true,
		ErrorMsg:     "",
		Ports:        portsForService,
	}
	convey.Convey("TeestTransferToPodForDaoSucc", t, func() {
		podForDbExpect := pod.transferToPodForDao()
		convey.So(podForDb, convey.ShouldResemble, podForDbExpect)
	})
}

func TestGetPodNetworkMessageDefault(t *testing.T) {
	bpnmExpect := createDefaultNetworkMessage("net_api")
	var mc *clients.ManagerClient
	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "GetDefaultNetWork",
		func(_ *clients.ManagerClient, _ string) (string, error) {
			return "net_api", nil
		})
	convey.Convey("TestGetPodNetworkMessageDefault", t, func() {
		annotations := make(map[string]string)
		bpnm, err := getPodNetworkMessage(annotations)
		convey.So(bpnm, convey.ShouldResemble, bpnmExpect)
		convey.So(err, convey.ShouldBeNil)
	})

}

func TestGetPodNetworkMessageNotDefault(t *testing.T) {
	bpnmExpect := &BluePrintNetworkMessage{
		[]BluePrintPort{
			BluePrintPort{AttachToNetwork: "net_api", Attributes: BluePrintAttributes{Accelerate: "false", Function: "std", NicName: "eth0", NicType: "normal"}},
			BluePrintPort{AttachToNetwork: "control", Attributes: BluePrintAttributes{Accelerate: "false", Function: "control", NicName: "nwrfcontrol", NicType: "normal"}},
			BluePrintPort{AttachToNetwork: "media", Attributes: BluePrintAttributes{Accelerate: "false", Function: "media", NicName: "nwrfmedia", NicType: "normal"}},
		},
	}
	convey.Convey("TestGetPodNetworkMessageDefault", t, func() {
		networks := "{\"ports\": " +
			"[{\"attach_to_network\": \"net_api\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"std\", \"nic_name\": \"eth0\", \"nic_type\": \"normal\"}}," +
			" {\"attach_to_network\": \"control\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"control\", \"nic_name\": \"nwrfcontrol\", \"nic_type\": \"normal\"}}, " +
			"{\"attach_to_network\": \"media\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"media\", \"nic_name\": \"nwrfmedia\", \"nic_type\": \"normal\"}}]}"
		annotations := map[string]string{
			"networks": networks,
		}
		bpnm, err := getPodNetworkMessage(annotations)
		convey.So(bpnm, convey.ShouldResemble, bpnmExpect)
		convey.So(err, convey.ShouldBeNil)
	})

}

func TestCreatePortByK8sAnnotations(t *testing.T) {
	monkey.Patch(getPodNetworkMessage, func(annotations map[string]string) (*BluePrintNetworkMessage, error) {
		return nil, nil
	})
	defer monkey.UnpatchAll()
	var service *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(service), "CreatePortsWithBluePrint", func(_ *portService, pod *PodForCreatPort, bpnm *BluePrintNetworkMessage) ([]*Port, error) {
		return nil, nil
	})
	annotations := make(map[string]string)
	podForCreatPort := &PodForCreatPort{}
	convey.Convey("TestCreatePortByK8sAnnotations", t, func() {
		_, err := createPortByK8sAnnotations(annotations, podForCreatPort)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestCreatePortByK8sAnnotationsGetFail(t *testing.T) {
	monkey.Patch(getPodNetworkMessage, func(annotations map[string]string) (*BluePrintNetworkMessage, error) {
		return nil, errors.New("get err")
	})
	defer monkey.UnpatchAll()

	annotations := make(map[string]string)
	podForCreatPort := &PodForCreatPort{}
	convey.Convey("TestCreatePortByK8sAnnotations", t, func() {
		_, err := createPortByK8sAnnotations(annotations, podForCreatPort)
		convey.So(err.Error(), convey.ShouldEqual, "get err")
	})
}

func TestCreatePortByK8sAnnotationsCreateFail(t *testing.T) {
	monkey.Patch(getPodNetworkMessage, func(annotations map[string]string) (*BluePrintNetworkMessage, error) {
		return nil, nil
	})
	defer monkey.UnpatchAll()
	var service *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(service), "CreatePortsWithBluePrint", func(_ *portService, pod *PodForCreatPort, bpnm *BluePrintNetworkMessage) ([]*Port, error) {
		return nil, errors.New("create err")
	})
	annotations := make(map[string]string)
	podForCreatPort := &PodForCreatPort{}
	convey.Convey("TestCreatePortByK8sAnnotationsCreateFail", t, func() {
		_, err := createPortByK8sAnnotations(annotations, podForCreatPort)
		convey.So(err.Error(), convey.ShouldEqual, "create err")
	})
}

func TestGetReportResult(t *testing.T) {
	convey.Convey("TestGetReportResult succ1\n", t, func() {
		reportPod := &clients.ReportPod{
			PodNs:   "pod_ns0",
			PodName: "pod_name0",
		}
		Pod2SyncStatusRepoCtx.Pod2SyncStatus = map[string]*exceptPodSyncStatus{
			"pod_ns0pod_name0": &exceptPodSyncStatus{
				PodNs:            "pod_ns0",
				PodName:          "pod_name0",
				IsSync2ManagerOk: false,
				IsSync2K8sOk:     true,
			},
			"pod1pod_name1": &exceptPodSyncStatus{
				PodNs:            "pod_ns1",
				PodName:          "pod_name1",
				IsSync2ManagerOk: false,
				IsSync2K8sOk:     true,
			},
		}
		monkey.Patch(reportPodToManager, func(reportPod *clients.ReportPod) error {
			return nil
		})
		defer monkey.UnpatchAll()
		isReportToManagerSucc, isReportToK8sSucc := getReportResult(reportPod)
		convey.So(isReportToManagerSucc, convey.ShouldBeTrue)
		convey.So(isReportToK8sSucc, convey.ShouldBeTrue)
	})
	convey.Convey("TestGetReportResult succ2\n", t, func() {
		reportPod := &clients.ReportPod{
			PodNs:   "pod_ns0",
			PodName: "pod_name0",
		}
		Pod2SyncStatusRepoCtx.Pod2SyncStatus = map[string]*exceptPodSyncStatus{
			"pod_ns0pod_name0": &exceptPodSyncStatus{
				PodNs:            "pod_ns0",
				PodName:          "pod_name0",
				IsSync2ManagerOk: true,
				IsSync2K8sOk:     false,
			},
			"pod1pod_name1": &exceptPodSyncStatus{
				PodNs:            "pod_ns1",
				PodName:          "pod_name1",
				IsSync2ManagerOk: false,
				IsSync2K8sOk:     false,
			},
		}
		monkey.Patch(reportPodToK8s, func(reportPod *clients.ReportPod) error {
			return nil
		})
		defer monkey.UnpatchAll()
		isReportToManagerSucc, isReportToK8sSucc := getReportResult(reportPod)
		convey.So(isReportToManagerSucc, convey.ShouldBeTrue)
		convey.So(isReportToK8sSucc, convey.ShouldBeTrue)
	})
	convey.Convey("TestGetReportResult succ3\n", t, func() {
		reportPod := &clients.ReportPod{
			PodNs:   "pod_ns0",
			PodName: "pod_name0",
		}
		Pod2SyncStatusRepoCtx.Pod2SyncStatus = map[string]*exceptPodSyncStatus{
			"pod_ns0pod_name0": &exceptPodSyncStatus{
				PodNs:            "pod_ns0",
				PodName:          "pod_name0",
				IsSync2ManagerOk: false,
				IsSync2K8sOk:     false,
			},
			"pod1pod_name1": &exceptPodSyncStatus{
				PodNs:            "pod_ns1",
				PodName:          "pod_name1",
				IsSync2ManagerOk: false,
				IsSync2K8sOk:     false,
			},
		}
		monkey.Patch(reportPodToManager, func(reportPod *clients.ReportPod) error {
			return nil
		})
		defer monkey.UnpatchAll()
		monkey.Patch(reportPodToK8s, func(reportPod *clients.ReportPod) error {
			return nil
		})
		isReportToManagerSucc, isReportToK8sSucc := getReportResult(reportPod)
		convey.So(isReportToManagerSucc, convey.ShouldBeTrue)
		convey.So(isReportToK8sSucc, convey.ShouldBeTrue)
	})
	convey.Convey("TestGetReportResult succ4\n", t, func() {
		reportPod := &clients.ReportPod{
			PodNs:   "pod_ns2",
			PodName: "pod_name2",
		}
		Pod2SyncStatusRepoCtx.Pod2SyncStatus = map[string]*exceptPodSyncStatus{
			"pod_ns0pod_name0": &exceptPodSyncStatus{
				PodNs:            "pod_ns0",
				PodName:          "pod_name0",
				IsSync2ManagerOk: false,
				IsSync2K8sOk:     false,
			},
			"pod1pod_name1": &exceptPodSyncStatus{
				PodNs:            "pod_ns1",
				PodName:          "pod_name1",
				IsSync2ManagerOk: false,
				IsSync2K8sOk:     false,
			},
		}
		monkey.Patch(reportPodToManager, func(reportPod *clients.ReportPod) error {
			return nil
		})
		defer monkey.UnpatchAll()
		monkey.Patch(reportPodToK8s, func(reportPod *clients.ReportPod) error {
			return nil
		})
		isReportToManagerSucc, isReportToK8sSucc := getReportResult(reportPod)
		convey.So(isReportToManagerSucc, convey.ShouldBeTrue)
		convey.So(isReportToK8sSucc, convey.ShouldBeTrue)
	})
	convey.Convey("TestGetReportResult fail1\n", t, func() {
		reportPod := &clients.ReportPod{
			PodNs:   "pod_ns0",
			PodName: "pod_name0",
		}
		Pod2SyncStatusRepoCtx.Pod2SyncStatus = map[string]*exceptPodSyncStatus{
			"pod_ns0pod_name0": &exceptPodSyncStatus{
				PodNs:            "pod_ns0",
				PodName:          "pod_name0",
				IsSync2ManagerOk: false,
				IsSync2K8sOk:     true,
			},
			"pod1pod_name1": &exceptPodSyncStatus{
				PodNs:            "pod_ns1",
				PodName:          "pod_name1",
				IsSync2ManagerOk: false,
				IsSync2K8sOk:     false,
			},
		}
		monkey.Patch(reportPodToManager, func(reportPod *clients.ReportPod) error {
			return errors.New("reportPodToManager fail")
		})
		defer monkey.UnpatchAll()
		isReportToManagerSucc, isReportToK8sSucc := getReportResult(reportPod)
		convey.So(isReportToManagerSucc, convey.ShouldBeFalse)
		convey.So(isReportToK8sSucc, convey.ShouldBeTrue)
	})
	convey.Convey("TestGetReportResult fail2\n", t, func() {
		reportPod := &clients.ReportPod{
			PodNs:   "pod_ns0",
			PodName: "pod_name0",
		}
		Pod2SyncStatusRepoCtx.Pod2SyncStatus = map[string]*exceptPodSyncStatus{
			"pod_ns0pod_name0": &exceptPodSyncStatus{
				PodNs:            "pod_ns0",
				PodName:          "pod_name0",
				IsSync2ManagerOk: true,
				IsSync2K8sOk:     false,
			},
			"pod1pod_name1": &exceptPodSyncStatus{
				PodNs:            "pod_ns1",
				PodName:          "pod_name1",
				IsSync2ManagerOk: false,
				IsSync2K8sOk:     false,
			},
		}
		monkey.Patch(reportPodToK8s, func(reportPod *clients.ReportPod) error {
			return errors.New("reportPodToK8s fail")
		})
		defer monkey.UnpatchAll()

		isReportToManagerSucc, isReportToK8sSucc := getReportResult(reportPod)
		convey.So(isReportToManagerSucc, convey.ShouldBeTrue)
		convey.So(isReportToK8sSucc, convey.ShouldBeFalse)
	})
	convey.Convey("TestGetReportResult fail3\n", t, func() {
		reportPod := &clients.ReportPod{
			PodNs:   "pod_ns0",
			PodName: "pod_name0",
		}
		Pod2SyncStatusRepoCtx.Pod2SyncStatus = map[string]*exceptPodSyncStatus{
			"pod_ns0pod_name0": &exceptPodSyncStatus{
				PodNs:            "pod_ns0",
				PodName:          "pod_name0",
				IsSync2ManagerOk: false,
				IsSync2K8sOk:     false,
			},
			"pod1pod_name1": &exceptPodSyncStatus{
				PodNs:            "pod_ns1",
				PodName:          "pod_name1",
				IsSync2ManagerOk: false,
				IsSync2K8sOk:     false,
			},
		}
		monkey.Patch(reportPodToManager, func(reportPod *clients.ReportPod) error {
			return errors.New("reportPodToManager fail")
		})
		defer monkey.UnpatchAll()
		monkey.Patch(reportPodToK8s, func(reportPod *clients.ReportPod) error {
			return errors.New("reportPodToK8s fail")
		})
		isReportToManagerSucc, isReportToK8sSucc := getReportResult(reportPod)
		convey.So(isReportToManagerSucc, convey.ShouldBeFalse)
		convey.So(isReportToK8sSucc, convey.ShouldBeFalse)
	})
	convey.Convey("TestGetReportResult fail4\n", t, func() {
		reportPod := &clients.ReportPod{
			PodNs:   "pod_ns2",
			PodName: "pod_name2",
		}
		Pod2SyncStatusRepoCtx.Pod2SyncStatus = map[string]*exceptPodSyncStatus{
			"pod_ns0pod_name0": &exceptPodSyncStatus{
				PodNs:            "pod_ns0",
				PodName:          "pod_name0",
				IsSync2ManagerOk: false,
				IsSync2K8sOk:     false,
			},
			"pod1pod_name1": &exceptPodSyncStatus{
				PodNs:            "pod_ns1",
				PodName:          "pod_name1",
				IsSync2ManagerOk: false,
				IsSync2K8sOk:     false,
			},
		}
		monkey.Patch(reportPodToManager, func(reportPod *clients.ReportPod) error {
			return errors.New("reportPodToManager fail")
		})
		defer monkey.UnpatchAll()
		monkey.Patch(reportPodToK8s, func(reportPod *clients.ReportPod) error {
			return nil
		})
		isReportToManagerSucc, isReportToK8sSucc := getReportResult(reportPod)
		convey.So(isReportToManagerSucc, convey.ShouldBeFalse)
		convey.So(isReportToK8sSucc, convey.ShouldBeTrue)
	})
}
