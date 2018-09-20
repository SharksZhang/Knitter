package services

import (
	"encoding/json"
	"errors"
	"github.com/bouk/monkey"
	"github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
	"github.com/ZTE/Knitter/knitter-monitor/infra/clients"
)

func TestPortService_CreatePortsWithBluePrint(t *testing.T) {

	networks := "{\"ports\": " +
		"[{\"attach_to_network\": \"net_api\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"std\", \"nic_name\": \"eth0\", \"nic_type\": \"normal\"}}," +
		" {\"attach_to_network\": \"control\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"control\", \"nic_name\": \"nwrfcontrol\", \"nic_type\": \"normal\"}}, " +
		"{\"attach_to_network\": \"media\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"media\", \"nic_name\": \"nwrfmedia\", \"nic_type\": \"normal\"}}]}"
	bpnm := &BluePrintNetworkMessage{}
	json.Unmarshal([]byte(networks), bpnm)
	mportsExpect := []*Port{&Port{}}
	var ps *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "FillPortsEagerAttrWithBluePrint",
		func(_ *portService, pod *PodForCreatPort, bpnm *BluePrintNetworkMessage) ([]*Port,
			error) {
			return mportsExpect, nil
		})
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "CreateBulkPorts", func(_ *portService, pod *PodForCreatPort, ports *clients.ManagerCreateBulkPortsReq) (resp clients.CreatePortsResp, err error) {
		return clients.CreatePortsResp{}, nil
	})
	monkey.Patch(fillPortsLazyAttr, func(resp clients.CreatePortsResp, ports []*Port, tenantID string) error {
		return nil
	})
	pod := &PodForCreatPort{}
	convey.Convey("TestPortService_CreatePortsWithBluePrint", t, func() {
		mports, err := GetPortService().CreatePortsWithBluePrint(pod, bpnm)
		convey.So(mports, convey.ShouldResemble, mportsExpect)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPortService_CreatePortsWithBluePrintFail(t *testing.T) {

	networks := "{\"ports\": " +
		"[{\"attach_to_network\": \"net_api\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"std\", \"nic_name\": \"eth0\", \"nic_type\": \"normal\"}}," +
		" {\"attach_to_network\": \"control\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"control\", \"nic_name\": \"nwrfcontrol\", \"nic_type\": \"normal\"}}, " +
		"{\"attach_to_network\": \"media\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"media\", \"nic_name\": \"nwrfmedia\", \"nic_type\": \"normal\"}}]}"
	bpnm := &BluePrintNetworkMessage{}
	json.Unmarshal([]byte(networks), bpnm)
	var ps *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "FillPortsEagerAttrWithBluePrint",
		func(_ *portService, pod *PodForCreatPort, bpnm *BluePrintNetworkMessage) ([]*Port,
			error) {
			return nil, errors.New("new eager attr err")
		})
	defer monkey.UnpatchAll()

	pod := &PodForCreatPort{}
	convey.Convey("TestPortService_CreatePortsWithBluePrint", t, func() {
		mports, err := GetPortService().CreatePortsWithBluePrint(pod, bpnm)
		convey.So(mports, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "new eager attr err")
	})
}

func TestPortService_CreatePortsWithBluePrintFail2(t *testing.T) {

	networks := "{\"ports\": " +
		"[{\"attach_to_network\": \"net_api\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"std\", \"nic_name\": \"eth0\", \"nic_type\": \"normal\"}}," +
		" {\"attach_to_network\": \"control\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"control\", \"nic_name\": \"nwrfcontrol\", \"nic_type\": \"normal\"}}, " +
		"{\"attach_to_network\": \"media\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"media\", \"nic_name\": \"nwrfmedia\", \"nic_type\": \"normal\"}}]}"
	bpnm := &BluePrintNetworkMessage{}
	json.Unmarshal([]byte(networks), bpnm)
	mportsExpect := []*Port{&Port{}}
	var ps *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "FillPortsEagerAttrWithBluePrint",
		func(_ *portService, pod *PodForCreatPort, bpnm *BluePrintNetworkMessage) ([]*Port,
			error) {
			return mportsExpect, nil
		})
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "CreateBulkPorts", func(_ *portService, pod *PodForCreatPort, ports *clients.ManagerCreateBulkPortsReq) (resp clients.CreatePortsResp, err error) {
		return clients.CreatePortsResp{}, errors.New("CreateBulkPorts err")
	})

	pod := &PodForCreatPort{}
	convey.Convey("TestPortService_CreatePortsWithBluePrint", t, func() {
		mports, err := GetPortService().CreatePortsWithBluePrint(pod, bpnm)
		convey.So(err.Error(), convey.ShouldEqual, "CreateBulkPorts err")
		convey.So(mports, convey.ShouldBeNil)
	})
}

