package main

import (
	"encoding/json"
	"flag"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	Token   string `json:"token"`
	Oid     string `json:"oid"`
	Address string `json:"address"`
	AppId   string `json:"appid"`
}

func loadConfig(path string) (*Config, error) {
	if strings.HasPrefix(path, "http") {
		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		config := &Config{}
		err = json.NewDecoder(resp.Body).Decode(config)
		if err != nil {
			return nil, err
		}
		return config, nil
	} else {
		bytes, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		config := &Config{}
		err = json.Unmarshal(bytes, config)
		if err != nil {
			return nil, err
		}
		return config, nil
	}
}

func main() {

	cfg := flag.String("config", "config.json", "config file")
	flag.Parse()

	config, err := loadConfig(*cfg)
	if err != nil {
		log.Fatal(err)
	}

	yzj := NewYunZhiJia(config.Token, config.Oid, config.AppId)
	server := gin.Default()
	status := map[string]ClockInTimeType{
		"start": ClockInTimeTypeStart,
		"end":   ClockInTimeTypeEnd,
	}
	for k, v := range status {
		t := v // fix: when use v in closure, it will always be the last value
		server.GET("/"+k, func(c *gin.Context) {
			ok, _ := yzj.IsClockInToday(t)
			if ok {
				c.String(200, "true")
			} else {
				c.String(200, "false")
			}
		})
	}
	if gin.Mode() == "debug" {
		server.GET("/raw", func(c *gin.Context) {
			flow, e := yzj.ClockInFlow()
			if e != nil {
				c.JSON(500, gin.H{
					"error": e.Error(),
				})
				return
			}
			c.JSON(200, flow)
		})
	}
	server.GET("/status", func(c *gin.Context) {
		c.String(200, "ok")
	})
	_ = server.Run(config.Address)
}
