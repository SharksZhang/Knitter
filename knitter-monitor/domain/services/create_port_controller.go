package services

import (
	"errors"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"math"
	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/knitter-monitor/infra/clients"
	"github.com/ZTE/Knitter/pkg/klog"
	"reflect"
	"github.com/ZTE/Knitter/knitter-monitor/daos"
)

var CreatePorts4PodController *CreatePortsController

func GetCreatePorts4PodController() *CreatePortsController {
	klog.Debugf("GetCreatePort4PodController")
	return CreatePorts4PodController

}

type CreatePortsController struct {
	clientSet       *kubernetes.Clientset
	podController   cache.Controller
	podStoreIndexer cache.Indexer
	PodEventMap     *PodAndEventMap
}

func NewCreatePortsController() (*CreatePortsController, error) {
	cpc := &CreatePortsController{}

	cpc.PodEventMap = &PodAndEventMap{
		Event: make(map[string]*operatingPod),
	}
	cpc.clientSet = clients.GetClientset()
	if cpc.clientSet == nil {
		klog.Errorf("newCreatePortForPodController: monitorcommon.GetClientSet() kubernetes clientSet is nil")
		return nil, errors.New("kubernetes clientSet is nil")
	}
	watchlist := cache.NewListWatchFromClient(cpc.clientSet.CoreV1().RESTClient(), "pods", v1.NamespaceAll,
		fields.Everything())

	cpc.podStoreIndexer, cpc.podController = cache.NewIndexerInformer(
		watchlist,
		&v1.Pod{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    cpc.enqueueCreatePod,
			DeleteFunc: cpc.enqueueDeletePod,
			UpdateFunc: cpc.updateFunc,
		},
		cache.Indexers{},
	)
	klog.Infof("create CreatePortForPodControlle successful")
	return cpc, nil
}

func (cpc *CreatePortsController) Run(workers int, stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	klog.Infof("CreatePortForPodController.Run : Starting serviceLookupController Manager ")
	go cpc.podController.Run(stopCh)
	var i int
	for i = 1; i < constvalue.WaitForCacheSyncTimes; i++ {
		if cache.WaitForCacheSync(stopCh, cpc.podController.HasSynced) {
			klog.Infof("cache.WaitForCacheSync(stopCh, cpc.podController.HasSynced) error")
			break
		}
		time.Sleep(time.Second * 1)
	}
	if i == constvalue.WaitForCacheSyncTimes {
		klog.Errorf("CreatePortForPodController.Run: cache.WaitForCacheSync(stopCh, cpc.podController.HasSynced:[%v]) error,", cpc.podController.HasSynced())
		return
	}

	SyncPod()
	go RecyclePortWorker()
	klog.Infof("CreatePortForPodController.Run : Started podWorker")

	<-stopCh
	klog.Infof("Shutting down Service Lookup Controller")

}

func (cpc *CreatePortsController) updateFunc(oldObj, newObj interface{}) {
	newPod := newObj.(*v1.Pod)
	klog.Debugf("updateFunc pod: [%v]", newPod)

	//todo modify evicted
	if newPod.DeletionTimestamp != nil || newPod.Status.Reason == "Evicted" {
		klog.Infof("updateFunc pod: [%v], prepare delete", newPod)

		oPod := NewOperatingPod(newObj.(*v1.Pod), constvalue.DeleteOperation)
		if oPod.ResourceManagerName == "" || oPod.ResourceManagerType == "" {
			oPod.ResourceManagerType, oPod.ResourceManagerName = SetOperatingPodResourceManager(oPod)
		}
		klog.Infof("@@@@updateFunc:pod name is [%v], id is [%v]", oPod.name, oPod.pod.UID)
		klog.Infof("@@@@pod become unknown, delete pod, oPod is [%v]", oPod)
		cpc.OperateOPod(oPod)
	}
}

func (cpc *CreatePortsController) enqueueCreatePod(obj interface{}) {
	if _, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		return
	}
	oPod := NewOperatingPod(obj.(*v1.Pod), constvalue.CreateOperation)
	klog.Infof("@@@@enqueueCreatePod:pod name is [%v], id is [%v]", oPod.name, oPod.pod.UID)
	klog.Debugf("CreatePortForPodController.enqueueCreatePod successfully :oPod is [%v], pod id is [%v]", oPod, oPod.pod.UID)
	klog.Infof("CreatePortForPodController.enqueueCreatePod successfully :oPod name is [%v], pod id is [%v]", oPod.name, oPod.pod.UID)
	cpc.OperateOPod(oPod)

}

