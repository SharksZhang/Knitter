package services

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"time"
	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/daos"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/knitter-monitor/infra/clients"
	"github.com/ZTE/Knitter/knitter-monitor/infra/db-config"
	"github.com/ZTE/Knitter/pkg/etcd"
	"github.com/ZTE/Knitter/pkg/klog"
)

func InitConfigurations4Monitor() error {
	beegoCfg := viper.GetString("beegoCfg")
	err := beego.LoadAppConfig("ini", beegoCfg)
	if err != nil {
		klog.Errorf("loadConfig: beego.LoadAppConfig(ini, beegoCfg:[%v]) error, err is [%v]", beegoCfg, err)
		return err
	}

	klogDir := viper.GetString("conf.monitor.log_dir")
	if klogDir == "" {
		klogDir = constvalue.DefaultMonitorLogDir
	}
	klog.ConfigLog(klogDir)

	err = initEtcd()
	if err != nil {
		klog.Errorf("InitConfigurations4Monitor: initEtcd error, error is [%v] ", err)
		return err
	}
	LoadAllResourcesToCache()
	err = clients.InitManagerClient()

	if err != nil {
		klog.Errorf("InitEnv4Monitor:Init cni manager error! Error:-%v", err)
		return fmt.Errorf("%v:InitEnv4Monitor:Init manager error", err)
	}

	waitManagerClient()

	err = clients.InitKubernetesClientset()
	if err != nil {
		klog.Errorf("infra.InitKubernetesClientset() err, error is [%v]", err)
		return err
	}

	klog.Info("InitConfigurations4Monitor: Init successful")
	return nil
}

func waitManagerClient() {
	for {
		managerClient := clients.GetManagerClient()
		checkErr := managerClient.CheckKnitterManager()
		if checkErr != nil {
			klog.Errorf("InitEnv4Monitor:CheckKnitterManager error! -%v",
				checkErr)
			time.Sleep(10 * time.Second)
		} else {
			break

		}
	}
}

func loadConfig() error {
	viper.BindPFlags(pflag.CommandLine)

	cfg := viper.GetString("cfg")
	viper.SetConfigFile(cfg)
	err := viper.ReadInConfig()
	if err != nil {
		klog.Errorf("loadConfig:viper.ReadInConfig() err, error is [%v]", err)
		return err
	}
	return nil

}

func initEtcd() error {

	etcdAPIVer := viper.GetInt64("conf.monitor.etcd.api_version")
	if etcdAPIVer == 0 {
		etcdAPIVer = int64(etcd.DefaultEtcdAPIVersion)
		klog.Warningf("InitEnv4Monitor: etcd api version is nil, use default:[%v]", etcdAPIVer)
	} else {
		klog.Infof("InitEnv4Manger: get etcd api version: %d", etcdAPIVer)
	}

	etcdURL := viper.GetString("conf.monitor.etcd.urls")
	if etcdURL == "" {
		klog.Errorf("InitEnv4Monitor: etcd service query url is null")
		return errors.New("InitEnv4Monitor: etcd service query url is null")
	}
	dbconfig.SetDataBase(etcd.NewEtcdWithRetry(int(etcdAPIVer), etcdURL))
	dbconfig.CheckDB()
	klog.Info("InitEnv4Manger: etcd service query url:", etcdURL)
	return nil
}

func LoadAllResourcesToCache() {
	for _, lroFunc := range LoadResourceObjectFuncs {
		LoadResouceObjectsLoop(lroFunc)
	}
	klog.Infof("LoadAllResourcesToCache: load all type resource SUCC")
}

type LoadResouceObjectFunc func() error

var LoadResourceObjectFuncs = []LoadResouceObjectFunc{
	daos.LoadAllPodForDBs,
}

func LoadResouceObjectsLoop(lroFunc LoadResouceObjectFunc) {
	for {
		var err error
		err = lroFunc()
		if err == nil || errobj.IsNotFoundError(err) {
			klog.Infof("LoadResouceObjectsLoop: lroFunc[%v] SUCC", lroFunc)
			return
		}
		klog.Warningf("LoadResouceObjectsLoop: lroFunc[%v] error: %v, just wait retry", lroFunc, err)
		time.Sleep(constvalue.GetLoadReourceRetryIntervalInSec * time.Second)
	}
}
