package main

import (
	"encoding/json"
	"flag"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

type Config struct {
	Token   string `json:"token"`
	Oid     string `json:"oid"`
	Address string `json:"address"`
	AppId   string `json:"appid"`
}

func main() {

	cfg := flag.String("c", "config.json", "config file")
	flag.Parse()
	bytes, err := os.ReadFile(*cfg)
	if err != nil {
		log.Panicf("read config file error: %v", err)
	}
	config := &Config{}
	err = json.Unmarshal(bytes, config)
	if err != nil {
		log.Panicf("parse config file error: %v", err)
	}
	log.Printf("config: %+v", config)

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
