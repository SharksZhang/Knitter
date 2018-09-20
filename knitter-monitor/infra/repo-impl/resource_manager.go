package repo_impl

import (
	"encoding/json"
	"strings"
	"sync"
	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/domain/services"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/knitter-monitor/infra/db-config"
	"k8s.io/client-go/tools/cache"
	"github.com/ZTE/Knitter/pkg/klog"
)

var resourceManagerRepoCache *ResourceManagerRepoCache
var resourceManagerRepoETCD *ResourceManagerRepoETCD

func init() {
	resourceManagerRepoCache = &ResourceManagerRepoCache{}
	resourceManagerRepoCache.init()
	resourceManagerRepoETCD = &ResourceManagerRepoETCD{}

}

type ResourceManagerRepository struct {
}

func (rm *ResourceManagerRepository) CreateAndGet(resourceManagerType, namespace, resourceManagerName string) (services.ResourceManagerInterface, error) {
	return resourceManagerRepoCache.CreateAndGet(resourceManagerType, namespace, resourceManagerName)
}

func (*ResourceManagerRepository) GetAll() []services.ResourceManagerInterface {
	return resourceManagerRepoCache.List()
}

func (*ResourceManagerRepository) Save(rm services.ResourceManagerInterface) error {
	err := resourceManagerRepoETCD.Save(rm)
	if err != nil {
		klog.Errorf("resourceManagerRepoETCD.Save(rm:[%v]) err, error is [%v]", rm, err)
		return err
	}
	err = resourceManagerRepoCache.indexer.Add(rm)
	if err != nil {
		klog.Errorf("resourceManagerRepoCache.indexer.Add(rm:[%v]) err, error is [%v]", rm, err)
		return err
	}
	klog.Infof("ResourceManagerRepository.Save success")
	return nil
}

func (*ResourceManagerRepository) Delete(rm services.ResourceManagerInterface) error {
	err := resourceManagerRepoETCD.Delete(rm)
	if err != nil {
		klog.Errorf("resourceManagerRepoETCD.Delete(rm:[%v]) err, error is [%v]", rm, err)
		return err
	}
	return resourceManagerRepoCache.delete(rm.GetKey())
}

type ResourceManagerRepoCache struct {
	indexer cache.Indexer
	locker  sync.RWMutex
}

func LoadResourceManagerRepoCache() error {
	klog.Infof("LoadResourceManagerRepoCache start")

	resourceManagers, err := resourceManagerRepoETCD.GetAll()
	if err != nil {
		klog.Errorf("LoadResourceManagerRepoCache: ResourceManagerRepo.GetAll() err, error is [%v]", err)
		return err
	}
	klog.Infof("resourceManagers:[%v]", resourceManagers)
	for _, resourceManager := range resourceManagers {
		err := resourceManagerRepoCache.add(resourceManager)
		if err != nil {
			klog.Warningf("resourceManagerRepoCache.add(resourceManager:[%v]) err, error is [%v]", err)
		}
	}
	klog.Infof("LoadResourceManagerRepoCache successfully")
	return nil
}

func ResourceManagerKeyFunc(obj interface{}) (string, error) {
	if obj == nil {
		klog.Error("ResourceManagerKeyFunc: obj arg is nil")
		return "", errobj.ErrObjectPointerIsNil
	}

	resourceManager, ok := obj.(services.ResourceManagerInterface)
	if !ok {
		klog.Error("ResourceManagerKeyFunc: obj arg is not type: *services.ResourceManagerInterface")
		return "", errobj.ErrArgTypeMismatch
	}

	return resourceManager.GetKey(), nil
}

func (rc *ResourceManagerRepoCache) init() {
	indexers := cache.Indexers{}
	rc.indexer = cache.NewIndexer(ResourceManagerKeyFunc, indexers)
}

