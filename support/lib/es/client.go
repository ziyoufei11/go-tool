package es

import (
	"crypto/tls"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net"
	"net/http"
	"sync"
	"time"
)

var isConnected bool
var Client *elasticsearch.Client
var once sync.Once

func Init() {
	once.Do(func() {
		err := connectToEs()
		if err != nil {
			color.Red("[ES] Init es error: %s", err)
			return
		}
		//心跳重连
		go heartbeat()
	})
}

func heartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		if !isConnected {
			v, err := Client.Ping()
			if err != nil || v.IsError() {
				fmt.Printf("[ES] Error: %v %v \n", v, err)
				err = connectToEs()
				if err == nil {
					fmt.Println("[ES] Reconnected")
				} else {
					fmt.Println("[ES] Reconnected Error", zap.Error(err))
				}
			}
		}
	}
}

func connectToEs() error {
	hosts := viper.GetStringSlice("es.addresses")
	user := viper.GetString("es.username")
	pass := viper.GetString("es.password")

	fmt.Printf("[ES] hosts %s, user %s, pass %s \n", hosts, user, pass)
	cfg := elasticsearch.Config{
		Addresses: hosts,
		Username:  user,
		Password:  pass,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   1000,
			ResponseHeaderTimeout: time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	var err error
	Client, err = elasticsearch.NewClient(cfg)
	if err != nil {
		isConnected = false
		return err
	}

	isConnected = true
	return nil
}
