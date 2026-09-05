// handlers.go
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Roelof-Jan/telebot/v3"
)

func startHandler(m *telebot.Message) {
	_, _ = bot.Send(m.Chat, `🌟 <b>Grok2API Telegram 图像生成工作台</b>

支持功能：
• /generate <提示词> [数量] - 普通图片生成（URL）
• /img2img <提示词> - 图生图（需上传图片）
• /transparent <提示词> - 透明背景图生成（PNG，Alpha通道）
• /providers - 查看提供商
• /addprovider - 添加中转代理

当前提供商：Grok2API (127.0.0.1:50066/v1)
模型：自动获取（支持手动指定 model=xxx）`)
}

func generateHandler(m *telebot.Message) {
	args := strings.Fields(m.Text)
	if len(args) < 2 {
		bot.ReplyTo(m, "请提供提示词，例如：/generate 一只猫在星空")
		return
	}

	prompt := strings.Join(args[1:], " ")
	num := 1
	if len(args) > 2 {
		num = parseNum(args[2])
	}

	_, _ = bot.Send(m.Chat, "🔄 正在生成图片...")
	req := &ImageGenRequest{
		Prompt:      prompt,
		NumImages:   num,
		Transparent: false,
	}
	urls, err := GetImageURLs(m, req)
	if err != nil {
		bot.ReplyTo(m, "❌ 生成失败："+err.Error())
		return
	}
	sendImages(m, urls)
}

func img2imgHandler(m *telebot.Message) {
	_, _ = bot.Send(m.Chat, "📸 请上传图片作为参考图，然后发送提示词")
	bot.Handle(telebot.OnPhoto, func(c telebot.Context) error {
		photo := c.Message().Photo
		if photo == nil {
			return c.Reply("未检测到图片")
		}

		file, err := bot.GetFile(photo.FileID)
		if err != nil {
			return err
		}
		photoURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", bot.Token, file.FilePath)

		resp, _ := http.Get(photoURL)
		body, _ := io.ReadAll(resp.Body)
		base64Str := base64.StdEncoding.EncodeToString(body)

		args := strings.Fields(c.Message().Text)
		prompt := strings.Join(args[1:], " ")
		if prompt == "" {
			prompt = "保持原图风格"
		}

		req := &ImageGenRequest{
			Prompt:        prompt,
			NumImages:     1,
			Img2imgBase64: base64Str,
			Transparent:   false,
		}

		urls, err := GetImageURLs(c.Message(), req)
		if err != nil {
			return c.Reply("❌ 生成失败：" + err.Error())
		}

		sendImages(c.Message(), urls)
		return nil
	})
}

func transparentHandler(m *telebot.Message) {
	_, _ = bot.Send(m.Chat, "🌟 正在为你生成透明背景图片...")
	args := strings.Fields(m.Text)
	if len(args) < 2 {
		return
	}

	prompt := strings.Join(args[1:], " ")
	req := &ImageGenRequest{
		Prompt:        prompt,
		NumImages:     1,
		Transparent:   true,
		ResponseFormat: "b64_json",
	}
	urls, err := GetImageURLs(m, req)
	if err != nil {
		bot.ReplyTo(m, "❌ 生成失败："+err.Error())
		return
	}
	sendImages(m, urls)
}

func GetImageURLs(m telebot.Message, req *ImageGenRequest) ([]string, error) {
	for _, provider := range Providers {
		urls, err := provider.Generate(context.Background(), req)
		if err == nil && len(urls) > 0 {
			return urls, nil
		}
	}
	return nil, fmt.Errorf("所有提供商生成失败")
}

func sendImages(m telebot.Message, urls []string) {
	for _, url := range urls {
		_, _ = bot.Send(m.Chat, &telebot.Photo{URL: url})
	}
}

func parseNum(s string) int {
	if n := parseInt(s); n > 0 && n <= 5 {
		return n
	}
	return 1
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
