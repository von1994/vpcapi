package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	vpc "github.com/onionpiece/vpcapi"
)

type Node struct {
	IP   string `json:"ip"`
	Name string `json:"name"`
}
type Instances struct {
	Instances []Node `json:"instances"`
}

var (
	vpcID      = "foo"
	interfaces vpc.DescribeInterfacesResponseData
	instances  Instances
	ipPool     = make(map[string]bool)
)

func main() {
	if data, err := ioutil.ReadFile("./interfaces.json"); err != nil {
		panic(err)
	} else if err := json.Unmarshal(data, &interfaces); err != nil {
		panic(err)
	}
	if data, err := ioutil.ReadFile("./instances.json"); err != nil {
		panic(err)
	} else if err := json.Unmarshal(data, &instances); err != nil {
		panic(err)
	}
	for i := 17; i != 255; i++ {
		ipPool[fmt.Sprintf("192.168.144.%d", i)] = false
	}

	http.HandleFunc("/", dispatch)

	log.Println("** Service Started on Port 8443 **")

	// Use ListenAndServeTLS() instead of ListenAndServe() which accepts two extra parameters.
	// We need to specify both the certificate file and the key file (which we've named
	// https-server.crt and https-server.key).
	if err := http.ListenAndServeTLS(":8443", "https-server.crt", "https-server.key", nil); err != nil {
		log.Fatal(err)
	}
}

func dispatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	url := r.URL.String()
	if strings.Contains(url, "DescribeNetworkInterfaces") {
		if strings.Contains(url, "instanceId") {
			getInstanceInterfaces(w, url)
		} else if strings.Contains(url, "networkInterfaceId") {
			getInterfaceByID(w, url)
		} else {
			getAllInterfaces(w)
		}
	} else if strings.Contains(url, "DescribeInstances") {
		getInstance(w, url)
	} else if strings.Contains(url, "AssignPrivateIpAddresses") {
		assigneIP(url, "", "")
		writeResponseCode(w)
	} else if strings.Contains(url, "UnassignPrivateIpAddresses") {
		releaseIP(url, "", "")
		writeResponseCode(w)
	} else if strings.Contains(url, "MigratePrivateIpAddress") {
		migrateIP(url)
		writeResponseCode(w)
	} else {
		fmt.Println(url)
	}
	return
}

func writeResponseCode(w http.ResponseWriter) {
	data := vpc.PrivateIPAddressesActionResponse{
		Data: vpc.PrivateIPAddressesActionResponseData{
			Code:     0,
			CodeDesc: "",
		},
		Message: "",
		Code:    0,
	}
	dataJSON, _ := json.Marshal(data)
	io.WriteString(w, string(dataJSON))
	return
}

func getAllInterfaces(w http.ResponseWriter) {
	data := &vpc.DescribeInterfacesResponse{
		Code:     0,
		CodeDesc: "",
		Data:     interfaces,
	}
	dataJSON, _ := json.Marshal(data)
	io.WriteString(w, string(dataJSON))
	return
}

func getInstance(w http.ResponseWriter, url string) {
	instanceID := ""
	ip := getURLValue(url, "Filters.0.Values.0")
	for _, instance := range instances.Instances {
		if instance.IP == ip {
			instanceID = instance.Name
			break
		}
	}
	fmt.Printf("getInstance: ip %s instanceID %s", ip, instanceID)
	data := &vpc.CVMDescribeInstancesResponse{
		Body: vpc.CVMDescribeInstancesResponseBody{
			TotalCount:  1,
			InstanceSet: []vpc.CVMDescribeInstancesInstance{},
		},
	}
	data.Body.InstanceSet = append(data.Body.InstanceSet, vpc.CVMDescribeInstancesInstance{InstanceID: instanceID})
	dataJSON, _ := json.Marshal(data)
	io.WriteString(w, string(dataJSON))
	return
}

