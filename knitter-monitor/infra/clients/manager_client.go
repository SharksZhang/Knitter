package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"

	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/spf13/viper"
)

var managerClient *ManagerClient

type ManagerClient struct {
	// Header *http.Header
	URLKnitterManager string
	VMID              string
}

func InitManagerClient() error {
	managerClient = &ManagerClient{}
	managerURL := viper.GetString("conf.monitor.manager.url")
	klog.Infof("cfg.GetString manager url: %v", managerURL)
	if managerURL == "" {
		klog.Errorf("InitClient:manager url is null")
		return errors.New("manager url is null")
	}
	managerClient.URLKnitterManager = managerURL
	return nil
}

func GetManagerClient() *ManagerClient {
	return managerClient
}

func (mc *ManagerClient) ReportPod(pod *ReportPod) error {
	klog.Debugf("ManagerClient.ReportPod: start ")
	url := mc.GetReportPodURL(pod.TenantID, pod.PodName)

	podByte, err := json.Marshal(pod)
	klog.Infof("ManagerClient.ReportPod: reportPod is [%v]", string(podByte))
	if err != nil {
		klog.Errorf("ManagerClient.ReportPod:json.Marshal(pod:[%v]), error ,err is [%v]", pod, err)
		return err
	}

	postStatusCode, rspByte, err := mc.PostBytes(url, podByte)
	if err != nil {
		klog.Errorf("ManagerClient.ReportPod: mc.Post(posturl: [%v], pod: [%v]) error, err is [%v]", url, string(podByte), err)
		return err
	} else if !IsHttpMethodStatusSuccess(postStatusCode) {
		klog.Errorf("ManagerClient.ReportPod: mc.Post(posturl: [%v], pod: [%v]) ok, but return status code is %v, "+
			"resp is [%v]", url, string(podByte), postStatusCode, string(rspByte))
		return errors.New(errobj.GetErrMsg(rspByte))
	}
	klog.Debugf("ManagerClient.ReportPod: END ")

	return nil
}

func (mc *ManagerClient) ReportDeletePod(podNs, podName string) error {
	klog.Debugf("ManagerClient.ReportDeletePod: start ")
	url := mc.GetReportPodURL(podNs, podName)
	rspByte, StatusCode, err := mc.Delete(url)
	if err != nil {
		klog.Errorf("ManagerClient.ReportDeletePod: "+
			"mc.Delete(posturl: [%v]) error, err is [%v]", url, err)
		return err
	} else if !IsHttpMethodStatusSuccess(StatusCode) {
		klog.Errorf("ManagerClient.ReportDeletePod: mc.Delete(url: [%v]) ok, but return status code is %v", url, StatusCode)
		return errors.New(errobj.GetErrMsg(rspByte))
	}
	klog.Debugf("ManagerClient.ReportDeletePod: END ")
	return nil

}

type ReportPod struct {
	PodName   string     `json:"pod_name"`
	PodNs     string     `json:"pod_ns"`
	ClusterID string     `json:"cluster_id"`
	TenantID  string     `json:"tenant_id"`
	Ports     []PortInfo `json:"ports"`
}

type PortInfo struct {
	PortID        string               `json:"port_id"`
	NetworkName   string               `json:"network_name"`
	NetworkPlanes string               `json:"network_planes"`
	FixIP         string               `json:"fix_ip"`
	FixedIPInfos  []agtmgr.FixedIPItem `json:"fixed_ip_items"`
}

