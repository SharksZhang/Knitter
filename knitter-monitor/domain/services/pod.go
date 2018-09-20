package services

import (
	"encoding/json"

	"k8s.io/api/core/v1"

	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"

	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/daos"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/knitter-monitor/infra/clients"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/klog"
)

func GetPodService() PodServiceInterface {
	klog.Debugf("GetPodService")
	return &podService{}
}

type PodServiceInterface interface {
	NewPodFromK8sPod(pod *v1.Pod) (*Pod, error)
	Save(pod *Pod) error
	Get(podNs, podName string) (*Pod, error)
	DeletePodAndPorts(podNs, podName string) error
	ReportPod(reportPod *clients.ReportPod) error
	PatchReportPod(podNS, podName string, req *agtmgr.AgentPodReq) error
	ReportDeletePod(podNs, podName string) error
	DeletePod(podNs, podName string) error
}

type podService struct {
}

func (ps *podService) NewPodFromK8sPod(k8sPod *v1.Pod) (*Pod, error) {
	klog.Infof("PodService.newPodFromK8sPod start podName is [%v]", k8sPod.Name)
	var resourceManagerName, resourceManagerType string
	if len(k8sPod.GetOwnerReferences()) > 0 {
		resourceManagerType = k8sPod.GetOwnerReferences()[0].Kind
		resourceManagerName = k8sPod.GetOwnerReferences()[0].Name
	}
	pod := &Pod{
		PodID:               string(k8sPod.GetObjectMeta().GetUID()),
		PodName:             k8sPod.GetObjectMeta().GetName(),
		PodNs:               k8sPod.GetObjectMeta().GetNamespace(),
		TenantId:            k8sPod.Namespace,
		IsSuccessful:        true,
		ResourceManagerName: resourceManagerName,
		ResourceManagerType: resourceManagerType,
	}
	var err error
	pod.Ports, err = createPortByK8sAnnotations(k8sPod.Annotations, pod.transferToPod4CreatePort())
	if err != nil {
		klog.Errorf("podService.NewPodFromK8sPod: createPortByK8sAnnotations err, err is [%v]", err)
		return pod, err
	}
	klog.Infof("NewPodFromK8sPod Successfully, pod is [%+v]", pod)
	return pod, nil
}

func getPodNetworkMessage(annotations map[string]string) (*BluePrintNetworkMessage, error) {
	networksStr, ok := annotations["networks"]
	klog.Debugf("getPodNetworkMessage:networksStr is [%v] ", networksStr)
	var bpnm *BluePrintNetworkMessage
	if isNotNetworkConfigExist(networksStr, ok) {
		klog.Infof("getPodNetworkMessage: no network message in blurPrint, use default")
		var err error
		bpnm, err = NewDefaultNetworkMessage()
		if err != nil {
			klog.Errorf("getPodNetworkMessage:GetDefaultNetworkConfig() err, error is [%v]", err)
			return nil, err
		}
		return bpnm, nil
	}
	bpnm = &BluePrintNetworkMessage{}
	err := json.Unmarshal([]byte(networksStr), bpnm)
	if err != nil {
		klog.Errorf("getPodNetworkMessage:Unmarshal(networksStr:[%v]) err, error is [%v]", networksStr, err)
		return nil, err
	}
	err = bpnm.checkPorts()
	if err != nil {
		klog.Errorf("getPodNetworkMessage: checkPorts error: %v", err)
		return nil, err
	}
	return bpnm, nil
}

func createPortByK8sAnnotations(annotations map[string]string, pod4CreatePort *PodForCreatPort) ([]*Port, error) {
	bpnm, err := getPodNetworkMessage(annotations)
	if err != nil {
		klog.Errorf("podService.NewPodFromK8sPod: getPodNetworkMessage() err , err is [%v]", err)
		return nil, err
	}
	ports, err := GetPortService().CreatePortsWithBluePrint(pod4CreatePort, bpnm)
	if err != nil {
		klog.Errorf("NewPodFromK8sPod:GetPortService().NewPortsWithEagerAttrAndLazyAttr(pod4CreatePort: [%v], nwJSONObj [%v]) err , err is [%v]", pod4CreatePort, bpnm, err)
		return nil, err
	}
	return ports, nil
}

