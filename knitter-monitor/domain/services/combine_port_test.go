package services

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCombinePortObjs(t *testing.T) {
	convey.Convey("TestCombinePortObjs_UseEth0First\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth1",
					VnicType:     "nomal",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "control",
					PortName:     "eth0",
					VnicType:     "nomal",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
		}

		aimPortObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "nomal",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
					Roles:        []string{"media", "control"},
				},
			},
		}

		portObjsResult, err := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldResemble, aimPortObjs)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestCombinePortObjs_VfNotAccelerateSucc\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "control",
					PortName:     "eth1",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "eio",
					PortName:     "eth2",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "oam",
					PortName:     "eth3",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
		}

		aimPortObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
					Roles:        []string{"media", "control", "eio", "oam"},
				},
			},
		}

		portObjsResult, err := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldResemble, aimPortObjs)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestCombinePortObjs_C0Succ\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "control",
					PortName:     "eth1",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "oam",
					PortName:     "eth2",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
		}
		aimPortObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
					Roles:        []string{"media", "control", "oam"},
				},
			},
		}

		portObjsResult, err := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldResemble, aimPortObjs)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestCombinePortObjs_CombineTrueAndFalseSucc\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "control",
					PortName:     "eth1",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "oam",
					PortName:     "eth2",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "oam",
					PortName:     "eth3",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "false",
					Metadata:     "test_metadata",
				},
			},
		}
		aimPortObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
					Roles:        []string{"media", "control", "oam"},
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "oam",
					PortName:     "eth3",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "false",
					Metadata:     "test_metadata",
					Roles:        []string{"oam"},
				},
			},
		}

		portObjsResult, err := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldResemble, aimPortObjs)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestCombinePortObjs_CombineTrueAndFalseSucc2\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "oam",
					PortName:     "eth1",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "false",
					Metadata:     "test_metadata",
				},
			},
		}
		aimPortObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
					Roles:        []string{"media"},
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "oam",
					PortName:     "eth1",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "false",
					Metadata:     "test_metadata",
					Roles:        []string{"oam"},
				},
			},
		}

		portObjsResult, err := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldResemble, aimPortObjs)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestCombinePortObjs_WithSameNetPlaneSucc\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "nomal",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth1",
					VnicType:     "nomal",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
		}
		aimPortObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "nomal",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
					Roles: []string{
						"media",
						"media",
					},
				},
			},
		}
		portObjsResult, err := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldResemble, aimPortObjs)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestCombinePortObjs_WithEioAndC0ConflictError\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "nomal",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "eio",
					PortName:     "eth1",
					VnicType:     "nomal",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
		}
		portObjsResult, err := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("TestCombinePortObjs_isSameC1RolesWithDiffCombineAttr\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "control",
					PortName:     "eth0",
					VnicType:     "nomal",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "control",
					PortName:     "eth1",
					VnicType:     "nomal",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "false",
					Metadata:     "test_metadata",
				},
			},
		}
		portObjsResult, err := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
	})
}
