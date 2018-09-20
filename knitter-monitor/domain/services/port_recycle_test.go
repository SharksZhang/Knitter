package services

import (
	"github.com/bouk/monkey"
	"github.com/smartystreets/goconvey/convey"
	"testing"
	"github.com/ZTE/Knitter/knitter-monitor/daos"
)

func TestNewRecyclePod(t *testing.T) {
	podDB := &daos.PodForDB{
		PodName: "pod1",
		PodNs:   "admin",
		Ports: []*daos.PortForDB{
			&daos.PortForDB{ID: "1"},
			&daos.PortForDB{ID: "2"},
			&daos.PortForDB{ID: "3"},
		},
	}

	recyclePodExpect := &RecyclePod{
		podNS:      podDB.PodNs,
		podName:    podDB.PodName,
		recycleKey: RecycleKey(podDB.PodNs + podDB.PodName),
		portIDs:    []string{"1", "2", "3"},
	}

	convey.Convey("TestNewRecyclePod", t, func() {
		recyclePod := NewRecyclePod(podDB)
		convey.So(recyclePod, convey.ShouldResemble, recyclePodExpect)
	})
}

func TestPortRecyclePool_AddAndRecycle(t *testing.T) {
	recyclePod := &RecyclePod{
		podName:    "recyclePod",
		podNS:      "admin",
		recycleKey: RecycleKey("recyclePod" + "admin"),
		portIDs:    []string{"1", "2", "3"},
	}
	monkey.Patch(operateRecycle, func(portRecyclePool *PortRecyclePool, recyclePod *RecyclePod) {
		return
	})
	convey.Convey("TestPortRecyclePool_AddAndRecycle", t, func() {
		GetPortRecyclePool().AddAndRecycle(recyclePod)
		_, ok := GetPortRecyclePool().inProcJobs[RecycleKey(recyclePod.podName+recyclePod.podNS)]
		convey.So(ok, convey.ShouldBeTrue)
		GetPortRecyclePool().Delete(RecycleKey(recyclePod.podName + recyclePod.podNS))
		_, ok = GetPortRecyclePool().inProcJobs[RecycleKey(recyclePod.podName+recyclePod.podNS)]
		convey.So(ok, convey.ShouldBeFalse)

	})
}