func TestPortService_CreatePortsWithBluePrintFail3(t *testing.T) {

	networks := "{\"ports\": " +
		"[{\"attach_to_network\": \"net_api\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"std\", \"nic_name\": \"eth0\", \"nic_type\": \"normal\"}}," +
		" {\"attach_to_network\": \"control\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"control\", \"nic_name\": \"nwrfcontrol\", \"nic_type\": \"normal\"}}, " +
		"{\"attach_to_network\": \"media\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"media\", \"nic_name\": \"nwrfmedia\", \"nic_type\": \"normal\"}}]}"
	bpnm := &BluePrintNetworkMessage{}
	json.Unmarshal([]byte(networks), bpnm)
	mportsExpect := []*Port{&Port{}}
	var ps *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "FillPortsEagerAttrWithBluePrint",
		func(_ *portService, pod *PodForCreatPort, bpnm *BluePrintNetworkMessage) ([]*Port,
			error) {
			return mportsExpect, nil
		})
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "CreateBulkPorts", func(_ *portService, pod *PodForCreatPort, ports *clients.ManagerCreateBulkPortsReq) (resp clients.CreatePortsResp, err error) {
		return clients.CreatePortsResp{}, nil
	})
	monkey.Patch(fillPortsLazyAttr, func(resp clients.CreatePortsResp, ports []*Port, tenantID string) error {
		return errors.New("fill err")
	})
	pod := &PodForCreatPort{}
	convey.Convey("TestPortService_CreatePortsWithBluePrint", t, func() {
		mports, err := GetPortService().CreatePortsWithBluePrint(pod, bpnm)
		convey.So(err.Error(), convey.ShouldEqual, "fill err")
		convey.So(mports, convey.ShouldBeNil)
	})
}

func TestPortService_FillPortsEagerAttrWithBluePrint(t *testing.T) {
	networks := "{\"ports\": " +
		"[{\"attach_to_network\": \"net_api\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"std\", \"nic_name\": \"eth0\", \"nic_type\": \"normal\"}}," +
		" {\"attach_to_network\": \"control\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"control\", \"nic_name\": \"nwrfcontrol\", \"nic_type\": \"normal\"}}, " +
		"{\"attach_to_network\": \"media\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"media\", \"nic_name\": \"nwrfmedia\", \"nic_type\": \"normal\"}}]}"
	bpnm := &BluePrintNetworkMessage{}
	json.Unmarshal([]byte(networks), bpnm)
	PodName := "test1"
	podNS := "admin"

	podForCreatePort := &PodForCreatPort{PodNs: podNS, PodName: PodName}

	portEagerAttrApi := PortEagerAttr{
		NetworkName:  "net_api",
		NetworkPlane: "std",
		PortName:     "eth0",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Combinable:   "true",
		Roles:        []string{"std"},
	}

	portEagerAttrControl := PortEagerAttr{
		NetworkName:  "control",
		NetworkPlane: "control",
		PortName:     "nwrfcontrol",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Combinable:   "true",
		Roles:        []string{"control"},
	}

	portEagerAttrMedia := PortEagerAttr{
		NetworkName:  "media",
		NetworkPlane: "media",
		PortName:     "nwrfmedia",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Combinable:   "true",
		Roles:        []string{"media"},
	}
	portsForService := []*Port{
		&Port{
			EagerAttr: portEagerAttrApi,
		},
		&Port{
			EagerAttr: portEagerAttrControl,
		},
		&Port{
			EagerAttr: portEagerAttrMedia,
		},
	}
	convey.Convey("TestPortService_FillPortsEagerAttrWithBluePrint", t, func() {
		ports, err := (&portService{}).FillPortsEagerAttrWithBluePrint(podForCreatePort, bpnm)
		convey.So(ports, convey.ShouldResemble, portsForService)
		convey.So(err, convey.ShouldBeNil)
	})

}

