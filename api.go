package vpcapi

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// PickInterface pick an interface from given interfaces with policy
func PickInterface(conf VPC, interfaces []DescribeInterfacesNetworkInterface) int {
	ipCount := 1000
	choseIndex := -1
	for idx, intf := range interfaces {
		if conf.Policy == VPCPolicyExclusive {
			if intf.Primary {
				continue
			}
			if len(intf.PrivateIPAddressSet) == 1 {
				return idx
			}
		} else {
			if conf.Policy != VPCPolicySharePrimary && intf.Primary {
				continue
			}
			curIPCount := len(intf.PrivateIPAddressSet)
			if curIPCount < ipCount {
				ipCount = curIPCount
				choseIndex = idx
			}
		}
	}
	if choseIndex >= 0 {
		log.Printf("VPC.API: chose interface %v", interfaces[choseIndex])
	}
	return choseIndex
}

// GetInstanceID get CVM instance ID based on given nodeIP
func GetInstanceID(conf VPC, nodeIP string) (string, error) {
	params := map[string]string{
		"Action":             "DescribeInstances",
		"Limit":              "1",
		"Version":            "2017-03-12",
		"Service":            "cvm",
		"private-ip-address": nodeIP,
	}
	resp, err := doRequest(conf, params)
	if err != nil {
		return "", err
	}
	cvmResp := &CVMDescribeInstancesResponse{}
	if err = json.Unmarshal(resp, cvmResp); err != nil {
		return "", err
	}
	instanceID := cvmResp.Body.InstanceSet[0].InstanceID
	return instanceID, nil
}

func getBaseParams(action, vpcID string) map[string]string {
	return map[string]string{
		"Action":  action,
		"Version": "2017-03-12",
		"Service": "vpc",
		"vpcId":   vpcID,
	}
}

func getDescribeNetworkInterfacesParams(vpcID string) map[string]string {
	return getBaseParams("DescribeNetworkInterfaces", vpcID)
}

// GetInterface get cvm network interface by given interface ID
func GetInterface(conf VPC, interfaceID string) (*DescribeInterfacesNetworkInterface, error) {
	params := getDescribeNetworkInterfacesParams(conf.VPCID)
	params["networkInterfaceId"] = interfaceID
	resp, err := doRequest(conf, params)
	if err != nil {
		return nil, err
	}
	describeInterfacesResp := &DescribeInterfacesResponse{}
	if err = json.Unmarshal(resp, describeInterfacesResp); err != nil {
		return nil, err
	}
	log.Printf("VPC.API: getInterface for %s: %v\n", interfaceID, describeInterfacesResp)
	if len(describeInterfacesResp.Data.Data) == 0 {
		return nil, nil
	}
	return &describeInterfacesResp.Data.Data[0], nil
}

// GetInterfaces get all network interfaces on the vpc
func GetInterfaces(conf VPC) ([]DescribeInterfacesNetworkInterface, error) {
	params := getDescribeNetworkInterfacesParams(conf.VPCID)
	resp, err := doRequest(conf, params)
	if err != nil {
		return nil, err
	}
	describeInterfacesResp := &DescribeInterfacesResponse{}
	if err = json.Unmarshal(resp, describeInterfacesResp); err != nil {
		return nil, err
	}
	return describeInterfacesResp.Data.Data, nil
}

// GetInterfaceIPs get network interface IPs by given interface ID
func GetInterfaceIPs(conf VPC, interfaceID string) ([]string, error) {
	intf, err := GetInterface(conf, interfaceID)
	if err != nil {
		return nil, err
	}
	ips := []string{}
	for _, ip := range intf.PrivateIPAddressSet {
		if !ip.Primary {
			ips = append(ips, ip.PrivateIPAddress)
		}
	}
	return ips, nil
}

