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

	userState := make(map[int]string)

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

		text := "Whatever, bro"
		if command := getCommand(update); command != "" {
			switch command {
			case "/add":
				ptf.addItems(strings.Fields(update.Message.Text)[1:])
				client.Index().Id(strconv.Itoa(userId)).Index("quotes").Type("portfolio").BodyJson(ptf).Do()
				text = "Added stock(s) to portfolio"
			case "/del":
				text = "Here's what you have in your portfolio :\n"
				for i, item := range ptf.Items {
					text += fmt.Sprintf("%v - %v\n", i, item)
				}
				text += "\nWhich index do you want to delete ?\n"
				userState[userId] = "deleting"
			case "/watchlist":
				q, err := quote.Retrieve(qs, ptf.Items)
				if err != nil {
					log.Fatalln("Couldn't get the prices:", err)
				}
				text = ""
				for _, sq := range q {
					text += fmt.Sprintf("%v:\t\t%v\n", sq.Symbol, sq.LastTradePrice)
				}
			case "/search":
				text = "Not yet implemented!"
			default:
				text = "Sup bro? Sorry but there's no such command!"
			}
		} else {
			words := strings.Fields(update.Message.Text)
			switch userState[userId] {
			case "deleting":
				for _, w := range words {
					idx, err := strconv.Atoi(w)
					if err == nil && idx < len(ptf.Items) {
						ptf.Items = append(ptf.Items[:idx], ptf.Items[idx+1:]...)
					}
				}
				client.Index().Id(strconv.Itoa(userId)).Index("quotes").Type("portfolio").BodyJson(ptf).Do()
				text = "Portfolio was updated"
			}
		}
		message := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		bot.Send(message)
	}
}

func getCommand(update tgbotapi.Update) (command string) {
	command = ""
	if update.Message.Entities != nil {
		for _, entity := range *update.Message.Entities {
			if entity.Type == "bot_command" {
				command = update.Message.Text[entity.Offset:entity.Offset + entity.Length]
				log.Println("Got command", command)
			}
		}
	}
	return command
}
