package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"loveHome/models"
	"loveHome/utils"
	"strconv"
	"time"
)

type OrderController struct {
	beego.Controller
}

type OrderRequest struct {
	House_id   string `json:"house_id"`
	Start_date string `json:"start_date"`
	End_date   string `json:"end_date"`
}

func (this *OrderController) RetData(resp map[string]interface{}) {
	this.Data["json"] = resp
	this.ServeJSON()
}

func (this *OrderController) PostOrderHouseData() {
	resp := make(map[string]interface{})
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)

	//1.根据session得到用户id
	user_id := this.GetSession("user_id")
	//2.得到用户请求的json数据，检测合法性，不合法就返回json数据
	var req OrderRequest
	if errDecode := json.Unmarshal(this.Ctx.Input.RequestBody, &req); nil != errDecode {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = "请求参数为空"
		return
	}
	//3.确定退房时间end_date必须在订房时间start_date之后
	start_time, _ := time.Parse("2006-01-02 15:04:05", req.Start_date+" 00:00:00")
	end_time, _ := time.Parse("2006-01-02 15:04:05", req.End_date+" 00:00:00")
	if end_time.Before(start_time) {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = "结束时间在开始时间之前"
		return
	}
	days := end_time.Sub(start_time).Hours()/24 + 1
	//4.根据order_id获取关联的房源信息
	house_id, _ := strconv.Atoi(req.House_id)
	house := models.House{Id: house_id}
	o := orm.NewOrm()
	if errRead := o.Read(&house); nil != errRead {
		resp["errno"] = models.RECODE_NODATA
		resp["errmsg"] = models.RecodeText(models.RECODE_NODATA)
		return
	}
	//5.确保当前的user_id不是房源信息所关联的user_id
	if user_id == house.User.Id {
		resp["errno"] = models.RECODE_ROLEERR
		resp["errmsg"] = models.RecodeText(models.RECODE_ROLEERR)
		return
	}
	//6.确保用户选择的房屋未被预定，日期没有冲突，如果已经被人预定返回错误信息
	//7.封装完整的order订单信息
	amount := days * float64((house.Price))
	order := models.OrderHouse{}
	user := models.User{Id: user_id.(int)}

	order.House = &house
	order.User = &user
	order.Begin_date = start_time
	order.End_date = end_time
	order.Days = int(days)
	order.House_price = house.Price
	order.Amount = int(amount)
	order.Status = models.ORDER_STATUS_WAIT_ACCEPT
	//8.将订单信息写入表中
	order_id, errInser := o.Insert(&order)
	if nil != errInser {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		beego.Debug("PostOrderHouseData step 8 error")
		return
	}
	//9.返回order_id的json给前端
	this.SetSession("user_id", user_id)
	respData := make(map[string]interface{})
	respData["order_id"] = order_id
	resp["data"] = respData
	return
}

func (this *OrderController) GetOrderData() {
	resp := make(map[string]interface{})
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)

	user_id := this.GetSession("user_id")

	role := this.GetString("role")
	if "" == role {
		resp["errno"] = models.RECODE_ROLEERR
		resp["errmsg"] = models.RecodeText(models.RECODE_ROLEERR)
		return
	}
	//3.根据角色执行不同逻辑
	o := orm.NewOrm()
	orders := []models.OrderHouse{}
	var order_list []interface{}
	if "landlord" == role {
		landlordHouse := []models.House{}
		o.QueryTable("house").Filter("user__id", user_id).All(&landlordHouse)
		housesIds := []int{-1}
		for _, house := range landlordHouse {
			housesIds = append(housesIds, house.Id)
		}
		beego.Debug("houseIds=", housesIds)
		o.QueryTable("order_house").Filter("house__id__in", housesIds).OrderBy("-ctime").All(&orders)
	} else {
		o.QueryTable("order_house").Filter("user__id", user_id).OrderBy("-ctime").All(&orders)
	}
	for _, order := range orders {
		o.LoadRelated(&order, "House")
		o.LoadRelated(&order, "User")
		order_list = append(order_list, order.To_order_info())
	}
	respData := make(map[string]interface{})
	respData["orders"] = order_list
	//7.返回查询结果json数据给前端
	resp["data"] = respData
	return
}

