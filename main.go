package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
)
// .env自定义配置数据存放字典
var customList map[string]interface{}
// 存储环境标识
var Env interface{}
// 获取.env自定义配置并写入数据字典
func viperGet(custom *viper.Viper) {
	// 查找并读取配置文件
	if err := custom.ReadInConfig(); err != nil {
		fmt.Println(fmt.Errorf("Fatal error config file: %s \n", err)) // 读取配置文件失败致命错误
	}
	customList = make(map[string]interface{})
	customAll := custom.AllKeys()
	// 有自定义数据就覆盖原有数据
	if len(customAll) > 0 {
		for _, index := range customAll {
			customList[index] = custom.Get(index)
		}
	}
}
// 合并自定义配置和现在配置
func viperSet(configuration *viper.Viper) {
	if len(customList) > 0 && customList["env"] != Env {
		ConfigNameChange(configuration, customList["env"])
	} else if len(customList) == 0 {
		ConfigNameChange(configuration, "prod")
	}
	if err := configuration.ReadInConfig(); err != nil {
		fmt.Println(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	nowConfigAll := configuration.AllKeys()
	if len(nowConfigAll) > 0 {
		if len(customList) > 0 {
			if err := configuration.MergeConfigMap(customList); err != nil {
				fmt.Println(fmt.Errorf("Failed to merge configuration files: %s \n", err))
			}
		}
	}
}
// 设置配置文件路径
func ConfigNameChange(configuration *viper.Viper, env interface{}) {
	Env = env
	// 读取预先定义好的配置文件
	configuration.AddConfigPath("config")
	switch Env {
	case "dev":
		configuration.SetConfigName("dev")
	case "prod":
		configuration.SetConfigName("prod")
	case "test":
		configuration.SetConfigName("test")
	default:
		configuration.SetConfigName("prod")
	}
	configuration.SetConfigType("yaml")
}
// 监控配置文件是否发生变化并处理数据合并
func WatchConfigChange(custom *viper.Viper, function func()) {
	custom.WatchConfig()
	custom.OnConfigChange(func(e fsnotify.Event) {
		function()
	})
}

func main() {
	// 自定义.env
	custom := viper.New()
	// 预定义配置文件
	configuration := viper.New()
	// 设定自定义.env路径和文件名称
	custom.SetConfigFile(".env")
	// 设定读取数据格式
	custom.SetConfigType("yaml")
	// 赋值自定义配置
	viperGet(custom)
	// 监控.env配置文件变化
	WatchConfigChange(custom, func() {
		viperGet(custom)
		viperSet(configuration)
	})
	// 获取运行环境
	_ = configuration.BindEnv("Env", "Env")
	env := custom.Get("Env")
	ConfigNameChange(configuration, env)
	viperSet(configuration)
	// 监控config配置文件变化
	WatchConfigChange(configuration, func() {
		viperSet(configuration)
	})

	r := gin.Default()
	// 访问/version的返回值会随配置文件的变化而变化
	r.GET("/version", func(c *gin.Context) {
		data := make([]interface{}, 0)
		ginMode := configuration.Get("ginmode")
		drivername := configuration.Get("database.drivername")
		data = append(data, ginMode)
		data = append(data, drivername)
		c.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg": "输出",
			"data": data,
		})
	})

	if err := r.Run("0.0.0.0:8000"); err != nil {
		fmt.Println(err)
	}
}
