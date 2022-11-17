package main

import (
	"context"

	"github.com/spf13/viper"

	"github.com/ben-turner/sarteam/sarteam"
)

// main is the entry point for the application. It creates a new SARTeam instance and starts it.
func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.sarteam")
	viper.AddConfigPath("/etc/sarteam")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	config := &sarteam.Config{}
	if err := viper.Unmarshal(config); err != nil {
		panic(err)
	}

	s, err := sarteam.New(config)
	if err != nil {
		panic(err)
	}

	err = s.Start(context.Background())
	if err != nil {
		panic(err)
	}
}
