package services

import (
	"github.com/bouk/monkey"
	"github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/util/workqueue"
	//"github.com/golang/mock/gomock"
	"errors"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"testing"
	"time"
	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/infra/clients"
)

func TestNewCreatePortForPodController(t *testing.T) {
	monkey.Patch(clients.GetClientset, func() *kubernetes.Clientset {
		return &kubernetes.Clientset{}
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestNewCreatePortForPodController", t, func() {
		_, err := NewCreatePortsController()
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestNewCreatePortForPodControllerKubernetesNilFail(t *testing.T) {
	convey.Convey("TestNewCreatePortForPodControllerKubernetesNilFail", t, func() {
		_, err := NewCreatePortsController()
		convey.So(err.Error(), convey.ShouldEqual, "kubernetes clientSet is nil")
	})
}

func TestEnqueueCreatePod(t *testing.T) {
	controller := &CreatePortsController{}
	controller.PodEventMap = &PodAndEventMap{
		Event: make(map[string]*operatingPod),
	}

	objectMeta := metav1.ObjectMeta{
		Name:      "pod1",
		Namespace: "admin",
		UID:       "testpod-1111",
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}
	pod := &operatingPod{
		pod:                 k8sPod,
		name:                k8sPod.GetName(),
		operation:           constvalue.CreateOperation,
		ResourceManagerName: "",
		ResourceManagerType: "",
		NameSpace:           k8sPod.GetNamespace(),
	}

	monkey.Patch(Operate, func(eventQueue *PodAndEventMap, podNs string, podName string) {
		return
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestEnqueueCreatePod", t, func() {
		controller.enqueueCreatePod(k8sPod)
		key := "adminpod1"
		oPod := controller.PodEventMap.Get(key)
		convey.So(oPod, convey.ShouldResemble, pod)
	})

}

//func TestEnqueueDeletePod(t *testing.T) {
//	controller := &CreatePortsController{}
//	controller.PodEventMap = &PodAndEventMap{
//		Event: make(map[string]*operatingPod),
//	}
//	objectMeta := metav1.ObjectMeta{
//		Name:      "pod1",
//		UID:       "testpod-1111",
//		Namespace: "admin",
//	}
//	k8sPod := &v1.Pod{
//		ObjectMeta: objectMeta,
//	}
//	pod := &operatingPod{
//		pod:                 k8sPod,
//		name:                k8sPod.GetName(),
//		operation:           constvalue.DeleteOperation,
//		NameSpace:           k8sPod.GetNamespace(),
//	}
//
//	monkey.Patch(Operate, func(eventQueue *PodAndEventMap, podNs string, podName string) {
//		return
//	})
//
//	defer monkey.UnpatchAll()
//
//	convey.Convey("TestEnqueueDeletePod", t, func() {
//		controller.enqueueDeletePod(k8sPod)
//		key := "adminpod1"
//
//		oPod := controller.PodEventMap.Get(key)
//
//		convey.So(oPod, convey.ShouldResemble, pod)
//	})
//
//}

func TestPodAndEventMap(t *testing.T) {
	pe := PodAndEventMap{
		Event: make(map[string]*operatingPod),
	}
	objectMeta := metav1.ObjectMeta{
		Name:      "pod1",
		UID:       "testpod-1111",
		Namespace: "admin",
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}
	pod1 := &operatingPod{
		operation: constvalue.DeleteOperation,
		pod:       k8sPod,
	}

	objectMeta2 := metav1.ObjectMeta{
		Name:      "pod1",
		UID:       "testpod-111",
		Namespace: "admin",
	}
	k8sPod2 := &v1.Pod{
		ObjectMeta: objectMeta2,
	}
	pod2 := &operatingPod{
		operation: constvalue.CreateOperation,
		pod:       k8sPod2,
	}
	key1 := objectMeta.Namespace + objectMeta.Name
	convey.Convey("TestPodAndEventMap", t, func() {
		pe.add(key1, pod1)
		pod1Expect := pe.Get(key1)
		convey.So(pod1Expect, convey.ShouldResemble, pod1)
		pe.delete(key1)
		pod1Expect = pe.Get(key1)
		convey.So(pod1Expect, convey.ShouldBeNil)

		exist := pe.CheckExistAndAdd(key1, pod1)
		convey.So(exist, convey.ShouldBeFalse)
		exist = pe.CheckExistAndAdd(key1, pod1)
		convey.So(exist, convey.ShouldBeTrue)
		equal := pe.CheckEqualAndDelete(key1, pod1)
		pod1Expect = pe.Get(key1)
		convey.So(equal, convey.ShouldBeTrue)
		convey.So(pod1Expect, convey.ShouldBeNil)

		pe.add(key1, pod2)
		equal = pe.CheckEqualAndDelete(key1, pod1)
		pod1Expect = pe.Get(key1)
		convey.So(equal, convey.ShouldBeFalse)
		convey.So(pod1Expect, convey.ShouldResemble, pod2)

	})
}

func TestOperateRunningState1(t *testing.T) {

	pe := &PodAndEventMap{
		Event: make(map[string]*operatingPod),
	}
	objectMeta := metav1.ObjectMeta{
		Name:      "pod1",
		UID:       "testpod-1111",
		Namespace: "admin",
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}
	pod1 := &operatingPod{
		operation: constvalue.DeleteOperation,
		pod:       k8sPod,
	}

	objectMeta2 := metav1.ObjectMeta{
		Name:      "pod2",
		UID:       "testpod-222",
		Namespace: "admin",
	}
	k8sPod2 := &v1.Pod{
		ObjectMeta: objectMeta2,
	}
	pod2 := &operatingPod{
		operation: constvalue.CreateOperation,
		pod:       k8sPod2,
	}
	key1 := objectMeta.Namespace + objectMeta.Name
	key2 := objectMeta2.Namespace + objectMeta2.Name

	pe.add(key1, pod1)
	pe.add(key2, pod2)
	var podservice *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "Get", func(_ *podService, podNs, podName string) (*Pod, error) {
		return nil, nil
	})

	monkey.Patch(deleteResource, func(oPod *operatingPod) error {

		return nil
	})
	monkey.Patch(createResource, func(oPod *operatingPod) error {
		return nil
	})
	defer monkey.UnpatchAll()
	Operate(pe, objectMeta.Namespace, objectMeta.Name)
}

func TestOperateRunningState(t *testing.T) {

	pe := &PodAndEventMap{
		Event: make(map[string]*operatingPod),
	}
	objectMeta := metav1.ObjectMeta{
		Name:      "pod1",
		UID:       "testpod-1111",
		Namespace: "admin",
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}
	pod1 := &operatingPod{
		operation: constvalue.DeleteOperation,
		pod:       k8sPod,
	}

	objectMeta2 := metav1.ObjectMeta{
		Name:      "pod2",
		UID:       "testpod-222",
		Namespace: "admin",
	}
	k8sPod2 := &v1.Pod{
		ObjectMeta: objectMeta2,
	}
	pod2 := &operatingPod{
		operation: constvalue.CreateOperation,
		pod:       k8sPod2,
	}
	key1 := objectMeta.Namespace + objectMeta.Name
	key2 := objectMeta2.Namespace + objectMeta2.Name

	pe.add(key1, pod1)
	pe.add(key2, pod2)
	var podservice *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "Get", func(_ *podService, podNs, podName string) (*Pod, error) {
		return nil, nil
	})

	monkey.Patch(deleteResource, func(oPod *operatingPod) error {
		return nil
	})

	monkey.Patch(createResource, func(oPod *operatingPod) error {
		return nil
	})
	defer monkey.UnpatchAll()
	Operate(pe, objectMeta2.Namespace, objectMeta2.Name)
}

func TestOperateCreatingState1(t *testing.T) {

	pe := &PodAndEventMap{
		Event: make(map[string]*operatingPod),
	}
	objectMeta := metav1.ObjectMeta{
		Name:      "pod1",
		UID:       "testpod-1111",
		Namespace: "admin",
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}
	pod1 := &operatingPod{
		operation: constvalue.DeleteOperation,
		pod:       k8sPod,
	}

	objectMeta2 := metav1.ObjectMeta{
		Name:      "pod2",
		UID:       "testpod-222",
		Namespace: "admin",
	}
	k8sPod2 := &v1.Pod{
		ObjectMeta: objectMeta2,
	}
	pod2 := &operatingPod{
		operation: constvalue.CreateOperation,
		pod:       k8sPod2,
	}
	key1 := objectMeta.Namespace + objectMeta.Name
	key2 := objectMeta2.Namespace + objectMeta2.Name

	pe.add(key1, pod1)
	pe.add(key2, pod2)
	var podservice *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "Get", func(_ *podService, podNs, podName string) (*Pod, error) {
		return nil, errors.New("Key not found")
	})

	monkey.Patch(deleteResource, func(oPod *operatingPod) error {
		return nil
	})

	monkey.Patch(createResource, func(oPod *operatingPod) error {
		return nil
	})
	defer monkey.UnpatchAll()
	Operate(pe, objectMeta.Namespace, objectMeta.Name)
}

