package main

import (
	"github.com/spf13/viper"

	"github.com/ben-turner/sarteam/internal/sarteam"
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

	s := sarteam.New(config)

	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
