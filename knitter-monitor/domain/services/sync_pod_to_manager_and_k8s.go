package services

import (
	"sync"
	"time"

	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/daos"
	"github.com/ZTE/Knitter/pkg/klog"
)

type exceptPodSyncStatus struct {
	PodNs            string `json:"pod_ns"`
	PodName          string `json:"pod_name"`
	Action           string `json:"action"`
	IsSync2ManagerOk bool   `json:"is_sync_to_manager_ok"`
	IsSync2K8sOk     bool   `json:"is_sync_to_k8s_ok"`
}

func (e exceptPodSyncStatus) IsAddAction() bool {
	if e.Action == constvalue.AddAction {
		return true
	}
	return false
}

func (e exceptPodSyncStatus) IsDelAction() bool {
	if e.Action == constvalue.DelAction {
		return true
	}
	return false
}

//Pod2SyncStatusRepo codes for reporting exceptional pods to k8s and manager recycle
type Pod2SyncStatusRepo struct {
	Pod2SyncStatus map[string]*exceptPodSyncStatus
	Mutex          sync.Mutex
}

var Pod2SyncStatusRepoCtx = Pod2SyncStatusRepo{
	Pod2SyncStatus: make(map[string]*exceptPodSyncStatus),
	Mutex:          sync.Mutex{},
}

func GetPod2SyncStatusRepo() *Pod2SyncStatusRepo {
	return &Pod2SyncStatusRepoCtx
}

func (p *Pod2SyncStatusRepo) Get(podNs, podName string) *exceptPodSyncStatus {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	key := getKey(podNs, podName)
	return p.Pod2SyncStatus[key]
}

func (p *Pod2SyncStatusRepo) GetAll() map[string]*exceptPodSyncStatus {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	return p.Pod2SyncStatus
}

func (p *Pod2SyncStatusRepo) Add(podNs, podName string, status *exceptPodSyncStatus) error {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	key := getKey(podNs, podName)
	p.Pod2SyncStatus[key] = status
	err := daos.GetPodSyncDao().Save(podNs, podName, transPodSyncToPodDB(status))
	if err != nil {
		return err
	}
	return nil
}

func (p *Pod2SyncStatusRepo) Del(podNs, podName string) error {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	key := getKey(podNs, podName)
	delete(p.Pod2SyncStatus, key)
	err := daos.GetPodSyncDao().Delete(podNs, podName)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pod2SyncStatusRepo) Init() error {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	podsDB, err := daos.GetPodSyncDao().GetAll()
	if err != nil {
		return err
	}
	for _, podDB := range podsDB {
		key := getKey(podDB.PodNs, podDB.PodName)
		p.Pod2SyncStatus[key] = transPodDBToPodSync(podDB)
	}
	return nil
}

func getKey(podNs, podName string) string {
	return podNs + podName
}

func transPodDBToPodSync(podDB *daos.PodSyncDB) *exceptPodSyncStatus {
	return &exceptPodSyncStatus{
		PodNs:            podDB.PodNs,
		PodName:          podDB.PodName,
		Action:           podDB.Action,
		IsSync2K8sOk:     podDB.IsSync2K8sOk,
		IsSync2ManagerOk: podDB.IsSync2ManagerOk,
	}
}

func transPodSyncToPodDB(podToSync *exceptPodSyncStatus) *daos.PodSyncDB {
	return &daos.PodSyncDB{
		PodNs:            podToSync.PodNs,
		PodName:          podToSync.PodName,
		Action:           podToSync.Action,
		IsSync2K8sOk:     podToSync.IsSync2K8sOk,
		IsSync2ManagerOk: podToSync.IsSync2ManagerOk,
	}
}

func ReportPodToManagerAndK8sRecycle() {
	const reportRetryInterval time.Duration = 4
	for {
		time.Sleep(reportRetryInterval * time.Second)
		processExceptionalSyncPods()
		time.Sleep(reportRetryInterval * time.Second)
	}
}

func processExceptionalSyncPods() {
	pod2SyncStatus := GetPod2SyncStatusRepo().Pod2SyncStatus
	for key, status := range pod2SyncStatus {
		klog.Infof("ReportPodToManagerAndK8sRecycle: key: %v, value: %+v", key, *status)
		time.Sleep(time.Second)

		if status.IsAddAction() {
			podDB, err := GetPodService().Get(status.PodNs, status.PodName)
			if err != nil {
				klog.Warningf("ReportPodToManagerAndK8sRecycle: GetPodService().Get(podNs: %v, podName: %v) error: %v",
					status.PodNs, status.PodName, err)
				continue
			}
			reportPod := podDB.transferToReportPod()
			GetPodService().ReportPod(reportPod)
		} else if status.IsDelAction() {
			GetPodService().ReportDeletePod(status.PodNs, status.PodName)
		}

		time.Sleep(time.Second)
	}
}
