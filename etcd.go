package vpcapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
)

const (
	// ETCDV3VPCPODKEYPREFIX is key prefix for vpc cni to store pod info into etcd
	ETCDV3VPCPODKEYPREFIX = "/vpc/pods/"
	// ETCDV3VPCIPKEYPREFIX is key prefix for vpc cni to store ip info into etcd
	ETCDV3VPCIPKEYPREFIX = "/vpc/ips/"
)

var (
	etcdClientTimeout    = 10 * time.Second
	etcdKeepaliveTime    = 30 * time.Second
	etcdKeepaliveTimeout = 10 * time.Second
)

// Etcdv3Client stands for a client for etcdv3
type Etcdv3Client struct {
	Client *clientv3.Client
}

// PodInfo defines struct pod info about vpc
type PodInfo struct {
	IP          string `json:"ip"`
	InterfaceID string `json:"interfaceID"`
	IPRetain    string `json:"ipRetain"`
}

// IPInfo defines struct ip info on vpc
type IPInfo struct {
	Namespace string `json:"ns"`
	Name      string `json:"name"`
}

// NewEtcdv3Client create a new etcdv3 client based on given netconf
func NewEtcdv3Client(caCertFile, certFile, keyFile, etcdEndpoints string) (*Etcdv3Client, error) {
	etcdLocation := strings.Split(etcdEndpoints, ",")
	if len(etcdLocation) == 0 {
		return nil, fmt.Errorf("no etcd endpoints specified")
	}
	tlsInfo := &transport.TLSInfo{
		CAFile:   caCertFile,
		CertFile: certFile,
		KeyFile:  keyFile,
	}
	tls, err := tlsInfo.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not initialize etcdv3 client: %+v", err)
	}
	cfg := clientv3.Config{
		Endpoints:            etcdLocation,
		TLS:                  tls,
		DialTimeout:          etcdClientTimeout,
		DialKeepAliveTime:    etcdKeepaliveTime,
		DialKeepAliveTimeout: etcdKeepaliveTimeout,
	}
	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}

	// test clientv3 connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ops := []clientv3.OpOption{
		clientv3.WithPrefix(),
		clientv3.WithLimit(1),
	}
	if _, err := client.Get(ctx, "/", ops...); err != nil {
		return nil, err
	}

	return &Etcdv3Client{Client: client}, nil
}

func getIPKey(ip string) string {
	return fmt.Sprintf("%s%s", ETCDV3VPCIPKEYPREFIX, ip)
}

func getPodKey(namespace, name string) string {
	return fmt.Sprintf("%s%s.%s", ETCDV3VPCPODKEYPREFIX, namespace, name)
}

// GetPodInfo get pod info with by given namespace and pod name
func (c *Etcdv3Client) GetPodInfo(namespace, name string) (*PodInfo, error) {
	resp, err := c.Client.Get(context.Background(), getPodKey(namespace, name))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	pod := &PodInfo{}
	if err := json.Unmarshal(resp.Kvs[0].Value, pod); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal value for pod %s.%s, since: %v", namespace, name, err)
	}
	return pod, nil
}

func (c *Etcdv3Client) _put(key string, data []byte, override bool) ([]byte, error) {
	ops := []clientv3.Op{clientv3.OpGet(key)}
	if override {
		ops = append(ops, clientv3.OpPut(key, string(data)))
	}
	resp, err := c.Client.Txn(context.Background()).If(
		clientv3.Compare(clientv3.Version(key), "=", 0)).Then(
		clientv3.OpPut(key, string(data))).Else(ops...).Commit()
	if err != nil {
		return nil, err
	}
	if !resp.Succeeded {
		return ((*clientv3.GetResponse)(resp.Responses[0].GetResponseRange())).Kvs[0].Value, nil
	}
	return nil, nil
}

// PutPodInfo will put pod info into etcd by given namespace, pod name and IP, interfaceID on which interface IP is,
// and whether IP is retained
func (c *Etcdv3Client) PutPodInfo(namespace, name, ip, interfaceID, ipRetain string) (string, string, error) {
	key := getPodKey(namespace, name)
	pod := &PodInfo{IP: ip, InterfaceID: interfaceID, IPRetain: ipRetain}
	data, err := json.Marshal(pod)
	if err != nil {
		return "", "", fmt.Errorf("Failed to marshal data for pod %s.%s, since: %v", namespace, name, err)
	}
	resp, err := c._put(key, data, true)
	if err != nil {
		return "", "", fmt.Errorf("Failed to do etcdv3 txn for pod %s.%s, since: %v", namespace, name, err)
	}
	if resp != nil {
		respPod := &PodInfo{}
		if err := json.Unmarshal(resp, respPod); err != nil {
			return "", "", fmt.Errorf("Failed to unmarshal etcdv3 get response, since: %v", err)
		}
		return respPod.IP, respPod.InterfaceID, nil
	}
	return "", "", nil
}

// DeletePodIPInfo delete both pod and IP info by given namespace, pod name and ip
func (c *Etcdv3Client) DeletePodIPInfo(namespace, name, ip string) error {
	ops := []clientv3.Op{
		clientv3.OpDelete(getPodKey(namespace, name)),
		clientv3.OpDelete(getIPKey(ip)),
	}
	_, err := c.Client.Txn(context.Background()).Then(ops...).Commit()
	return err
}

// DeleteIPInfo deletes IP info from etcd
func (c *Etcdv3Client) DeleteIPInfo(ip string) error {
	_, err := c.Client.Delete(context.Background(), getIPKey(ip))
	return err
}

// PutIPInfo will put IP info into etcd based on given namespace, pod name and IP
func (c *Etcdv3Client) PutIPInfo(namespace, name, ip string) (string, string, error) {
	key := getIPKey(ip)
	info := &IPInfo{Namespace: namespace, Name: name}
	data, err := json.Marshal(info)
	if err != nil {
		return "", "", fmt.Errorf("Failed to marshal data for ip %s: since: %v", ip, err)
	}
	resp, err := c._put(key, data, false)
	if err != nil {
		return "", "", fmt.Errorf("Failed to do etcdv3 txn for ip %s, since: %v", ip, err)
	}
	if resp != nil {
		respIP := &IPInfo{}
		if err := json.Unmarshal(resp, respIP); err != nil {
			return "", "", fmt.Errorf("Failed to unmarshal etcdv3 get response, since: %v", err)
		}
		return respIP.Namespace, respIP.Name, nil
	}
	return "", "", nil
}

// ValidateAndRecordIP will validate IP, and try to put IP info into etcd, with IP as key, owner(namespace and name) as value
func (c *Etcdv3Client) ValidateAndRecordIP(namespace, name, ip string) (bool, error) {
	if net.ParseIP(ip) == nil {
		return false, fmt.Errorf("Invalide IP %s for pod", ip)
	}
	ownerNamespace, ownerName, err := c.PutIPInfo(namespace, name, ip)
	if err != nil {
		return false, fmt.Errorf("Failed to registry VPC IP info into etcd, since: %v", err)
	}
	if (ownerNamespace != "" && ownerNamespace != namespace) || (ownerName != "" && ownerName != name) {
		return false, nil
	}
	return true, nil
}
