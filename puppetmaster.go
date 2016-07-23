package main

import (
	"database/sql"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"fmt"
	_ "github.com/mattn/go-yql"
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
	db, _ := sql.Open("yql", "")
	go http.ListenAndServe("127.0.0.1:9080", nil)
	for update := range updates {
		stmt, _ := db.Query(
			"select * from yahoo.finance.quotes where symbol = ?",
			"TSLA")
		for stmt.Next() {
			var data interface{}
			stmt.Scan(&data)
			// text := fmt.Sprintf("Hello %s\nI'm The Puppet Master. But you can call me Master.", update.Message.From.FirstName)
			text := fmt.Sprintf("%v", data)
			var message = tgbotapi.NewMessage(update.Message.Chat.ID, text)
			bot.Send(message)
		}
	}
}

func getQuotes() {

}
