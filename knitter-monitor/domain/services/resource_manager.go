package services

import (
	"sync"

	"k8s.io/apimachinery/pkg/api/errors"
	"strings"
	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/daos"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"time"
)

type ResourceManagerInterface interface {
	Alloc(pod *operatingPod) error
	Free(oPod *operatingPod) error
	GetKey() string
}

type ResourceManager struct {
	Key                 string `json:"Key"`
	ResourceManagerType string `json:"resource_manager_type"`
	Namespace           string `json:"Namespace"`
	Name                string `json:"Name"`
}

func (r *ResourceManager) GetKey() string {
	return r.Key
}

type DefaultResourceManager struct {
	*ResourceManager
}

func NewDefaultResourceManager(resourceManagerType string, namespace string, name string) *DefaultResourceManager {
	return &DefaultResourceManager{
		&ResourceManager{
			Key:                 resourceManagerType + namespace + name,
			ResourceManagerType: resourceManagerType,
			Namespace:           namespace,
			Name:                name,
		},
	}
}

func (r *DefaultResourceManager) Alloc(oPod *operatingPod) error {
	return createResource(oPod)
}

//todo refactor
func createResource(oPod *operatingPod) error {
	k8sPod := oPod.pod
	klog.Debugf("@@@@createResource: oPod Name is [%v]", k8sPod.Name)

	pod, err := GetPodService().NewPodFromK8sPod(k8sPod)
	klog.Debugf(" GetPodService().NewPodFromK8sPod(k8sPod) ,Pod is [%v]", pod)

	if err != nil {
		klog.Errorf("createResource.createPodWorker: GetPodService().NewPodFromK8sPod(k8sPod:[%v]) err,error is [%v]", k8sPod, err)
		return err
	}
	//todo need roll back
	err = GetPodService().Save(pod)
	if err != nil {
		portIDs := make([]string, 0)
		for _, port := range pod.Ports {
			portIDs = append(portIDs, port.LazyAttr.ID)
		}
		err1 := GetPortService().DeleteBulkPorts(pod.PodNs, portIDs)
		if err1 != nil {
			klog.Warningf("GetPortService().DeleteBulkPorts(pod.PodNs, portIDs) err, error is [%v]", err1)
		}

		klog.Errorf("createResource.createPodWorker:GetPodService().Save( pod:[%v] ) err ,error is [%v]", pod, err)
		return err
	}
	reportPod := pod.transferToReportPod()
	err = GetPodService().ReportPod(reportPod)
	if err != nil {
		klog.Warningf("createResource.createPodWorker:GetPodService().ReportPod(k8sPod.Namespace, k8sPod.Name)")
	}
	klog.Infof("createResource.createPodWorker:create pod end, pod is [%+v]", pod)
	return nil
}

func (r *DefaultResourceManager) Free(oPod *operatingPod) error {
	return deleteResource(oPod)
}

func deleteResource(oPod *operatingPod) error {
	klog.Debugf("@@@@deleteResource:oPod Key :[%v], pod id is [%v]", oPod.pod.Name, oPod.pod.UID)
	pod := oPod.pod
	err := GetPodService().DeletePodAndPorts(pod.Namespace, pod.Name)
	if err != nil {
		klog.Warningf("deleteResource.deletePodWorker:GetPodService().DeletePodAndPorts(podName:[%v]) err, error is [%v]", pod.Name, err)
		return err
	}

	err = GetPodService().ReportDeletePod(pod.Namespace, pod.Name)
	if err != nil {
		klog.Warningf("deleteResource:GetPodService().ReportDeletePod(podName:[%v]) err, error is [%v]", pod.Name, err)
	}
	klog.Infof("deleteResource: successfully pod Name is [%v], pod id is [%v]", pod.Name, pod.UID)
	return nil
}

type DuplicateResourceManager struct {
	LocalResource       int       `json:"local_resource"`
	UnusedResource      [][]*Port `json:"unused_resource"`
	UsedResourcePodName []string  `json:"used_resource_pod_name"`
}

func (drm *DuplicateResourceManager) GetLocalResource() int {
	return drm.LocalResource
}

func (drm *DuplicateResourceManager) SetLocalResource(localResource int) {
	drm.LocalResource = localResource
}

