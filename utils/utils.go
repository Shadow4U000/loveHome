package utils

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
)

//定义config变量
var (
	G_real_addr      string //本地ip
	G_server_addr    string //服务器IP
	G_server_port    string //服务器端口
	G_redis_addr     string //redisIP
	G_redis_port     string //redis端口
	G_redis_dbnum    string //redis db编号
	G_mysql_addr     string //mysql IP
	G_mysql_port     string //mysql Port
	G_mysql_dbname   string //mysql db库名
	G_fdfs_http_addr string //fdfs nginx ip地址
)

func InitConfig() {
	//从配置文件读取配置信息
	appconf, err := config.NewConfig("ini", "conf/app.conf")
	if nil != err {
		beego.Debug(err)
		return
	}
	G_real_addr = appconf.String("realaddr")
	G_server_addr = appconf.String("httpaddr")
	G_server_port = appconf.String("httpport")
	G_redis_addr = appconf.String("redisaddr")
	G_redis_port = appconf.String("redisport")
	G_redis_dbnum = appconf.String("redisdbnum")
	G_mysql_addr = appconf.String("mysqladdr")
	G_mysql_port = appconf.String("mysqlport")
	G_mysql_dbname = appconf.String("mysqldbname")
	G_fdfs_http_addr = appconf.String("fdfs_http_addr")
}

func init() {
	InitConfig()
}
