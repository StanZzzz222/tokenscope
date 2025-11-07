package models

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

/*
   Created by zyx
   Date Time: 2024/8/2
   File: base_response.go
*/

type Result struct {
	Code int    `json:"code"`
	Data any    `json:"data"`
	Err  string `json:"error"`
	ctx  *gin.Context
}
type IResult interface {
	Success(code int, data any)
	Image(data []byte)
	Error(code int, err error)
}

func Response(ctx *gin.Context) IResult {
	return &Result{
		ctx: ctx,
	}
}

func (r *Result) Success(code int, data any) {
	r.ctx.JSON(code, &Result{
		Code: code,
		Data: data,
		Err:  "",
	})
	r.ctx.Abort()
}

func (r *Result) Image(data []byte) {
	r.ctx.Data(http.StatusOK, "image/png", data)
	r.ctx.Abort()
}

func (r *Result) Error(code int, err error) {
	r.ctx.JSON(code, &Result{
		Code: code,
		Data: nil,
		Err:  err.Error(),
	})
	r.ctx.Abort()
}
