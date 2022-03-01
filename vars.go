package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var Version = "development"
var LogLevel = "TRACE"
var ConfigFilePath = ""
var Config *configBase

const (
	ReplyTimeout = 15 * time.Second
)

type configEndpoint struct {
	URL       string `yaml:"url"`
	VerifySSL bool   `yaml:"verify_ssl"`
	Timeout   string `yaml:"timeout"`
}

type configBase struct {
	Server struct {
		BindAddress string `yaml:"bind_address"`
		AutoTLS     struct {
			Email    string   `yaml:"email"`
			Enabled  bool     `yaml:"enabled"`
			Redirect bool     `yaml:"Redirect"`
			Domains  []string `yaml:"domains"`
		} `yaml:"autotls"`
	} `yaml:"server"`
	Main     configEndpoint   `yaml:"main"`
	Replicas []configEndpoint `yaml:"replicas"`
}

func LoadConfig() {

	configSetDefaults()
	viper.AutomaticEnv()

	runtimeDir := getRuntimeDir()
	if ConfigFilePath == "" {
		ConfigFilePath = runtimeDir + "/proxies.yaml"
	}

	viper.SetConfigName("proxies")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Dir(ConfigFilePath))

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	cfg := &configBase{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		log.Fatalln("unable to decode into config struct, %v", err)
	}
	Config = cfg

	//
	if IsDevelopment() {
		gin.SetMode(gin.DebugMode)
		log.SetLevel(log.TraceLevel)
	} else {
		// default
		gin.SetMode(gin.ReleaseMode)
		log.SetLevel(log.WarnLevel)

		// custom
		switch strings.ToUpper(LogLevel) {
		case "PANIC":
			log.SetLevel(log.PanicLevel)
		case "FATAL":
			log.SetLevel(log.FatalLevel)
		case "ERR", "ERROR":
			log.SetLevel(log.ErrorLevel)
		case "WARN", "WARNING":
			log.SetLevel(log.WarnLevel)
		case "INFO":
			log.SetLevel(log.InfoLevel)
		case "DEBUG":
			log.SetLevel(log.DebugLevel)
			gin.SetMode(gin.DebugMode)
		case "TRACE":
			log.SetLevel(log.TraceLevel)
			gin.SetMode(gin.DebugMode)
		}

		fmt.Println("LogLevel: ", log.GetLevel())
	}

	// Log
	log.SetFormatter(&log.TextFormatter{})

	log.Debug("============================", "configuration", "============================")
	log.Debug(Config)

}

func getRuntimeDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func IsDevelopment() bool {

	if Version != "development" {
		return false
	}
	return true
}

func configSetDefaults() {

	viper.SetDefault("server.bind_address", ":http")
	viper.SetDefault("server.certs.enabled", false)
	viper.SetDefault("master.url", "http://localhost:8080")
	viper.SetDefault("master.verify_ssl", false)
	viper.SetDefault("replicas", map[string]interface{}{})

}