func (cpc *CreatePortsController) enqueueDeletePod(obj interface{}) {
	if _, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		return
	}
	oPod := NewOperatingPod(obj.(*v1.Pod), constvalue.DeleteOperation)
	klog.Infof("@@@@enqueueDeletePod:pod name is [%v], id is [%v]", oPod.name, oPod.pod.UID)
	klog.Infof("delete pod: [%v]", obj.(*v1.Pod))

	if oPod.ResourceManagerName == "" || oPod.ResourceManagerType == "" {
		oPod.ResourceManagerType, oPod.ResourceManagerName = SetOperatingPodResourceManager(oPod)
	}
	klog.Infof("CreatePortForPodController.enqueueDeletePod successfully :oPod is [%v], pod id is [%v]", oPod, oPod.pod.UID)
	cpc.OperateOPod(oPod)
}

func SetOperatingPodResourceManager(oPod *operatingPod) (string, string) {
	for {
		pod, err := GetPodService().Get(oPod.NameSpace, oPod.name)
		if err != nil && !errobj.IsNotFoundError(err) {
			klog.Warningf("GetPodService().Get(oPod.NameSpace, oPod.name) err , error is [%v]", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err != nil && errobj.IsNotFoundError(err) {
			klog.Warningf("GetPodService().Get(oPod.NameSpace, oPod.name) err , error is [%v]", err)
			return "", ""
		}

		return pod.ResourceManagerType, pod.ResourceManagerName

	}
}

func (cpc *CreatePortsController) OperateOPod(oPod *operatingPod) {
	key := oPod.pod.Namespace + oPod.pod.Name
	klog.Infof("OperateOPod: pod Name and operate are [%v],id is [%v] ", oPod.pod.Name+oPod.operation, oPod.pod.UID)

	exist := cpc.PodEventMap.CheckExistAndAdd(key, oPod)
	if !exist {
		klog.Debugf("CreatePortForPodController.enqueueDeletePod: go Operate")
		go Operate(cpc.PodEventMap, oPod.pod.Namespace, oPod.pod.Name)
	}
}

type PodAndEventMap struct {
	lock  sync.RWMutex
	Event map[string]*operatingPod
}

func (pe *PodAndEventMap) add(key string, value *operatingPod) {
	pe.lock.Lock()
	defer pe.lock.Unlock()
	pe.Event[key] = value
	klog.Infof("PodAndEventMap.add Key is [%v], value is [%v]", key, value)

}

func (pe *PodAndEventMap) CheckExistAndAdd(key string, value *operatingPod) bool {
	pe.lock.Lock()
	defer pe.lock.Unlock()
	exist := true
	events, ok := pe.Event[key]
	if !ok || events == nil {
		exist = false
	}
	pe.Event[key] = value
	klog.Infof("PodAndEventMap.CheckExistAndAdd Key is [%v], value is [%v]", key, value)
	return exist
}

func (pe *PodAndEventMap) CheckEqualAndDelete(key string, value *operatingPod) bool {
	pe.lock.Lock()
	defer pe.lock.Unlock()
	currentValue := pe.Event[key]
	if currentValue == value {
		delete(pe.Event, key)
		return true
	}
	klog.Infof("PodAndEventMap.CheckEqualAndDelete Key is [%v], value is [%v]", key, value)
	return false
}

func (pe *PodAndEventMap) delete(key string) {
	pe.lock.Lock()
	defer pe.lock.Unlock()
	events, exists := pe.Event[key]
	if !exists || events == nil {
		return
	}
	delete(pe.Event, key)
	klog.Infof("PodAndEventMap.delete Key is [%v]", key)

}

func (pe *PodAndEventMap) Get(key string) *operatingPod {
	pe.lock.Lock()
	defer pe.lock.Unlock()
	return pe.Event[key]
}

type operatingPod struct {
	pod                 *v1.Pod
	name                string
	operation           string
	failedTimes         int
	ResourceManagerName string
	ResourceManagerType string
	NameSpace           string
}

func NewOperatingPod(pod *v1.Pod, operation string) *operatingPod {
	var resourceManagerName, resourceManagerType string
	if len(pod.GetOwnerReferences()) > 0 {
		resourceManagerType = pod.GetOwnerReferences()[0].Kind
		resourceManagerName = pod.GetOwnerReferences()[0].Name
	}
	return &operatingPod{
		pod:                 pod,
		operation:           operation,
		ResourceManagerName: resourceManagerName,
		ResourceManagerType: resourceManagerType,
		NameSpace:           pod.Namespace,
		name:                pod.Name,
	}

}

func Operate(eventQueue *PodAndEventMap, podNs string, podName string) {
	defer errobj.RecoverPanic()
	key := podNs + podName
	klog.Infof("@@@@Operate start:pod ns is [%v], name is [%v]", podNs, podName)
	var initState string
	for {
		_, err := GetPodService().Get(podNs, podName)
		if errobj.IsNotFoundError(err) {
			initState = constvalue.CreatingState
			break
		}
		if err == nil {
			initState = constvalue.RunningState
			break
		}
		time.Sleep(time.Second * constvalue.GetPodStatusIntervalTime)
	}

	klog.Debugf("@@@@Operate: pod ns is [%v], name is [%v], id is [%v]", podNs, podName)

	stateMachine := newPodStateMachine(eventQueue, key, initState)

	for {
		oPod := eventQueue.Get(key)
		if oPod == nil {
			klog.Warningf("Operate:Opod is nil, should return ")
			return
		}

		klog.Infof("@@@@Operate:Pod and pod name is [%v], ID is [%v], pod status is [%v], operation is [%v], failedTimes:[%v]",
			key,oPod.pod.UID, reflect.TypeOf(stateMachine.state), oPod.operation, oPod.failedTimes)

		if oPod.failedTimes != 0 {
			time.Sleep(GetSleepDurationByTimes(oPod.failedTimes, 30))
		}
		var stop bool
		if oPod.operation == constvalue.CreateOperation {
			stop = stateMachine.create(oPod)

		} else {
			stop = stateMachine.delete(oPod)
		}
		if stop {
			return
		}
	}
}

func GetSleepDurationByTimes(retryTimes int, maxSecond int) time.Duration {
	var sleepSecond int
	//prevent int overflow
	if retryTimes >32 {
		retryTimes = 32
	}
	if int(math.Pow(2, float64(retryTimes))) < maxSecond {
		sleepSecond = int(math.Pow(2, float64(retryTimes)))
	} else {
		sleepSecond = maxSecond
	}
	return time.Second * time.Duration(sleepSecond)

}

//state pattern
type PodStateMachine struct {
	creatingState *CreatingState
	runningState  *RunningState
	podEventMap   *PodAndEventMap
	key           string
	state         State
}

func newPodStateMachine(podEventMap *PodAndEventMap, key string, initState string) *PodStateMachine {
	operater := &PodStateMachine{}
	creatingState := &CreatingState{
		worker: operater,
	}
	runningState := &RunningState{
		worker: operater,
	}
	operater.creatingState = creatingState
	operater.runningState = runningState
	operater.podEventMap = podEventMap
	operater.key = key
	if initState == constvalue.CreatingState {
		operater.state = creatingState
	} else {
		operater.state = runningState
	}
	return operater
}

func (opw *PodStateMachine) create(oPod *operatingPod) bool {
	return opw.state.create(oPod)
}

func (opw *PodStateMachine) delete(oPod *operatingPod) bool {
	return opw.state.delete(oPod)
}

type State interface {
	create(oPod *operatingPod) bool
	delete(oPod *operatingPod) bool
}

type CreatingState struct {
	worker *PodStateMachine
}

func (cs CreatingState) create(Opod *operatingPod) bool {
	klog.Debugf("CreatingState.create:pod Name is [%v]", Opod.pod.Name)
	var err error

	managerInterface, err := ResourceManagerRepo.CreateAndGet(Opod.ResourceManagerType, Opod.NameSpace,
		Opod.ResourceManagerName)
	if err != nil {
		klog.Errorf("ResourceManagerRepo.CreateAndGet(%v) err, error is [%v]", Opod.ResourceManagerType+Opod.NameSpace+
			Opod.ResourceManagerName, err)
		return false
	}
	err = managerInterface.Alloc(Opod)
	if err != nil {
		klog.Errorf("CreatingState.create(opod:[%v]) err, error is [%v]", Opod, err)
		cs.worker.state = cs.worker.creatingState
		Opod.failedTimes++
		return false
	}


	key := Opod.pod.Namespace + Opod.pod.Name
	cs.worker.state = cs.worker.runningState
	stop := cs.worker.podEventMap.CheckEqualAndDelete(key, Opod)
	return stop
}

func (cs CreatingState) delete(Opod *operatingPod) bool {
	klog.Warningf("CreatingState.delete: should not delete, pod Name is [%v]", Opod.pod.Name)
	key := Opod.pod.Namespace + Opod.pod.Name

	return cs.worker.podEventMap.CheckEqualAndDelete(key, Opod)
}

type RunningState struct {
	worker *PodStateMachine
}

func (rs RunningState) create(Opod *operatingPod) bool {
	klog.Warningf("CreatingState.create: should not create,pod Name is [%v]", Opod.pod.Name)
	key := Opod.pod.Namespace + Opod.pod.Name
	for {
		podForDB, err := daos.GetPodDao().Get(Opod.NameSpace, Opod.name)
		if errobj.IsNotFoundError(err){
			klog.Warningf("daos.GetPodDao().Get(Opod.NameSpace:[%v], Opod.name:[%v]) err, error is [%v]", Opod.NameSpace, Opod.name, err)
			return rs.worker.podEventMap.CheckEqualAndDelete(key, Opod)
		}
		if err != nil && errobj.IsNotFoundError(err) {
			klog.Warningf("daos.GetPodDao().Get(Opod.NameSpace:[%v], Opod.name:[%v]) err, error is [%v]", Opod.NameSpace, Opod.name, err)
			time.Sleep(constvalue.EtcdRetryTimes * time.Second)
			continue
		}

		if podForDB.PodID != string(Opod.pod.UID){
			podForDB.PodID = string(Opod.pod.UID)
			err := daos.GetPodDao().Save(podForDB)
			if err != nil {
				klog.Warningf("daos.GetPodDao().Save(podForDB) err, error is [%v]", err)
				time.Sleep(constvalue.EtcdRetryTimes * time.Second)
				continue
			}
		}
		break
	}
	return rs.worker.podEventMap.CheckEqualAndDelete(key, Opod)

}

func (rs RunningState) delete(Opod *operatingPod) bool {
	klog.Infof("RunningState.delete:pod Name is [%v]", Opod.pod.Name)
	var err error

	for {
		podForDB, err := daos.GetPodDao().Get(Opod.NameSpace, Opod.name)
		key := Opod.pod.Namespace + Opod.pod.Name
		if errobj.IsNotFoundError(err){
			klog.Warningf("daos.GetPodDao().Get(Opod.NameSpace:[%v], Opod.name:[%v]) err, error is [%v]",Opod.NameSpace,Opod.name,err)
			return rs.worker.podEventMap.CheckEqualAndDelete(key, Opod)
		}
		if err != nil && !errobj.IsNotFoundError(err){
			klog.Warningf("daos.GetPodDao().Get(Opod.NameSpace:[%v], Opod.name:[%v]) err, error is [%v]",Opod.NameSpace,Opod.name,err)
			time.Sleep(constvalue.EtcdRetryTimes * time.Second)
			continue
		}
		if podForDB.PodID != string(Opod.pod.UID){
			return rs.worker.podEventMap.CheckEqualAndDelete(key, Opod)
		}
		break
	}
	managerInterface, err := ResourceManagerRepo.CreateAndGet(Opod.ResourceManagerType, Opod.NameSpace,
		Opod.ResourceManagerName)
	if err != nil {
		klog.Errorf("ResourceManagerRepo.CreateAndGet(%v) err, error is [%v]", Opod.ResourceManagerType+Opod.NameSpace+
			Opod.ResourceManagerName, err)
		return false
	}
	err = managerInterface.Free(Opod)
	if err != nil {
		klog.Errorf("RunningState.delete(opod:[%v]) err, error is [%v]", Opod, err)
		rs.worker.state = rs.worker.runningState
		Opod.failedTimes++
		return false
	}
	key := Opod.pod.Namespace + Opod.pod.Name
	rs.worker.state = rs.worker.creatingState
	stop := rs.worker.podEventMap.CheckEqualAndDelete(key, Opod)
	return stop
}
