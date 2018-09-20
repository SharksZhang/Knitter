package constvalue

const (
	ComponentName                    = "nwmonitor"
	DefaultMonitorLogDir             = "/root/info/logs/nwmonitor"
	DefaultKubeconfig                = "/etc/kubernetes/kubectl.kubeconfig"
	DefaultConfDir                   = "./"
	DefaultConfFile                  = "knitter.json"
	DefaultWorkerNumber              = 20
	DefaultBeegoConfPath             = "conf/app.conf"
	DefaultPortRecycleInterval       = 120
	DefaultPodSyncInterval           = 120
	GetLoadReourceRetryIntervalInSec = 30
	EtcdRetryTimes                   = 30
)

const (
	GetPodStatusIntervalTime = 10
)

const (
	WaitForCacheSyncTimes = 100
)
const (
	LogicalPortDefaultVnicType = "normal"
)
const HTTPDefaultTimeoutInSec = 60 // default http GET/POST request timeout in second

const PaaSTenantAdminDefaultUUID = "admin"
const (
	reportPodURL = "/po"
)

const (
	NetPlaneStd     = "std"
	NetPlaneEio     = "eio"
	NetPlaneControl = "control"
	NetPlaneMedia   = "media"
	NetPlaneOam     = "oam"
)

const (
	DefaultNetworkPlane = "std"
	DefaultPortName     = "eth0"
	DefaultVnicType     = "normal"
	DefaultIsAccelerate = "false"
)

const (
	CreateOperation = "create"
	DeleteOperation = "delete"
	CreatingState   = "creating"
	RunningState    = "running"
)

const (
	MechDriverOvs      = "normal"
	MechDriverSriov    = "direct"
	MechDriverPhysical = "physical"
)

const (
	TypeReplicationController = "replicationcontroller"
	TypeReplicaSet            = "replicaset"
	TypeStatefulSet           = "statefulset"
)

const (
	AddAction = "add"
	DelAction = "del"
)
