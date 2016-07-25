package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"fmt"
	"github.com/sckor/quote"
	_ "github.com/sckor/yahoo"
	"github.com/olivere/elastic"
	"strconv"
)

func main() {
	const token = "TOKEN"
	const botUrl = "BOTURL"

	/* Bot setup */
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


	/* Yahoo Finance API connection */
	qs, err := quote.Open("yahoo", "")
	if err != nil {
		log.Fatalln("Can't open Yahoo API: %v", err)
	}

	/* Elasticsearch connection */
	client, err := elastic.NewClient()
	if err != nil {
		log.Fatalln("Couldn't connect to Elasticsearch : %v", err)
	}

	/* Update processing loop */
	for update := range updates {
		userId := update.Message.From.ID
		log.Println("Got message from user.ID = %v", userId)
		_, err := client.Get().Id(strconv.Itoa(userId)).Do()
		if err != nil {
			log.Println("Couldn't find user portfolio for %v : %v",
				userId,
				err)
		}
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

