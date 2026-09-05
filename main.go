// main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Roelof-Jan/telebot/v3"
)

var bot *telebot.Bot

func main() {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("❌ TELEGRAM_TOKEN 环境变量未设置")
	}

	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	var err error
	bot, err = telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	LoadProviders()

	bot.Handle("/start", startHandler)
	bot.Handle("/generate", generateHandler)
	bot.Handle("/img2img", img2imgHandler)
	bot.Handle("/transparent", transparentHandler)
	bot.Handle("/providers", providersHandler)
	bot.Handle("/addprovider", addProviderHandler)

	log.Println("🤖 Grok2API Telegram 图像生成工作台（自动模型 + 透明图）已启动")
	bot.Start()
}
