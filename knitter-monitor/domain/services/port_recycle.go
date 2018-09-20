package services

import (
	"encoding/json"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"sync"
	"time"

	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/daos"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/knitter-monitor/infra/db-config"
	"github.com/ZTE/Knitter/pkg/klog"
	"k8s.io/apimachinery/pkg/types"
)

func init() {
	recyclePool = &PortRecyclePool{
		inProcJobs: make(map[RecycleKey]bool),
		Locker:     &sync.Mutex{},
	}
}

type RecycleKey string

func RecyclePortWorker() {
	defer errobj.RecoverPanic()
	for {
		err := collectRecyclePods()
		if err != nil {
			klog.Warningf("RecyclePortWorker err, error is [%v]", err)
		}
		time.Sleep(time.Second * constvalue.DefaultPortRecycleInterval)
	}

}

func collectRecyclePods() error {
	nodes, err := dbconfig.GetDataBase().ReadDir(dbconfig.GetKeyOfMonitorPortRecycle())
	if err != nil {
		klog.Warningf("CollectRecyclePod: ReadDir err, error is [%v]", err)
		return err
	}

	for _, node := range nodes {
		ns := strings.TrimPrefix(node.Key, dbconfig.GetKeyOfMonitorPortRecycle()+"/")
		nodes1, err := dbconfig.GetDataBase().ReadDir(dbconfig.GetKeyOfMonitorPortRecycleNS(ns))
		if err != nil {
			klog.Warningf("CollectRecyclePod: ReadDir(GetKeyOfMonitorPortRecycleNS) err error is[%v] ", err)
			continue
		}
		for _, node1 := range nodes1 {
			PodDB := &daos.PodForDB{}
			err := json.Unmarshal([]byte(node1.Value), PodDB)
			if err != nil {
				klog.Warningf("json.Unmarshal([]byte[node1.Value], PodDB) err, error is [%v]", err)
				continue
			}
			GetPortRecyclePool().AddAndRecycle(NewRecyclePod(PodDB))
		}

	}
	return nil
}

func SyncPod() {
	defer errobj.RecoverPanic()

	k8sPods := CollectK8sPods()
	PodDbs := CollectMonitorPods()
	createPods := GetCreatePod(k8sPods, PodDbs)
	deletePods := GetDeletePod(k8sPods, PodDbs)
	for _, createPod := range createPods {
		klog.Infof("cratePod is [%+v]", createPod)
		GetCreatePorts4PodController().OperateOPod(createPod)
	}
	for _, deletePod := range deletePods {
		klog.Infof("deletePod is [%+v]", deletePod)
		GetCreatePorts4PodController().OperateOPod(deletePod)
	}
}

type RecyclePod struct {
	podNS      string
	podName    string
	recycleKey RecycleKey
	portIDs    []string
}

func NewRecyclePod(podForDb *daos.PodForDB) *RecyclePod {
	portIDs := []string{}
	for _, port := range podForDb.Ports {
		if port != nil && port.ID != "" {
			portIDs = append(portIDs, port.ID)
		}
	}
	return &RecyclePod{
		podNS:      podForDb.PodNs,
		podName:    podForDb.PodName,
		recycleKey: RecycleKey(podForDb.PodNs + podForDb.PodName),
		portIDs:    portIDs,
	}

}

var recyclePool *PortRecyclePool

func GetPortRecyclePool() *PortRecyclePool {
	klog.Debugf("GetPortRecyclePool:")
	return recyclePool
}

type PortRecyclePool struct {
	Locker     sync.Locker
	inProcJobs map[RecycleKey]bool
}

func (pr *PortRecyclePool) AddAndRecycle(recyclePod *RecyclePod) {
	pr.Locker.Lock()
	defer pr.Locker.Unlock()
	_, ok := pr.inProcJobs[recyclePod.recycleKey]
	if ok {
		return
	}
	pr.inProcJobs[recyclePod.recycleKey] = true
	go operateRecycle(pr, recyclePod)

}

func (pr *PortRecyclePool) Delete(recycleKey RecycleKey) {
	pr.Locker.Lock()
	defer pr.Locker.Unlock()
	delete(pr.inProcJobs, recycleKey)
}