func TestPortService_DeleteBulkPorts(t *testing.T) {
	var mc *clients.ManagerClient
	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "DeleteNeutronPort",
		func(mc *clients.ManagerClient, tenantID string, portID string) (e error) {
			return nil
		})
	convey.Convey("TestPortService_DeleteBulkPorts", t, func() {
		portIDs := []string{"1", "2"}
		err := GetPortService().DeleteBulkPorts("admin", portIDs)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPortService_DeleteBulkPortsFail(t *testing.T) {
	var mc *clients.ManagerClient
	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "DeleteNeutronPort",
		func(mc *clients.ManagerClient, tenantID string, portID string) (e error) {
			return errors.New("delete err")
		})
	convey.Convey("TestPortService_DeleteBulkPortsFail", t, func() {
		portIDs := []string{"1", "2"}
		err := GetPortService().DeleteBulkPorts("admin", portIDs)
		convey.So(err.Error(), convey.ShouldEqual, "delete ports error")
	})
}

func TestPortService_CreateBulkPorts(t *testing.T) {
	PodName := "test1"
	podNS := "admin"
	podForCreatePort := &PodForCreatPort{PodNs: podNS, PodName: PodName}

	portEagerAttrApi := PortEagerAttr{
		NetworkName:  "net_api",
		NetworkPlane: "std",
		PortName:     "eth0",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"std"},
	}

	portEagerAttrControl := PortEagerAttr{
		NetworkName:  "control",
		NetworkPlane: "control",
		PortName:     "nwrfcontrol",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"control"},
	}

	portEagerAttrMedia := PortEagerAttr{
		NetworkName:  "media",
		NetworkPlane: "media",
		PortName:     "nwrfmedia",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"media"},
	}
	portsForService := []*Port{
		&Port{
			EagerAttr: portEagerAttrApi,
		},
		&Port{
			EagerAttr: portEagerAttrControl,
		},
		&Port{
			EagerAttr: portEagerAttrMedia,
		},
	}

	resp := clients.CreatePortsResp{[]clients.CreatePortInfo{
		clients.CreatePortInfo{
			Name:      "eth0",
			NetworkID: "net_api",
		},
	}}

	var mc *clients.ManagerClient
	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "CreateNeutronBulkPorts",
		func(mc *clients.ManagerClient, reqID string, req *clients.ManagerCreateBulkPortsReq, tenantID string) (clients.CreatePortsResp, error) {
			return resp, nil
		})
	defer monkey.UnpatchAll()
	monkey.Patch(RollBackCreateBulkPorts, func(err error, pod *PodForCreatPort, resp clients.CreatePortsResp) {

	})
	convey.Convey("TestPortService_CreateBulkPorts", t, func() {
		ps := &portService{}
		reqs := ps.buildBulkPortsReq(podForCreatePort, portsForService)
		_, err := ps.CreateBulkPorts(podForCreatePort, reqs)
		convey.So(err, convey.ShouldBeNil)
	})

}

func TestPortService_CreateBulkPortsFail(t *testing.T) {
	PodName := "test1"
	podNS := "admin"
	podForCreatePort := &PodForCreatPort{PodNs: podNS, PodName: PodName}

	portEagerAttrApi := PortEagerAttr{
		NetworkName:  "net_api",
		NetworkPlane: "std",
		PortName:     "eth0",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"std"},
	}

	portEagerAttrControl := PortEagerAttr{
		NetworkName:  "control",
		NetworkPlane: "control",
		PortName:     "nwrfcontrol",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"control"},
	}

	portEagerAttrMedia := PortEagerAttr{
		NetworkName:  "media",
		NetworkPlane: "media",
		PortName:     "nwrfmedia",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"media"},
	}
	portsForService := []*Port{
		&Port{
			EagerAttr: portEagerAttrApi,
		},
		&Port{
			EagerAttr: portEagerAttrControl,
		},
		&Port{
			EagerAttr: portEagerAttrMedia,
		},
	}

	var mc *clients.ManagerClient
	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "CreateNeutronBulkPorts",
		func(_ *clients.ManagerClient, reqID string, req *clients.ManagerCreateBulkPortsReq, tenantID string) (clients.CreatePortsResp, error) {
			return clients.CreatePortsResp{}, errors.New("create err")
		})
	defer monkey.UnpatchAll()
	monkey.Patch(RollBackCreateBulkPorts, func(err error, pod *PodForCreatPort, resp clients.CreatePortsResp) {

	})
	convey.Convey("TestPortService_CreateBulkPorts", t, func() {
		ps := &portService{}
		reqs := ps.buildBulkPortsReq(podForCreatePort, portsForService)
		_, err := ps.CreateBulkPorts(podForCreatePort, reqs)
		convey.So(err.Error(), convey.ShouldEqual, "create err")
	})

}

func TestFillPortLazyAttr(t *testing.T) {
	PodName := "testpod"
	podNS := "admin"

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
		ID:         "1111",
		Name:       "eth0",
		TenantID:   podNS,
		MacAddress: "0000",
		Cidr:       "192.0.0.0/8",
	}

	portsForServiceExpect := []*Port{
		&Port{
			EagerAttr: portEagerAttrApi,
			LazyAttr:  portLazyAttrApi,
		},
	}
	portsForService := []*Port{
		&Port{
			EagerAttr: portEagerAttrApi,
		},
	}

	resp := clients.CreatePortsResp{[]clients.CreatePortInfo{
		clients.CreatePortInfo{
			Name:       "eth0",
			NetworkID:  "net_api",
			MacAddress: "0000",
			Cidr:       "192.0.0.0/8",
			PortID:     "1111",
		},
	}}
	convey.Convey("TestFillPortLazyAttr", t, func() {
		fillPortsLazyAttr(resp, portsForService, "admin")
		convey.So(portsForService, convey.ShouldResemble, portsForServiceExpect)
	})

}