func (drm *DuplicateResourceManager) GetUnusedResource() [][]*Port {
	return drm.UnusedResource
}

func (drm *DuplicateResourceManager) AppendUnusedResource(ports []*Port) {
	drm.UnusedResource = append(drm.UnusedResource, ports)

}

func (drm *DuplicateResourceManager) DeleteUnusedResource() {
	drm.UnusedResource = drm.UnusedResource[1:]
}

func (drm *DuplicateResourceManager) AppendUsedResourcePodName(podName string) {
	drm.UsedResourcePodName = append(drm.UsedResourcePodName, podName)
}

func (drm *DuplicateResourceManager) DeleteUsedResourcePodName(podName string) {
	for index, name := range drm.UsedResourcePodName {
		if name == podName {
			if index == len(drm.UsedResourcePodName)-1 {
				drm.UsedResourcePodName = drm.UsedResourcePodName[0:index]
				return
			}
			drm.UsedResourcePodName = append(drm.UsedResourcePodName[0:index], drm.UsedResourcePodName[index+1:]...)
			return
		}
	}
}

func (drm *DuplicateResourceManager) IsPodUsedResource(podName string) bool {
	for _, name := range drm.UsedResourcePodName {
		if name == podName {
			return true
		}
	}
	return false
}

type ReplicationControllersResourceManager struct {
	*ResourceManager          `json:"ResourceManager"`
	*DuplicateResourceManager `json:"DuplicateResourceManager"`
	locker                    sync.Mutex
}

func NewReplicationControllersResourceManager(resourceManagerType string, namespace string, name string) *ReplicationControllersResourceManager {
	return &ReplicationControllersResourceManager{
		ResourceManager: &ResourceManager{
			Key:                 resourceManagerType + namespace + name,
			ResourceManagerType: resourceManagerType,
			Namespace:           namespace,
			Name:                name,
		},
		DuplicateResourceManager: &DuplicateResourceManager{
			LocalResource:       0,
			UnusedResource:      make([][]*Port, 0),
			UsedResourcePodName: make([]string, 0),
		},
	}
}

func (r *ReplicationControllersResourceManager) Alloc(oPod *operatingPod) error {
	r.locker.Lock()
	defer r.locker.Unlock()
	return duplicateAlloc(oPod, r.GetReplica, r.DuplicateResourceManager, r)
}

func (r *ReplicationControllersResourceManager) Free(oPod *operatingPod) error {
	r.locker.Lock()
	defer r.locker.Unlock()
	return duplicateFree(oPod, r.GetReplica, r.DuplicateResourceManager, r)
}

func (r *ReplicationControllersResourceManager) GetReplica() (int, error) {
	return GetMonitorK8sClient().GetReplicationControllerReplicas(r.Namespace, r.Name)
}

type StatefulSetResourceManager struct {
	*ResourceManager
	*DuplicateResourceManager
}

func NewStatefulSetResourceManager(resourceManagerType string, namespace string, name string) *StatefulSetResourceManager {
	return &StatefulSetResourceManager{
		ResourceManager: &ResourceManager{
			Key:                 resourceManagerType + namespace + name,
			ResourceManagerType: resourceManagerType,
			Namespace:           namespace,
			Name:                name,
		},
		DuplicateResourceManager: &DuplicateResourceManager{
			LocalResource:       0,
			UnusedResource:      make([][]*Port, 0),
			UsedResourcePodName: make([]string, 0),
		},
	}
}

func (r *StatefulSetResourceManager) Alloc(pod *operatingPod) error {
	return duplicateAlloc(pod, r.GetReplica, r.DuplicateResourceManager, r)
}

func (r *StatefulSetResourceManager) Free(oPod *operatingPod) error {
	return duplicateFree(oPod, r.GetReplica, r.DuplicateResourceManager, r)
}

func (r *StatefulSetResourceManager) GetReplica() (int, error) {
	return GetMonitorK8sClient().GetStatefulSetReplicas(r.Namespace, r.Name)
}

type ReplicaSetResourceManager struct {
	*ResourceManager
	*DuplicateResourceManager
}

