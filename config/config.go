package config

import (
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	App   `mapstructure:"app"`
	Log   `mapstructure:"log"`
	Mysql `mapstructure:"mysql"`
	Redis `mapstructure:"redis"`
	Sound `mapstructure:"sound"`
}

type App struct {
	Port int `mapstructure:"port"`
}

type Log struct {
	FileName     string `mapstructure:"fileName"`
	MaxSize      int    `mapstructure:"maxSize"`
	MaxBackups   int    `mapstructure:"maxBackups"`
	MaxAge       int    `mapstructure:"maxAge"`
	IsConsoleLog bool   `mapstructure:"consoleLog"`
}

type Mysql struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	UserName string `mapstructure:"userName"`
	PassWord string `mapstructure:"passWord"`
	DbName   string `mapstructure:"dbName"`
}

type Redis struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type Sound struct {
	Token   string `mapstructure:"token"`
	Url     string `mapstructure:"url"`
	Version int    `mapstructure:"version"`
}

var AppConfig Config

func Init() {
	//设置配置文件路径
	var configFile string
	// 设置 flag ; -c /app/config.yaml 指定文件地址
	flag.StringVar(&configFile, "c", "", "choose config file.")
	flag.Parse()
	if len(configFile) == 0 {
		configFile = "./app.yaml"
	}

	// os env todo
	// if configEnv := os.Getenv("VIPER_CONFIG"); configEnv != "" {
	// 	configFile = configEnv
	// }

	//初始化viper
	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// 监听配置文件
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		// 重载配置
		if err := v.Unmarshal(&AppConfig); err != nil {
			fmt.Printf("WatchConfig :unable to decode into struct, %v", err)
		}

		fmt.Printf("changed AppConfig: %v\n", AppConfig)
	})

	//值解组到结构体
	err2 := v.Unmarshal(&AppConfig)
	if err2 != nil {
		fmt.Printf("unable to decode into struct, %v", err)
	}

}