// GetInterfaceByIP get network interface by given interface IP
func GetInterfaceByIP(conf VPC, ip string) (*DescribeInterfacesNetworkInterface, error) {
	params := getDescribeNetworkInterfacesParams(conf.VPCID)
	resp, err := doRequest(conf, params)
	if err != nil {
		return nil, err
	}
	describeInterfacesResp := &DescribeInterfacesResponse{}
	if err = json.Unmarshal(resp, describeInterfacesResp); err != nil {
		return nil, err
	}
	for _, intf := range describeInterfacesResp.Data.Data {
		for _, intfIP := range intf.PrivateIPAddressSet {
			if intfIP.PrivateIPAddress == ip {
				return &intf, nil
			}
		}
	}
	return nil, nil
}

// GetInstanceInterfaces get instance network interfaces by given instance ID
func GetInstanceInterfaces(conf VPC, instanceID string) ([]DescribeInterfacesNetworkInterface, error) {
	params := getDescribeNetworkInterfacesParams(conf.VPCID)
	params["instanceId"] = instanceID
	resp, err := doRequest(conf, params)
	if err != nil {
		return nil, err
	}
	describeInterfacesResp := &DescribeInterfacesResponse{}
	if err = json.Unmarshal(resp, describeInterfacesResp); err != nil {
		return nil, err
	}
	log.Printf("VPC.API: getInstanceInterfaces for %s: %v\n", instanceID, describeInterfacesResp)
	return describeInterfacesResp.Data.Data, nil
}

func assignInferfaceSecondaryIP(conf VPC, interfaceID string) error {
	params := getBaseParams("AssignPrivateIpAddresses", conf.VPCID)
	params["networkInterfaceId"] = interfaceID
	params["secondaryPrivateIpAddressCount"] = "1"
	resp, err := doRequest(conf, params)
	if err != nil {
		return fmt.Errorf("VPC.API: assignInferfaceSecondaryIP doRequest failed with: %v", err)
	}
	assignResp := PrivateIPAddressesActionResponse{}
	if err := json.Unmarshal(resp, &assignResp); err != nil {
		return fmt.Errorf("VPC.API: assignInferfaceSecondaryIP failed to do json unmarshal, since: %v", err)
	}
	if assignResp.Code != 0 {
		return fmt.Errorf("VPC.API: assignInferfaceSecondaryIP response error, code %d, message %s", assignResp.Code, assignResp.Message)
	}
	return err
}
func releaseInterfaceSecondaryIP(conf VPC, interfaceID, podIP string) error {
	params := getBaseParams("UnassignPrivateIpAddresses", conf.VPCID)
	params["networkInterfaceId"] = interfaceID
	params["privateIpAddress.0"] = podIP
	resp, err := doRequest(conf, params)
	if err != nil {
		return fmt.Errorf("VPC.API: releaseInferfaceSecondaryIP doRequest failed with: %v", err)
	}
	releaseResp := PrivateIPAddressesActionResponse{}
	if err := json.Unmarshal(resp, &releaseResp); err != nil {
		return fmt.Errorf("VPC.API: releaseInferfaceSecondaryIP failed to do json unmarshal, since: %v", err)
	}
	if releaseResp.Code != 0 {
		return fmt.Errorf("VPC.API: releaseInferfaceSecondaryIP response error, code %d, message %s", releaseResp.Code, releaseResp.Message)
	}
	return err
}

func migrateInterfaceSecondaryIP(conf VPC, podIP, oldInterfaceID, newInterfaceID string) error {
	params := getBaseParams("MigratePrivateIpAddress", conf.VPCID)
	params["privateIpAddress"] = podIP
	params["oldNetworkInterfaceId"] = oldInterfaceID
	params["newNetworkInterfaceId"] = newInterfaceID
	resp, err := doRequest(conf, params)
	if err != nil {
		return fmt.Errorf("VPC.API: migrateInferfaceSecondaryIP doRequest failed with: %v", err)
	}
	migrateResp := PrivateIPAddressesActionResponse{}
	if err := json.Unmarshal(resp, &migrateResp); err != nil {
		return fmt.Errorf("VPC.API: migrateInferfaceSecondaryIP failed to do json unmarshal, since: %v", err)
	}
	if migrateResp.Code != 0 {
		return fmt.Errorf("VPC.API: migrateInferfaceSecondaryIP response error, code %d, message %s", migrateResp.Code, migrateResp.Message)
	}

	for i := 0; i != conf.IPMigrate.PostCheckRetry; i++ {
		intf, err := GetInterface(conf, newInterfaceID)
		if err != nil {
			return fmt.Errorf("VPC.API: migrateInferfaceSecondaryIP failed to getInterface to detect, since: %v", err)
		}
		for _, ip := range intf.PrivateIPAddressSet {
			if ip.PrivateIPAddress == podIP {
				return nil
			}
		}
		time.Sleep(time.Duration(conf.IPMigrate.PostCheckInterval) * time.Millisecond)
	}
	return fmt.Errorf("VPC.API: after %d * %dms detect, ip failed to migrate to new interface", conf.IPMigrate.PostCheckRetry, conf.IPMigrate.PostCheckInterval)
}

