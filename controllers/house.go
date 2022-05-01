package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"github.com/keonjeo/fdfs_client"
	"loveHome/models"
	"loveHome/utils"
	"path"
	"strconv"
	"time"
)

var facilitiy_name = []string{
	"无线网络", "热水淋浴", "空调", "暖气", "允许吸烟", "饮水设备", "牙具", "香皂", "拖鞋", "手纸", "毛巾", "沐浴露、洗发露", "冰箱", "洗衣机", "电梯", "允许做饭", "允许带宠物", "允许聚会", "门禁系统", "停车位", "有线网络", "电视", "浴缸", "吃鸡", "打台球"}

type HouseController struct {
	beego.Controller
}

func (this *HouseController) RetData(resp map[string]interface{}) {
	this.Data["json"] = resp
	this.ServeJSON()
}

func (this *HouseController) GetHouseData() {

	resp := make(map[string]interface{})
	json.Unmarshal(this.Ctx.Input.RequestBody, &resp)

	//1.从session字段获取user_id
	defer this.RetData(resp)
	user_id := this.GetSession("user_id")

	//2.从数据库中拿到user_id对应的user
	houses := []models.House{}
	o := orm.NewOrm()
	qs := o.QueryTable("house")
	num, err := qs.Filter("user_id", user_id.(int)).All(&houses)

	if nil != err {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	if 0 == num {
		resp["errno"] = models.RECODE_NODATA
		resp["errmsg"] = models.RecodeText(models.RECODE_NODATA)
		return
	}
	respData := make(map[string]interface{})
	respData["houses"] = houses

	resp["data"] = respData
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
}

func (this *HouseController) PostHouseData() {
	//1.从前端拿到数据
	resp := make(map[string]interface{})

	defer this.RetData(resp)
	reqData := make(map[string]interface{})
	json.Unmarshal(this.Ctx.Input.RequestBody, &reqData)

	//2.判断前端数据的合法性
	house := models.House{}
	house.Title = reqData["title"].(string)
	price, _ := strconv.Atoi(reqData["price"].(string))
	room_count, _ := strconv.Atoi(reqData["room_count"].(string))
	minDays, _ := strconv.Atoi(reqData["min_days"].(string))
	maxDays, _ := strconv.Atoi(reqData["max_days"].(string))
	area_id, _ := strconv.Atoi(reqData["area_id"].(string))
	user_id, _ := this.GetSession("user_id").(int)
	house.Price = price
	house.Room_count = room_count
	house.Min_days = minDays
	house.Max_days = maxDays
	house.Unit = reqData["unit"].(string)
	house.Beds = reqData["beds"].(string)
	house.Address = reqData["address"].(string)
	house.User = &models.User{Id: user_id}
	house.Area = &models.Area{Id: area_id}
	facilities := []*models.Facility{}
	for _, fid := range reqData["facility"].([]interface{}) {
		f_id, _ := strconv.Atoi(fid.(string))
		facility := models.Facility{Id: f_id}
		facilities = append(facilities, &facility)
	}

	o := orm.NewOrm()
	house_id, errInsert := o.Insert(&house)
	if nil != errInsert {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		beego.Info("1-----", errInsert, house)
		return
	}

	m2m := o.QueryM2M(&house, "Facilities")
	beego.Info("facilities=", facilities)
	for _, facility := range facilities {
		num, errM2M := m2m.Add(facility)
		if nil != errM2M || 0 == num {
			resp["errno"] = models.RECODE_DBERR
			resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
			beego.Info("2-----", errM2M)
			return
		}
	}

	//m2m := o.QueryM2M(&house, "Facilities")
	//beego.Info("facilities=", facilities)
	//num, errM2M := m2m.Add(&facilities)
	//if nil != errM2M || 0 == num {
	//	resp["errno"] = models.RECODE_DBERR
	//	resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
	//	beego.Info("2-----", errM2M)
	//	return
	//}

	respData := make(map[string]interface{})
	respData["house_id"] = &house_id
	resp["data"] = respData
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)

}

