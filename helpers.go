/*
 * Copyright (C) 2019 Gregor Pogačnik
 */

package TokensApi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fiksn/TokensApi/entities"

	"github.com/golang/glog"
)

type CredentialsConfig struct {
	APIKey    string `json:"key" description:"Api key."`
	APISecret string `json:"secret" description:"Api secret."`
}

const (
	Timeout = time.Duration(5 * time.Second)
)

var (
	Credentials  *CredentialsConfig
	NullHookFunc = func(code int, reason string) {}
	HookFunc     = NullHookFunc
)

type ErrorHook func(int, string)

func InstallErrorHook(f ErrorHook) {
	HookFunc = f
}

func UninstallErrorHook() {
	HookFunc = NullHookFunc
}

func Init(key, secret string) {
	Credentials = &CredentialsConfig{
		APIKey:    key,
		APISecret: secret,
	}
}

func parseJsonCfg(configPath string) (*CredentialsConfig, error) {
	jsonBlob, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not parse config: %v", err))
	}
	var cfg CredentialsConfig
	err = json.Unmarshal(jsonBlob, &cfg)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to unmarshal json blob: %v", string(jsonBlob)))
	}

	return &cfg, nil
}

func request(url string) []byte {
	glog.V(6).Infof("request url: %v\n", url)

	client := http.Client{
		Timeout: Timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		glog.Warningf("request error %v", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Warningf("request error %v", err)
		return nil
	}

	return body
}

func calcHmac(key string, message string) string {
	sig := hmac.New(sha256.New, []byte(key))
	sig.Write([]byte(message))

	return strings.ToUpper(hex.EncodeToString(sig.Sum(nil)))
}

func requestAuthPost(url string, data url.Values) []byte {
	if Credentials == nil {
		glog.Warningf("requestAuthPost - credentials were not set, you might need to call Init()")
		return nil
	}

	unixTime := strconv.Itoa(int(time.Now().UnixNano() / 1000))

	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		glog.Warningf("requestAuthPost error %v", err)
		return nil
	}

	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("key", Credentials.APIKey)
	req.Header.Set("nonce", unixTime)
	signature := calcHmac(Credentials.APISecret, unixTime+Credentials.APIKey)
	req.Header.Set("signature", signature)

	client := &http.Client{Timeout: Timeout}

	glog.V(6).Infof("requestAuthPost url: %v data %v\n", url, data)

	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf("requestAuthPost error %v", err)
		return nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Warningf("requestAuthPost error %v", err)
		return nil
	}

	return body
}

func requestAuth(url string) []byte {
	if Credentials == nil {
		glog.Warningf("requestAuth - credentials were not set, you might need to call Init()")
		return nil
	}

	unixTime := strconv.Itoa(int(time.Now().UnixNano() / 1000))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		glog.Warningf("requestAuth error %v", err)
		return nil
	}

	req.Header.Set("key", Credentials.APIKey)
	req.Header.Set("nonce", unixTime)

	signature := calcHmac(Credentials.APISecret, unixTime+Credentials.APIKey)

	req.Header.Set("signature", signature)

	client := &http.Client{Timeout: Timeout}

	glog.V(6).Infof("requestAuth url: %v\n", url)

	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf("requestAuth error %v", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Warningf("requestAuth error %v", err)
		return nil
	}

	return body
}

func deserialize(jsonBlob []byte, resp entities.Statuser) (err error) {

	var errorEntity entities.ErrorResp

	err = json.Unmarshal(jsonBlob, resp)

	if err != nil {
		glog.Warningf("Unable to unmarshal json blob: %v (%v)", string(jsonBlob), err)
		resp.SetStatus("error")
		return err
	}

	if resp.GetStatus() != "ok" {
		err = json.Unmarshal(jsonBlob, &errorEntity)
		if err == nil {
			HookFunc(errorEntity.ErrorCode, errorEntity.Reason)
			return fmt.Errorf("%v %v", errorEntity.ErrorCode, errorEntity.Reason)
		}

		HookFunc(0, resp.GetStatus())
		return errors.New(resp.GetStatus())
	}

	return nil
}
