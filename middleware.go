package util

import (
	"context"
	"net/http"
	"net/url"

	"github.com/zeromicro/go-zero/core/logx"
)

type BasicContext string

const (
	Token      BasicContext = "token"
	UUID       BasicContext = "uuid"
	Name       BasicContext = "name"
	PageURI    BasicContext = "pageURI"
	RequestURI BasicContext = "requestURI"
	SystemCode BasicContext = "systemCode"
)

func BaseCors(w http.ResponseWriter) {
	// 接受网关转发时的headers参数
	w.Header().Add("Access-Control-Allow-Headers", "x-page-uri")
	w.Header().Add("Access-Control-Allow-Headers", "x-request-uri")
	w.Header().Add("Access-Control-Allow-Headers", "x-request-uuid")
	w.Header().Add("Access-Control-Allow-Headers", "x-request-name")
	w.Header().Add("Access-Control-Allow-Headers", "x-request-system")
	w.Header().Add("Access-Control-Allow-Headers", "x-request-service")
	w.Header().Add("Access-Control-Allow-Headers", "x-request-token")
}

func BaseMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := r.Header.Get("x-request-uuid")
		name := url.QueryEscape(r.Header.Get("x-request-name"))
		requestURI := url.QueryEscape(r.Header.Get("x-request-uri"))
		pageURI := url.QueryEscape(r.Header.Get("x-page-uri"))
		systemCode := r.Header.Get("x-request-system")
		token := r.Header.Get("x-request-token")
		ctx := context.WithValue(r.Context(), UUID, uuid)
		ctx = context.WithValue(ctx, Name, name)
		ctx = context.WithValue(ctx, SystemCode, systemCode)
		ctx = context.WithValue(ctx, Token, token)
		uuidField := logx.LogField{
			Key:   "uuid",
			Value: uuid,
		}
		nameField := logx.LogField{
			Key:   "name",
			Value: name,
		}
		systemField := logx.LogField{
			Key:   "system",
			Value: systemCode,
		}
		pageURIField := logx.LogField{
			Key:   "pageURI",
			Value: pageURI,
		}
		requestURIField := logx.LogField{
			Key:   "requestURI",
			Value: requestURI,
		}
		ctx = logx.WithFields(ctx, systemField, pageURIField, requestURIField, uuidField, nameField)
		next(w, r.WithContext(ctx))
	}
}
