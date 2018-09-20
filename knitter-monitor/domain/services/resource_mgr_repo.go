package services

var ResourceManagerRepo ResourceManagerRepositoryInterface

type ResourceManagerRepositoryInterface interface {
	Save(manager ResourceManagerInterface) error
	Delete(manager ResourceManagerInterface) error
	GetAll() []ResourceManagerInterface
	CreateAndGet(resourceManagerType, namespace, resourceManagerName string) (ResourceManagerInterface, error)
}
