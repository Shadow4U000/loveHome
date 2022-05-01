package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	_ "github.com/astaxie/beego/cache/redis"
	"github.com/astaxie/beego/orm"
	"loveHome/models"
	"loveHome/utils"
	"time"
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
	defer c.RetData(resp)

	var areas []models.Area
	//从redis缓存中拿数据
	redis_config_map := map[string]string{
		"key":   "lovehome2",
		"conn":  utils.G_redis_addr + ":" + utils.G_redis_port,
		"dbNum": utils.G_redis_dbnum,
	}
	redis_config, _ := json.Marshal(redis_config_map)
	cache_conn, err := cache.NewCache("redis", string(redis_config))

	if areaData := cache_conn.Get("area"); nil != areaData {
		//var area_info interface{}
		//json.Unmarshal(areaData.([]byte), &area_info)
		//resp["data"] = area_info
		//beego.Info("cache data get,area=", resp["data"])
		//return
		var area_info []models.Area
		json.Unmarshal(areaData.([]byte), &area_info)
		resp["data"] = area_info
		beego.Info("cache data get,area=", resp["data"])
		return
	}

	//从mysql数据库拿到area数据
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

	//把数据转换成json格式存入缓存
	json_str, err_json := json.Marshal(areas)
	if nil != err_json {
		beego.Info("encoding error")
		return
	}
	beego.Info("areas =", areas)
	cache_conn.Put("area", json_str, time.Second*3600)

	//打包或json返回给前端
	beego.Info("query data success,resp=", resp, "num=", num)

}
