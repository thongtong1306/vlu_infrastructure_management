package controllers

import (
	beegoctx "github.com/beego/beego/v2/server/web/context"
)

func jsonOK(ctx *beegoctx.Context, v interface{}) {
	ctx.Output.Header("Cache-Control", "no-store")
	_ = ctx.Output.JSON(v, false, false)
}

func jsonErr(ctx *beegoctx.Context, code int, msg string) {
	ctx.Output.SetStatus(code)
	_ = ctx.Output.JSON(map[string]string{"error": msg}, false, false)
}