func (rc *ResourceManagerRepoCache) add(managerInterface services.ResourceManagerInterface) error {
	rc.locker.Lock()
	defer rc.locker.Unlock()
	err := rc.indexer.Add(managerInterface)
	if err != nil {
		klog.Errorf("ResourceManagerRepoCache.Add: Add obj[%v] to repo FAILED, error: %v", managerInterface, err)
		return err
	}
	klog.Infof("ResourceManagerRepoCache.Add: Add obj[%v] to repo success", managerInterface)
	return nil
}

func (rc *ResourceManagerRepoCache) delete(ID string) error {
	rc.locker.Lock()
	defer rc.locker.Unlock()
	err := rc.indexer.Delete(ID)
	if err != nil {
		klog.Errorf("ResourceManagerRepoCache.delete: delete obj[%v] to repo FAILED, error: %v", ID, err)
		return err
	}
	klog.Infof("ResourceManagerRepoCache.delete: delete obj[%v] to repo success", ID)
	return nil
}

func (rc *ResourceManagerRepoCache) CreateAndGet(resourceManagerType, namespace, resourceManagerName string) (services.ResourceManagerInterface, error) {
	rc.locker.Lock()
	defer rc.locker.Unlock()
	ID := resourceManagerType + namespace + resourceManagerName
	item, exists, err := rc.indexer.GetByKey(ID)
	if err != nil {
		klog.Errorf("ResourceManagerRepoCache.CreateAndGet: ID[%s]'s object FAILED, error: %v", ID, err)
		return nil, err
	}
	if !exists {
		klog.Warningf("ResourceManagerRepoCache.CreateAndGet: ID[%s]'s object not found, init resource manager", ID)
		resourceManager := services.ResourceManagerFactory(resourceManagerType, namespace, resourceManagerName)
		rc.indexer.Add(resourceManager)
		return resourceManager, nil
	}

	resourceManager, ok := item.(services.ResourceManagerInterface)
	if !ok {
		klog.Errorf("ResourceManagerRepoCache.CreateAndGet: ID[%s]'s object[%v] type not match ResourceManagerInterface", ID, item)
		return nil, errobj.ErrObjectTypeMismatch
	}
	klog.Infof("ResourceManagerRepoCache.CreateAndGet: ID[%s]'s object[%v] success", ID, resourceManager)
	return resourceManager, nil
}

func (rc *ResourceManagerRepoCache) List() []services.ResourceManagerInterface {
	rc.locker.RLock()
	defer rc.locker.RUnlock()
	objs := rc.indexer.List()
	resourceManagers := make([]services.ResourceManagerInterface, 0)

	for _, obj := range objs {
		resourceManager, ok := obj.(services.ResourceManagerInterface)
		if !ok {
			klog.Errorf("NetworkObjectRepo.List: List result object: %v is not type *NetworkObject, skip", obj)
			continue
		}
		resourceManagers = append(resourceManagers, resourceManager)
	}
	return resourceManagers
}

type ResourceManagerRepoETCD struct {
}

func (*ResourceManagerRepoETCD) Save(rm services.ResourceManagerInterface) error {
	key := dbconfig.GetKeyOfResourceManager(rm.GetKey())
	value, err := json.Marshal(rm)
	if err != nil {
		klog.Errorf("json.Marshal(rm:[%v]) err, error is [%v]", err)
		return err
	}
	err = dbconfig.GetDataBase().SaveLeaf(key, string(value))
	if err != nil {
		klog.Errorf(" dbconfig.GetDataBase().SaveLeaf(key[%v], string(value:[%v])) err, error is [%v]", key, string(value), err)
		return err
	}
	return nil
}

func (*ResourceManagerRepoETCD) Delete(rm services.ResourceManagerInterface) error {
	key := dbconfig.GetKeyOfResourceManager(rm.GetKey())

	err := dbconfig.GetDataBase().DeleteLeaf(key)
	if err != nil && !errobj.IsNotFoundError(err) {
		klog.Errorf(" dbconfig.GetDataBase().SaveLeaf(Key[%v], string(value:[%v])) err, error is [%v]", key, err)
		return err
	}
	klog.Infof("ResourceManagerRepoETCD.Delete: key is [%v] to repo success", rm.GetKey())

	return nil
}

