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
	"strings"
)

type Portfolio struct {
	Items	[]string `json:"items"`
}

func (p *Portfolio) addItems(items []string) {
	hItems := make(map[string] int)

	for _, item := range p.Items {
		hItems[item] = 1
	}
	for _, item := range items {
		hItems[item] = 1
	}
	p.Items = *new([]string)
	for k, _ := range(hItems) {
		p.Items = append(p.Items, k)
	}
}

func main() {
	const token = "TOKEN"
	const botUrl = "BOTURL"

	/* Bot setup */
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Can't connect to Telegram Bot API", err)
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
		userId := update.Message.From.ID

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

		if command := getCommand(update); command != "" {
			switch command {
			case "/add":
				log.Println("Message is :", update.Message.Text)
				ptf.addItems(strings.Fields(update.Message.Text)[1:])
			case "/del":
			}
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

func getCommand(update tgbotapi.Update) (command string) {
	command = ""
	for _, entity := range *update.Message.Entities {
		log.Println(fmt.Sprintf("Entity is %#v", entity))
		if entity.Type == "bot_command" {
			command = update.Message.Text[entity.Offset:entity.Offset + entity.Length]
			log.Println("Got command", command)
		}
	}
	return command
}
