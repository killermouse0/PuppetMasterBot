package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"fmt"
	"github.com/sckor/quote"
	_ "github.com/sckor/yahoo"
)

func main() {
	const token = "TOKEN"
	const botUrl = "BOTURL"

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(botUrl))
	if err != nil {
		log.Panic(err)
	}

	updates := bot.ListenForWebhook("/")
	go http.ListenAndServe("127.0.0.1:9080", nil)


	qs, err := quote.Open("yahoo", "")
	if err != nil {
		log.Fatalln("Can't open Yahoo API: %v", err)
	}

	for update := range updates {
		// text := fmt.Sprintf("Hello %s\nI'm The Puppet Master. But you can call me Master.", update.Message.From.FirstName)

		q, err := quote.Retrieve(qs, []string{update.Message.Text})
		if err != nil {
			log.Fatalln("Couldn't get the prices: %+v", err)
		}
		log.Println(update.Message.Text)

		text := fmt.Sprintf("%v", q)

		var message = tgbotapi.NewMessage(update.Message.Chat.ID, text)
		bot.Send(message)
	}
}