func NewReplicaSetResourceManager(resourceManagerType string, namespace string, name string) *ReplicaSetResourceManager {
	return &ReplicaSetResourceManager{
		ResourceManager: &ResourceManager{
			Key:                 resourceManagerType + namespace + name,
			ResourceManagerType: resourceManagerType,
			Namespace:           namespace,
			Name:                name,
		},
		DuplicateResourceManager: &DuplicateResourceManager{
			LocalResource:       0,
			UnusedResource:      make([][]*Port, 0),
			UsedResourcePodName: make([]string, 0),
		},
	}
}

func (r *ReplicaSetResourceManager) Alloc(pod *operatingPod) error {
	return duplicateAlloc(pod, r.GetReplica, r.DuplicateResourceManager, r)
}

func (r *ReplicaSetResourceManager) Free(oPod *operatingPod) error {
	return duplicateFree(oPod, r.GetReplica, r.DuplicateResourceManager, r)
}

func (r *ReplicaSetResourceManager) GetReplica() (int, error) {
	return GetMonitorK8sClient().GetReplicaSetReplicas(r.Namespace, r.Name)
}

func ResourceManagerFactory(resourceManagerType string, namespace string, name string) ResourceManagerInterface {
	klog.Infof("ResourceManagerFactory: ResourceManagerType [%v], Namespace:[%v], Name:[%v]", resourceManagerType,
		namespace, name)
	if strings.ToLower(resourceManagerType) == constvalue.TypeReplicationController {
		klog.Infof("NewReplicationControllersResourceManager")
		return NewReplicationControllersResourceManager(resourceManagerType, namespace, name)
	}

	if strings.ToLower(resourceManagerType) == constvalue.TypeReplicaSet {
		klog.Infof("NewReplicaSetResourceManager")
		return NewReplicaSetResourceManager(resourceManagerType, namespace, name)
	}

	if strings.ToLower(resourceManagerType) == constvalue.TypeStatefulSet {
		klog.Infof("NewStatefulSetResourceManager")
		return NewStatefulSetResourceManager(resourceManagerType, namespace, name)
	}
	klog.Infof("NewDefaultResourceManager")
	return NewDefaultResourceManager(resourceManagerType, namespace, name)
}


func duplicateAlloc(oPod *operatingPod, getReplica func() (int, error), manager *DuplicateResourceManager,
	managerInterface ResourceManagerInterface) error {
	replica, err := getReplica()
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("Alloc:GetReplica err, error is [%v]", err)
		return err
	}
	klog.Infof("ReplicationControllersResourceManager.Alloc: replica is [%v], manager.LocalResource is [%v]", replica, manager.GetLocalResource())
	if replica <= manager.GetLocalResource() {
		if len(manager.GetUnusedResource()) == 0 {
			return errobj.ErrResourceInUse
		}
		pod := &Pod{
			TenantId:            oPod.pod.Namespace,
			PodID:               string(oPod.pod.UID),
			PodName:             oPod.pod.Name,
			PodNs:               oPod.pod.Namespace,
			PodType:             "",
			IsSuccessful:        true,
			ErrorMsg:            "",
			Ports:               manager.GetUnusedResource()[0],
			ResourceManagerType: oPod.ResourceManagerType,
			ResourceManagerName: oPod.ResourceManagerName,
		}
		err := GetPodService().Save(pod)
		if err != nil {
			klog.Errorf("Alloc:GetPodService().Save(pod[%v]) err, error is [%v] ", pod, err)
			return err
		}
		manager.DeleteUnusedResource()
		manager.AppendUsedResourcePodName(pod.PodName)

		err = ResourceManagerRepo.Save(managerInterface)
		if err != nil {
			klog.Errorf("SaveResourceManagerToETCD(managerInterface) err, error is [%v]", err)
			manager.AppendUnusedResource(pod.Ports)
			manager.DeleteUsedResourcePodName(pod.PodName)
			for {
				err1 := daos.GetPodDao().Delete(pod.PodNs, pod.PodName)
				if err1 != nil && ! errobj.IsNotFoundError(err1) {
					klog.Warningf("duplicateAlloc: rollback DeletePod err, error is [%v]", err1)
					continue
				}
				break
			}
			return err
		}

		reportPod := pod.transferToReportPod()
		err = GetPodService().ReportPod(reportPod)
		if err != nil {
			klog.Warningf("podService.PatchReportPod: ps.ReportPod(podNs:[%v], PodName:[%v]) error, err is [%v]", reportPod.PodNs, reportPod.PodName, err)
		}

	} else {
		if err := createResource(oPod); err != nil {
			klog.Errorf("ReplicationControllersResourceManager.Alloc err")
			return err
		}
		manager.SetLocalResource(manager.GetLocalResource() + 1)
		manager.AppendUsedResourcePodName(oPod.name)

		err = ResourceManagerRepo.Save(managerInterface)
		if err != nil {
			manager.SetLocalResource(manager.GetLocalResource() - 1)
			manager.DeleteUsedResourcePodName(oPod.name)
			err1 := GetPodService().DeletePod(oPod.NameSpace, oPod.name)
			if err1 != nil {
				klog.Warningf("duplicateAlloc: rollback GetPodService().DeletePod err, error is [%v]", err1)
			}

			klog.Errorf("SaveResourceManagerToETCD(managerInterface) err, error is [%v]", err)
			return err
		}
	}
	return nil
}


