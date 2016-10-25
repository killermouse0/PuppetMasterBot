package mybot

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

type UserState struct {
	Command string
	State string
	Stash map[string]interface{}
}

func New() (res *UserState) {
	res = new(UserState)
	res.Stash = make(map[string]interface{})
	return res
}

func GetCommand(update tgbotapi.Update) (command string) {
	command = ""
	if update.Message.Entities != nil {
		for _, entity := range *update.Message.Entities {
			if entity.Type == "bot_command" {
				command = update.Message.Text[entity.Offset : entity.Offset+entity.Length]
				log.Println("Got command", command)
			}
		}
	}
	return command
}