func (*ResourceManagerRepoETCD) GetAll() ([]services.ResourceManagerInterface, error) {
	key := dbconfig.GetKeyOfResourceManagers()
	resourceManagers := make([]services.ResourceManagerInterface, 0)
	nodes, err := dbconfig.GetDataBase().ReadDir(key)
	if err != nil {
		klog.Errorf("common.GetDataBase().ReadDir(Key:[%v]) FAILED, error is %v", key, err)
		return nil, err
	}
	for _, node := range nodes {
		klog.Infof("etcd key is :[%v], value is [%v] ", node.Key, node.Value)
		if strings.Index(strings.ToLower(node.Key), constvalue.TypeReplicationController) > -1 {
			resourceManager := &services.ReplicationControllersResourceManager{
				ResourceManager:          &services.ResourceManager{},
				DuplicateResourceManager: &services.DuplicateResourceManager{},
			}
			err = json.Unmarshal([]byte(node.Value), resourceManager)
			klog.Infof("ReplicationControllersResourceManager.ResourceManager:[%v], DuplicateResourceManager:[%v]",
				resourceManager.ResourceManager, resourceManager.DuplicateResourceManager)
			for _, ports := range resourceManager.DuplicateResourceManager.UnusedResource {
				for _, port := range ports {
					klog.Infof("port:[%v]", port)

				}
			}
			if err != nil {
				klog.Errorf("json.Unmarshal([]byte(node.Value), resourceManager) FAILED, error is %v", node.Value, err)
				return nil, err
			}
			klog.Infof("node.value:[%v]", node.Value)
			klog.Infof("resourceManager:[%+v]", resourceManager)
			resourceManagers = append(resourceManagers, resourceManager)
		} else if strings.Index(strings.ToLower(node.Key), constvalue.TypeReplicaSet) > -1 {
			resourceManager := &services.ReplicaSetResourceManager{
				ResourceManager:          &services.ResourceManager{},
				DuplicateResourceManager: &services.DuplicateResourceManager{},
			}
			err = json.Unmarshal([]byte(node.Value), resourceManager)
			klog.Infof("resourceManager.ResourceManager:[%v], DuplicateResourceManager:[%v]",
				resourceManager.ResourceManager, resourceManager.DuplicateResourceManager)
			if err != nil {
				klog.Errorf("json.Unmarshal([]byte(node.Value), TypeReplicaSet) FAILED, error is %v", node.Value, err)
				return nil, err
			}
			resourceManagers = append(resourceManagers, resourceManager)

		} else if strings.Index(strings.ToLower(node.Key), constvalue.TypeStatefulSet) > -1 {
			resourceManager := &services.StatefulSetResourceManager{
				ResourceManager:          &services.ResourceManager{},
				DuplicateResourceManager: &services.DuplicateResourceManager{},
			}
			err = json.Unmarshal([]byte(node.Value), resourceManager)
			klog.Infof("resourceManager.ResourceManager:[%v], DuplicateResourceManager:[%v]",
				resourceManager.ResourceManager, resourceManager.DuplicateResourceManager)
			if err != nil {
				klog.Errorf("json.Unmarshal([]byte(node.Value), StatefulSetResourceManager) FAILED, error is %v", node.Value, err)
				return nil, err
			}
			resourceManagers = append(resourceManagers, resourceManager)

		} else {
			resourceManager := &services.DefaultResourceManager{
				ResourceManager: &services.ResourceManager{},
			}
			err = json.Unmarshal([]byte(node.Value), resourceManager)
			klog.Infof("resourceManager.ResourceManager:[%v],",
				resourceManager.ResourceManager)
			if err != nil {
				klog.Errorf("json.Unmarshal([]byte(node.Value), StatefulSetResourceManager) FAILED, error is %v", node.Value, err)
				return nil, err
			}
			resourceManagers = append(resourceManagers, resourceManager)
		}

	}
	return resourceManagers, nil
}
