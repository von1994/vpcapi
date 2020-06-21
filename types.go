package vpcapi

const (
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
