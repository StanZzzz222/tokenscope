package common

import (
	"fmt"
	"github.com/spf13/viper"
)

/*
   Created by zyx
   Date Time: 2025/9/17
   File: config.go
*/

// InitConfig Init config
func InitConfig() {
	viper.SetConfigFile("./config.toml")
	viper.SetConfigType("toml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("Init config falied, %v", err.Error()))
	}
	viper.WatchConfig()
}
