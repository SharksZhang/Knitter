package services

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPod2SyncStatusRepo_Get(t *testing.T) {
	convey.Convey("TestPod2SyncStatusRepo_Get succ\n", t, func() {
		Pod2SyncStatusRepoCtx.Pod2SyncStatus = map[string]*exceptPodSyncStatus{
			"pod_ns0pod_name0": &exceptPodSyncStatus{
				PodNs:   "pod_ns0",
				PodName: "pod_name0",
			},
			"pod1pod_name1": &exceptPodSyncStatus{
				PodNs:   "pod_ns1",
				PodName: "pod_name1",
			},
		}
		status := GetPod2SyncStatusRepo().Get("pod_ns0", "pod_name0")
		convey.So(status.PodNs, convey.ShouldEqual, "pod_ns0")
		convey.So(status.PodName, convey.ShouldEqual, "pod_name0")
	})
	convey.Convey("TestPod2SyncStatusRepo_Get fail\n", t, func() {
		Pod2SyncStatusRepoCtx.Pod2SyncStatus = map[string]*exceptPodSyncStatus{
			"pod_ns0pod_name0": &exceptPodSyncStatus{
				PodNs:   "pod_ns0",
				PodName: "pod_name0",
			},
			"pod1pod_name1": &exceptPodSyncStatus{
				PodNs:   "pod_ns1",
				PodName: "pod_name1",
			},
		}
		status := GetPod2SyncStatusRepo().Get("pod_ns3", "pod_name3")
		convey.So(status, convey.ShouldBeNil)
	})
}

func TestPod2SyncStatusRepo_GetAll(t *testing.T) {
	convey.Convey("TestPod2SyncStatusRepo_GetAll succ\n", t, func() {
		Pod2SyncStatusRepoCtx.Pod2SyncStatus = map[string]*exceptPodSyncStatus{
			"pod_ns0pod_name0": &exceptPodSyncStatus{
				PodNs:   "pod_ns0",
				PodName: "pod_name0",
			},
			"pod_ns1pod_name1": &exceptPodSyncStatus{
				PodNs:   "pod_ns1",
				PodName: "pod_name1",
			},
		}
		status := GetPod2SyncStatusRepo().GetAll()
		convey.So(status["pod_ns0pod_name0"].PodName, convey.ShouldEqual, "pod_name0")
		convey.So(status["pod_ns1pod_name1"].PodNs, convey.ShouldEqual, "pod_ns1")

	})
}