func operateRecycle(portRecyclePool *PortRecyclePool, recyclePod *RecyclePod) {
	defer errobj.RecoverPanic()
	defer portRecyclePool.Delete(recyclePod.recycleKey)
	klog.Infof("@@@@RecycleWorker:pod Name is [%v]", recyclePod.podName)
	err := GetPortService().DeleteBulkPorts(recyclePod.podNS, recyclePod.portIDs)
	if err != nil {
		klog.Warningf("RecycleWorker:DeleteBulkPorts(podNs:[%v], portIDs:[%v] ) err, error is [%v]", recyclePod.podNS, recyclePod.portIDs, err)
		return
	}

	err = dbconfig.GetDataBase().DeleteLeaf(dbconfig.GetKeyOfMonitorPortRecyclePod(recyclePod.podNS, recyclePod.podName))
	if err != nil && !errobj.IsNotFoundError(err) {
		klog.Errorf("RecycleWorker:Delete(podNs:[%v], podName[%v]) err, error is [%v]", recyclePod.podNS, recyclePod.podName, err)
		return

	}
}

func GetCreatePod(k8sPods []*v1.Pod, podDBs []*daos.PodForDB) []*operatingPod {
	createPods := []*operatingPod{}
	for _, k8sPod := range k8sPods {
		exist := false
		for _, podDb := range podDBs {
			if k8sPod.Namespace == podDb.PodNs && k8sPod.Name == podDb.PodName {
				exist = true
				break
			}
		}
		if !exist {
			oPod := NewOperatingPod(k8sPod,constvalue.CreateOperation)
			createPods = append(createPods, oPod)

		}
	}
	return createPods
}

func GetDeletePod(k8sPods []*v1.Pod, podDBs []*daos.PodForDB) []*operatingPod {
	deletePods := []*operatingPod{}
	for _, podDb := range podDBs {
		var exist = false
		for _, k8sPod := range k8sPods {
			if k8sPod.Namespace == podDb.PodNs && k8sPod.Name == podDb.PodName {
				exist = true
				break
			}
		}

		if !exist {
			objectMeta := metav1.ObjectMeta{
				Namespace: podDb.PodNs,
				Name:      podDb.PodName,
				UID:  	   types.UID(podDb.PodID),
			}
			k8sPod := &v1.Pod{
				ObjectMeta: objectMeta,
			}

			opod := NewOperatingPod(k8sPod,constvalue.DeleteOperation)

			deletePods = append(deletePods, opod)
		}
	}
	return deletePods
}

func CollectK8sPods() []*v1.Pod {
	k8sPods := []*v1.Pod{}
	podObjs := GetCreatePorts4PodController().podStoreIndexer.List()
	for _, podObj := range podObjs {
		if k8sPod, ok := podObj.(*v1.Pod); ok {
			k8sPods = append(k8sPods, k8sPod)
			klog.Debugf("CollectK8sPods: pod Name is [%v]", k8sPod.Name)
		}
	}
	return k8sPods
}

func CollectMonitorPods() []*daos.PodForDB {

	nodes, err := dbconfig.GetDataBase().ReadDir(dbconfig.GetKeyOfMonitorPods())
	if err != nil {
		klog.Warningf("CollectRecyclePod: ReadDir err, error is [%v]", err)
	}
	podForDBs := []*daos.PodForDB{}
	for _, node := range nodes {
		ns := strings.TrimPrefix(node.Key, dbconfig.GetKeyOfMonitorPods()+"/")
		nodes1, err := dbconfig.GetDataBase().ReadDir(dbconfig.GetKeyOfMonitorPodNs(ns))
		if err != nil {
			klog.Warningf("CollectRecyclePod: ReadDir(GetKeyOfMonitorPortRecycleNS) err error is[%v] ", err)
			continue
		}
		for _, node1 := range nodes1 {
			PodDB := &daos.PodForDB{}
			err := json.Unmarshal([]byte(node1.Value), PodDB)
			if err != nil {
				klog.Warningf("json.Unmarshal([]byte[node1.Value], PodDB) err, error is [%v]", err)
				continue
			}
			klog.Debugf("CollectMonitorPods: PodDb Name is [%v]", PodDB.PodName)
			podForDBs = append(podForDBs, PodDB)
		}

	}
	return podForDBs
}
