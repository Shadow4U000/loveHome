package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/keonjeo/fdfs_client"
	"loveHome/models"
	"path"
)

type UserController struct {
	beego.Controller
}

func (this *UserController) RetData(resp map[string]interface{}) {
	this.Data["json"] = resp
	this.ServeJSON()
}

func (this *UserController) Reg() {

	resp := make(map[string]interface{})
	json.Unmarshal(this.Ctx.Input.RequestBody, &resp)

	//从session字段获取
	defer this.RetData(resp)

	o := orm.NewOrm()
	user := models.User{}
	user.Password_hash = resp["password"].(string)
	user.Name = resp["mobile"].(string)
	user.Mobile = resp["mobile"].(string)

	id, err := o.Insert(&user)
	if err != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	beego.Info("reg success id=", id)
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	this.SetSession("name", user.Name)
}

func (this *UserController) Postavatar() {

	resp := make(map[string]interface{})
	defer this.RetData(resp)

	//1.获取前端的文件
	fileData, hd, err_File := this.GetFile("avatar")
	if nil != err_File {
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
		return
	}

	//2.得到文件后缀
	suffix := path.Ext(hd.Filename)

	//3.存储文件到fastdfs上
	fdfsClient, errClient := fdfs_client.NewFdfsClient("conf/client.conf")
	if nil != errClient {
		beego.Info("New FdfsClient error %s", errClient.Error())
		return
	}
	fileBuffer := make([]byte, hd.Size)
	_, errBuffer := fileData.Read(fileBuffer)
	if nil != errBuffer {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}

	DataResponse, errUpload := fdfsClient.UploadByBuffer(fileBuffer, suffix[1:])
	if nil != errUpload {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}

	//4.从session里拿到user_id
	user_id := this.GetSession("user_id")
	var user models.User

	//5.更新用户数据库中的内容
	o := orm.NewOrm()
	qs := o.QueryTable("user")
	qs.Filter("Id", user_id).One(&user)
	user.Avatar_url = DataResponse.RemoteFileId

	_, errUpdate := o.Update(&user)
	if nil != errUpdate {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}

	urlMap := make(map[string]string)
	urlMap["avatar_url"] = DataResponse.RemoteFileId
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = urlMap
}