func (mc *ManagerClient) CreateNeutronBulkPorts(reqID string, req *ManagerCreateBulkPortsReq, tenantID string) (CreatePortsResp, error) {
	defer errobj.RecoverPanic()
	resp := CreatePortsResp{}
	reqJSON, err := json.Marshal(&req)
	if err != nil {
		klog.Errorf("CreateNeutronBulkPorts: marshall http request body:%v failed, error: -%v", req, err)
		return resp, err
	}

	postPortURL := mc.GetCreatePortURL(tenantID) + MakeURLReqIDSuffix(reqID)
	klog.Infof("CreateNeutronBulkPorts: Post url: %s, body: %v", postPortURL, string(reqJSON))
	postStatusCode, rspByte, err := mc.PostBytes(postPortURL, reqJSON)
	if err != nil {
		klog.Errorf("CreateNeutronBulkPorts: masterClient.Post(posturl: %s, nil) error! -%v", postPortURL, err)
		return resp, err
	} else if !IsHttpMethodStatusSuccess(postStatusCode) {
		klog.Errorf("CreateNeutronBulkPorts: masterClient.Post(posturl: %v, %v) ok, but return status code is %v", postPortURL, string(reqJSON), postStatusCode)
		return resp, errors.New(errobj.GetErrMsg(rspByte))
	}

	err = json.Unmarshal(rspByte, &resp)
	if err != nil {
		klog.Errorf("CreateNeutronBulkPorts: json.Unmarshal failed, err: %v", err)
		return resp, err
	}
	return resp, nil
}

func (mc *ManagerClient) GetReportPodURL(tenantId, podName string) string {
	return mc.URLKnitterManager + "/pods/" + tenantId + "/" + podName
}
func (mc *ManagerClient) GetCreatePortURL(tenantID string) string {
	return mc.URLKnitterManager + "/port"
}

func (mc *ManagerClient) GetHealthURL() string {
	return mc.URLKnitterManager + "/tenants/admin" + "/health"
}

func (mc *ManagerClient) Get(postURL string) (int, []byte, error) {
	resp, err := HTTPGet(postURL)
	if err != nil {
		klog.Errorf("masterclient post error! -%v", err)
		return http.StatusInternalServerError, nil, err
	}
	defer HTTPClose(resp)
	body, err := HTTPReadAll(resp)
	if err != nil {
		klog.Errorf("ManagerClient.Get: HTTPReadAll(resp) error, error is [%v]", err)
		return http.StatusInternalServerError, nil, err
	}
	statuscode := resp.StatusCode
	return statuscode, body, nil
}

func (mc *ManagerClient) GetDeletePortURL(tenantID, portID string) string {
	return mc.GetCreatePortURL(tenantID) + "/" + portID
}

func (mc *ManagerClient) PostBytes(postURL string, postData []byte) (int, []byte, error) {

	contentType := "application/json"
	resp, err := HTTPPost(postURL, contentType, postData)
	klog.Infof("ManagerClient.PostBytes:postURL is [%v]", postURL)
	if err != nil {
		klog.Errorf("masterclient post error! -%v", err)
		return http.StatusInternalServerError, nil, err
	}
	defer HTTPClose(resp)
	body, _ := HTTPReadAll(resp)
	statuscode := resp.StatusCode
	return statuscode, body, nil
}

func (mc *ManagerClient) Delete(deleteURL string) (b []byte, statusCode int, e error) {
	resp, err := HTTPDelete(deleteURL)
	if err != nil {
		klog.Errorf("##masterclient delete http.DefaultClient.Do error! -%v", err)
		return nil, resp.StatusCode, err
	}
	defer HTTPClose(resp)
	body, _ := HTTPReadAll(resp)
	return body, resp.StatusCode, nil
}

