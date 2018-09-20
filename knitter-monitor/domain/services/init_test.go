package services

import (
	"github.com/bouk/monkey"
	"github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"testing"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/knitter-monitor/infra/clients"
	"github.com/ZTE/Knitter/knitter-monitor/infra/db-config"
)

//func TestInitConfigurations4Monitor(t *testing.T) {
//	monkey.Patch(initEtcd, func() error {
//		return nil
//	})
//
//	monkey.Patch(infra.InitManagerClient, func() error {
//		return nil
//	})
//	defer monkey.UnpatchAll()
//
//	monkey.Patch(waitManagerClient, func() {
//
//	})
//
//	monkey.Patch(beego.LoadAppConfig, func(adapterName, configPath string) error {
//		return nil
//	})
//
//
//	convey.Convey("TestInitConfigurations4Monitor", t, func() {
//		err := InitConfigurations4Monitor()
//		//convey.So(err, convey.ShouldBeNil)
//		convey.So(viper.GetString("conf.monitor.log_dir"), convey.ShouldEqual, "/root/info/logs/nwnode")
//		convey.So(viper.GetInt("conf.monitor.etcd.api_version"), convey.ShouldEqual, 3)
//		convey.So(viper.GetString("cfg"), convey.ShouldEqual, constvalue.DefaultConfDir+constvalue.DefaultConfFile)
//		convey.So(viper.GetString("kubeconfig"), convey.ShouldEqual, constvalue.DefaultKubeconfig)
//		convey.So(viper.GetString("beegoCfg"),convey.ShouldEqual, constvalue.DefaultBeegoConfPath)
//	})
//}

func TestInitConfigurations4MonitorBeegoErr(t *testing.T) {
	monkey.Patch(initEtcd, func() error {
		return nil
	})

	monkey.Patch(clients.InitManagerClient, func() error {
		return nil
	})
	defer monkey.UnpatchAll()

	monkey.Patch(waitManagerClient, func() {

	})

	convey.Convey("TestInitConfigurations4MonitorBeegoErr", t, func() {
		err := InitConfigurations4Monitor()
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestInitEtcd(t *testing.T) {
	monkey.Patch(dbconfig.SetDataBase, func(i dbaccessor.DbAccessor) error {
		return nil
	})



	monkey.Patch(dbconfig.CheckDB, func() error {
		return nil
	})
	defer monkey.UnpatchAll()


	convey.Convey("TestInitEtcd", t, func() {
		err := initEtcd()
		convey.So(err, convey.ShouldBeNil)
	})
}
