package util

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	ServiceInfoKeyPrefix   = "service.info."
	ServiceConfigKeyPrefix = "service.config."
)

type ModeVsersion struct {
	Dev  int64 `json:"dev"`  // 开发环境
	Test int64 `json:"test"` // 测试环境
	Pre  int64 `json:"pre"`  // 预生产环境
	Pro  int64 `json:"pro"`  // 生产环境
}

type Service struct {
	Typ            string       `json:"type"`           // 服务类型, api、rpc
	Summary        string       `json:"summary"`        // 服务概述
	CurrentVersion ModeVsersion `json:"currentVersion"` // 当前版本
}

func (s Service) getVersion(mode string) (version int64) {
	switch mode {
	case "dev":
		version = int64(s.CurrentVersion.Dev)
	case "test":
		version = int64(s.CurrentVersion.Test)
	case "pre":
		version = int64(s.CurrentVersion.Pre)
	default:
		version = int64(s.CurrentVersion.Pro)
	}
	return
}

type ServiceConfig struct {
	Content    string `json:"content"`    // 版本创建时间
	CreateTime string `json:"createTime"` // 版本创建时间
	CreateUser string `json:"createUser"` // 创建人
	UpdateTime string `json:"updateTime"` // 最近版本落地时间
}

func (cli Client) GetConfig(serviceName string, mode string, prevision int64) (content string, version int64, err error) {
	res, err := cli.Get(ServiceInfoKeyPrefix + serviceName)
	if err != nil {
		return
	}
	var info Service
	err = json.Unmarshal(res, &info)
	if err != nil {
		err = fmt.Errorf("服务%s的服务信息内容错误", serviceName)
		return
	}
	version = info.getVersion(mode)

	if version == 0 {
		err = fmt.Errorf("暂无服务%s在%s环境下的配置版本应用信息", serviceName, mode)
		return
	}

	if prevision == version {
		err = fmt.Errorf("暂无服务%s在%s环境下的配置版本仍然是%d", serviceName, mode, prevision)
		return
	}

	content, err = cli.getConfigVal(serviceName, version)

	return
}

func (cli Client) getConfigVal(serviceName string, version int64) (content string, err error) {
	key := fmt.Sprintf("%s%s.%d", ServiceConfigKeyPrefix, serviceName, version)
	value, err := cli.Get(key)
	if err != nil {
		err = fmt.Errorf("服务%s的v%d版本配置不存在", serviceName, version)
		return
	}
	var configInfo ServiceConfig

	err = json.Unmarshal(value, &configInfo)
	if err != nil {
		err = fmt.Errorf("服务%s的v%d版本配置内容错误", serviceName, version)
		return
	}
	content = configInfo.Content
	return
}

func (cli Client) InitMerge(name, mode string, cPoint any) int64 {
	serviceNameField := logx.LogField{
		Key:   "serviceName",
		Value: name,
	}
	modeField := logx.LogField{
		Key:   "mode",
		Value: mode,
	}
	logx.AddGlobalFields(serviceNameField, modeField)
	val, version, err := cli.GetConfig(name, mode, 0)
	if err != nil {
		logx.Error("合并配置中心服务配置失败, 原因:", err)
	} else {
		err = conf.LoadFromJsonBytes([]byte(val), cPoint)
		if err != nil {
			logx.Errorf("配置中心服务配置错误格式错误: %v", err)
		}
		logx.Errorf("启动程序，合并配置中心服务配置成功, 配置version: %d", version)
	}
	return version
}

func (cli Client) Watch(serviceName, mode string, version int64, action func(string)) {
	if cli.cli == nil {
		return
	}
	watcher := clientv3.NewWatcher(cli.cli)
	// timeout := 5 * time.Second
	// ctx, cancel := context.WithTimeout(context.Background(), timeout)
	for {
		watchRespChan := watcher.Watch(context.Background(), ServiceInfoKeyPrefix+serviceName)
		newVersion := version
		for watchResp := range watchRespChan {
			for _, event := range watchResp.Events {
				switch event.Type {
				case mvccpb.PUT:
					var info Service
					err := json.Unmarshal(event.Kv.Value, &info)
					if err != nil {
						err = fmt.Errorf("服务%s的服务信息内容错误", serviceName)
						logx.Error(err)
					} else {
						currentVersion := info.getVersion(mode)
						if currentVersion != newVersion {
							content, err := cli.getConfigVal(serviceName, currentVersion)
							if err == nil {
								action(content)
								newVersion = currentVersion
								logx.Infof("服务%s的版本配置更换为v%d", serviceName, currentVersion)
							} else {
								err = fmt.Errorf("服务%s的v%d版本配置内容错误", serviceName, currentVersion)
								logx.Error(err)
							}
						}
					}
				}
			}
		}
	}
}
