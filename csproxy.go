package main

import (
	"github.com/spf13/viper"
	"log"
	"os"
)

func checkError(err error) {
	if err != nil {
		log.Fatalf("Fatal error: %v", err.Error())
		os.Exit(1)
	}
}

func main() {
	viper.SetConfigName("csproxy")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	checkError(err)

	// Hacky workaround due to https://github.com/spf13/viper/issues/158
	if viper.Get("listeners.carbon.address") == nil {
		viper.Set("listeners.carbon.address", "127.0.0.1")
	}
	if viper.Get("listeners.carbon.port") == nil {
		viper.Set("listeners.carbon.port", 2003)
	}
	if viper.Get("listeners.http.address") == nil {
		viper.Set("listeners.http.address", "127.0.0.1")
	}
	if viper.Get("listeners.http.port") == nil {
		viper.Set("listeners.http.port", 9080)
	}

	done := make(chan bool, 1)
	metrics := make(chan *carbonMetric)
	go carbonListener(
		viper.GetString("listeners.carbon.address"),
		viper.GetInt("listeners.carbon.port"),
		done, metrics)

	// for status
	go httpListener(
		viper.GetString("listeners.http.address"),
		viper.GetInt("listeners.http.port"),
		done)

	go statsdWriter(
		viper.GetString("outputs.statsd.address"),
		viper.GetInt("outputs.statsd.port"),
		done, metrics)

	<-done
}
