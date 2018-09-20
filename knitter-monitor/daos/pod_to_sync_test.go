package daos

import (
	"encoding/json"
	"testing"

	"github.com/bouk/monkey"
	"github.com/coreos/etcd/client"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"

	"github.com/ZTE/Knitter/knitter-manager/tests/mock/db-mock"
	"github.com/ZTE/Knitter/knitter-monitor/infra/db-config"
	"github.com/ZTE/Knitter/pkg/db-accessor"
)

func TestPodSyncDao_Get(t *testing.T) {
	convey.Convey("TestPodSyncDao_Get succ\n", t, func() {
		podName := "pod_name"
		podNs := "pod_ns"
		podInDB := &PodSyncDB{
			PodNs:            "pod_ns",
			PodName:          "pod_name",
			Action:           "action",
			IsSync2ManagerOk: true,
			IsSync2K8sOk:     true,
		}
		mockController := gomock.NewController(t)
		defer mockController.Finish()
		dbmock := mockdbaccessor.NewMockDbAccessor(mockController)
		monkey.Patch(dbconfig.GetDataBase, func() dbaccessor.DbAccessor {
			return dbmock
		})
		defer monkey.UnpatchAll()
		podBytes, _ := json.Marshal(podInDB)
		dbmock.EXPECT().ReadLeaf(gomock.Any()).Return(string(podBytes), nil)
		pod, err := GetPodSyncDao().Get(podNs, podName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(pod, convey.ShouldResemble, podInDB)
	})
}

func TestPodSyncDao_GetAll(t *testing.T) {
	convey.Convey("TestPodSyncDao_GetAll succ\n", t, func() {
		podInDB := &PodSyncDB{
			PodNs:            "pod_ns",
			PodName:          "pod_name",
			Action:           "action",
			IsSync2ManagerOk: true,
			IsSync2K8sOk:     true,
		}
		mockController := gomock.NewController(t)
		defer mockController.Finish()
		dbmock := mockdbaccessor.NewMockDbAccessor(mockController)
		monkey.Patch(dbconfig.GetDataBase, func() dbaccessor.DbAccessor {
			return dbmock
		})
		defer monkey.UnpatchAll()
		podBytes, _ := json.Marshal(podInDB)
		nodes := []*client.Node{
			{
				Key:   dbconfig.GetKeyOfPodsSyncToManagerAndK8s(),
				Value: string(podBytes),
			},
		}
		dbmock.EXPECT().ReadDir(gomock.Any()).Return(nodes, nil)
		pod, err := GetPodSyncDao().GetAll()
		convey.So(err, convey.ShouldBeNil)
		convey.So(pod, convey.ShouldResemble, []*PodSyncDB{podInDB})
	})
}

func TestPodSyncDao_Save(t *testing.T) {
	convey.Convey("TestPodSyncDao_Save succ\n", t, func() {
		podName := "pod_name"
		podNs := "pod_ns"
		podInDB := &PodSyncDB{
			PodNs:            "pod_ns",
			PodName:          "pod_name",
			Action:           "action",
			IsSync2ManagerOk: true,
			IsSync2K8sOk:     true,
		}
		mockController := gomock.NewController(t)
		defer mockController.Finish()
		dbmock := mockdbaccessor.NewMockDbAccessor(mockController)
		monkey.Patch(dbconfig.GetDataBase, func() dbaccessor.DbAccessor {
			return dbmock
		})
		defer monkey.UnpatchAll()

		dbmock.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil)
		err := GetPodSyncDao().Save(podNs, podName, podInDB)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPodSyncDao_Delete(t *testing.T) {
	convey.Convey("TestPodSyncDao_Delete succ\n", t, func() {
		podName := "pod_name"
		podNs := "pod_ns"
		mockController := gomock.NewController(t)
		defer mockController.Finish()
		dbmock := mockdbaccessor.NewMockDbAccessor(mockController)
		monkey.Patch(dbconfig.GetDataBase, func() dbaccessor.DbAccessor {
			return dbmock
		})
		defer monkey.UnpatchAll()

		dbmock.EXPECT().DeleteLeaf(gomock.Any()).Return(nil)
		err := GetPodSyncDao().Delete(podNs, podName)
		convey.So(err, convey.ShouldBeNil)
	})
}
