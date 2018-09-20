package services

import (
	"errors"
	"testing"

	"github.com/bouk/monkey"
	"github.com/smartystreets/goconvey/convey"
)

func TestCheckPortsInBP(t *testing.T) {
	convey.Convey("TestCheckPortsInBP_succ\n", t, func() {
		monkey.Patch(isPortNameReused, func(ports []BluePrintPort) error {
			return nil
		})
		defer monkey.UnpatchAll()
		monkey.Patch(isPhysicalPortAccelerateForNonEio, func(ports []BluePrintPort) error {
			return nil
		})

		bpnm := &BluePrintNetworkMessage{}
		err := bpnm.checkPorts()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestCheckPortsInBP_isPortNameReused_false\n", t, func() {
		monkey.Patch(isPortNameReused, func(ports []BluePrintPort) error {
			return errors.New("isPortNameReused false")
		})
		defer monkey.UnpatchAll()
		bpnm := &BluePrintNetworkMessage{}
		err := bpnm.checkPorts()
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("TestCheckPortsInBP_isPhysicalPortAccelerateForNonEio_false\n", t, func() {
		monkey.Patch(isPortNameReused, func(ports []BluePrintPort) error {
			return nil
		})
		defer monkey.UnpatchAll()
		monkey.Patch(isPhysicalPortAccelerateForNonEio, func(ports []BluePrintPort) error {
			return errors.New("isPhysicalPortAccelerateForNonEio false")
		})
		bpnm := &BluePrintNetworkMessage{}
		err := bpnm.checkPorts()
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestIsPortNameReused(t *testing.T) {
	convey.Convey("TestIsPortNameReused_true\n", t, func() {
		ports := []BluePrintPort{
			{
				Attributes: BluePrintAttributes{
					NicName: "portname1",
				},
			},
			{
				Attributes: BluePrintAttributes{
					NicName: "portname1",
				},
			},
		}
		err := isPortNameReused(ports)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("TestIsPortNameReused_false\n", t, func() {
		ports := []BluePrintPort{
			{
				Attributes: BluePrintAttributes{
					NicName: "portname1",
				},
			},
			{
				Attributes: BluePrintAttributes{
					NicName: "portname2",
				},
			},
		}
		err := isPortNameReused(ports)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestIsPhysicalPortAccelerateForNonEio(t *testing.T) {
	convey.Convey("TestIsPhysicalPortAccelerateForNonEio_true\n", t, func() {
		ports := []BluePrintPort{
			{
				Attributes: BluePrintAttributes{
					NicType:    "physical",
					Accelerate: "true",
					Function:   "eio",
				},
			},
			{
				Attributes: BluePrintAttributes{
					NicType:    "physical",
					Accelerate: "true",
					Function:   "control",
				},
			},
		}
		err := isPhysicalPortAccelerateForNonEio(ports)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("TestIsPhysicalPortAccelerateForNonEio_false\n", t, func() {
		ports := []BluePrintPort{
			{
				Attributes: BluePrintAttributes{
					NicType:    "physical",
					Accelerate: "true",
					Function:   "eio",
				},
			},
			{
				Attributes: BluePrintAttributes{
					NicType:    "physical",
					Accelerate: "false",
					Function:   "eio",
				},
			},
		}
		err := isPhysicalPortAccelerateForNonEio(ports)
		convey.So(err, convey.ShouldBeNil)
	})
}
