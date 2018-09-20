package dbconfig

import (
	"github.com/bouk/monkey"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"testing"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/knitter-manager/tests/mock/db-mock"
)

func TestGetDatabaseSetDataBase(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(controller)

	convey.Convey("TestGetDatabaseSetDataBase", t, func() {
		SetDataBase(dbmock)
		db := GetDataBase()
		convey.So(db, convey.ShouldEqual, dbmock)
	})
}

func TestCheckDB(t *testing.T) {

	monkey.Patch(dbaccessor.CheckDataBase, func(Db dbaccessor.DbAccessor) error {
		return nil
	})
	defer monkey.UnpatchAll()
	monkey.Patch(GetClusterUUID, func() string {
		return "1111"
	})

	convey.Convey("TestCheckDB", t, func() {
		err := CheckDB()
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetKeyOfMonitorPortRecyclePod(t *testing.T) {
	podns := "admin"
	podName := "pod1"
	keyExpect := "/knitter/monitor/portrecycle/"+podns + "/" + podName
	convey.Convey("TestGetKeyOfMonitorPortRecyclePod", t, func() {
		key := GetKeyOfMonitorPortRecyclePod(podns, podName)
		convey.So(key, convey.ShouldEqual, keyExpect)
	})
}