func (ps *podService) Save(pod *Pod) error {
	klog.Debugf("podService.Save start pod is [%v]", pod)
	podForDB := pod.transferToPodForDao()
	err := daos.GetPodDao().Save(podForDB)
	if err != nil {
		klog.Errorf("podService.Save:daos.GetPodDao().Save(podForDB:[%v]) err, error is ", podForDB, err)
		return err
	}
	return nil

}

func (ps *podService) Get(podNs, podName string) (*Pod, error) {
	podForDB, err := daos.GetPodDao().Get(podNs, podName)
	if errobj.IsNotFoundError(err) {
		return nil, err
	}
	if err != nil {
		klog.Errorf("podService.Get:daos.GetPodDao().Get(podName:[%v]) err, error is [%v]", podName, err)
		return nil, err
	}
	pod := newPodFromPodForDB(podForDB)
	return pod, nil
}

func (ps *podService) DeletePodAndPorts(podNs, podName string) error {
	klog.Infof("podService.DeletePodAndPorts: start, podNs :[%v], podName:[%v]", podNs, podName)
	podForDB, err := daos.GetPodDao().Move(podNs, podName)
	if err != nil && errobj.IsNotFoundError(err) {
		return nil
	}
	if err != nil && !errobj.IsNotFoundError(err) {
		klog.Errorf("podService.DeletePodAndPorts:Move(podNs:[%v], podName[:%v]) err, error is [%v]", podNs, podName, err)
		return err
	}
	GetPortRecyclePool().AddAndRecycle(NewRecyclePod(podForDB))
	return nil

}

func (ps *podService) ReportPod(reportPod *clients.ReportPod) error {
	klog.Infof("podService.ReportPod start ")

	isReportToManagerSucc, isReportToK8sSucc := getReportResult(reportPod)

	if !isReportToManagerSucc || !isReportToK8sSucc {
		statusNew := &exceptPodSyncStatus{
			PodNs:            reportPod.PodNs,
			PodName:          reportPod.PodName,
			Action:           constvalue.AddAction,
			IsSync2ManagerOk: isReportToManagerSucc,
			IsSync2K8sOk:     isReportToK8sSucc,
		}
		err := GetPod2SyncStatusRepo().Add(reportPod.PodNs, reportPod.PodName, statusNew)
		if err != nil {
			return err
		}
	} else {
		statusOld := GetPod2SyncStatusRepo().Get(reportPod.PodNs, reportPod.PodName)
		if statusOld != nil {
			GetPod2SyncStatusRepo().Del(reportPod.PodNs, reportPod.PodName)
		}
	}

	klog.Infof("podService.ReportPod end ")
	return nil
}


func (ps *podService) DeletePod(podNs, podName string) error {
	pod, err := daos.GetPodDao().Get(podNs, podName)
	if err != nil && !errobj.IsNotFoundError(err) {
		klog.Errorf("daos.GetPodDao().Get(podNs:[%v], podName[:%v]) err, error is [%v]", podNs, podName, err)
		return err
	}
	if errobj.IsNotFoundError(err) {
		klog.Warningf("daos.GetPodDao().Get(podNs:[%v], podName[:%v]) not found", podNs, podName)
		return nil
	}
	portIDs := make([]string, 0)
	for _, port := range pod.Ports {
		portIDs = append(portIDs, port.ID)
	}
	err = GetPortService().DeleteBulkPorts(podNs, portIDs)
	if err != nil {
		klog.Errorf("GetPortService().DeleteBulkPorts(podNs:[%v], portIDs:[%v] ) err, error is [%v]", podNs, portIDs, err)
		return err
	}

	err = daos.GetPodDao().Delete(podNs, podName)
	if err != nil && !errobj.IsNotFoundError(err) {
		klog.Errorf("daos.GetPodDao().Delete(podNs:[%v], podName[%v]) err, error is [%v]", podNs, podName, err)
		return err
	}

	return nil

}