func (this *HouseController) PostHouseImage() {

	resp := make(map[string]interface{})
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)

	//1.从用户请求中获取到图片数据
	fileData, hd, err_File := this.GetFile("house_image")
	defer fileData.Close()
	if nil == fileData {
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
		return
	}
	if nil != err_File {
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
		return
	}

	//2.将用户二进制数据存到fdfs中。得到fileid
	suffix := path.Ext(hd.Filename)
	if ".jpg" != suffix && ".png" != suffix && ".gif" != suffix && ".jpeg" != suffix {
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
		return
	}
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
	//3.从请求的url中获得house_id
	house_id := this.Ctx.Input.Param(":id")
	var house models.House
	house.Id, _ = strconv.Atoi(house_id)
	//4.查看该房屋的index_image_url主图是否为空
	o := orm.NewOrm()
	errRead := o.Read(&house)
	if nil != errRead {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	//image_url := "http://121.36.84.103:8080/" + DataResponse.RemoteFileId
	image_url := DataResponse.RemoteFileId

	if "" == house.Index_image_url {
		house.Index_image_url = image_url
	}

	//5.主图不为空，将该图片的fileid字段追加（关联查询）到houseimage字段中到house_image表中
	house_image := models.HouseImage{House: &house, Url: image_url}
	house.Images = append(house.Images, &house_image)
	if _, errInsert := o.Insert(&house_image); nil != errInsert {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	if _, errUpdate := o.Update(&house); nil != errUpdate {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	//6.拼接完整域名
	respData := make(map[string]string)
	respData["url"] = image_url

	resp["data"] = respData
}

func (this *HouseController) GetDetailHouseData() {
	resp := make(map[string]interface{})
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)
	//var respData []interface{}
	respData := make(map[string]interface{})
	house_id, _ := strconv.Atoi(this.Ctx.Input.Param(":id"))
	//1.从缓存服务器中请求“house_page_data”字段
	//redis_config_map := map[string]string{
	//	"key":   "lovehome2",
	//	"conn":  utils.G_redis_addr + ":" + utils.G_redis_port,
	//	"dbNum": utils.G_redis_dbnum,
	//}
	//redis_config, _ := json.Marshal(redis_config_map)
	////cache_conn, errCache := cache.NewCache("redis", string(redis_config))
	//if nil != errCache {
	//	resp["errno"] = models.RECODE_REQERR
	//	resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
	//	return
	//}
	//house_page_key := "house_page_data"
	//house_page_value := cache_conn.Get(house_page_key)
	user_id := this.GetSession("user_id")
	//if nil != house_page_value {
	//	beego.Debug("======get house page info from CACHE!======")
	//	respData["user_id"] = user_id
	//	house_info := make(map[string]interface{})
	//	json.Unmarshal(house_page_value.([]byte), &house_info)
	//	respData["house"] = house_info
	//	resp["data"] = respData
	//	return
	//}

	//2.如果缓存没有，需要从数据库中查询到房屋列表

	house := models.House{Id: house_id}
	o := orm.NewOrm()
	o.Read(&house)
	o.LoadRelated(&house, "Area")
	o.LoadRelated(&house, "User")
	o.LoadRelated(&house, "Images")
	o.LoadRelated(&house, "Facilities")

	//if err := o.QueryTable("house").Filter("id", house_id).One(&house); nil == err {
	//	//o.LoadRelated(&house, "Area")
	//	//o.LoadRelated(&house, "User")
	//	//o.LoadRelated(&house, "Images")
	//	//o.LoadRelated(&house, "Facilities")
	//	respData["houses"] = house
	//
	//}

	//house_page_value_new, _ := json.Marshal(house.To_one_house_desc())
	//cache_conn.Put(house_page_key, house_page_value_new, 3600*time.Second)
	respData["user_id"] = user_id
	respData["house"] = house.To_one_house_desc()

	resp["data"] = respData
	return
}

func (this *HouseController) GetHouseSearchData() {

	resp := make(map[string]interface{})
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)
	respData := make(map[string]interface{})
	//1.获取用户发来的参数，aid,sd,ed,sk,p
	var aid int
	this.Ctx.Input.Bind(&aid, "aid")
	var sd string
	this.Ctx.Input.Bind(&sd, "sd")
	var ed string
	this.Ctx.Input.Bind(&ed, "ed")
	var sk string
	this.Ctx.Input.Bind(&sk, "sk")
	var page int
	this.Ctx.Input.Bind(&page, "p")
	//2.检验开始时间一定要早于结束时间
	start_time, _ := time.Parse("2006-01-02 15:04:05", sd+" 00:00:00")
	end_time, _ := time.Parse("2006-01-02 15:04:05", ed+" 00:00:00")
	if end_time.Before(start_time) {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = "离店时间需在入住时间之后"
		return
	}
	//3.判断p的合法性，一定要大于0的整数
	if page <= 0 {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = "页数不能小于或等于0"
		return
	}
	//4.尝试从缓存中获取数据，返回查询结果json
	redis_config_map := map[string]string{
		"key":   "lovehome2",
		"conn":  utils.G_redis_addr + ":" + utils.G_redis_port,
		"dbNum": utils.G_redis_dbnum,
	}
	redis_config, _ := json.Marshal(redis_config_map)
	cache_conn, errCache := cache.NewCache("redis", string(redis_config))
	if nil != errCache {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}
	house_search_key := "house_search_data"
	house_search_value := cache_conn.Get(house_search_key)
	if nil != house_search_value {
		beego.Debug("======get house_search_datafrom CACHE!======")
		house_Info := make(map[string]interface{})
		json.Unmarshal(house_search_value.([]byte), &house_Info)
		respData["house"] = house_Info
		respData["total_page"] = 10
		respData["current_page"] = 1
		resp["data"] = respData
		return
	}
	//5.如果缓存中没有数据，从数据库中查询
	houses := []models.House{}
	o := orm.NewOrm()
	qs := o.QueryTable("house")
	//查询指定城区的所有房源，按发布时间降序排列
	num, errFilter := qs.Filter("area_id", aid).OrderBy("-ctime").All(&houses)
	if nil != errFilter {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	//求出所有分页
	total_page := int(num)/models.HOUSE_LIST_PAGE_CAPACITY + 1
	house_page := 1
	var house_list []interface{}
	for _, house := range houses {
		o.LoadRelated(&house, "Area")
		o.LoadRelated(&house, "User")
		o.LoadRelated(&house, "Images")
		o.LoadRelated(&house, "Facilities")
		house_list = append(house_list, house.To_house_info())
	}
	//（此处过于复杂，可以暂时以发布时间顺序查询）

	//6.将查询条件存储到缓存
	house_search_list, _ := json.Marshal(house_list)
	cache_conn.Put(house_search_key, house_search_list, 3600*time.Second)
	respData["houses"] = house_list
	respData["total_page"] = total_page
	respData["current_page"] = house_page
	//7.返回查询结果json数据给前端
	resp["data"] = respData
	return
}
