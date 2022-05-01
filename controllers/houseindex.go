package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"loveHome/models"
	"loveHome/utils"
	"time"
)

type HouseIndexController struct {
	beego.Controller
}

func (this *HouseIndexController) RetData(resp map[string]interface{}) {
	this.Data["json"] = resp
	this.ServeJSON()
}

func (this *HouseIndexController) GetHouseIndex() {

	resp := make(map[string]interface{})
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)
	var respData []interface{}

	//1 从缓存服务器中请求 "home_page_data" 字段,如果有值就直接返回
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
		beego.Debug("------cache error 1")
		return
	}
	home_page_key := "home_page_data"
	home_page_value := cache_conn.Get(home_page_key)
	if nil != home_page_value {
		beego.Debug("======get house page info from CACHE!======")
		json.Unmarshal(home_page_value.([]byte), &respData)
		resp["data"] = respData
		beego.Debug("------cache error 2")
		return
	}

	//2 如果缓存没有,需要从数据库中查询到房屋列表
	houses := []models.House{}
	o := orm.NewOrm()
	if _, errQuery := o.QueryTable("house").Limit(models.HOME_PAGE_MAX_HOUSES).All(&houses); nil == errQuery {
		for _, v := range houses {
			//o.LoadRelated(&v, "Area")
			//o.LoadRelated(&v, "User")
			//o.LoadRelated(&v, "Images")
			//o.LoadRelated(&v, "Facilities")
			respData = append(respData, v.To_house_info())
		}
	}

	//3.将data存入缓存中
	home_page_value, _ = json.Marshal(respData)
	cache_conn.Put(home_page_key, home_page_value, 3600*time.Second)
	beego.Debug("respData=", respData)
	//4.返回前端data
	resp["data"] = respData
}
