package services

import (
	"reflect"
	"time"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/pkg/klog"
)

const SyncResourceManagerPortsInterval = 120

func SyncResourceManagerPorts() {
	defer errobj.RecoverPanic()
	for {
		time.Sleep(time.Second * SyncResourceManagerPortsInterval)
		listAndDeleteResourceMgrPorts()
	}
}

func listAndDeleteResourceMgrPorts() {
	resourceManagers := ResourceManagerRepo.GetAll()
	for _, resourceManager := range resourceManagers {
		if reflect.TypeOf(resourceManager) == reflect.TypeOf(&ReplicationControllersResourceManager{}) {
			rcResourceManager := resourceManager.(*ReplicationControllersResourceManager)
			klog.Debugf("listAndDeleteResourceMgrPorts:rcResourceManager.ResourceManager:[%v]", rcResourceManager.ResourceManager)

			exist, err := GetMonitorK8sClient().IsReplicationControllerExist(rcResourceManager.Namespace, rcResourceManager.Name)
			if err != nil {
				klog.Warningf("GetMonitorK8sClient().IsReplicationControllerExist"+
					"(Namespace:[%v], Name:[name]) error, err is [%v]", rcResourceManager.Namespace, rcResourceManager.Name, err)
				continue
			}
			DeleteDuplicateResourcePorts(rcResourceManager, rcResourceManager.DuplicateResourceManager, exist)
		}

		if reflect.TypeOf(resourceManager) == reflect.TypeOf(&StatefulSetResourceManager{}) {
			statefulSetManager := resourceManager.(*StatefulSetResourceManager)
			klog.Debugf("listAndDeleteResourceMgrPorts:statefulSetManager.ResourceManager:[%v]", statefulSetManager.ResourceManager)

			exist, err := GetMonitorK8sClient().IsStatefulSetExist(statefulSetManager.Namespace, statefulSetManager.Name)
			if err != nil {
				klog.Warningf("GetMonitorK8sClient().IsStatefulSetExist"+
					"(Namespace:[%v], Name:[name]) error, err is [%v]", statefulSetManager.Namespace, statefulSetManager.Name, err)
				continue
			}
			DeleteDuplicateResourcePorts(statefulSetManager, statefulSetManager.DuplicateResourceManager, exist)

		}
		if reflect.TypeOf(resourceManager) == reflect.TypeOf(&ReplicaSetResourceManager{}) {
			replicaSetResourceManager := resourceManager.(*ReplicaSetResourceManager)
			klog.Debugf("listAndDeleteResourceMgrPorts:replicaSetResourceManager.ResourceManager:[%v]", replicaSetResourceManager.ResourceManager)
			exist, err := GetMonitorK8sClient().IsReplicaSetExist(replicaSetResourceManager.Namespace, replicaSetResourceManager.Name)
			if err != nil {
				klog.Warningf("GetMonitorK8sClient().IsReplicaSetExist"+
					"(Namespace:[%v], Name:[name]) error, err is [%v]", replicaSetResourceManager.Namespace, replicaSetResourceManager.Name, err)
				continue
			}
			DeleteDuplicateResourcePorts(replicaSetResourceManager, replicaSetResourceManager.DuplicateResourceManager, exist)
		}
	}

}

func DeleteDuplicateResourcePorts(r ResourceManagerInterface, drm *DuplicateResourceManager, exist bool) {
	duplicateResourceManagerValue := *(drm)
	if !exist && len(duplicateResourceManagerValue.UsedResourcePodName) == 0 {
		klog.Infof("DeleteDuplicateResourcePorts: duplicateResourceManager:[%v], ResourceManagerInterface[%v]",
			duplicateResourceManagerValue, r.GetKey())
		var databaseDeleteFlag = true
		if len(duplicateResourceManagerValue.UnusedResource) > 0 {
			databaseDeleteFlag = makeAndDeletePorts(duplicateResourceManagerValue, databaseDeleteFlag)
		}
		if databaseDeleteFlag {
			ResourceManagerRepo.Delete(r)
		}
	}
}

func makeAndDeletePorts(duplicateResourceManagerValue DuplicateResourceManager, databaseDeleteFlag bool) bool {
	for _, ports := range duplicateResourceManagerValue.UnusedResource {
		var podNs string
		var portIDs = make([]string, 0)
		for _, port := range ports {
			klog.Infof("DeleteDuplicateResourcePorts:port:[%+v]", port)
			portID := port.LazyAttr.ID
			podNs = port.EagerAttr.PodNs
			portIDs = append(portIDs, portID)
		}
		err := GetPortService().DeleteBulkPorts(podNs, portIDs)
		if err != nil {
			klog.Warningf("GetPortService().DeleteBulkPorts(podNs:[%v], portIDs:[%v]) err, "+
				"error is [%v]", podNs, portIDs, err)
			databaseDeleteFlag = false
		}
	}
	return databaseDeleteFlag
}
