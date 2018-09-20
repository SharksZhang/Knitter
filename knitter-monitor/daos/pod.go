package daos

import (
	"encoding/json"

	"strings"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/knitter-monitor/infra/db-config"
	"k8s.io/client-go/tools/cache"
	"github.com/ZTE/Knitter/pkg/klog"
)

func GetPodDao() PodDaoInterface {
	//do not delete, for monkey ut
	klog.Debugf("GetPodDao")
	return &podDao{}
}

type PodDaoInterface interface {
	Get(podNs, podName string) (*PodForDB, error)
	Save(podForDB *PodForDB) error
	Delete(podNs, podName string) error
	Move(podNs, podName string) (*PodForDB, error)
}

type podDao struct {
}

func (pd *podDao) Get(podNs, podName string) (*PodForDB, error) {
	klog.Debugf("podDao.get start,pod name is [%v] ", podName)
	podForDB, err := GetPodForDBRepoSingleton().Get(podNs + podName)
	if errobj.IsNotFoundError(err) {
		return nil, err
	}
	if err != nil {
		klog.Errorf("GetPodForDBRepoSingleton().Get(podNs + podName:[%v]) err, error is [%v]", podNs+podName, err)
		return nil, err
	}
	klog.Debugf("podDao.get SUCC, podForDB is [%v]", podForDB)
	return podForDB, nil

}

func (pd *podDao) Save(podForDB *PodForDB) error {
	klog.Debugf("podDao.Save start : pod forDB is [%v]", podForDB)
	key := dbconfig.GetKeyOfMonitorPod(podForDB.PodNs, podForDB.PodName)
	podByte, err := json.Marshal(podForDB)
	if err != nil {
		klog.Errorf("podDao.Save: json.Marshal(podForDB :[%v]) err, error is [%v]", podForDB, err)
		return err
	}
	err = dbconfig.GetDataBase().SaveLeaf(key, string(podByte))
	if err != nil {
		klog.Errorf("podDao.Save: dbconfig.GetDataBase().SaveLeaf(key, string(podByte)) err, error is [%v]", err)
		return err
	}
	err = GetPodForDBRepoSingleton().Add(podForDB)
	if err != nil {
		klog.Warningf("GetPodForDBRepoSingleton().Add(podForDB:[%v]) err, error is [%v]", podForDB, err)
	}
	klog.Debugf("podDao.Save END : pod forDB is [%v]", podForDB)

	return nil
}

func (pd *podDao) Delete(podNs, podName string) error {
	klog.Debugf("podDao delete start,pod name is [%v] ", podName)
	key := dbconfig.GetKeyOfMonitorPod(podNs, podName)

	err := dbconfig.GetDataBase().DeleteLeaf(key)
	if err != nil && !errobj.IsNotFoundError(err) {
		klog.Errorf("dbconfig.GetDataBase().DeleteLeaf(key)")
		return err
	}
	err = GetPodForDBRepoSingleton().Del(podNs + podName)
	if err != nil && errobj.IsNotFoundError(err) {
		klog.Warningf("GetPodForDBRepoSingleton().Del(podNs + podName)")
	}
	klog.Debugf("podDao.delete:podDao delete SUCC")
	return nil

}

func (pd *podDao) Move(podNs, podName string) (*PodForDB, error) {
	//todo
	key := dbconfig.GetKeyOfMonitorPod(podNs, podName)

	value, err := dbconfig.GetDataBase().ReadLeaf(key)
	if err != nil && errobj.IsNotFoundError(err) {
		err1 := GetPodForDBRepoSingleton().Del(podNs + podName)
		if err1 != nil {
			klog.Warningf("GetPodForDBRepoSingleton().Del(podNs + podName) err, error is [%v]", err)
		}
		return nil, err
	}
	if err != nil && !errobj.IsNotFoundError(err) {
		klog.Errorf("PodDao.Move:ReadLeaf(key:[%v]) err, error is [%v]", key, err)
		return nil, err
	}

	podForDB := &PodForDB{}
	err = json.Unmarshal([]byte(value), podForDB)
	if err != nil {
		klog.Errorf("PodDao.Move:json.Unmarshal([]byte(value:[%v]), podForDB: [%v]) err, error is [%v]", value, podForDB, err)
		return nil, err
	}

	err = dbconfig.GetDataBase().DeleteAndSaveLeaf(key, dbconfig.GetKeyOfMonitorPortRecyclePod(podNs, podName), value)
	if err != nil {
		klog.Errorf("PodDao.Move:DeleteAndSaveLeaf(key: [%v], savekey: [%v], value:[%v] ) err, error is [%v]",
			key, dbconfig.GetKeyOfMonitorPod(podNs, podName), value, err)
		return nil, err
	}
	err = GetPodForDBRepoSingleton().Del(podNs + podName)
	if err != nil && !errobj.IsNotFoundError(err) {
		klog.Errorf("GetPodForDBRepoSingleton().Del(podNs + podName) err, error is [%v]", err)
	}
	klog.Infof("PodDao.Move: success, podName is [%v]", podName)
	return podForDB, nil

}

type PodForDB struct {
	TenantId            string       `json:"tenant_id"`
	PodID               string       `json:"pod_id"`
	PodName             string       `json:"pod_name"`
	PodNs               string       `json:"pod_ns"`
	PodType             string       `json:"pod_type"`
	IsSuccessful        bool         `json:"is_successful"`
	ErrorMsg            string       `json:"error_msg"`
	Ports               []*PortForDB `json:"ports"`
	ResourceManagerType string       `json:"resource_manager_type"`
	ResourceManagerName string       `json:"resource_manager_name"`
}

