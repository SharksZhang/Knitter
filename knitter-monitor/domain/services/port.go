package services

import (
	"errors"
	"strings"

	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"

	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/daos"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/knitter-monitor/infra/clients"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/knitter-monitor/infra/db-config"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
)

func GetPortService() PortServiceInterface {
	return &portService{}
}

type PortServiceInterface interface {
	CreatePortsWithBluePrint(pod *PodForCreatPort, bpnm *BluePrintNetworkMessage) ([]*Port, error)
	DeleteBulkPorts(tenantID string, portIDs []string) error
}

type portService struct {
}

func newPortFromPortForDB(db *daos.PortForDB) *Port {
	portEagerAttr := PortEagerAttr{
		NetworkName:  db.NetworkName,
		NetworkPlane: db.NetworkPlane,
		PortName:     db.PortName,
		VnicType:     db.VnicType,
		Accelerate:   db.Accelerate,
		PodName:      db.PodName,
		PodNs:        db.PodNs,
		FixIP:        db.FixIP,
		IPGroupName:  db.IPGroupName,
		Metadata:     db.Metadata,
		Combinable:   db.Combinable,
		Roles:        db.Roles,
	}
	portLazyAttr := PortLazyAttr{
		ID:           db.ID,
		Name:         db.LazyName,
		TenantID:     db.TenantID,
		MacAddress:   db.MacAddress,
		FixedIps:     db.FixedIps,
		FixedIPInfos: db.FixedIPInfos,
		Cidr:         db.Cidr,
	}
	return &Port{
		EagerAttr: portEagerAttr,
		LazyAttr:  portLazyAttr,
	}
}

//todo refactor
func (ps *portService) CreatePortsWithBluePrint(pod *PodForCreatPort, bpnm *BluePrintNetworkMessage) ([]*Port, error) {
	klog.Debugf("CreatePortsWithBluePrint start, podName is [%v]", pod.PodName)
	ports, err := ps.FillPortsEagerAttrWithBluePrint(pod, bpnm)
	if err != nil {
		klog.Errorf("ps.FillPortsEagerAttrWithBluePrint(pod :[%v] , nwJSON:[%v]) err, error is [%v]", pod, bpnm, err)
		return nil, err
	}

	reqs := ps.buildBulkPortsReq(pod, ports)
	if len(reqs.Ports) == 0 {
		klog.Warningf("no port should be created")
		return ports, err
	}

	createBulkPortsResp, err := ps.CreateBulkPorts(pod, reqs)
	if err != nil {
		klog.Errorf("ps.NewPortsWithEagerAttrFromK8s() err, error is [%v]", err)
		return nil, err
	}
	err = fillPortsLazyAttr(createBulkPortsResp, ports, pod.TenantID)
	if err != nil {
		portIDs := make([]string, 0)
		for _, portResp := range createBulkPortsResp.Ports {
			portIDs = append(portIDs, portResp.PortID)
		}
		err1 := GetPortService().DeleteBulkPorts(pod.PodNs, portIDs)
		if err1 != nil {
			klog.Warningf("GetPortService().DeleteBulkPorts(pod.PodNs, portIDs) err, error is [%v]", err1)
		}

		klog.Errorf("fillPortLazyAttr(createBulkPortsResp, mports, pod.TenantID) err, error is [%v]", err)
		return nil, err
	}
	klog.Infof("CreatePortsWithBluePrint successfully, podName is [%v]", pod.PodName)
	return ports, nil

}

