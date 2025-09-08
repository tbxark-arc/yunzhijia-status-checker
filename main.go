package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sphere/confstore"
	"github.com/go-sphere/confstore/codec"
	"github.com/go-sphere/confstore/provider"
	"github.com/go-sphere/confstore/provider/file"
	"github.com/go-sphere/confstore/provider/http"
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

	config, err := confstore.Load[Config](provider.NewSelect(*conf,
		provider.If(file.IsLocalPath, func(s string) provider.Provider {
			return file.New(s)
		}),
		provider.If(http.IsRemoteURL, func(s string) provider.Provider {
			return http.New(s, http.WithTimeout(10*time.Second))
		}),
	), codec.JsonCodec())
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