func TestOperateCreatingState(t *testing.T) {

	pe := &PodAndEventMap{
		Event: make(map[string]*operatingPod),
	}
	objectMeta := metav1.ObjectMeta{
		Name:      "pod1",
		UID:       "testpod-1111",
		Namespace: "admin",
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}
	pod1 := &operatingPod{
		operation: constvalue.DeleteOperation,
		pod:       k8sPod,
	}

	objectMeta2 := metav1.ObjectMeta{
		Name:      "pod2",
		UID:       "testpod-222",
		Namespace: "admin",
	}
	k8sPod2 := &v1.Pod{
		ObjectMeta: objectMeta2,
	}
	pod2 := &operatingPod{
		operation: constvalue.CreateOperation,
		pod:       k8sPod2,
	}
	key1 := objectMeta.Namespace + objectMeta.Name
	key2 := objectMeta2.Namespace + objectMeta2.Name

	pe.add(key1, pod1)
	pe.add(key2, pod2)
	var podservice *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "Get", func(_ *podService, podNs, podName string) (*Pod, error) {
		return nil, errors.New("Key not found")
	})

	monkey.Patch(deleteResource, func(oPod *operatingPod) error {
		return nil
	})
	monkey.Patch(createResource, func(oPod *operatingPod) error {
		return nil
	})
	defer monkey.UnpatchAll()
	Operate(pe, objectMeta2.Namespace, objectMeta2.Name)
}

func TestGetSleepDurationByTimes(t *testing.T) {
	convey.Convey("TestGetSleepDurationByTimes", t, func() {
		sleepDuration := GetSleepDurationByTimes(2, 30)
		convey.So(time.Second*time.Duration(4), convey.ShouldResemble, sleepDuration)
		sleepDuration = GetSleepDurationByTimes(5, 30)
		convey.So(time.Second*time.Duration(30), convey.ShouldResemble, sleepDuration)

		sleepDuration = GetSleepDurationByTimes(62, 30)
		convey.So(time.Second*time.Duration(30), convey.ShouldResemble, sleepDuration)

	})
}
