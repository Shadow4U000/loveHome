package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"loveHome/models"
)

type SessionController struct {
	beego.Controller
}

func (this *SessionController) RetData(resp map[string]interface{}) {
	this.Data["json"] = resp
	this.ServeJSON()
}

func (this *SessionController) GetSessionData() {

	resp := make(map[string]interface{})
	defer this.RetData(resp)
	user := models.User{}

	resp["errno"] = models.RECODE_DBERR
	resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)

	name := this.GetSession("name")
	if nil != name {
		user.Name = name.(string)
		resp["errno"] = models.RECODE_OK
		resp["errmsg"] = models.RecodeText(models.RECODE_OK)
		resp["data"] = user
	}
}

func (this *SessionController) DeleteSessionData() {

	resp := make(map[string]interface{})
	defer this.RetData(resp)
	this.DelSession("name")
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)

}

func (this *SessionController) Login() {

	//1.得到用户信息
	resp := make(map[string]interface{})
	defer this.RetData(resp)
	//获取前端传过来的json数据
	json.Unmarshal(this.Ctx.Input.RequestBody, &resp)

	//2.判断是否合法
	if nil == resp["mobile"] || nil == resp["password"] {
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
		beego.Info("one")
		return
	}

	//3.与数据库匹配判断账号密码正确
	o := orm.NewOrm()
	//var user models.User
	user := models.User{Mobile: resp["mobile"].(string)}
	err_Query := o.QueryTable("user").Filter("mobile", resp["mobile"].(string)).One(&user)
	if orm.ErrNoRows == err_Query {
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
		return
	}
	if user.Password_hash != resp["password"] {
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
		beego.Info("three")
		return
	}

	//4.添加session
	this.SetSession("name", user.Name)
	this.SetSession("mobile", user.Mobile)
	this.SetSession("user_id", user.Id)

	//5返回json数据给前端
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	return
}
