package routers

import (
	"github.com/astaxie/beego"
	"loveHome/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/api/v1.0/areas", &controllers.AreaController{}, "get:GetArea")
	beego.Router("/api/v1.0/session", &controllers.SessionController{}, "get:GetSessionData;delete:DeleteSessionData")
	beego.Router("/api/v1.0/sessions", &controllers.SessionController{}, "post:Login")
	beego.Router("/api/v1.0/users", &controllers.UserController{}, "post:Reg")
	beego.Router("/api/v1.0/user", &controllers.UserController{}, "get:GetUserData")
	beego.Router("/api/v1.0/user/name", &controllers.UserController{}, "put:UpdateName")
	beego.Router("/api/v1.0/user/avatar", &controllers.UserController{}, "post:Postavatar")
	beego.Router("/api/v1.0/user/auth", &controllers.UserController{}, "get:AuthGet;post:AuthPost")
	beego.Router("/api/v1.0/user/houses", &controllers.HouseController{}, "get:GetHouseData")
	beego.Router("/api/v1.0/houses", &controllers.HouseController{}, "get:GetHouseSearchData;post:PostHouseData")
	beego.Router("/api/v1.0/houses/index", &controllers.HouseIndexController{}, "get:GetHouseIndex")
	beego.Router("/api/v1.0/houses/?:id/images", &controllers.HouseController{}, "post:PostHouseImage")
	beego.Router("/api/v1.0/houses/?:id", &controllers.HouseController{}, "get:GetDetailHouseData")
	beego.Router("/api/v1.0/orders", &controllers.OrderController{}, "post:PostOrderHouseData")
	beego.Router("/api/v1.0/orders/:id/status", &controllers.OrderController{}, "put:Orderstatus")
	beego.Router("/api/v1.0/user/orders", &controllers.OrderController{}, "get:GetOrderData")
	beego.Router("/api/v1.0/orders/:id/comment", &controllers.OrderController{}, "put:OrderComment")
}
