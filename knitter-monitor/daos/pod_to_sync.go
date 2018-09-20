package daos

import (
	"encoding/json"

	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/knitter-monitor/infra/db-config"
	"github.com/ZTE/Knitter/pkg/klog"
)

func GetPodSyncDao() PodSyncDaoInterface {
	klog.Debugf("GetPodSyncDao")
	return &podSyncDao{}
}

type PodSyncDaoInterface interface {
	Get(podNs, podName string) (*PodSyncDB, error)
	Save(podNs, podName string, podInDB *PodSyncDB) error
	Delete(podNs, podName string) error
	GetAll() ([]*PodSyncDB, error)
}

type podSyncDao struct {
}

func (pd *podSyncDao) Get(podNs, podName string) (*PodSyncDB, error) {
	key := dbconfig.GetKeyOfPodSyncToManagerAndK8s(podNs, podName)

	value, err := dbconfig.GetDataBase().ReadLeaf(key)
	if errobj.IsNotFoundError(err) {
		return nil, err
	}
	if err != nil && !errobj.IsNotFoundError(err) {
		klog.Errorf("dbconfig.GetDataBase().ReadLeaf(key:[%v]) err, error is [%v]", key, err)
		return nil, err
	}
	podInDB := &PodSyncDB{}
	err = json.Unmarshal([]byte(value), podInDB)
	if err != nil {
		klog.Errorf("json.Unmarshal([]byte(value:[%v]), podInDB: [%v]) err, error is [%v]", value, podInDB, err)
		return nil, err
	}
	klog.Debugf("podSyncDao.get SUCC, podInDB is [%v]", podInDB)
	return podInDB, nil

}

func (pd *podSyncDao) GetAll() ([]*PodSyncDB, error) {
	key := dbconfig.GetKeyOfPodsSyncToManagerAndK8s()

	nodes, err := dbconfig.GetDataBase().ReadDir(key)
	if errobj.IsNotFoundError(err) {
		return nil, err
	}
	if err != nil && !errobj.IsNotFoundError(err) {
		klog.Errorf("dbconfig.GetDataBase().ReadLeaf(key:[%v]) err, error is [%v]", key, err)
		return nil, err
	}
	podInDBs := make([]*PodSyncDB, 0)
	for _, node := range nodes {
		podInDB := &PodSyncDB{}
		err = json.Unmarshal([]byte(node.Value), podInDB)
		if err != nil {
			klog.Errorf("json.Unmarshal([]byte(value:[%v]), podInDB: [%v]) err, error is [%v]", node.Value, podInDB, err)
			return nil, err
		}
		podInDBs = append(podInDBs, podInDB)
	}

	return podInDBs, nil
}

func (pd *podSyncDao) Save(podNs, podName string, podInDB *PodSyncDB) error {
	klog.Debugf("podSyncDao.Save start : podInDB is [%v]", podInDB)
	key := dbconfig.GetKeyOfPodSyncToManagerAndK8s(podNs, podName)
	podByte, err := json.Marshal(podInDB)
	if err != nil {
		klog.Errorf("podSyncDao.Save: json.Marshal(podInDB :[%v]) err, error is [%v]", podInDB, err)
		return err
	}
	err = dbconfig.GetDataBase().SaveLeaf(key, string(podByte))
	if err != nil {
		klog.Errorf("podSyncDao.Save: dbconfig.GetDataBase().SaveLeaf(key, string(podByte)) err, error is [%v]", err)
		return err
	}
	klog.Debugf("podSyncDao.Save END : podInDB is [%v]", podInDB)

	return nil
}

func (pd *podSyncDao) Delete(podNs, podName string) error {
	klog.Debugf("podSyncDao delete start, pod name is [%v] ", podName)
	key := dbconfig.GetKeyOfPodSyncToManagerAndK8s(podNs, podName)

	err := dbconfig.GetDataBase().DeleteLeaf(key)
	if err != nil && !errobj.IsNotFoundError(err) {
		klog.Errorf("dbconfig.GetDataBase().DeleteLeaf(key)")
		return err
	}
	klog.Debugf("podSyncDao.delete: podDao delete SUCC")
	return nil

}

type PodSyncDB struct {
	PodNs            string `json:"pod_ns"`
	PodName          string `json:"pod_name"`
	Action           string `json:"action"`
	IsSync2ManagerOk bool   `json:"is_sync_to_manager_ok"`
	IsSync2K8sOk     bool   `json:"is_sync_to_k8s_ok"`
}
