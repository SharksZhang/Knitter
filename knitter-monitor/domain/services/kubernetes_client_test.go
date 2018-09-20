package services

//todo refactor
//func TestGetKubernetesClient(t *testing.T) {
//	convey.Convey("TestGetClient", t, func() {
//		restClient := GetMonitorK8sClient()
//		convey.So(restClient, convey.ShouldHaveSameTypeAs, &MonitorK8sClient{})
//		convey.So(restClient, convey.ShouldNotBeNil)
//		convey.So(restClient.client, convey.ShouldNotBeNil)
//		var k8sClient *MonitorK8sClient
//		replica := int32(5)
//		monkey.PatchInstanceMethod(reflect.TypeOf(k8sClient), "GetReplicationController", func(_ *MonitorK8sClient,nameSpace string, Name string) (*v1.ReplicationController, error) {
//			return &v1.ReplicationController{
//				Spec:v1.ReplicationControllerSpec{
//					Replicas:&replica,
//				},
//			}, nil
//		})
//		defer monkey.UnpatchAll()
//		rc, err :=restClient.GetReplicationControllerReplicas("admin", "1111")
//		convey.So(err,convey.ShouldBeNil)
//		convey.So(rc, convey.ShouldEqual, 5)
//
//	})
//
//}
