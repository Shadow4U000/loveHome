package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/keonjeo/fdfs_client"
	"loveHome/models"
	"loveHome/utils"
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
	if nil != err {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	user.Id = int(id)
	beego.Info("reg success id=", id)
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)

	this.SetSession("name", user.Name)
	this.SetSession("mobile", user.Mobile)
	this.SetSession("user_id", user.Id)
}

func (this *UserController) Postavatar() {

	resp := make(map[string]interface{})
	defer this.RetData(resp)

	//1.获取前端的文件
	fileData, hd, err_File := this.GetFile("avatar")
	defer fileData.Close()
	if nil != err_File {
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
		return
	}

	//2.得到文件后缀
	suffix := path.Ext(hd.Filename)
	if ".jpg" != suffix && ".png" != suffix && ".gif" != suffix && ".jpeg" != suffix {
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
		return
	}
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
	qs.Filter("id", user_id).One(&user)

	urlMap := make(map[string]string)
	//urlMap["avatar_url"] = "http://121.36.84.103:8080/" + DataResponse.RemoteFileId
	urlMap["avatar_url"] = utils.AddDomain2Url(DataResponse.RemoteFileId)
	user.Avatar_url = DataResponse.RemoteFileId

	_, errUpdate := o.Update(&user)
	if nil != errUpdate {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}

	beego.Info("avatar_url=", urlMap["avatar_url"])
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = urlMap
}

func (this *UserController) GetUserData() {

	resp := make(map[string]interface{})
	defer this.RetData(resp)

	//1.从数据库中拿到user_id对应的user
	user_id := this.GetSession("user_id")

	//2.从数据库中拿到user_id对应的user
	user := models.User{Id: user_id.(int)}
	o := orm.NewOrm()
	err_Read := o.Read(&user)
	if nil != err_Read {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = &user
}

func (this *UserController) UpdateName() {

	resp := make(map[string]interface{})
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)

	//1.获得session中的user_id
	user_id := this.GetSession("user_id")

	//2.获得前端传过来的数据
	UserName := make(map[string]string)
	json.Unmarshal(this.Ctx.Input.RequestBody, &UserName)
	beego.Info("get userName =", UserName["name"])

	//3.更新user_id对应的name
	user := models.User{Id: user_id.(int)}

	o := orm.NewOrm()
	user.Name = UserName["name"]

	if _, err_Update := o.Update(&user, "name"); nil != err_Update {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	//4.把session中的name字段更新
	this.SetSession("user_id", user_id)
	this.SetSession("name", UserName["name"])
	//5.把数据打包返回给前端

	resp["data"] = &user
}

func (this *UserController) AuthGet() {

	resp := make(map[string]interface{})
	defer this.RetData(resp)

	//1.从数据库中拿到user_id对应的user
	user_id := this.GetSession("user_id")
	beego.Info("user_id=", user_id)
	//2.从数据库中拿到user_id对应的user
	user := models.User{Id: user_id.(int)}
	o := orm.NewOrm()
	err_Read := o.Read(&user)
	if nil != err_Read {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	this.SetSession("user_id", user.Id)

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = &user
}

func (this *UserController) AuthPost() {

	resp := make(map[string]interface{})
	defer this.RetData(resp)

	//1.获得session中的user_id
	user_id := this.GetSession("user_id")

	//2.获得前端传过来的数据
	UserInfo := make(map[string]string)
	json.Unmarshal(this.Ctx.Input.RequestBody, &UserInfo)
	beego.Info("real_name =", UserInfo["real_name"], "id_card=", UserInfo["id_card"])

	//3.更新user_id对应的name
	user := models.User{Id: user_id.(int)}
	o := orm.NewOrm()

	if err_Read := o.Read(&user); nil != err_Read {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	user.Real_name = UserInfo["real_name"]
	user.Id_card = UserInfo["id_card"]

	if _, err_Update := o.Update(&user); nil != err_Update {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	//4.把session中的name字段更新
	this.SetSession("user_id", user.Id)
	//5.把数据打包返回给前端
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = &user
}
