package services

import (
	"errors"

	"strconv"
	"strings"
	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/pkg/klog"
)

func combinePortObjs(ports []*Port) ([]*Port, error) {
	for i := 0; i < len(ports); i++ {
		ports[i].EagerAttr.Roles = append(ports[i].EagerAttr.Roles, ports[i].EagerAttr.NetworkPlane)
		for j := i + 1; j < len(ports); j++ {
			if isSameRole(ports[i], ports[j]) {
				err := isSameRolesIllegal(ports[i], ports[j])
				if err != nil {
					return nil, err
				}
				setPortNameToDefaultName(ports[j], ports[i])
				ports[i].EagerAttr.Roles = append(ports[i].EagerAttr.Roles, ports[j].EagerAttr.NetworkPlane)
				ports = delEleOfSliceByIndex(ports, j)
				//back one step because of delete j'st element in slice
				j--
			} else if err := isSameC1PortWithDiffCombineAttr(ports[i], ports[j]); err != nil {
				return nil, err
			}
		}
	}
	portStrs := make([]string, 0)
	for i, port := range ports {
		portTmp := "port[" + strconv.Itoa(i) + "] role is " + strings.Join(port.EagerAttr.Roles, ",")
		portStrs = append(portStrs, portTmp)
	}
	klog.Infof("Post combination [%v]", strings.Join(portStrs, "|"))
	return ports, nil

}

func isSameRole(port1, port2 *Port) bool {
	return port1.EagerAttr.Combinable == "true" &&
		port2.EagerAttr.Combinable == "true" &&
		port1.EagerAttr.NetworkName == port2.EagerAttr.NetworkName &&
		port1.EagerAttr.VnicType == port2.EagerAttr.VnicType &&
		port1.EagerAttr.Accelerate == port2.EagerAttr.Accelerate &&
		port1.EagerAttr.PodName == port2.EagerAttr.PodName &&
		port1.EagerAttr.PodNs == port2.EagerAttr.PodNs &&
		port1.EagerAttr.IPGroupName == port2.EagerAttr.IPGroupName
	//port1.EagerAttr.Metadata == port2.EagerAttr.Metadata
}

func isSameRolesIllegal(port1, port2 *Port) error {
	if isEioAndC0NetPlaneConflict(port1, port2) {
		err := errors.New("try to combine netplanes: " + port1.EagerAttr.NetworkPlane + " and " +
			port2.EagerAttr.NetworkPlane + " with same direct and accelerate attr error")
		klog.Errorf("IsSameRolesIllegal: %v", err)
		return err
	}
	return nil
}

func isEioAndC0NetPlaneConflict(port1, port2 *Port) bool {
	return (IsCTNetPlane(port1.EagerAttr.NetworkPlane) && port2.EagerAttr.NetworkPlane == constvalue.NetPlaneEio ||
		IsCTNetPlane(port2.EagerAttr.NetworkPlane) && port1.EagerAttr.NetworkPlane == constvalue.NetPlaneEio) &&
		port1.EagerAttr.Accelerate == "true" && port2.EagerAttr.Accelerate == "true"
}

func setPortNameToDefaultName(port1, port2 *Port) {
	if port1.EagerAttr.PortName == constvalue.DefaultPortName {
		port2.EagerAttr.PortName = constvalue.DefaultPortName
	}
}

func delEleOfSliceByIndex(slice []*Port, index int) []*Port {
	return append(slice[:index], slice[index+1:]...)
}

func isSameC1PortWithDiffCombineAttr(port1, port2 *Port) error {
	if isSameC1RolesWithDiffCombineAttrBool(port1, port2) {
		err := errors.New("netplanes: " +
			port1.EagerAttr.NetworkPlane + " and " + port2.EagerAttr.NetworkPlane +
			" are all using the same C0's port with the same NetworkName: " +
			port1.EagerAttr.NetworkName + ", they must be combined")
		klog.Errorf("isSameC1PortWithDiffCombineAttr: %v", err)
		return err
	}
	return nil
}

func isSameC1RolesWithDiffCombineAttrBool(port1, port2 *Port) bool {
	return port1.EagerAttr.Combinable != port2.EagerAttr.Combinable &&
		port1.EagerAttr.NetworkName == port2.EagerAttr.NetworkName &&
		IsCTNetPlane(port1.EagerAttr.NetworkPlane) &&
		IsCTNetPlane(port2.EagerAttr.NetworkPlane) &&
		port1.EagerAttr.Accelerate == "true" &&
		port2.EagerAttr.Accelerate == "true"
}

func IsCTNetPlane(netPlane string) bool {
	return netPlane == constvalue.NetPlaneControl ||
		netPlane == constvalue.NetPlaneMedia ||
		netPlane == constvalue.NetPlaneOam
}

func LinkStrOfSliceToStrWithDot(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	var linkedSlice string = slice[0]
	for i := 1; i < len(slice); i++ {
		linkedSlice = linkedSlice + "," + slice[i]
	}
	return linkedSlice
}
