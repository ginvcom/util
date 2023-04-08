package util

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type ErrorMsg struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *ErrorMsg) Error() string {
	return e.Msg
}

type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func ErrResponse(w http.ResponseWriter, errMsg ErrorMsg) {
	body := Body{
		Code: errMsg.Code,
		Msg:  errMsg.Msg,
	}
	httpx.OkJson(w, body)
}

func Response(w http.ResponseWriter, resp interface{}, err error, code ...int) {
	var body Body
	if err != nil {
		if len(code) > 0 {
			body.Code = code[0]
		} else {
			body.Code = -1
		}
		body.Msg = err.Error()
	} else {
		body.Msg = "OK"
		body.Data = resp
	}
	httpx.OkJson(w, body)
}
