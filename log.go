package util

import "github.com/zeromicro/go-zero/core/logx"

func AddGlobalFields(name, mode string) {
	serviceNameField := logx.LogField{
		Key:   "serviceName",
		Value: name,
	}
	modeField := logx.LogField{
		Key:   "mode",
		Value: mode,
	}
	logx.AddGlobalFields(serviceNameField, modeField)
}
