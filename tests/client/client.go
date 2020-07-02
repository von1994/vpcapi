package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"

	"github.com/onionpiece/vpcapi"
)

func main() {
	data, err := ioutil.ReadFile("./config")
	if err != nil {
		panic(err)
	}
	conf := vpcapi.VPC{}
	if err := json.Unmarshal(data, &conf); err != nil {
		panic(err)
	}
	if len(os.Args) < 2 {
		fmt.Println("not enough parameters")
		fmt.Println("getInterfaceByIP <podIP/interfaceIP>\nallocateIP <nodeIP>\nreleaseIP <interfaceID> <podIP>\nmigrateIP <podIP> <oldInterfaceID> <newInterfaceID>")
		fmt.Println("getInterfaces")
		return
	}
	switch os.Args[1] {
	case "getInterfaceByIP":
		{
			ip := os.Args[2]
			intf, err := vpcapi.GetInterfaceByIP(conf, ip)
			if err != nil {
				panic(err)
			}
			fmt.Println(intf)
			return
		}
	case "allocateIP":
		{
			nodeIP := os.Args[2]
			instanceID, err := vpcapi.GetInstanceID(conf, nodeIP)
			if err != nil {
				panic(err)
			}
			fmt.Printf("instanceID: %s\n", instanceID)
			interfaces, err := vpcapi.GetInstanceInterfaces(conf, instanceID)
			if err != nil {
				panic(err)
			}
			chosenIdx := vpcapi.PickInterface(conf, interfaces)
			fmt.Printf("chosen interface: %s\n", interfaces[chosenIdx].NetworkInterfaceID)
			originIPs := []string{}
			for _, ip := range interfaces[chosenIdx].PrivateIPAddressSet {
				originIPs = append(originIPs, ip.PrivateIPAddress)
			}
			fmt.Printf("origin ips: %v\n", originIPs)
			if err := vpcapi.AssignIP(conf, interfaces[chosenIdx].NetworkInterfaceID); err != nil {
				panic(err)
			}
			ips, _ := vpcapi.GetInterfaceIPs(conf, interfaces[chosenIdx].NetworkInterfaceID)
			fmt.Printf("ips: %v\n", ips)
			for _, ip := range ips {
				exists := false
				for _, _ip := range originIPs {
					if ip == _ip {
						exists = true
					}
				}
				if !exists {
					fmt.Printf("New IP: %s, instanceID: %s, MAC: %s\n",
						ip, interfaces[chosenIdx].NetworkInterfaceID, interfaces[chosenIdx].MacAddress)
				}
			}
		}
	case "releaseIP":
		{
			interfaceID := os.Args[2]
			podIP := os.Args[3]
			if err := vpcapi.ReleaseIP(conf, interfaceID, podIP); err != nil {
				panic(err)
			}
		}
	case "migrateIP":
		{
			podIP := os.Args[2]
			oldInterfaceID := os.Args[3]
			newInterfaceID := os.Args[4]
			if err := vpcapi.MigrateIP(conf, podIP, oldInterfaceID, newInterfaceID); err != nil {
				panic(err)
			}
		}
	case "getInterfaces":
		{
			interfaces, err := vpcapi.GetInterfaces(conf)
			if err != nil {
				panic(err)
			}
			for _, intf := range interfaces {
				primaryIP := ""
				secondaryIPs := []string{}
				for _, ip := range intf.PrivateIPAddressSet {
					if ip.Primary {
						primaryIP = ip.PrivateIPAddress
					} else {
						secondaryIPs = append(secondaryIPs, ip.PrivateIPAddress)
					}
				}
				fmt.Printf("InterfaceID:%s\tMAC:%s\tPrimaryIP:%s\n", intf.NetworkInterfaceID, intf.MacAddress, primaryIP)
				fmt.Printf("\tsecondaryIPs: %v\n", secondaryIPs)
			}
		}
	}
}
