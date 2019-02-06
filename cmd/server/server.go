package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/rc452860/vnet/common/config"
	"github.com/rc452860/vnet/common/log"
	"github.com/rc452860/vnet/db"
	"github.com/rc452860/vnet/proxy/server"
	"github.com/rc452860/vnet/utils/datasize"
)

func main() {
	conf, err := config.LoadDefault()
	log.Info("cpu core: %d", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err != nil {
		log.Err(err)
		return
	}

	ctx, cancle := context.WithCancel(context.Background())
	if conf.Mode == "db" {

		db.DbStarted(ctx)
	} else {
		BareStarted()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	for {
		data := <-c
		log.Error("recive signal: %s", data.String())
		if data == os.Interrupt {
			cancle()
			return
		}
	}
}

func BareStarted() {
	host := flag.String("host", "0.0.0.0", "shadowsocks server host")
	method := flag.String("method", "aes-128-cfb", "shadowsocks method")
	password := flag.String("password", "killer", "shadowsocks password")
	port := flag.Int("port", 1090, "shadowsocks port")
	limit := flag.String("limit", "", "shadowsocks traffic limit exp:4MB")
	timeout := flag.Int("timeout", 3000, "connect timeout (Millisecond)")
	flag.Parse()
	limitNumberical := datasize.MustParse(*limit)
	shadowsocks, err := server.NewShadowsocks(*host, *method, *password, *port, server.ShadowsocksArgs{
		Limit:          limitNumberical,
		ConnectTimeout: time.Duration(*timeout) * time.Millisecond,
	})
	if err != nil {
		log.Err(err)
		return
	}
	if err := shadowsocks.Start(); err != nil {
		log.Err(err)
		return
	}

	go func() {
		tick := time.Tick(1 * time.Second)
		for {
			<-tick
			upSpeed, _ := datasize.HumanSize(shadowsocks.UpSpeed)
			downSpeed, _ := datasize.HumanSize(shadowsocks.DownSpeed)
			upBytes, _ := datasize.HumanSize(shadowsocks.UpBytes)
			downBytes, _ := datasize.HumanSize(shadowsocks.DownBytes)
			log.Info("[upspeed: %s] [downspeed: %s] [up: %s] [down: %s]", upSpeed, downSpeed, upBytes, downBytes)
		}
	}()

}
