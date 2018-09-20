package services

import (
	"fmt"
	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/infra/clients"
	"github.com/ZTE/Knitter/pkg/klog"
)

type BluePrintNetworkMessage struct {
	Ports []BluePrintPort `json:"ports"`
}

func NewDefaultNetworkMessage() (*BluePrintNetworkMessage, error) {
	bpnm := &BluePrintNetworkMessage{}
	netName, err := clients.GetManagerClient().GetDefaultNetWork(constvalue.PaaSTenantAdminDefaultUUID)
	if err != nil {
		klog.Errorf("NewDelaultNetworkMessage() err: %v", err)
		return nil, err
	}
	ports := make([]BluePrintPort, 0)
	port := BluePrintPort{
		AttachToNetwork: netName,
		Attributes: BluePrintAttributes{
			Accelerate: constvalue.DefaultIsAccelerate,
			Function:   constvalue.DefaultNetworkPlane,
			NicName:    constvalue.DefaultPortName,
			NicType:    constvalue.DefaultVnicType,
		},
	}
	ports = append(ports, port)
	bpnm.Ports = ports
	return bpnm, nil
}

type BluePrintPort struct {
	AttachToNetwork string              `json:"attach_to_network"`
	Attributes      BluePrintAttributes `json:"attributes"`
}

type BluePrintAttributes struct {
	Accelerate  string      `json:"accelerate"`
	Function    string      `json:"function"`
	NicName     string      `json:"nic_name"`
	NicType     string      `json:"nic_type"`
	MetaData    interface{} `json:"metadata"`
	IpGroupName string      `json:"ip_group_name"`
	Combinable  string      `json:"combinable"`
}

func (netMsgInBP *BluePrintNetworkMessage) checkPorts() error {
	if err := isPortNameReused(netMsgInBP.Ports); err != nil {
		return err
	}
	if err := isPhysicalPortAccelerateForNonEio(netMsgInBP.Ports); err != nil {
		return err
	}
	return nil
}

func isPortNameReused(ports []BluePrintPort) error {
	for i := 0; i < len(ports); i++ {
		for j := i + 1; j < len(ports); j++ {
			if ports[i].Attributes.NicName == ports[j].Attributes.NicName {
				return fmt.Errorf("port Name: %v is reused", ports[i].Attributes.NicName)
			}
		}
	}

	return nil
}

func isPhysicalPortAccelerateForNonEio(ports []BluePrintPort) error {
	for _, port := range ports {
		if port.Attributes.NicType == constvalue.MechDriverPhysical &&
			port.Attributes.Accelerate == "true" &&
			port.Attributes.Function != constvalue.NetPlaneEio {
			return fmt.Errorf("physical is accelerate for not eio netplane")
		}
	}
	return nil
}
