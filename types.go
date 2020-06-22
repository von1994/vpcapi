package vpcapi

const (
	// AnnoKeyVPCIPAM is used to enable IPAM for VPC.  To enable it set value to true-like.
	// TODO: it's necessary to add support for VPC in IPClaim CRD, since user may want to known IPs before they
	// deploy Pods.
	AnnoKeyVPCIPAM = "alcor.io/vpc-cni.ipam"
	// AnnoKeyVPCIPRetain tells cni not to call VPC api to release IP when pod deleted, mostly beside of enabled key
	// this is the only one can set in annotations for a new created pod
	AnnoKeyVPCIPRetain = "alcor.io/vpc-cni.ipRetain"
	// AnnoKeyVPCIP records IP allocated from VPC, cni will patch IP with this key to pod annotations.
	// IF a controller want to make pod IP retained, beside of setting key "ipRetain" to true-like, the contoller
	// itself should watch this key, and records it value. So when pod is deleted by accident, controller can create
	// a new pod with previous IP with this key as a pod annotation.
	AnnoKeyVPCIP = "alcor.io/vpc-cni.ip"
	// AnnoKeyVPCNICMAC records MAC address of network interface on which pod traffic will go through.
	// Since, inside of CVM, MAC address is the only identity to locate an interface, so this key is used by CNI itself.
	AnnoKeyVPCNICMAC = "alcor.io/vpc-cni.nicMAC"
	// AnnoKeyVPCNICID records ID of network interface on which pod traffic will go through, overlay APIs like describe
	// interface, assign IP, release IP, migrate IP will use interface ID. Only useful to CNI itself.
	AnnoKeyVPCNICID = "alcor.io/vpc-cni.nicID"
	// AnnoKeyVPCInstanceID records which CVM pod is running on. This key will help CNI to determine whether "pod migration"
	// happend, which means old pod is deleted by any reason, while a new pod with the same IP and hostname created to
	// replace the deleted one. For case, if controller try to make the new created pod scheduled to the same node, it's OK,
	// since, CNI will find nothing changed; but if pod is scheduled to another node, on that node, CNI will find that
	// instanceID in pod annotations not match instanceID of node, which means it's "pod migration". So for "pod migration",
	// CNI need to call overlay API to do ip migration.
	AnnoKeyVPCInstanceID = "alcor.io/vpc-cni.instanceID"

	// VPCPolicyExclusive policy will make sure each pod have a separate CVM network interface to use
	VPCPolicyExclusive = "Exclusive"
	// VPCPolicyShare policy as default policy, will make pods share CVM network interfaces with each other, except the primary interface
	VPCPolicyShare = "Share"
	// VPCPolicySharePrimary will make pods share CVM network interfaces with each other, including the primary interface
	VPCPolicySharePrimary = "SharePrimary"
)

// CVMDescribeInstancesInstance is member of "InstanceSet" in response body of cmv reqeust DescribeInstances
type CVMDescribeInstancesInstance struct {
	InstanceID string `json:"InstanceId"`
}

// CVMDescribeInstancesResponseBody is "Response" in response body of cvm request DescribeInstances
type CVMDescribeInstancesResponseBody struct {
	TotalCount  int                            `json:"TotalCount"`
	InstanceSet []CVMDescribeInstancesInstance `json:"InstanceSet"`
}

// CVMDescribeInstancesResponse is response of cvm request DescribeInstances
type CVMDescribeInstancesResponse struct {
	Body CVMDescribeInstancesResponseBody `json:"Response"`
}

// DescribeInterfacesPrivateIPAddresses is member of data.data.privateIpAddressesSet in response of vpc request DescribeNetworkInterfaces
type DescribeInterfacesPrivateIPAddresses struct {
	Primary          bool   `json:"primary"`
	PrivateIPAddress string `json:"privateIpAddress"`
}

// DescribeInterfacesInstance is data.data.instanceSet in response of vpc request DescribeNetworkInterfaces
type DescribeInterfacesInstance struct {
	InstanceID string `json:"instanceId"`
}