func getReportResult(reportPod *clients.ReportPod) (bool, bool) {
	var isReportToManagerSucc, isReportToK8sSucc bool = true, true
	statusOld := GetPod2SyncStatusRepo().Get(reportPod.PodNs, reportPod.PodName)
	if statusOld != nil && !statusOld.IsSync2ManagerOk || statusOld == nil {
		errOfReportToManager := reportPodToManager(reportPod)
		if errOfReportToManager != nil {
			klog.Warningf("getReportResult: reportPodToManager error: %v", errOfReportToManager)
			isReportToManagerSucc = false
		}
	}
	if statusOld != nil && !statusOld.IsSync2K8sOk || statusOld == nil {
		errOfReportTok8S := reportPodToK8s(reportPod)
		if errOfReportTok8S != nil {
			klog.Warningf("getReportResult: reportPodToK8s error: %v", errOfReportTok8S)
			isReportToK8sSucc = false
		}
	}
	return isReportToManagerSucc, isReportToK8sSucc
}

func reportPodToManager(reportPod *clients.ReportPod) error {
	err := clients.GetManagerClient().ReportPod(reportPod)
	if err != nil {
		klog.Errorf("podService.ReportPod:reportPodToManager(reportPod:[%v]) error: %v", reportPod, err)
		return err
	}
	return nil
}

func (ps *podService) PatchReportPod(podNs, PodName string, req *agtmgr.AgentPodReq) error {
	klog.Infof("podService.PatchReportPod:  agentPodReq is [%v]", req)
	pod, err := ps.Get(podNs, PodName)
	if err != nil {
		klog.Errorf("podService.PatchReportPod: ps.Get(podNs: [%v], PodName: [%v]) error, err is [%v]", podNs, PodName, err)
		return err
	}
	klog.Infof("podService.PatchReportPod:pod is [%+v]", pod)
	var flagNeedPatch = false
	for _, port := range pod.Ports {
		for _, reqPort := range req.Ports {
			klog.Debugf("podService.PatchReportPod:pod.port is [%v], reqPort is [%v]", port, reqPort)
			if needCoverPort(*port, reqPort) {
				klog.Debugf("podService.PatchReportPod: patch ")
				ip := ports.IP{
					IPAddress: reqPort.FixIP,
				}
				ips := make([]ports.IP, 0)
				ips = append(ips, ip)
				port.LazyAttr.FixedIps = ips
				port.LazyAttr.ID = reqPort.PortId
				port.LazyAttr.FixedIPInfos[0] = reqPort.FixedIPInfos
				flagNeedPatch = true
				klog.Debugf("podService.PatchReportPod: flagNeedPatch:[%v] ", flagNeedPatch)
			}
		}
	}
	for i, port := range pod.Ports {
		klog.Debugf("podService.PatchReportPod:after change port[%v] is [%+v]", i, *port)
	}
	if flagNeedPatch {
		err = ps.Save(pod)
		if err != nil {
			klog.Errorf("podService.PatchReportPod: ps.Save(pod: [%v]) error, err is [%v]", pod, err)
			return err
		}
		reportPod := pod.transferToReportPod()
		err = ps.ReportPod(reportPod)
		if err != nil {
			klog.Warningf("podService.PatchReportPod: ps.ReportPod(podNs:[%v], PodName:[%v]) error, err is [%v]", podNs, PodName, err)
		}
	}
	klog.Infof("podService.PatchReportPod: successful")
	return nil
}

func needCoverPort(dbPort Port, reqPort agtmgr.PortInfo) bool {
	return dbPort.EagerAttr.NetworkPlane == reqPort.NetworkPlane &&
		dbPort.EagerAttr.NetworkName == reqPort.NetworkName &&
		(IsPortForC0(dbPort) || dbPort.EagerAttr.VnicType == "host")
}

func IsPortForC0(port Port) bool {
	return IsCTNetPlane(port.EagerAttr.NetworkPlane) &&
		port.EagerAttr.Accelerate == "true"
}

