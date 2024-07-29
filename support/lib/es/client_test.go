package es

import (
	"github.com/spf13/viper"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../../")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name string
	}{
		{
			name: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init()
			time.Sleep(10 * time.Second)
		})
	}
}
