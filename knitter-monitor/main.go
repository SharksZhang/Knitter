package main

import (
	"flag"

	"github.com/astaxie/beego"

	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/domain/services"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/knitter-monitor/infra/repo-impl"

	_ "github.com/ZTE/Knitter/knitter-monitor/routers"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/version"
	"github.com/spf13/viper"
	"errors"
	"github.com/spf13/pflag"
)

func main() {
	defer errobj.RecoverPanic()

	err := loadConfigMessage()
	if err != nil {
		klog.Errorf("loadConfigMessage() err, error is [%v]", err)
		return
	}

	err = services.InitConfigurations4Monitor()
	if err != nil {
		klog.Errorf("services.InitConfigurations4Monitor() err, error is [%v]", err)
		return
	}


	services.ResourceManagerRepo = &repo_impl.ResourceManagerRepository{}


	err = repo_impl.LoadResourceManagerRepoCache()
	if err != nil {
		klog.Errorf("services.loadResourceManagerRepoCache() err, error is [%v]", err)
		return
	}

	createPortForPodController, err := services.NewCreatePortsController()
	services.CreatePorts4PodController = createPortForPodController
	if err != nil {
		klog.Errorf("services.NewCreatePortForPodController() err, error is [%v]", err)
		return
	}

	err = services.GetPod2SyncStatusRepo().Init()
	if err != nil {
		klog.Errorf("services.GetPod2SyncStatusRepo().Init() error: %v", err)
		return
	}

	go services.ReportPodToManagerAndK8sRecycle()

	var stopCh <-chan struct{}
	go createPortForPodController.Run(constvalue.DefaultWorkerNumber, stopCh)

	beego.ErrorHandler("404", errobj.PageNotFound)
	beego.BeeApp.Server.MaxHeaderBytes = 16 * (1 << 10)
	go services.SyncResourceManagerPorts()
	klog.Infof("Run knitter monitor successful")
	beego.Run()
}

func loadConfigMessage() error {
	flag.String("cfg", constvalue.DefaultConfDir+constvalue.DefaultConfFile, "config file path")
	flag.String("kubeconfig", constvalue.DefaultKubeconfig, "absolute path to the kubeconfig file")
	flag.String("beegoCfg", constvalue.DefaultBeegoConfPath, "beego config path")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	flag.Parse()
	if version.HasVerFlag() {
		version.PrintVersion()
		return errors.New("print version ")
	}

	viper.BindPFlags(pflag.CommandLine)

	cfg := viper.GetString("cfg")
	viper.SetConfigFile(cfg)
	err := viper.ReadInConfig()
	if err != nil {
		klog.Errorf("loadConfigMessage:viper.ReadInConfig() err, error is [%v]", err)
		return err
	}
	return nil

}

