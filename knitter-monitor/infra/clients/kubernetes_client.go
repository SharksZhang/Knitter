package clients

import (
	v1beta1app "k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"github.com/ZTE/Knitter/pkg/klog"
)

var monitorK8sClient = &MonitorK8sClient{}

func GetMonitorK8sClient() *MonitorK8sClient {
	return monitorK8sClient
}

type MonitorK8sClient struct {
}

func (mc *MonitorK8sClient) GetReplicationController(nameSpace string, name string) (*v1.ReplicationController, error) {
	return GetClientset().CoreV1().ReplicationControllers(nameSpace).Get(name, meta_v1.GetOptions{})
}

func (mc *MonitorK8sClient) GetPodsPatchController(nameSpace string, name string, pt types.PatchType, data []byte, subresources ...string) (*v1.Pod, error) {
	return GetClientset().CoreV1().Pods(nameSpace).Patch(name, pt, data)
}

func (mc *MonitorK8sClient) GetReplicationControllerReplicas(namespace string, name string) (int, error) {
	replicaController, err := mc.GetReplicationController(namespace, name)
	if err != nil {
		klog.Errorf("MonitorK8sClient.GetReplicationControllerReplicas: GetReplicationControllerReplicas error, err is [%v]", err)
		return 0, err
	}
	return int(*replicaController.Spec.Replicas), nil
}

func (mc *MonitorK8sClient) GetReplicaSet(nameSpace string, name string) (*v1beta1.ReplicaSet, error) {
	return GetClientset().ExtensionsV1beta1().ReplicaSets(nameSpace).Get(name, meta_v1.GetOptions{})

}

func (mc *MonitorK8sClient) GetReplicaSetReplicas(namespace string, name string) (int, error) {
	replicaController, err := mc.GetReplicaSet(namespace, name)
	if err != nil {
		klog.Errorf("MonitorK8sClient.GetReplicaSetReplicas: GetReplicaSetReplicas error, err is [%v]", err)
		return 0, err
	}
	return int(*replicaController.Spec.Replicas), nil
}

func (mc *MonitorK8sClient) GetStatefulSet(nameSpace string, name string) (*v1beta1app.StatefulSet, error) {
	return GetClientset().AppsV1beta1().StatefulSets(nameSpace).Get(name, meta_v1.GetOptions{})

}

func (mc *MonitorK8sClient) GetStatefulSetReplicas(namespace string, name string) (int, error) {
	replicaController, err := mc.GetStatefulSet(namespace, name)
	if err != nil {
		klog.Errorf("MonitorK8sClient.GetStatefulSetReplicas: GetStatefulSet error, err is [%v]", err)
		return 0, err
	}
	return int(*replicaController.Spec.Replicas), nil
}

func (mc *MonitorK8sClient) IsReplicationControllerExist(namespace string, name string) (bool, error) {
	replicaController, err := mc.GetReplicationController(namespace, name)
	if err != nil && errors.IsNotFound(err) {
		klog.Debugf("GetReplicationController: errors.IsNotFound(err:[%v]):[%v]", err, errors.IsNotFound(err))
		return false, nil
	}
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("MonitorK8sClient.IsReplicationControllerExist: GetReplicationControllerReplicas error, err is [%v]", err)
		return true, err
	}
	klog.Debugf("IsReplicationControllerExist:replicaController:[%+v]", replicaController)
	return true, nil
}

func (mc *MonitorK8sClient) IsStatefulSetExist(namespace string, name string) (bool, error) {
	statefulSet, err := mc.GetStatefulSet(namespace, name)
	if err != nil && errors.IsNotFound(err) {
		klog.Infof("errors.IsNotFound(err:[%v]):[%v]", err, errors.IsNotFound(err))
		return false, nil
	}
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("mc.GetStatefulSet(namespace, name) error, err is [%v]", err)
		return true, err
	}
	klog.Debugf("IsStatefulSetExist:statefulSet:[%+v]", statefulSet)
	return true, nil
}

func (mc *MonitorK8sClient) IsReplicaSetExist(namespace string, name string) (bool, error) {
	replicaSet, err := mc.GetReplicaSet(namespace, name)
	if err != nil && errors.IsNotFound(err) {
		klog.Infof("errors.IsNotFound(err:[%v]):[%v]", err, errors.IsNotFound(err))
		return false, nil
	}
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("mc.GetReplicaSet(namespace, name) error, err is [%v]", err)
		return true, err
	}
	klog.Debugf("IsReplicaSetExist:replicaSet:[%+v]", replicaSet)
	return true, nil
}