func duplicateFree(oPod *operatingPod, getReplica func() (int, error), manager *DuplicateResourceManager,
	managerInterface ResourceManagerInterface) error {
	klog.Infof("duplicateFree start, pod Name is [%v], pod ID is [%v]", oPod.pod.Name, oPod.pod.UID)
	klog.Infof("DuplicateResourceManager:[%+v]", manager)
	if !manager.IsPodUsedResource(oPod.name) {
		klog.Warningf("duplicateFree: Pod has been released")
		return nil
	}
	replica, err := getReplica()
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("Free:GetReplica err, error is [%v]", err)
		return err
	}
	klog.Infof("duplicateFree:replica:[%v], manager.LocalResource[%v]", replica, manager.GetLocalResource())
	if replica < manager.GetLocalResource() {
		if err := deleteResource(oPod); err != nil {
			klog.Errorf("Free:deleteResource err, error is [%v]", err)
			return err
		}
		manager.SetLocalResource(manager.GetLocalResource() - 1)
		manager.DeleteUsedResourcePodName(oPod.name)
	} else {
		klog.Infof("Free:delete database only")
		pod, err := GetPodService().Get(oPod.pod.Namespace, oPod.pod.Name)
		if err != nil {
			klog.Errorf("Free:Get(ns:[%v], Name:[%v]) err, error is [%v]", oPod.pod.Namespace, oPod.pod.Name, err)
			return err
		}
		//todo refactor
		err = daos.GetPodDao().Delete(oPod.pod.Namespace, oPod.pod.Name)
		if err != nil {
			klog.Errorf("deleteResource: Delete(ns:[%v], Name[%v]) err, error is [%v]", oPod.pod.Namespace, oPod.pod.Name, err)
			return err
		}
		manager.UnusedResource = append(manager.UnusedResource, pod.Ports)
		manager.DeleteUsedResourcePodName(oPod.name)
		err = GetPodService().ReportDeletePod(oPod.pod.Namespace, oPod.pod.Name)
		if err != nil {
			klog.Warningf("deleteResource: GetPodService().ReportDeletePod(ns:[%v], Name[%v]) err, error is [%v]", oPod.pod.Namespace, oPod.pod.Name, err)
		}
	}
	if manager.GetLocalResource() == 0 && replica == 0 {
		for{
			err := ResourceManagerRepo.Delete(managerInterface)
			if err != nil && !errobj.IsNotFoundError(err) {
				klog.Warningf("DeleteResourceManagerToETCD(managerInterface[%v]) err, error is [%v]", managerInterface, err)
				time.Sleep(constvalue.EtcdRetryTimes * time.Second)
				continue
			}
			return nil
		}

	}

	for {
		err = ResourceManagerRepo.Save(managerInterface)
		if err != nil {
			klog.Warningf("SaveResourceManagerToETCD(managerInterface) err, error is [%v]", err)
			time.Sleep(constvalue.EtcdRetryTimes * time.Second)
			continue
		}
		break
	}

	return nil
}


