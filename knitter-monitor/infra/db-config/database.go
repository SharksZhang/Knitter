package dbconfig

import (
	"time"

	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/uuid"
)

var etcdDataBase dbaccessor.DbAccessor = nil

//todo how to use managerUUID
var managerUUID = ""

func SetDataBase(i dbaccessor.DbAccessor) error {
	etcdDataBase = i
	return nil
}

var GetDataBase = func() dbaccessor.DbAccessor {
	return etcdDataBase
}

func CheckDB() error {
	for {
		/*wait for etcd database OK*/
		e := dbaccessor.CheckDataBase(etcdDataBase)
		if e != nil {
			klog.Error("DataBase Config is ERROR", e)
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	for {
		SetClusterID()
		if GetClusterID() != uuid.NIL.String() {
			break
		}
	}

	return nil
}

var uuidCluster string

func SetClusterID() {
	uuidCluster = GetClusterUUID()
}
func GetClusterID() string {
	return uuidCluster
}


func GetKeyOfClusterUUID() string {
	return GetKeyOfKnitter() + "/cluster_uuid"
}

func GetClusterUUID() string {
	key := GetKeyOfClusterUUID()
	id, err := GetDataBase().ReadLeaf(key)
	if err == nil {
		return id
	}
	if !errobj.IsNotFoundError(err) {
		return uuid.NIL.String()
	}
	id = uuid.NewUUID()
	err = GetDataBase().SaveLeaf(key, id)
	if err == nil {
		return id
	}
	return uuid.NIL.String()
}


func GetKeyOfKnitter() string {
	return "/knitter"
}

func GetKeyOfMonitor() string {
	return GetKeyOfKnitter() + "/monitor"
}

func GetKeyOfMonitorPods() string {
	return GetKeyOfMonitor() + "/pods"
}

func GetKeyOfPodsSyncToManagerAndK8s() string {
	return GetKeyOfMonitor() + "/pods_sync"
}

func GetKeyOfPodSyncToManagerAndK8s(podNS, podName string) string {
	return GetKeyOfPodsSyncToManagerAndK8s() + "/" + podNS + podName
}

func GetKeyOfMonitorPodNs(podNS string) string {
	return GetKeyOfMonitorPods() + "/" + podNS
}

func GetKeyOfMonitorPod(podNS, podName string) string {
	return GetKeyOfMonitorPodNs(podNS) + "/" + podName
}

func GetKeyOfMonitorPortRecycle() string {
	return GetKeyOfMonitor() + "/portrecycle"
}
func GetKeyOfMonitorPortRecycleNS(ns string) string {
	return GetKeyOfMonitorPortRecycle() + "/" + ns
}

func GetKeyOfMonitorPortRecyclePod(podNS, podName string) string {
	return GetKeyOfMonitorPortRecycleNS(podNS) + "/" + podName
}

func GetKeyOfResourceManagers() string {
	return "/knitter/monitor/resource_managers"
}

func GetKeyOfResourceManager(resourceManagerKey string) string {
	return GetKeyOfResourceManagers() + "/" + resourceManagerKey
}