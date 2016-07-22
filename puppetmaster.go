package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
)

func main() {
	const token = "252613835:AAFvcaqHZKwefdaeq-w0Cm1fOUCqYFOvefo"

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	_, err = bot.SetWebhook(tgbotapi.NewWebhook("https://brainless.me/dkshasdrfiweyrausydfjg"))
	if err != nil {
		log.Panic(err)
	}

	updates := bot.ListenForWebhook("/")
	go http.ListenAndServe("127.0.0.1:9080", nil)
	for update := range updates {
		var message = tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		bot.Send(message)
	}
}
