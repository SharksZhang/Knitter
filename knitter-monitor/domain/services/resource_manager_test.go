package services

import (
	"errors"
	"github.com/bouk/monkey"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/knitter-monitor/infra/clients"
)

func TestDefaultResourceManager(t *testing.T) {
	convey.Convey("TestDefaultResourceManager", t, func() {
		var rm *ResourceManagerInterface
		rmType := "replicationcontrollers"
		namespace := "admin"
		name := "replication-11111"

		rcrm := NewDefaultResourceManager(rmType, namespace, name)
		rcrmExpect := &DefaultResourceManager{
			&ResourceManager{
				ResourceManagerType: rmType,
				Name:                name,
				Namespace:           namespace,
				Key:                 rmType + namespace + name},
		}
		convey.So(rcrm, convey.ShouldImplement, rm)
		convey.So(rcrm, convey.ShouldResemble, rcrmExpect)
	})

}

func TestResourceManager_Alloc(t *testing.T) {

	_, pod := createPod("pod1", "admin")
	opod := &operatingPod{
		pod: &v1.Pod{},
	}
	var podservice *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "NewPodFromK8sPod", func(_ *podService, k8sPod *v1.Pod) (*Pod, error) {
		return pod, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "Save", func(_ *podService, pod *Pod) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "ReportPod", func(_ *podService, reportPod *clients.ReportPod) error {
		return nil
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestResourceManager_Alloc", t, func() {
		err := (&DefaultResourceManager{}).Alloc(opod)
		convey.So(err, convey.ShouldBeNil)
	})

}

func TestResourceManager_Alloc_NewPodFromK8sPodFail(t *testing.T) {

	_, pod := createPod("pod1", "admin")
	opod := &operatingPod{
		pod: &v1.Pod{},
	}
	var podservice *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "NewPodFromK8sPod", func(_ *podService, k8sPod *v1.Pod) (*Pod, error) {
		return pod, errors.New("new error")
	})

	defer monkey.UnpatchAll()

	convey.Convey("TestResourceManager_Alloc_NewPodFromK8sPodFail", t, func() {
		err := (&DefaultResourceManager{}).Alloc(opod)
		convey.So(err.Error(), convey.ShouldEqual, "new error")
	})

}

func TestResourceManager_Alloc_SaveFail(t *testing.T) {

	_, pod := createPod("pod1", "admin")
	opod := &operatingPod{
		pod: &v1.Pod{},
	}
	var podservice *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "NewPodFromK8sPod", func(_ *podService, k8sPod *v1.Pod) (*Pod, error) {
		return pod, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "Save", func(_ *podService, pod *Pod) error {
		return errors.New("save error")
	})

	defer monkey.UnpatchAll()

	convey.Convey("TestResourceManager_Alloc", t, func() {
		err := (&DefaultResourceManager{}).Alloc(opod)
		convey.So(err.Error(), convey.ShouldEqual, "save error")
	})

}

func TestResourceManager_Alloc_ReportPodFail(t *testing.T) {

	_, pod := createPod("pod1", "admin")
	opod := &operatingPod{
		pod: &v1.Pod{},
	}
	var podservice *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "NewPodFromK8sPod", func(_ *podService, k8sPod *v1.Pod) (*Pod, error) {
		return pod, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "Save", func(_ *podService, pod *Pod) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "ReportPod", func(_ *podService, reportPod *clients.ReportPod) error {
		return errors.New("ReportPod err")
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestResourceManager_Alloc", t, func() {
		err := (&DefaultResourceManager{}).Alloc(opod)
		convey.So(err, convey.ShouldBeNil)
	})

}