func (ps *podService) ReportDeletePod(podNs, podName string) error {
	klog.Infof("podService.DeleteReportPod:  START podNs is [%v], podName is [%v]", podNs, podName)
	var isReportToManagerSucc, isReportToK8sSucc bool = true, true
	err := reportDeletePodToManager(podNs, podName)
	if err != nil {
		klog.Warningf("ReportDeletePod: reportDeletePodToManager error: %v", err)
		isReportToManagerSucc = false
	}
	statusNew := &exceptPodSyncStatus{
		PodNs:            podNs,
		PodName:          podName,
		Action:           constvalue.DelAction,
		IsSync2ManagerOk: isReportToManagerSucc,
		IsSync2K8sOk:     isReportToK8sSucc,
	}
	if !isReportToManagerSucc {
		statusOld := GetPod2SyncStatusRepo().Get(podNs, podName)
		if statusOld != nil && statusOld.IsAddAction() {
			GetPod2SyncStatusRepo().Del(podNs, podName)
		} else {
			GetPod2SyncStatusRepo().Add(podNs, podName, statusNew)
		}
	} else {
		statusOld := GetPod2SyncStatusRepo().Get(podNs, podName)
		if statusOld != nil {
			GetPod2SyncStatusRepo().Del(podNs, podName)
		}
	}

	klog.Infof("podService.DeleteReportPod: END, podNs is [%v], podName is [%v]", podNs, podName)
	return nil

}

func reportDeletePodToManager(podNs, podName string) error {
	err := clients.GetManagerClient().ReportDeletePod(podNs, podName)
	if err != nil {
		klog.Errorf("podService.DeleteReportPod: clients.GetManagerClient().ReportDeletePod(podNs:[%v], podName:[%v]) error ,"+
			"err is [%v]", podNs, podName, err)
		return err
	}
	return nil
}

func isNotNetworkConfigExist(networksStr string, ok bool) bool {
	return networksStr == "" || !ok || networksStr == "\"\""
}

type Pod struct {
	TenantId            string
	PodID               string
	PodName             string
	PodNs               string
	PodType             string //todo podtype
	IsSuccessful        bool
	ErrorMsg            string
	Ports               []*Port
	ResourceManagerType string
	ResourceManagerName string
}

func newPodFromPodForDB(db *daos.PodForDB) *Pod {
	mports := make([]*Port, 0)
	for _, portForDB := range db.Ports {
		mport := newPortFromPortForDB(portForDB)
		mports = append(mports, mport)
	}
	return &Pod{
		TenantId:            db.TenantId,
		PodID:               db.PodID,
		PodName:             db.PodName,
		PodNs:               db.PodNs,
		PodType:             db.PodType,
		IsSuccessful:        db.IsSuccessful,
		ErrorMsg:            db.ErrorMsg,
		Ports:               mports,
		ResourceManagerType: db.ResourceManagerType,
		ResourceManagerName: db.ResourceManagerName,
	}
}

func (p *Pod) transferToPod4CreatePort() *PodForCreatPort {
	return &PodForCreatPort{
		TenantID: p.TenantId,
		PodID:    p.PodID,
		PodName:  p.PodName,
		PodNs:    p.PodNs,
	}

}

func (p *Pod) transferToPodForDao() *daos.PodForDB {
	portsForDB := make([]*daos.PortForDB, 0)
	for _, port := range p.Ports {
		portForDB := port.transferToPortForDB()
		portsForDB = append(portsForDB, portForDB)
	}
	return &daos.PodForDB{
		TenantId:            p.TenantId,
		PodID:               p.PodID,
		PodNs:               p.PodNs,
		PodName:             p.PodName,
		PodType:             p.PodType,
		ErrorMsg:            p.ErrorMsg,
		IsSuccessful:        p.IsSuccessful,
		Ports:               portsForDB,
		ResourceManagerType: p.ResourceManagerType,
		ResourceManagerName: p.ResourceManagerName,
	}

}

func (p *Pod) transferToReportPod() *clients.ReportPod {

	portInfos := make([]clients.PortInfo, 0)
	for _, port := range p.Ports {
		networkPlanes := LinkStrOfSliceToStrWithDot(port.EagerAttr.Roles)
		portInfo := clients.PortInfo{
			PortID:        port.LazyAttr.ID,
			NetworkName:   port.EagerAttr.NetworkName,
			NetworkPlanes: networkPlanes,
			FixedIPInfos:  port.LazyAttr.FixedIPInfos,
		}
		if len(port.LazyAttr.FixedIps) > 0 {
			portInfo.FixIP = port.LazyAttr.FixedIps[0].IPAddress
		}
		portInfos = append(portInfos, portInfo)
	}

	return &clients.ReportPod{
		PodName:  p.PodName,
		PodNs:    p.PodNs,
		TenantID: p.TenantId,
		Ports:    portInfos,
	}

}

type PodForCreatPort struct {
	TenantID string
	PodID    string
	PodName  string
	PodNs    string
}