func (ps *portService) FillPortsEagerAttrWithBluePrint(pod *PodForCreatPort, bpnm *BluePrintNetworkMessage) ([]*Port,
	error) {

	if len(bpnm.Ports) == 0 {
		klog.Errorf("Get port config error! %v", errobj.ErrGetPortConfigError)
		return nil, errobj.ErrGetPortConfigError
	}
	var mports []*Port
	for _, bpPort := range bpnm.Ports {
		port := &Port{}
		err := port.fillPortEagerAttr(pod.PodNs, pod.PodName, bpPort)
		if err != nil {
			return nil, err
		}
		mports = append(mports, port)
	}
	combinedPortObjs, err := combinePortObjs(mports)
	if err != nil {
		return nil, err
	}
	return combinedPortObjs, nil
}
func (ps *portService) DeleteBulkPorts(podNS string, portIDs []string) error {
	managerClient := clients.GetManagerClient()
	errs := make([]error, 0)
	for _, portID := range portIDs {
		klog.Infof("portService.DeleteBulkPorts: delete port, port id is [%v]", portID)
		err := managerClient.DeleteNeutronPort(podNS, portID)
		if err != nil && !errobj.IsNotFoundError(err) {
			klog.Errorf("DeleteBulkPorts:managerClient.DeleteNeutronPort(tenantID :[%v], portID:[%v]) error: [%v]", podNS, portID, err)
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		klog.Errorf("DeleteBulkPorts: Delete [%v] Ports error", len(errs))
		return errobj.ErrDeleteBulkPortsError
	}
	return nil
}

func (ps *portService) CreateBulkPorts(pod *PodForCreatPort, createBulkPortsReq *clients.ManagerCreateBulkPortsReq) (resp clients.CreatePortsResp, err error) {
	defer RollBackCreateBulkPorts(err, pod, resp)
	managerClient := clients.GetManagerClient()
	if len(createBulkPortsReq.Ports) == 0 {
		klog.Errorf("No need to creat ports!")
		return clients.CreatePortsResp{}, errors.New("CreateBulkPortReqs is nil")
	}
	resp, err = managerClient.CreateNeutronBulkPorts(pod.PodID, createBulkPortsReq, pod.TenantID)
	if err != nil {
		klog.Errorf("CreateNeutronBulkPorts: agtCtx.Mc.CreateNeutronBulkPorts failed, error! -%v", err)
		return resp, err
	}
	klog.Infof("CreateNeutronBulkPorts: create and Unmarshal result: [%v]", resp)
	return resp, nil

}

func RollBackCreateBulkPorts(err error, pod *PodForCreatPort, resp clients.CreatePortsResp) {
	if err != nil {
		for _, port := range resp.Ports {
			err := clients.GetManagerClient().DeleteNeutronPort(pod.TenantID, port.PortID)
			if err != nil {
				klog.Warningf("RollBackCreateBulkPorts:DeleteNeutronPort err, err is [%v]", err)
			}
		}
	}
}

func (ps *portService) buildBulkPortsReq(pod *PodForCreatPort, ports []*Port) *clients.ManagerCreateBulkPortsReq {
	reqs := clients.ManagerCreateBulkPortsReq{Ports: make([]clients.ManagerCreatePortReq, 0)}
	for _, port := range ports {
		klog.Debugf("portService.buildBulkPortsReq:  portt is [%v]", port)
		if isC0Port(port) {
			continue
		}
		managerPortReq := port.TransformToMangerCreatePortReq(pod)
		reqs.Ports = append(reqs.Ports, *managerPortReq)
	}
	klog.Debugf("portService.buildBulkPortsReq:  reqs is [%v]", reqs)

	return &reqs
}

func isC0Port(port *Port) bool {
	return port.EagerAttr.Accelerate == "true" && IsCTNetPlane(port.EagerAttr.NetworkPlane)
}

type Port struct {
	EagerAttr PortEagerAttr `json:"eager_attr"`
	LazyAttr  PortLazyAttr  `json:"lazy_attr"`
}

func (p *Port) transferToPortForDB() *daos.PortForDB {
	return &daos.PortForDB{
		NetworkName:  p.EagerAttr.NetworkName,
		NetworkPlane: p.EagerAttr.NetworkPlane,
		PortName:     p.EagerAttr.PortName,
		VnicType:     p.EagerAttr.VnicType,
		Accelerate:   p.EagerAttr.Accelerate,
		PodName:      p.EagerAttr.PodName,
		PodNs:        p.EagerAttr.PodNs,
		FixIP:        p.EagerAttr.FixIP,
		IPGroupName:  p.EagerAttr.IPGroupName,
		Metadata:     p.EagerAttr.Metadata,
		Combinable:   p.EagerAttr.Combinable,
		Roles:        p.EagerAttr.Roles,

		ID:           p.LazyAttr.ID,
		LazyName:     p.LazyAttr.Name,
		TenantID:     p.LazyAttr.TenantID,
		MacAddress:   p.LazyAttr.MacAddress,
		FixedIps:     p.LazyAttr.FixedIps,
		FixedIPInfos: p.LazyAttr.FixedIPInfos,
		Cidr:         p.LazyAttr.Cidr,
	}
}

func fillPortsLazyAttr(resp clients.CreatePortsResp, ports []*Port, tenantID string) error {
	for i, port := range ports {
		var createPortInfo clients.CreatePortInfo
		var flag = false
		for _, createPortInfo = range resp.Ports {
			if strings.HasSuffix(createPortInfo.Name, port.EagerAttr.PortName) {
				flag = true
				break
			}
		}
		if flag {
			ports[i].LazyAttr.ID = createPortInfo.PortID
			ports[i].LazyAttr.Name = createPortInfo.Name
			ports[i].LazyAttr.MacAddress = createPortInfo.MacAddress
			ports[i].LazyAttr.FixedIps = createPortInfo.FixedIps
			ports[i].LazyAttr.FixedIPInfos = createPortInfo.FixedIPInfos
			ports[i].LazyAttr.TenantID = tenantID
			ports[i].LazyAttr.Cidr = createPortInfo.Cidr
		} else if port.EagerAttr.Accelerate == "false" || port.EagerAttr.NetworkPlane == "eio" {
			klog.Errorf("createPortInfo:%v is not in bulkports, we must destory bulkports", port.EagerAttr.PortName)
			return errobj.ErrPortNotFound
		}

	}
	return nil
}

func (p *Port) fillPortEagerAttr(podNs, podName string, bpPort BluePrintPort) error {
	eagerPort, err := NewEagerPort(bpPort)
	if err != nil {
		return err
	}
	p.EagerAttr.NetworkName = eagerPort.NetworkName
	p.EagerAttr.NetworkPlane = eagerPort.NetworkPlane
	p.EagerAttr.PortName = eagerPort.PortName
	p.EagerAttr.VnicType = eagerPort.VnicType
	p.EagerAttr.Accelerate = eagerPort.Accelerate
	p.EagerAttr.FixIP = eagerPort.FixIP
	p.EagerAttr.IPGroupName = eagerPort.IPGroupName
	p.EagerAttr.PodNs = podNs
	p.EagerAttr.PodName = podName
	p.EagerAttr.Combinable = eagerPort.Combinable
	p.EagerAttr.Metadata = eagerPort.MetaDate
	return nil
}

func (p *Port) TransformToMangerCreatePortReq(pod *PodForCreatPort) *clients.ManagerCreatePortReq {
	return &clients.ManagerCreatePortReq{
		TenantID:    pod.TenantID,
		NetworkName: p.EagerAttr.NetworkName,
		PortName:    p.EagerAttr.PortName,
		//logic p has no mode direct and physical
		VnicType: constvalue.LogicalPortDefaultVnicType,
		//todo  how to get nodeID
		NodeID: dbconfig.GetClusterID(),
		PodNs:  p.EagerAttr.PodNs,
		//todo delete pod Name
		PodName:     "FakePodName",
		FixIP:       p.EagerAttr.FixIP,
		IPGroupName: p.EagerAttr.IPGroupName,
	}
}

type PortEagerAttr struct {
	NetworkName  string      `json:"network_name"`
	NetworkPlane string      `json:"network_plane"`
	PortName     string      `json:"port_name"`
	VnicType     string      `json:"vnic_type"`
	Accelerate   string      `json:"accelerate"`
	PodName      string      `json:"pod_name"`
	PodNs        string      `json:"pod_ns"`
	FixIP        string      `json:"fix_ip"`
	IPGroupName  string      `json:"ip_group_name"`
	Combinable   string      `json:"combinable"`
	Metadata     interface{} `json:"metadata"`
	Roles        []string    `json:"roles"`
}

type PortLazyAttr struct {
	//NetworkID      string
	ID           string                     `json:"id"`
	Name         string                     `json:"name"`
	TenantID     string                     `json:"tenant_id"`
	MacAddress   string                     `json:"mac_address"`
	FixedIps     []ports.IP                 `json:"fixed_ips"`
	FixedIPInfos []agtmgr.FixedIPItem `json:"fixed_ip_infos"`
	Cidr         string                     `json:"cidr"`
}

type EagerPort struct {
	NetworkName  string
	NetworkPlane string
	PortName     string
	VnicType     string
	Accelerate   string
	FixIP        string
	IPGroupName  string
	PortFunc     string
	Combinable   string
	MetaDate     interface{}
}

func NewEagerPort(bpPort BluePrintPort) (*EagerPort, error) {
	eagerPort := &EagerPort{}
	if bpPort.AttachToNetwork == "" {
		err := errors.New("port attach network is blank")
		klog.Error("port attach network is blank")
		return nil, err
	}
	portName := bpPort.Attributes.NicName
	if portName == "" {
		portName = "eth_" + bpPort.AttachToNetwork
		if len(portName) > 12 {
			rs := []rune(portName)
			portName = string(rs[0:12])
		}
	}
	if len(portName) > 12 {
		klog.Errorf("Lenth of port Name is greater than 12")
		return nil, errors.New("lenth of port Name is illegal")
	}
	portFunc := bpPort.Attributes.Function
	if portFunc == "" {
		portFunc = "std"
	}

	portType := bpPort.Attributes.NicType
	if portType == "" {
		portType = "normal"
	}

	isUseDpdk := bpPort.Attributes.Accelerate
	if isUseDpdk != "true" {
		isUseDpdk = "false"
	}

	ipGroupName := bpPort.Attributes.IpGroupName

	combinable := bpPort.Attributes.Combinable
	if combinable != "false" {
		combinable = "true"
	}

	metadataObj := bpPort.Attributes.MetaData

	eagerPort.NetworkName = bpPort.AttachToNetwork
	eagerPort.NetworkPlane = portFunc
	eagerPort.PortName = bpPort.Attributes.NicName
	eagerPort.VnicType = portType
	eagerPort.Accelerate = isUseDpdk
	eagerPort.IPGroupName = ipGroupName
	eagerPort.Combinable = combinable
	eagerPort.MetaDate = metadataObj
	return eagerPort, nil
}
