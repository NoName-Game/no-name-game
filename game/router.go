package game

import (
	"errors"
	"log"
	"reflect"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	funcs = map[string]interface{}{
		"test": testerFunc,
	}
)

func routing(update tgbotapi.Update) {
	var route string
	if update.Message != nil {
		route = parseMessage(update.Message)
	}

	Call(funcs, route, update)
}

// Call - Method to call another func
func Call(m map[string]interface{}, name string, params ...interface{}) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of params is not adapted")
		return
	}

	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}

	result = f.Call(in)
	return
}

func parseMessage(message *tgbotapi.Message) string {
	parsed := message.Text
	if message.IsCommand() == true {
		parsed = message.Command()
	}

	return parsed
}

func testerFunc(update tgbotapi.Update) {
	log.Println("Here Yeeee")
}