// AssignIP will invoke VPC API to assign an IP to given interface
func AssignIP(conf VPC, interfaceID string) error {
	// the API is weak, if we invoke it frequently, error like "您操作的资源正在执行其他操作，请稍后重试" will raise
	ok := false
	var err error
	for i := 0; i != conf.IPAssign.Retry; i++ {
		err = assignInferfaceSecondaryIP(conf, interfaceID)
		if err == nil {
			ok = true
			break
		}
		time.Sleep(time.Duration(conf.IPAssign.Interval) * time.Millisecond)
	}
	if !ok {
		return fmt.Errorf("VPC.API: failed to assign secondary IP for interface %s, since: %v", interfaceID, err)
	}
	return nil
}

// ReleaseIP will invoke VPC API to release IP on given interface
func ReleaseIP(conf VPC, interfaceID, podIP string) error {
	// the API is weak, if we invoke it frequently, error like "您操作的资源正在执行其他操作，请稍后重试" will raise
	ok := false
	var err error
	for i := 0; i != conf.IPRelease.Retry; i++ {
		err = releaseInterfaceSecondaryIP(conf, interfaceID, podIP)
		if err == nil {
			ok = true
			break
		}
		time.Sleep(time.Duration(conf.IPRelease.Interval) * time.Millisecond)
	}
	if !ok {
		return fmt.Errorf("VPC.API: failed to release IP %s on %s, since: %v", podIP, interfaceID, err)
	}
	return nil
}

func CheckMigrateIPStatus(conf VPC, podIP, oldInterfaceID, newInterfaceID string) error {
	intf, err := GetInterfaceByIP(conf, podIP)
	if err != nil {
		return fmt.Errorf("VPC.API: getInterfaceByIP doRequest failed with: %v", err)
	}
	if intf.NetworkInterfaceID == oldInterfaceID {
		return fmt.Errorf("VPC.API: Migrate IP %s failed, retry once more", podIP)
	} else if intf.NetworkInterfaceID == newInterfaceID {
		return nil
	} else {
		return fmt.Errorf("VPC.API: IP not between in %s and %s", oldInterfaceID, newInterfaceID)
	}
}

// MigrateIP will invoke VPC API to migrate IP from old interface to new interface
func MigrateIP(conf VPC, ip, oldInterfaceID, newInterfaceID string) error {
	// the API is weak, if we invoke it frequently, error like "您操作的资源正在执行其他操作，请稍后重试" will raise
	ok := false
	var err error
	for i := 0; i != conf.IPMigrate.Retry; i++ {
		if i > 0 {
			err = CheckMigrateIPStatus(conf, ip, oldInterfaceID, newInterfaceID)
			if err == nil {
				log.Printf("VPC.API: at No.%s retry, migrate IP already done.", strconv.Itoa(i))
				ok = true
				break
			}
		}
		err = migrateInterfaceSecondaryIP(conf, ip, oldInterfaceID, newInterfaceID)
		if err == nil {
			ok = true
			break
		}
		time.Sleep(time.Duration(conf.IPMigrate.Interval) * time.Millisecond)
	}
	if !ok {
		return fmt.Errorf("VPC.API: failed to migrate Pod IP %s between intefaces, %s => %s, since: %v", ip, oldInterfaceID, newInterfaceID, err)
	}
	return nil
}
