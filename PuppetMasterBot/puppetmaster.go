package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-yql"
	"gopkg.in/olivere/elastic.v3"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/killermouse0/PuppetMasterBot/mybot"
)

type Portfolio struct {
	Items []string `json:"items"`
}

func (p *Portfolio) addItems(items []string) {
	hItems := make(map[string]int)

	for _, item := range p.Items {
		hItems[item] = 1
	}
	for _, item := range items {
		hItems[item] = 1
	}
	p.Items = *new([]string)
	for k, _ := range hItems {
		p.Items = append(p.Items, k)
	}
}

func truncToN(s string, n int) (res string) {
	if len(s) > n {
		res = s[:n-1]
	} else {
		res = s
	}
	return res
}

func main() {
	const token = "TOKEN"
	const botUrl = "BOTURL"

	userState := make(map[int]*mybot.UserState)

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
	db, err := sql.Open("yql", "||store://datatables.org/alltableswithkeys")
	if err != nil {
		log.Panic(err)
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
		var ptf Portfolio
		if res != nil {
			json.Unmarshal(*res.Source, &ptf)
		}

		text := "Whatever, bro"
		if command := mybot.GetCommand(update); command != "" {
			delete(userState, userId)
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
				text += "\nWhich indces do you want to delete ? (You can delete more than one at once)\n"
				userState[userId] = mybot.New()
				userState[userId].State = "deleting"
			case "/watchlist":
				stocks := strings.Join(ptf.Items, `","`)
				stocks = `"` + stocks + `"`
				query := "select * from yahoo.finance.quotes where symbol in (" + stocks + ")"
				log.Println(query)
				stmt, err := db.Query(query)
				if err != nil {
					log.Println("YQL query failed :", err)
				} else {
					log.Println("YQL query succeeded")
					text = "<pre>"
					for stmt.Next() {
						var data map[string]interface{}
						stmt.Scan(&data)
						text += fmt.Sprintf("%s\n    |%10s|%10s|\n",
							data["Name"],
							data["ChangeinPercent"],
							data["LastTradePriceOnly"])
					}
					text += "</pre>"
				}
			case "/search":
				text = "Not yet implemented!"
			case "/newtracker":
				args := strings.Fields(update.Message.Text)
				userState[userId].State = "gettingTrackerType"
				text = "OK, let's create that " + args[1] + " tracker!\nWhat's its type? (event, numeric, geo, duration)"
			case "/track":
				//				args := strings.Fields(update.Message.Text)
				text = "Not yet implemented!"
			default:
				text = "Sup bro? Sorry but there's no such command!"
			}
		} else {
			words := strings.Fields(update.Message.Text)
			switch userState[userId].State {
			case "deleting":
				for _, w := range words {
					idx, err := strconv.Atoi(w)
					log.Println("Going to delete", idx)
					if err == nil && idx < len(ptf.Items) {
						ptf.Items[idx] = ""
					}
				}
				var tmpItems []string
				for _, v := range ptf.Items {
					if v != "" {
						tmpItems = append(tmpItems, v)
					}
				}
				ptf.Items = tmpItems
				client.Index().Id(strconv.Itoa(userId)).Index("quotes").Type("portfolio").BodyJson(ptf).Do()
				text = "Portfolio was updated"
				delete(userState, userId)
			}
		}
		message := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		message.ParseMode = "HTML"
		bot.Send(message)
	}
}