// DescribeInterfacesNetworkInterface is member of data.data in response of vpc request DescribeNetworkInterfaces
type DescribeInterfacesNetworkInterface struct {
	Instance            DescribeInterfacesInstance             `json:"instanceSet"`
	MacAddress          string                                 `json:"macAddress"`
	NetworkInterfaceID  string                                 `json:"networkInterfaceId"`
	Primary             bool                                   `json:"primary"`
	PrivateIPAddressSet []DescribeInterfacesPrivateIPAddresses `json:"privateIpAddressesSet"`
	SubnetID            string                                 `json:"subnetId"`
	VpcID               string                                 `json:"vpcId"`
	VpcName             string                                 `json:"vpcName"`
}

// DescribeInterfacesResponseData is data.data of response of vpc request DescribeNetworkInterfaces
type DescribeInterfacesResponseData struct {
	Data []DescribeInterfacesNetworkInterface `json:"data"`
}

// DescribeInterfacesResponse is response of vpc request DescribeNetworkInterfaces
type DescribeInterfacesResponse struct {
	Code     int                            `json:"code"`
	CodeDesc string                         `json:"codeDesc"`
	Data     DescribeInterfacesResponseData `json:"data"`
}

// PrivateIPAddressesActionResponseData is data of response of vpc request AssignPrivateIpAddresses
type PrivateIPAddressesActionResponseData struct {
	Code     int    `json:"code"`
	CodeDesc string `json:"codeDesc"`
}

// PrivateIPAddressesActionResponse is response of vpc request AssignPrivateIpAddresses
type PrivateIPAddressesActionResponse struct {
	Data    PrivateIPAddressesActionResponseData `json:"data"`
	Message string                               `json:"message"`
	Code    int                                  `json:"code"`
}

// IPAssign defines parameters for ip assignment API for vpc
type IPAssign struct {
	Retry    int `json:"retry,omitempty"`
	Interval int `json:"interval,omitempty"`
}

// IPRelease defines parameters for ip unassignment API for vpc
type IPRelease struct {
	Retry    int `json:"retry,omitempty"`
	Interval int `json:"interval,omitempty"`
}

// IPMigrate defines parameters for ip migration API for vpc
type IPMigrate struct {
	Retry             int `json:"retry,omitempty"`
	Interval          int `json:"interval,omitempty"`
	PostCheckRetry    int `json:"postCheckRetry,omitempty"`
	PostCheckInterval int `json:"postCheckInterval,omitempty"`
}

// IPDetect defines struct for detect ip for vpc
type IPDetect struct {
	Delay    int `json:"delay,omitempty"`
	Retry    int `json:"retry,omitempty"`
	Interval int `json:"interval,omitempty"`
}

// VPC defines struct for vpc, the cni for TX Cloud overlay
type VPC struct {
	SecretID       string    `json:"secretID"`
	SecretKey      string    `json:"secretKey"`
	Region         string    `json:"region"`
	VPCID          string    `json:"vpcID"`
	MTU            int       `json:"MTU"`
	Policy         string    `json:"policy,omitempty"`
	CVMAPIVersion  string    `json:"cvmAPIVersion"`
	VPCAPIVersion  string    `json:"vpcAPIVersion"`
	CVMAPIEndpoint string    `json:"cvmAPIEndpoint"`
	VPCAPIEndpoint string    `json:"vpcAPIEndpoint"`
	V2URI          string    `json:"v2URL"`
	V3URI          string    `json:"v3URL"`
	InstanceID     string    `json:"instanceID,omitempty"`
	NodeInterface  string    `json:"nodeInterface,omitempty"`
	NodeIfPrefix   string    `json:"nodeIfPrefix,omitempty"`
	IPAssign       IPAssign  `json:"ipAssign,omitempty"`
	IPRelease      IPRelease `json:"ipRelease,omitempty"`
	IPDetect       IPDetect  `json:"ipDetect,omitempty"`
	IPMigrate      IPMigrate `json:"ipMigrate,omitempty"`
}
