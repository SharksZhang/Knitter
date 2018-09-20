package clients

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/viper"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/pkg/klog"
)

var clientSet *kubernetes.Clientset

func InitKubernetesClientset() error {
	masterURL := viper.GetString("conf.monitor.k8s.url")
	if masterURL == "" {
		klog.Errorf("InitKubernetesClientset: [%v]", errobj.ErrK8sMasterUrlIsNil)
		return errobj.ErrK8sMasterUrlIsNil
	}

	config, err := clientcmd.BuildConfigFromFlags(masterURL, "")
	if err != nil {
		klog.Errorf("clientcmd.BuildConfigFromFlags( ) err, error is [%v]", err)
		return err
	}

	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorf("kubernetes.NewForConfig(config) err, error is [%v]", err)
		return err
	}
	klog.Infof("InitKubernetesClientset : Init successful")
	return nil
}

func GetClientset() *kubernetes.Clientset {
	klog.Debugf("GetClientset")
	return clientSet
}
