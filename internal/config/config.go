package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Metrics struct {
	LoadAvg              bool `mapstructure:"load_avg"`
	CPUAvg               bool `mapstructure:"cpu_avg"`
	DiskIO               bool `mapstructure:"disk_io"`
	DiskUsage            bool `mapstructure:"disk_usage"`
	NetConnections       bool `mapstructure:"net_connections"`
	NetConnectionsStates bool `mapstructure:"net_connections_states"`
	NetTopByProtocol     bool `mapstructure:"net_top_by_protocol"`
	NetTopByClients      bool `mapstructure:"net_top_by_connection"`
}

type SystemPoints struct {
	Proc          string `mapstructure:"proc"`
	Sys           string `mapstructure:"sys"`
	Dev           string `mapstructure:"dev"`
	Run           string `mapstructure:"run"`
	TCP           string `mapstructure:"tcp"`
	TCP6          string `mapstructure:"tcp6"`
	UDP           string `mapstructure:"udp"`
	UDP6          string `mapstructure:"udp6"`
	ProcMountInfo string `mapstructure:"procMountInfo"`
	Interface     string `mapstructure:"interface"`
}

type DaemonConfig struct {
	Host     string       `mapstructure:"host"`
	Port     string       `mapstructure:"port"`
	Metrics  Metrics      `mapstructure:"metrics"`
	System   SystemPoints `mapstructure:"system"`
	LogLevel string       `mapstructure:"log_level"`
}

func (d *DaemonConfig) Validate() error {
	validate := validator.New()
	err := validate.Var(d.LogLevel, "oneof=DEBUG INFO WARNING ERROR CRITICAL")
	if err != nil {
		var fe validator.ValidationErrors
		errors.As(err, &fe)
		return fmt.Errorf("invalid log level value: %s", fe[0].Value())
	}
	return nil
}

func Load(path string, cfg interface{}) error {
	setDefaults()
	viper.SetConfigType("yaml")
	viper.SetConfigFile(path)

	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("load config error: %v\n", err.Error())
		log.Println("use default values")
	}
	for _, k := range viper.AllKeys() {
		value := viper.GetString(k)
		isEnvVar := strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}")
		if isEnvVar {
			key := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
			viper.Set(k, os.Getenv(key))
		}
	}
	if err = viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("%w: %s", ErrConfigBindingFail, err.Error())
	}

	return nil
}
