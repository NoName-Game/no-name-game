package app

import (
	"errors"
	"reflect"
	"strconv"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	funcs = map[string]interface{}{
		"the-answer-is":    theAnswerIs,
		"test-multi-state": testMultiState,
	}
)

// Routing - Check message type and call if exist the correct function
func routing(update tgbotapi.Update) {
	if update.Message != nil {
		if ok := checkUser(update.Message); ok {
			var route string

			route = parseMessage(update.Message)

			//FIXME:
			routeCache, _ := services.Redis.Get(strconv.FormatUint(uint64(player.ID), 10)).Result()
			if routeCache != "" {
				route = routeCache
			}

			// if player.State.Function != "" {
			// 	route = player.State.Function
			// }

			// Check if command exist.
			if _, ok := funcs[route]; ok {
				Call(funcs, route, update)
			}
		}
	}
}

// Call - Method to call another func and check needed parameters
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

// Parse message text, if command it's like telegram format the message will be parsed and return simple text without "/" char
func parseMessage(message *tgbotapi.Message) (parsed string) {
	parsed = message.Text
	if message.IsCommand() == true {
		parsed = message.Command()
	}

	return
}