func getURLValue(url, key string) string {
	ifName := ""
	for _, sub := range strings.Split(strings.Split(url, "?")[1], "&") {
		if strings.HasPrefix(sub, key) {
			ifName = strings.Split(sub, "=")[1]
			break
		}
	}
	return ifName
}

func getIfName(url string) string {
	return getURLValue(url, "networkInterfaceId")
}

func getInstanceName(url string) string {
	return getURLValue(url, "instanceId")
}

func getIP(url string) string {
	return getURLValue(url, "privateIpAddress.0")
}

func assigneIP(url, ifName, ip string) {
	if ifName == "" {
		ifName = getIfName(url)
	}
	if ip == "" {
		for _ip, used := range ipPool {
			if !used {
				ipPool[_ip] = true
				ip = _ip
				break
			}
		}
	}
	for idx := range interfaces.Data {
		if interfaces.Data[idx].NetworkInterfaceID == ifName {
			interfaces.Data[idx].PrivateIPAddressSet = append(interfaces.Data[idx].PrivateIPAddressSet, vpc.DescribeInterfacesPrivateIPAddresses{Primary: false, PrivateIPAddress: ip})
			fmt.Printf("After assign: %v\n", interfaces.Data[idx].PrivateIPAddressSet)
			break
		}
	}
}

func releaseIP(url, ifName, ip string) {
	if ip == "" {
		ip = getIP(url)
		ipPool[ip] = false
	}
	if ifName == "" {
		ifName = getIfName(url)
	}
	for idx := range interfaces.Data {
		if interfaces.Data[idx].NetworkInterfaceID == ifName {
			ipIdx := 0
			for _idx := range interfaces.Data[idx].PrivateIPAddressSet {
				if interfaces.Data[idx].PrivateIPAddressSet[_idx].PrivateIPAddress == ip {
					ipIdx = _idx
					break
				}
			}
			interfaces.Data[idx].PrivateIPAddressSet = append(interfaces.Data[idx].PrivateIPAddressSet[:ipIdx], interfaces.Data[idx].PrivateIPAddressSet[ipIdx+1:len(interfaces.Data[idx].PrivateIPAddressSet)]...)
			fmt.Printf("After release: %v\n", interfaces.Data[idx].PrivateIPAddressSet)
			break
		}
	}
}

func migrateIP(url string) {
	ip := getURLValue(url, "privateIpAddress")
	oldIfName := getURLValue(url, "oldNetworkInterfaceId")
	newIfName := getURLValue(url, "newNetworkInterfaceId")
	releaseIP(url, oldIfName, ip)
	assigneIP(url, newIfName, ip)
}

func getInterfaceByID(w http.ResponseWriter, url string) {
	ifName := getIfName(url)
	data := &vpc.DescribeInterfacesResponse{
		Code:     0,
		CodeDesc: "",
		Data: vpc.DescribeInterfacesResponseData{
			Data: []vpc.DescribeInterfacesNetworkInterface{},
		},
	}
	for _, intf := range interfaces.Data {
		if intf.NetworkInterfaceID == ifName {
			data.Data.Data = append(data.Data.Data, intf)
			break
		}
	}
	dataJSON, _ := json.Marshal(data)
	io.WriteString(w, string(dataJSON))
	return
}

func getInstanceInterfaces(w http.ResponseWriter, url string) {
	instName := getInstanceName(url)
	data := &vpc.DescribeInterfacesResponse{
		Code:     0,
		CodeDesc: "",
		Data: vpc.DescribeInterfacesResponseData{
			Data: []vpc.DescribeInterfacesNetworkInterface{},
		},
	}
	for _, intf := range interfaces.Data {
		if strings.HasPrefix(intf.NetworkInterfaceID, instName) {
			data.Data.Data = append(data.Data.Data, intf)
		}
	}
	fmt.Printf("getInstanceInterfaces will response interfaces: %v\n", data.Data.Data)
	dataJSON, _ := json.Marshal(data)
	io.WriteString(w, string(dataJSON))
	return
}
