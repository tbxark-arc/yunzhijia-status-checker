package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/TBXark/confstore"
	"github.com/gin-gonic/gin"
)

var BuildVersion = "dev"

type Config struct {
	Token   string `json:"token"`
	Oid     string `json:"oid"`
	Address string `json:"address"`
	AppId   string `json:"appid"`
}

func main() {

	conf := flag.String("config", "config.json", "config file")
	help := flag.Bool("help", false, "show help")
	flag.Parse()

	if *help {
		fmt.Printf("version: %s\n", BuildVersion)
		flag.Usage()
		return
	}

	config, err := confstore.Load[Config](*conf)
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
