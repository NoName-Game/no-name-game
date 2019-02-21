package app

import (
	"errors"
	"reflect"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Routing - Check message type and call if exist the correct function
func routing(update tgbotapi.Update) {
	if update.Message != nil {
		if player := helpers.CheckUser(update.Message); player.ID >= 1 {
			route := parseMessage(update.Message)

			if helpers.InArray(route, breakerRoutes) != true {
				routeCache := helpers.GetRedisState(player)
				if routeCache != "" {
					route = routeCache
				}
			}

			// Check if command exist.
			if _, ok := routes[route]; ok {
				Call(routes, route, update, player)
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