func TestResourceManager_Free(t *testing.T) {
	objectMeta := metav1.ObjectMeta{
		Name:      "pod1",
		UID:       "testpod-1111",
		Namespace: "admin",
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}
	oPod := &operatingPod{
		operation: constvalue.DeleteOperation,
		pod:       k8sPod,
	}
	var podservice *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "DeletePodAndPorts", func(_ *podService, podNs, podName string) error {
		return nil
	})
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "ReportDeletePod", func(_ *podService, podNs, podName string) error {
		return nil
	})
	convey.Convey("TestResourceManager_Free", t, func() {
		err := (&DefaultResourceManager{}).Free(oPod)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestResourceManager_Free_DeletePodAndPortsFail(t *testing.T) {
	objectMeta := metav1.ObjectMeta{
		Name:      "pod1",
		UID:       "testpod-1111",
		Namespace: "admin",
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}
	oPod := &operatingPod{
		operation: constvalue.DeleteOperation,
		pod:       k8sPod,
	}
	var podservice *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "DeletePodAndPorts", func(_ *podService, podNs, podName string) error {
		return errors.New("DeletePodAndPorts err")
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestResourceManager_Free_DeletePodAndPortsFail", t, func() {
		err := (&DefaultResourceManager{}).Free(oPod)
		convey.So(err.Error(), convey.ShouldEqual, "DeletePodAndPorts err")
	})
}

func TestResourceManager_Free_ReportDeletePodFail(t *testing.T) {
	objectMeta := metav1.ObjectMeta{
		Name:      "pod1",
		UID:       "testpod-1111",
		Namespace: "admin",
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}
	oPod := &operatingPod{
		operation: constvalue.DeleteOperation,
		pod:       k8sPod,
	}
	var podservice *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "DeletePodAndPorts", func(_ *podService, podNs, podName string) error {
		return nil
	})
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "ReportDeletePod", func(_ *podService, podNs, podName string) error {
		return errors.New("ReportDeletePod err")
	})
	convey.Convey("TestResourceManager_Free_ReportDeletePodFail", t, func() {
		err := (&DefaultResourceManager{}).Free(oPod)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestReplicationControllersResourceManager(t *testing.T) {
	convey.Convey("TestNewReplicationControllersResourceManager", t, func() {
		var rm *ResourceManagerInterface
		rmType := constvalue.TypeReplicationController
		namespace := "admin"
		name := "replication-11111"

		rcrm := NewReplicationControllersResourceManager(rmType, namespace, name)
		rcrmExpect := &ReplicationControllersResourceManager{
			ResourceManager: &ResourceManager{
				ResourceManagerType: rmType,
				Name:                name,
				Namespace:           namespace,
				Key:                 rmType + namespace + name,
			},
			DuplicateResourceManager: &DuplicateResourceManager{
				LocalResource:       0,
				UnusedResource:      make([][]*Port, 0),
				UsedResourcePodName: make([]string, 0),
			},
		}
		convey.So(rcrm, convey.ShouldImplement, rm)
		convey.So(rcrm, convey.ShouldResemble, rcrmExpect)
	})

}

func TestResourceManagerFactory(t *testing.T) {
	convey.Convey("TestResourceManagerFactory", t, func() {
		namespace := "admin"
		name := "111"
		convey.Convey("TestResourceManagerFactory_default", func() {
			resourceManagerType := ""
			rm := ResourceManagerFactory(resourceManagerType, namespace, name)
			var rmi *ResourceManagerInterface
			convey.So(rm, convey.ShouldImplement, rmi)
			convey.So(rm, convey.ShouldHaveSameTypeAs, &DefaultResourceManager{})
			convey.So(rm, convey.ShouldResemble, &DefaultResourceManager{
				&ResourceManager{
					ResourceManagerType: resourceManagerType,
					Name:                name,
					Namespace:           namespace,
					Key:                 resourceManagerType + namespace + name,
				},
			})
		})
		convey.Convey("TestResourceManagerFactory_ReplicationControllers", func() {
			ressourceManagerType := constvalue.TypeReplicationController
			rm := ResourceManagerFactory(ressourceManagerType, namespace, name)
			convey.So(rm, convey.ShouldHaveSameTypeAs, &ReplicationControllersResourceManager{
				ResourceManager: &ResourceManager{
					ResourceManagerType: ressourceManagerType,
					Name:                name,
					Namespace:           namespace,
					Key:                 ressourceManagerType + namespace + name},
			})
		})
	})
}

func TestReplicationControllersResourceManager_GetReplica(t *testing.T) {
	convey.Convey("TestReplicationControllersResourceManager_GetReplica", t, func() {
		rc := NewReplicationControllersResourceManager(constvalue.TypeReplicationController, "admin", "111")
		var mc *MonitorK8sClient
		monkey.PatchInstanceMethod(reflect.TypeOf(mc), "GetReplicationControllerReplicas", func(_ *MonitorK8sClient, namespace string,
			name string) (int, error) {
			return 6, nil
		})
		defer monkey.UnpatchAll()
		replica, err := rc.GetReplica()
		convey.So(err, convey.ShouldBeNil)
		convey.So(replica, convey.ShouldEqual, 6)

	})

}

func TestReplicationControllersResourceManager_GetReplica_Fail(t *testing.T) {
	convey.Convey("TestReplicationControllersResourceManager_GetReplica", t, func() {
		rc := NewReplicationControllersResourceManager(constvalue.TypeReplicationController, "admin", "111")
		var mc *MonitorK8sClient
		monkey.PatchInstanceMethod(reflect.TypeOf(mc), "GetReplicationControllerReplicas", func(_ *MonitorK8sClient, namespace string,
			name string) (int, error) {
			return -1, errors.New("get err")
		})
		defer monkey.UnpatchAll()
		replica, err := rc.GetReplica()
		convey.So(err, convey.ShouldResemble, errors.New("get err"))
		convey.So(replica, convey.ShouldEqual, -1)

	})

}

/*func TestReplicationControllersResourceManager_Alloc(t *testing.T) {

	convey.Convey("replica > LocalResource and createResource succ", t, func() {
		monkey.Patch(createResource, func(oPod *operatingPod) error {
			return nil
		})
		defer monkey.UnpatchAll()
		rc := NewReplicationControllersResourceManager(constvalue.TypeReplicationController, "admin", "1111")
		monkey.PatchInstanceMethod(reflect.TypeOf(rc), "GetReplica", func(_ *ReplicationControllersResourceManager) (int, error) {
			return 1, nil
		})
		monkey.Patch(SaveResourceManagerToETCD, func(rm ResourceManagerInterface) error  {
			return nil
		})
		podId := "pod1"
		opod := &operatingPod{
			pod: &v1.Pod{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					UID:       types.UID(podId),
				},
			},
		}
		err := rc.Alloc(opod)
		convey.So(rc.UnusedResource, convey.ShouldHaveSameTypeAs, [][]*Port{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(rc.LocalResource, convey.ShouldEqual, 1)
		convey.So(podId, convey.ShouldBeIn, rc.UsedResourcePodName)
	})
}
*/
func TestReplicationControllersResourceManager_AllocReplica(t *testing.T) {
	convey.Convey("Replica<localResource_unused==0", t, func() {
		monkey.Patch(createResource, func(oPod *operatingPod) error {
			return nil
		})
		defer monkey.UnpatchAll()
		rc := NewReplicationControllersResourceManager(constvalue.TypeReplicationController, "admin", "1111")
		monkey.PatchInstanceMethod(reflect.TypeOf(rc), "GetReplica", func(_ *ReplicationControllersResourceManager) (int, error) {
			return 1, nil
		})
		rc.SetLocalResource(6)
		err := rc.Alloc(nil)
		convey.So(err, convey.ShouldResemble, errobj.ErrResourceInUse)
	})
}

func TestReplicationControllersResourceManager_AllocErr(t *testing.T) {
	convey.Convey("replica>LocalResource Alloc_Err", t, func() {
		monkey.Patch(createResource, func(oPod *operatingPod) error {
			return errors.New("createResource err")
		})
		defer monkey.UnpatchAll()
		rc := NewReplicationControllersResourceManager(constvalue.TypeReplicationController, "admin", "1111")
		monkey.PatchInstanceMethod(reflect.TypeOf(rc), "GetReplica", func(_ *ReplicationControllersResourceManager) (int, error) {
			return 1, nil
		})
		err := rc.Alloc(nil)
		convey.So(rc.LocalResource, convey.ShouldEqual, 0)
		convey.So(err, convey.ShouldResemble, errors.New("createResource err"))
	})
}

/*func TestReplicationControllersResourceManager_AllocSucc2(t *testing.T) {
	convey.Convey("Replica<local and Resource_unused>0", t, func() {
		podName := "pod1"
		podId := "1111"
		namespace := "admin"
		rc := NewReplicationControllersResourceManager(constvalue.TypeReplicationController, namespace, podName)
		monkey.PatchInstanceMethod(reflect.TypeOf(rc), "GetReplica", func(_ *ReplicationControllersResourceManager) (int, error) {
			return 1, nil
		})
		monkey.Patch(SaveResourceManagerToETCD, func(rm ResourceManagerInterface) error {
			return nil
		})
		var podservice *podService
		monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "ReportPod", func(_ *podService, reportPod *infra.ReportPod) error {
			return nil
		})
		defer monkey.UnpatchAll()
		rc.LocalResource = 6
		ports := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName: podId,
				},
				LazyAttr: PortLazyAttr{
					ID: "11111",
				},
			},
		}
		rc.UnusedResource = [][]*Port{ports}

		opod := &operatingPod{
			pod: &v1.Pod{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
					UID:       types.UID(podId),
				},
			},
		}

		podExpect := &Pod{
			PodName:  podName,
			TenantId: namespace,
			PodNs:    namespace,
			PodID:    podId,
			Ports:    ports,
			IsSuccessful:true,
		}
		var service *podService
		monkey.PatchInstanceMethod(reflect.TypeOf(service), "Save", func(_ *podService, pod *Pod) error {
			if convey.ShouldResemble(pod, podExpect) == "" {
				return nil
			}
			return errors.New("unExpect pod err")
		})

		err := rc.Alloc(opod)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(rc.UnusedResource), convey.ShouldEqual, 0)
	})
}
*/
/*func TestReplicationControllersResourceManager_Free(t *testing.T) {
	convey.Convey("GetReplicaErr", t, func() {
		podName := "pod1"
		namespace := "admin"
		podId := "pod1"
		opod := &operatingPod{
			pod: &v1.Pod{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					UID:       types.UID(podId),
					Name:      podName,
					Namespace: namespace,
				},
			},
		}
		rc := NewReplicationControllersResourceManager(constvalue.TypeReplicationController, namespace, podName)
		monkey.PatchInstanceMethod(reflect.TypeOf(rc), "GetReplica", func(_ *ReplicationControllersResourceManager) (int, error) {
			return 0, errors.New("get replica err")
		})
		defer monkey.UnpatchAll()
		rc.AppendUsedResourcePodName(podId)
		err := rc.Free(opod)
		convey.So(err, convey.ShouldResemble, errors.New("get replica err"))
	})
}
*/
/*func TestReplicationControllersResourceManager_Free2(t *testing.T) {
	convey.Convey("replica < LocalResource and succ", t, func() {
		podName := "pod1"
		namespace := "admin"
		podId := "pod1"
		opod := &operatingPod{
			pod: &v1.Pod{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					UID:       types.UID(podId),
					Name:podName,
					Namespace:namespace,
				},
			},
		}
		rc := NewReplicationControllersResourceManager(constvalue.TypeReplicationController, namespace, podName)
		localResource := 3
		rc.LocalResource = localResource
		monkey.PatchInstanceMethod(reflect.TypeOf(rc), "GetReplica", func(_ *ReplicationControllersResourceManager) (int, error) {
			return 0, nil
		})
		monkey.Patch(SaveResourceManagerToETCD, func(rm ResourceManagerInterface) error {
			return nil
		})
		monkey.Patch(deleteResource, func(oPod *operatingPod) error {
			return nil
		})
		defer monkey.UnpatchAll()
		rc.AppendUsedResourcePodName(podId)
		err := rc.Free(opod)
		convey.So(err, convey.ShouldBeNil)
		convey.So(rc.LocalResource, convey.ShouldEqual, localResource-1)
		convey.So(len(rc.UsedResourcePodName), convey.ShouldEqual, 0)
	})
}
*/
/*func TestReplicationControllersResourceManager_Free3(t *testing.T) {
	convey.Convey("replica < LocalResource and deleteResource Err", t, func() {
		podName := "pod1"
		namespace := "admin"
		podId := "pod1"
		opod := &operatingPod{
			pod: &v1.Pod{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					UID:       types.UID(podId),
					Name:      podName,
					Namespace: namespace,
				},
			},
		}
		rc := NewReplicationControllersResourceManager(constvalue.TypeReplicationController, namespace, podName)
		localResource := 3
		rc.LocalResource = localResource
		monkey.PatchInstanceMethod(reflect.TypeOf(rc), "GetReplica", func(_ *ReplicationControllersResourceManager) (int, error) {
			return 0, nil
		})
		defer monkey.UnpatchAll()
		monkey.Patch(deleteResource, func(oPod *operatingPod) error {
			return errors.New("deleteResource err")
		})
		rc.AppendUsedResourcePodName(podId)
		err := rc.Free(opod)
		convey.So(err, convey.ShouldResemble, errors.New("deleteResource err"))
		convey.So(rc.LocalResource, convey.ShouldEqual, localResource)
	})
}
*/
/*func TestReplicationControllersResourceManager_Free4(t *testing.T) {
	convey.Convey("replica > LocalResource ", t, func() {
		podName := "pod1"
		podId := "1111"
		namespace := "admin"
		ports := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName: podId,
				},
				LazyAttr: PortLazyAttr{
					ID: "11111",
				},
			},
		}

		opod := &operatingPod{
			pod: &v1.Pod{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
					UID:       types.UID(podId),
				},
			},
		}

		podExpect := &Pod{
			PodName:  podName,
			TenantId: namespace,
			PodNs:    namespace,
			PodID:    podId,
			Ports:    ports,
		}
		rc := NewReplicationControllersResourceManager(constvalue.TypeReplicationController, namespace, podName)
		localResource := 0
		rc.LocalResource = localResource
		rc.AppendUsedResourcePodName(podId)
		monkey.PatchInstanceMethod(reflect.TypeOf(rc), "GetReplica", func(_ *ReplicationControllersResourceManager) (int, error) {
			return 5, nil
		})

		var podservice *podService
		monkey.PatchInstanceMethod(reflect.TypeOf(podservice), "ReportDeletePod", func(_ *podService, podNs, podName string) error {
			return nil
		})
		defer monkey.UnpatchAll()
		var service *podService
		monkey.PatchInstanceMethod(reflect.TypeOf(service), "Get", func(_ *podService, podNs, podName string) (*Pod, error) {
			if podNs == opod.pod.Namespace && podName == opod.pod.Name {
				return podExpect, nil
			}
			return nil, errors.New("unexpected err")
		})

		//todo refactor test. add podService.Delete
		mockController := gomock.NewController(t)
		defer mockController.Finish()
		dbmock := mockdbaccessor.NewMockDbAccessor(mockController)
		dbmock.EXPECT().DeleteLeaf("/knitter/monitor/pods/admin/pod1").Return(nil)
		dbmock.EXPECT().SaveLeaf(gomock.Any(),gomock.Any()).Return(nil)

		monkey.Patch(infra.GetDataBase, func() dbaccessor.DbAccessor {
			return dbmock
		})
		defer monkey.UnpatchAll()

		err := rc.Free(opod)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(rc.UnusedResource), convey.ShouldEqual, 1)
		convey.So(rc.UnusedResource[0], convey.ShouldResemble, ports)

	})
}
*/
/*func TestReplicationControllersResourceManager_Free5(t *testing.T) {
	convey.Convey("replica > LocalResource ", t, func() {
		podName := "pod1"
		podId := "1111"
		namespace := "admin"
		ports := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName: podId,
				},
				LazyAttr: PortLazyAttr{
					ID: "11111",
				},
			},
		}

		opod := &operatingPod{
			pod: &v1.Pod{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
					UID:       types.UID(podId),
				},
			},
		}

		podExpect := &Pod{
			PodName:  podName,
			TenantId: namespace,
			PodNs:    namespace,
			PodID:    podId,
			Ports:    ports,
		}
		rc := NewReplicationControllersResourceManager(constvalue.TypeReplicationController, namespace, podName)
		localResource := 0
		rc.SetLocalResource(localResource)
		rc.AppendUsedResourcePodName(podId)
		monkey.PatchInstanceMethod(reflect.TypeOf(rc), "GetReplica", func(_ *ReplicationControllersResourceManager) (int, error) {
			return 5, nil
		})
		defer monkey.UnpatchAll()
		var service *podService
		monkey.PatchInstanceMethod(reflect.TypeOf(service), "Get", func(_ *podService, podNs, podName string) (*Pod, error) {
			return podExpect, nil
		})

		//todo refactor test. add podService.Delete
		mockController := gomock.NewController(t)
		defer mockController.Finish()
		dbmock := mockdbaccessor.NewMockDbAccessor(mockController)
		dbmock.EXPECT().DeleteLeaf("/knitter/monitor/pods/admin/pod1").Return(errors.New("delete err"))

		monkey.Patch(dbconfig.GetDataBase, func() dbaccessor.DbAccessor {
			return dbmock
		})
		defer monkey.UnpatchAll()

		err := rc.Free(opod)
		convey.So(err, convey.ShouldResemble, errors.New("delete err"))
		convey.So(len(rc.UnusedResource), convey.ShouldEqual, 0)

	})
}
*/