type PodForDBRepo struct {
	indexer cache.Indexer
}

var podForDBRepo PodForDBRepo

func GetPodForDBRepoSingleton() *PodForDBRepo {
	return &podForDBRepo
}

func PodForDBKeyFunc(obj interface{}) (string, error) {
	if obj == nil {
		klog.Error("PodForDBKeyFunc: obj arg is nil")
		return "", errobj.ErrObjectPointerIsNil
	}

	podForDB, ok := obj.(*PodForDB)
	if !ok {
		klog.Error("PodForDBKeyFunc: obj arg is not type: *PodForDB")
		return "", errobj.ErrArgTypeMismatch
	}

	return podForDB.PodNs + podForDB.PodName, nil
}

func (p *PodForDBRepo) Init() {
	indexers := cache.Indexers{}
	p.indexer = cache.NewIndexer(PodForDBKeyFunc, indexers)
}

func (p *PodForDBRepo) Add(podForDB *PodForDB) error {
	err := p.indexer.Add(podForDB)
	if err != nil {
		klog.Errorf("PodForDBRepo.Add: add obj[%v] to repo FAILED, error: %v", podForDB, err)
		return err
	}
	klog.Infof("PodForDBRepo.Add: add obj[%v] to repo SUCC", podForDB)
	return nil
}

func (p *PodForDBRepo) Del(ID string) error {
	err := p.indexer.Delete(ID)
	if err != nil {
		klog.Errorf("PodForDBRepo.Del: DeleteByKey(ID: %s) FAILED, error: %v", ID, err)
		return err
	}

	klog.Infof("PodForDBRepo.Del: DeleteByKey(ID: %s) SUCC", ID)
	return nil
}

func (p *PodForDBRepo) Get(ID string) (*PodForDB, error) {
	item, exists, err := p.indexer.GetByKey(ID)
	if err != nil {
		klog.Errorf("PodForDBRepo.Get: PodForDB[%s]'s object FAILED, error: %v", ID, err)
		return nil, err
	}
	if !exists {
		klog.Debugf("PodForDBRepo.Get: PodForDB[%s]'s err, error is [%v]", ID, errobj.ErrRecordNotExist)
		return nil, errobj.ErrRecordNotExist
	}

	podForDB, ok := item.(*PodForDB)
	if !ok {
		klog.Errorf("PodForDBRepo.Get: ID[%s]'s object[%v] type not match PodForDB", ID, item)
		return nil, errobj.ErrObjectTypeMismatch
	}
	klog.Infof("PodForDBRepo.Get: ID[%s]'s object[%v] SUCC", ID, podForDB)
	return podForDB, nil
}

func init() {
	GetPodForDBRepoSingleton().Init()
}

func LoadAllPodForDBs() error {
	podForDBs, err := GetAllPodForDBs()
	if err != nil {
		klog.Errorf("LoadAllPodForDBs: GetAllPodForDBs FAILED, error: %v", err)
		return err
	}

	for _, podForDB := range podForDBs {
		err = GetPodForDBRepoSingleton().Add(podForDB)
		if err != nil {
			klog.Errorf("LoadAllPodForDBs: GetPodForDBRepoSingleton().Add(podForDB) FAILED, error: %v",
				podForDB, err)
			return err
		}
		klog.Tracef("LoadAllPodForDBs: GetPodForDBRepoSingleton().Add(podForDB: %v) SUCC",
			podForDB)
	}
	return nil
}

func GetAllPodForDBs() ([]*PodForDB, error) {
	key := dbconfig.GetKeyOfMonitorPods()
	nodes, err := dbconfig.GetDataBase().ReadDir(key)
	if err != nil {
		klog.Errorf("GetAllPodForDBs: ReadDir(key: %s) FAILED, error: %v", key, err)
		return nil, err
	}

	podForDBs := make([]*PodForDB, 0)
	klog.Infof("nodes:[%v]", nodes)
	for _, node := range nodes {
		klog.Infof("node:[%v]", node)
		podNS := strings.TrimPrefix(node.Key, key+"/")
		klog.Infof("podNs:[%v]", podNS)
		podNodes, err := dbconfig.GetDataBase().ReadDir(dbconfig.GetKeyOfMonitorPodNs(podNS))
		klog.Infof("podNodes:[%v]", podNodes)
		if err != nil {
			klog.Errorf("dbconfig.GetDataBase().ReadDir(dbconfig.GetKeyOfMonitorPodNs(podNS):[%v]) err, error is [%v] ",
				dbconfig.GetKeyOfMonitorPodNs(podNS), err)
			return nil, err
		}
		for _, podNode := range podNodes {
			klog.Infof("podNode:[%v]", podNode)
			podForDB, err := UnmarshalPodForDB([]byte(podNode.Value))
			if err != nil {
				klog.Errorf("UnmarshalPodForDB([]byte(podNode.Value:[%v])) err, error is [%v]", node.Value, err)
				return nil, err
			}
			podForDBs = append(podForDBs, podForDB)
		}
	}

	klog.Tracef("GetAllPodForDBs: get all logical networks: %v SUCC", podForDBs)
	return podForDBs, nil
}

var UnmarshalPodForDB = func(value []byte) (*PodForDB, error) {
	var podForDB PodForDB
	err := json.Unmarshal([]byte(value), &podForDB)
	if err != nil {
		klog.Errorf("UnmarshalPodForDB: json.Unmarshal(%s) FAILED, error: %v", string(value), err)
		return nil, errobj.ErrUnmarshalFailed
	}

	klog.Infof("UnmarshalPodForDB: PodForDB[%v] SUCC", podForDB)
	return &podForDB, nil
}
