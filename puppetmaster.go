package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"fmt"
)

func main() {
	const token = "TOKEN"
	const botUrl = "BOTURL"

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(botURL))
	if err != nil {
		log.Panic(err)
	}

	updates := bot.ListenForWebhook("/")
	go http.ListenAndServe("127.0.0.1:9080", nil)
	for update := range updates {
		text := fmt.Sprintf("Hello %s\nI'm The Puppet Master. But you can call me Master.", update.Message.From.FirstName)
		var message = tgbotapi.NewMessage(update.Message.Chat.ID, text)
		bot.Send(message)
	}
}