func (mc *ManagerClient) CheckKnitterManager() error {
	getURL := mc.GetHealthURL()
	klog.Infof("CheckKnitterManager: Get url: %s", getURL)
	getStatusCode, healthByte, err := mc.Get(getURL)
	if err != nil {
		klog.Errorf("CheckKnitterManager: masterClient.Get(getUrl: %s) error! -%v", getURL, err)
		return fmt.Errorf("%v:CheckKnitterManager: masterClient.Get return error", err)
	}
	if getStatusCode != 200 {
		klog.Errorf("CheckKnitterManager: masterClient.Get(getUrl: %v) ok, but return status code is %v",
			getURL, getStatusCode)
		return errors.New("checkKnitterManager: masterClient.Get return status code is not 200 error")
	}
	healthJSON, err := jason.NewObjectFromBytes(healthByte)
	if err != nil {
		klog.Errorf("CheckKnitterManager: jason.NewObjectFromBytes(healthByte) error, err is %v",
			err)
		return fmt.Errorf("%v:CheckKnitterManager: jason.NewObjectFromBytes(healthByte) error", err)
	}
	healthState, err := healthJSON.GetString("state")
	if err != nil {
		klog.Errorf("CheckKnitterManager: healthJson.GetString(state) error, err is %v",
			err)
		return fmt.Errorf("%v:CheckKnitterManager: healthJson.GetString(state) error", err)
	}
	if healthState != "good" {
		klog.Errorf("CheckKnitterManager: Knitter-manager is not service state!")
		return errors.New("checkKnitterManager: Knitter-manager is not service state")
	}
	klog.Info("CheckKnitterManager Successful!")
	return nil
}

func (mc *ManagerClient) DeleteNeutronPort(tenantID string, portID string) (e error) {
	defer errobj.RecoverPanic()
	deleteurl := mc.GetDeletePortURL(tenantID, portID) + MakeURLReqIDSuffix(portID)
	klog.Debugf("DeleteNeutronPort: delete url is: %s", deleteurl)
	bRsp, statusCode, err := mc.Delete(deleteurl)
	if err != nil {
		klog.Errorf("DeleteNeutronPort: masterClient.Delete(deleteurl: %s) error ! err is %v", deleteurl, err)
		return err
	}
	if statusCode == http.StatusNotFound {
		klog.Infof("DeleteNeutronPort: masterClient.Delete statusCode is :%d", statusCode)
		return nil
	}
	if !IsHttpMethodStatusSuccess(statusCode) {
		klog.Errorf("DeleteNeutronPort: masterClient.Delete(deleteurl: %s) error! statscode: %v  response data: %v", deleteurl, statusCode, string(bRsp))
		return fmt.Errorf("DeleteNeutronPort: masterClient.Delete error! statusCode: %v  response data: %v", statusCode, string(bRsp))
	}

	klog.Infof("DeleteNeutronPort: delete url: %s response data: %v statusCode: %d", deleteurl, string(bRsp), statusCode)
	return nil
}

var HTTPGet = func(url string) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 1; i < MaxRetryTimesForHTTPReq; i++ {
		resp, err = get(url)
		if err == nil {
			return resp, nil
		}
		time.Sleep(time.Second * time.Duration(i))
	}
	return resp, err
}

var HTTPDelete = func(url string) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 1; i < MaxRetryTimesForHTTPReq; i++ {
		resp, err = Delete(url)
		if err == nil {
			return resp, nil
		}
		time.Sleep(time.Second * time.Duration(i))
	}
	return resp, err
}

var Delete = func(url string) (resp *http.Response, err error) {
	client := &http.Client{Timeout: constvalue.HTTPDefaultTimeoutInSec * time.Second}
	req, errreq := http.NewRequest("DELETE", url, nil)
	if errreq != nil {
		klog.Errorf("http.NewRequest(DELETE,url,nil) error! -%v", errreq)
		return nil, errreq
	}

	resp, errresp := client.Do(req)
	if errresp != nil {
		klog.Errorf("http.DefaultClient.Do(req) error! -%v", errresp)
		return nil, errresp
	}
	klog.Debugf("HttpDelete success! errreq and errresp are nil")
	return resp, nil
}

func get(url string) (*http.Response, error) {
	client := &http.Client{Timeout: constvalue.HTTPDefaultTimeoutInSec * time.Second}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		klog.Error("Get: http NewRequest error: ", err.Error())
		return nil, fmt.Errorf("%v:http NewRequest error", err)
	}

	response, err := client.Do(request)
	if err != nil {
		klog.Error("Get: client.Do error: ", err.Error())
		return nil, fmt.Errorf("%v:http client.Do error", err)
	}
	return response, err
}