func (this *OrderController) Orderstatus() {
	resp := make(map[string]interface{})
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)

	//1.根据session拿到user_id
	user_id := this.GetSession("user_id")

	//2.通过当前url参数得到当前订单id
	order_id := this.Ctx.Input.Param(":id")

	//3.解析客户端请求的json数据，得到action数据
	var req map[string]interface{}
	if errDecode := json.Unmarshal(this.Ctx.Input.RequestBody, &req); nil != errDecode {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}

	//4.检验action是否合法，不合法，返回json错误信息
	action := req["action"]
	if "accept" != action && "reject" != action {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}

	//5.查找订单表，找到该订单并确定当前订单状态是WAIT_ACCEPT
	o := orm.NewOrm()
	order := models.OrderHouse{}
	if errQuery := o.QueryTable("order_house").Filter("id", order_id).Filter("status", models.ORDER_STATUS_WAIT_ACCEPT).One(&order); nil != errQuery {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		beego.Info(" Orderstatus step 5 error")
		return
	}

	//6.检验该订单的user_id是否是当前用户user_id
	if _, errRelated := o.LoadRelated(&order, "House"); nil != errRelated {
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)

		return
	}
	house := order.House
	if house.User.Id != user_id {
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = "订单用户不匹配，操作无效"
		return
	}
	//7.判断action
	if "accept" == action {
		//如果action为accept(接单)，更换该订单status为WAIT_COMMENT等待用户评价
		order.Status = models.ORDER_STATUS_WAIT_COMMENT
	} else if "reject" == action {
		order.Status = models.ORDER_STATUS_REJECTED
		reason := req["reason"]
		order.Comment = reason.(string)
	}

	if _, errUpdate := o.Update(&order); nil != errUpdate {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		beego.Info(" Orderstatus step 7 error")
		return
	}

	return
}

func (this *OrderController) OrderComment() {
	resp := make(map[string]interface{})
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)

	//1.根据session拿到user_id
	user_id := this.GetSession("user_id")

	//2.通过当前url参数得到当前订单id
	order_id := this.Ctx.Input.Param(":id")

	//3.解析客户端请求的json数据，得到action数据
	var req map[string]interface{}
	if errDecode := json.Unmarshal(this.Ctx.Input.RequestBody, &req); nil != errDecode {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}

	//4.检验评价信息是否合法
	comment := req["comment"].(string)
	if comment == "" {
		resp["errno"] = models.RECODE_PARAMERR
		resp["errmsg"] = models.RecodeText(models.RECODE_PARAMERR)
		return
	}

	//5.查询数据库，订单必须存在，订单状态必须为待评价状态
	o := orm.NewOrm()
	order := models.OrderHouse{}
	if errQuery := o.QueryTable("order_house").Filter("id", order_id).Filter("status", models.ORDER_STATUS_WAIT_COMMENT).One(&order); nil != errQuery {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		beego.Info(" Orderstatus step 5 error")
		return
	}

	//6.确保订单所关联的用户和该用户是同一个人
	if _, errRelated := o.LoadRelated(&order, "User"); nil != errRelated {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)

		return
	}
	if order.User.Id != user_id {
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = "该订单并不属于本人"
		return
	}
	//7.关联查询order订单所关联的house信息
	if _, errRelated := o.LoadRelated(&order, "House"); nil != errRelated {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	house := order.House
	//8.将order和house完整数据更新到数据库中，指定只更新status，comment字段的数据
	order.Status = models.ORDER_STATUS_COMPLETE
	order.Comment = comment
	house.Order_count++
	if _, errUpdate := o.Update(&order, "status", "comment"); nil != errUpdate {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	if _, errUpdate := o.Update(house, "order_count"); nil != errUpdate {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	//9.将house_info_[house_id]存的redis的key删除(因为已经修改了订单数量)
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
	house_info_key := strconv.Itoa(house.Id)
	if errDelete := cache_conn.Delete(house_info_key); nil != errDelete {
		beego.Error("delete", house_info_key, "error,err=", errDelete)
	}

	return
}
