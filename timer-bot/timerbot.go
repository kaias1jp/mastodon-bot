package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"./config"
	m "github.com/kaias1jp/go-mastodon"
)

func main() {
	const layout = "15:04:05"

	confDir := "./env/"
	appMode := os.Getenv("APP_MODE") // 動作環境をAPP_MODEという形で環境変数に格納する
	if appMode == "" {
		panic("failed to get application mode, check whether APP_MODE is set.")
	}
	conf, err := config.NewConfig(confDir, appMode) // 引数に渡す
	if err != nil {
		panic(err.Error())
	}

	config := &m.Config{
		Server:       conf.APP.Server,
		ClientID:     conf.APP.ClientID,
		ClientSecret: conf.APP.ClientSecret,
	}

	c := m.NewClient(config)

	err = c.Authenticate(context.Background(), conf.APP.Email, conf.APP.Password)
	if err != nil {
		log.Fatal(err)
	}

	t := time.NewTicker(1 * time.Second) // 1秒おきに通知
	defer t.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer signal.Stop(sig)
	for {
		select {
		case now := <-t.C:
			// 1秒経過した。ここで何かを行う。
			if now.Minute() == 00 && now.Second() == 00 {
				message := now.Format(layout) + "をお知らせします"
				fmt.Println(message)
				c.PostStatus(context.Background(), &m.Toot{
					Status:     message,
					Visibility: "unlisted",
				})
			}
		case s := <-sig:
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				fmt.Println("Stop!")
				return
			}
		}
	}
}