func IsHttpMethodStatusSuccess(statusCode int) bool {
	if statusCode >= 200 && statusCode < 300 {
		return true
	}
	return false
}

//func Post(url string, contentType string, body interface{}) ([]byte, error) {
func post(url string, contentType string, postData []byte) (*http.Response, error) {
	client := &http.Client{Timeout: constvalue.HTTPDefaultTimeoutInSec * time.Second}
	body := bytes.NewReader(postData)
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		klog.Error("Post: http NewRequest error: ", err.Error())
		return nil, fmt.Errorf("%v:http NewRequest error", err)
	}

	request.Header.Set("Content-Type", contentType)
	response, err := client.Do(request)
	if err != nil {
		klog.Error("Post: client.Do error: ", err.Error())
		return nil, errors.New("http client.Do error")
	}

	return response, nil
}

func NewGUID(podID string) string {
	return podID
}

var HTTPClose = func(resp *http.Response) error {
	return resp.Body.Close()
}

var HTTPReadAll = func(resp *http.Response) ([]byte, error) {

	return ioutil.ReadAll(resp.Body)
}

const MaxRetryTimesForHTTPReq = 6

var HTTPPost = func(url string, bodyType string, postData []byte) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 1; i < MaxRetryTimesForHTTPReq; i++ {
		resp, err = post(url, bodyType, postData)
		if err == nil {
			return resp, nil
		}
		time.Sleep(time.Second * time.Duration(i))
	}
	return resp, err
}

func MakeURLReqIDSuffix(reqID string) string {
	return "?" + "req_id=" + reqID
}

type ManagerCreateBulkPortsReq struct {
	Ports []ManagerCreatePortReq `json:"ports"`
}

type ManagerCreatePortReq struct {
	TenantID    string `json:"tenant_id"`
	NetworkName string `json:"network_name"`
	PortName    string `json:"port_name"`
	VnicType    string `json:"vnic_type"` // only used by physical port create-attach, logical port create ignore it
	NodeID      string `json:"node_id"`   // node id which send request
	PodNs       string `json:"pod_ns"`
	PodName     string `json:"pod_name"`
	FixIP       string `json:"ip_addr"`
	ClusterID   string `json:"cluster_id"`
	IPGroupName string `json:"ip_group_name"`
}

type CreatePortInfo struct {
	Name         string               `json:"name"`
	NetworkID    string               `json:"network_id"`
	MacAddress   string               `json:"mac_address"`
	FixedIps     []ports.IP           `json:"fixed_ips"`
	FixedIPInfos []agtmgr.FixedIPItem `json:"fixed_ip_items"`
	Cidr         string               `json:"cidr"`
	PortID       string               `json:"id"`
}

type CreatePortsResp struct {
	Ports []CreatePortInfo `json:"ports"`
}

func (mc *ManagerClient) GetDefaultNetWork(tenantID string) (string, error) {
	url := mc.GetDefaultNetworkURL(tenantID)
	statusCode, netInfoByte, err := mc.Get(url)
	if err != nil || statusCode != 200 {
		klog.Errorf("DefaultNetworkRole: get network from knitter-manager error! %v", err)
		return "", err
	}

	networkInfoJSON, err := jason.NewObjectFromBytes(netInfoByte)
	if err != nil {
		klog.Errorf("DefaultNetworkRole:Get:jason.NewObjectFromBytes err: %v, networkInfoByte: %v",
			err, netInfoByte)
		return "", errobj.ErrJasonNewObjectFailed
	}

	networkName, err := networkInfoJSON.GetString("name")
	if err != nil {
		klog.Errorf("DefaultNetworkRole: get network name error! %v", err)
		return "", errobj.ErrJasonGetStringFailed
	}
	return networkName, nil
}

func (mc ManagerClient) GetDefaultNetworkURL(tenantID string) string {
	return mc.URLKnitterManager + "/networks/placeholder" + "?default=true"
}
