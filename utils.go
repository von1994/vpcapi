package vpcapi

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	// according to https://cloud.tencent.com/document/api/213/15753#Filter, filters is f**king useless
	usableFilters = []string{"private-ip-address"}
)

func getKeys() []string {
	return []string{"Nonce", "Region", "SecretId", "Timestamp", "SignatureMethod", "RequestClient"}
}

func formatFilter(params map[string]string) {
	index := 0
	for _, k := range usableFilters {
		if v, ok := params[k]; ok && v != "" {
			params[fmt.Sprintf("Filters.%d.Name", index)] = k
			params[fmt.Sprintf("Filters.%d.Values.0", index)] = params[k]
			delete(params, k)
			index++
		}
	}
}

func getConfValue(conf VPC, k string) string {
	switch k {
	case "SecretId":
		return conf.SecretID
	case "Region":
		return conf.Region
	case "VPCEndpoint":
		return conf.VPCAPIEndpoint
	case "CVMEndpoint":
		return conf.CVMAPIEndpoint
	case "Nonce":
		return strconv.Itoa(rand.Int())
	case "Timestamp":
		return strconv.FormatInt(time.Now().Unix(), 10)
	case "SignatureMethod":
		return "HmacSHA1"
	case "RequestClient":
		return "SDK_PYTHON_2.0.15"
	default:
		return ""
	}
}

func mergeKeys(conf VPC, params map[string]string) {
	confKeys := getKeys()
	for _, k := range confKeys {
		if v, ok := params[k]; !ok || v == "" {
			params[k] = getConfValue(conf, k)
		}
	}
}

func getSHA1Signature(conf VPC, params map[string]string, requestMethod, endpoint string) string {
	keys := []string{}
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	params["Nonce"] = getConfValue(conf, "Nonce")
	params["Timestamp"] = getConfValue(conf, "Timestamp")

	reqStrs := []string{}
	for _, k := range keys {
		reqStrs = append(reqStrs, fmt.Sprintf("%s=%s", k, params[k]))
	}

	reqStr := strings.Join(reqStrs, "&")
	signReqStr := fmt.Sprintf("%s%s?%s", requestMethod, endpoint, reqStr)
	hashed := hmac.New(sha1.New, []byte(conf.SecretKey))
	hashed.Write([]byte(signReqStr))
	rawSign := base64.StdEncoding.EncodeToString(hashed.Sum(nil))
	return url.QueryEscape(rawSign)
}

func setReqQuery(params map[string]string, req *http.Request) {
	ret := []string{}
	for k, v := range params {
		ret = append(ret, fmt.Sprintf("%s=%s", k, v))
	}
	url := strings.Join(ret, "&")
	req.URL.RawQuery = url
}

func doRequest(conf VPC, params map[string]string) ([]byte, error) {
	requestMethod := "GET"
	endpoint := ""
	if params["Service"] == "vpc" {
		endpoint = conf.VPCAPIEndpoint + conf.V2URI
		delete(params, "Servcie")
	} else if params["Service"] == "cvm" {
		endpoint = conf.CVMAPIEndpoint + conf.V3URI
		delete(params, "Service")
	}
	formatFilter(params)
	mergeKeys(conf, params)
	sign := getSHA1Signature(conf, params, requestMethod, endpoint)
	params["Signature"] = sign

	req, _ := http.NewRequest(requestMethod, "https://"+endpoint, nil)
	setReqQuery(params, req)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Fail exec http client Do,err:%s\n", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
