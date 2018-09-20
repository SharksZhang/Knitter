package daos

import (
	"encoding/json"
	"github.com/bouk/monkey"
	"github.com/golang/mock/gomock"
	"errors"
	"github.com/smartystreets/goconvey/convey"
	"testing"
	"github.com/ZTE/Knitter/knitter-manager/tests/mock/db-mock"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/knitter-monitor/infra/db-config"
)

//func TestPodDao_Get(t *testing.T) {
//	value := `{	"tenant_id":"admin",
//				"pod_id":"1111",
//				"pod_name":"222"}`
//	mockController := gomock.NewController(t)
//	defer mockController.Finish()
//	dbmock := mockdbaccessor.NewMockDbAccessor(mockController)
//	dbmock.EXPECT().ReadLeaf("/knitter/monitor/pods/admin/222").Return(value, nil)
//
//	monkey.Patch(infra.GetDataBase, func() dbaccessor.DbAccessor {
//		return dbmock
//	})
//	defer monkey.UnpatchAll()
//
//	podExpect := &PodForDB{
//		TenantId: "admin",
//		PodID:    "1111",
//		PodName:  "222",
//	}
//	convey.Convey("TestPodDao_Get", t, func() {
//		pod, err := GetPodDao().Get("admin", "222")
//		convey.So(pod, convey.ShouldResemble, podExpect)
//		convey.So(err, convey.ShouldBeNil)
//	})
//
//}

//func TestPodDao_GetFail(t *testing.T) {
//	mockController := gomock.NewController(t)
//	defer mockController.Finish()
//	dbmock := mockdbaccessor.NewMockDbAccessor(mockController)
//	dbmock.EXPECT().ReadLeaf("/knitter/monitor/pods/admin/222").Return("", errors.New("get err"))
//
//	monkey.Patch(infra.GetDataBase, func() dbaccessor.DbAccessor {
//		return dbmock
//	})
//	defer monkey.UnpatchAll()
//
//	convey.Convey("TestPodDao_Get", t, func() {
//		pod, err := GetPodDao().Get("admin", "222")
//		convey.So(pod, convey.ShouldBeNil)
//		convey.So(err.Error(), convey.ShouldEqual, "get err")
//	})
//
//}

func TestPodDao_Save(t *testing.T) {
	podExpect := &PodForDB{
		TenantId: "admin",
		PodID:    "1111",
		PodName:  "222",
		PodNs:    "admin",
	}
	value, _ := json.Marshal(podExpect)
	mockController := gomock.NewController(t)
	defer mockController.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(mockController)
	dbmock.EXPECT().SaveLeaf("/knitter/monitor/pods/admin/222", string(value)).Return(nil)

	monkey.Patch(dbconfig.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodDao_Save", t, func() {
		err := GetPodDao().Save(podExpect)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPodDao_SaveFail(t *testing.T) {
	podExpect := &PodForDB{
		TenantId: "admin",
		PodID:    "1111",
		PodName:  "222",
		PodNs:    "admin",
	}
	value, _ := json.Marshal(podExpect)
	mockController := gomock.NewController(t)
	defer mockController.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(mockController)
	dbmock.EXPECT().SaveLeaf("/knitter/monitor/pods/admin/222", string(value)).Return(errors.New("save err"))

	monkey.Patch(dbconfig.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodDao_Save", t, func() {
		err := GetPodDao().Save(podExpect)
		convey.So(err.Error(), convey.ShouldEqual, "save err")
	})
}

func TestPodDao_Delete(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(mockController)
	dbmock.EXPECT().DeleteLeaf("/knitter/monitor/pods/admin/222").Return(nil)

	monkey.Patch(dbconfig.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodDao_Delete", t, func() {
		err := GetPodDao().Delete("admin", "222")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPodDao_DeleteFail(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(mockController)
	dbmock.EXPECT().DeleteLeaf("/knitter/monitor/pods/admin/222").Return(errors.New("delete err"))

	monkey.Patch(dbconfig.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodDao_DeleteFail", t, func() {
		err := GetPodDao().Delete("admin", "222")
		convey.So(err.Error(), convey.ShouldEqual, "delete err")
	})
}

func TestPodDao_DeleteKeyNotFound(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(mockController)
	dbmock.EXPECT().DeleteLeaf("/knitter/monitor/pods/admin/222").Return(errors.New(errobj.EtcdKeyNotFound))

	monkey.Patch(dbconfig.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodDao_DeleteKeyNotFound", t, func() {
		err := GetPodDao().Delete("admin", "222")
		convey.So(err, convey.ShouldBeNil)
	})
}
