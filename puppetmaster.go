package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"fmt"
	"github.com/sckor/quote"
	_ "github.com/sckor/yahoo"
	"gopkg.in/olivere/elastic.v3"
	"strconv"
	"encoding/json"
)

type Portfolio struct {
	Items	[]string `json:"items"`
}

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
		log.Fatalln("Can't open Yahoo API:", err)
	}

	/* Elasticsearch connection */
	client, err := elastic.NewClient()
	if err != nil {
		log.Fatalln("Couldn't connect to Elasticsearch :", err)
	}

	/* Update processing loop */
	for update := range updates {
		var command string
		userId := update.Message.From.ID

		for _, entity := range *update.Message.Entities {
			if entity.Type == "bot_command" {
				command = update.Message.Text[entity.Offset:entity.Offset + entity.Length]
				log.Println("Got command", command)
			}
		}
		log.Println("Got message from user.ID =", userId)
		res, err := client.Get().
			Index("quotes").
			Type("portfolio").
			Id(strconv.Itoa(userId)).
			Do()
		if err != nil {
			log.Println("Couldn't find user portfolio for user=",
				userId, ":", err)
		}
		var ptf Portfolio;
		if res != nil {
			json.Unmarshal(*res.Source, &ptf)
		}

		log.Println(fmt.Sprintf("got ptf = %#v", ptf.Items))
		q, err := quote.Retrieve(qs, []string{update.Message.Text})
		if err != nil {
			log.Fatalln("Couldn't get the prices:", err)
		}
		log.Println(update.Message.Text)

		text := fmt.Sprintf("%v", q)

		var message = tgbotapi.NewMessage(update.Message.Chat.ID, text)
		bot.Send(message)
	}
}

