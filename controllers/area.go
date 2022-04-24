package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"loveHome/models"
)

type AreaController struct {
	beego.Controller
}

func (this *AreaController) RetData(resp map[string]interface{}) {
	this.Data["json"] = resp
	this.ServeJSON()
}

func (c *AreaController) GetArea() {
	beego.Info("connect success")

	resp := make(map[string]interface{})
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	//从session字段获取
	defer c.RetData(resp)
	//从mysql数据库拿到area数据
	var areas []models.Area
	o := orm.NewOrm()
	num, err := o.QueryTable("Area").All(&areas)

	if err != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	if 0 == num {
		resp["errno"] = models.RECODE_NODATA
		resp["errmsg"] = models.RecodeText(models.RECODE_NODATA)
		return
	}

	resp["data"] = areas
	//打包或json返回给前端
	beego.Info("query data success,resp=", resp, "num=", num)

}
