package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	_ "loveHome/models"
	_ "loveHome/routers"
	"net/http"
	"strings"
)

func TransparentsStatic(ctx *context.Context) {
	orpath := ctx.Request.URL.Path
	beego.Debug("request url:", orpath)
	if strings.Index(orpath, "api") >= 0 {
		return
	}
	http.ServeFile(ctx.ResponseWriter, ctx.Request, "static/html/"+ctx.Request.URL.Path)
}

func ignoreStaticPath() {

	beego.SetStaticPath("group1/M00/", "fdfs/storage_data/data/")
	beego.InsertFilter("/", beego.BeforeRouter, TransparentsStatic)
	beego.InsertFilter("/*", beego.BeforeRouter, TransparentsStatic)
}

func main() {
	ignoreStaticPath()
	beego.Run()
}
