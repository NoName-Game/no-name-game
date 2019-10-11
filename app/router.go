package app

import (
	"errors"
	"reflect"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
)

// Routing - Check message type and call if exist the correct function
func routing(update tgbotapi.Update) {
	// ************************
	// Switch by message type
	// ************************
	var callingRoute string
	if update.Message != nil {
		callingRoute = parseMessage(update.Message)
	} else if update.CallbackQuery != nil {
		callingRoute = parseCallback(update.CallbackQuery)
	}

	// ******************************************
	// Check if callingRoute it's breaker routes
	// ******************************************
	isBreakerRoute, route := inRoutes(callingRoute, BreakerRoutes)
	if isBreakerRoute {
		Invoke(BreakerRoutes[route], "Handle", update)
		return
	}

	// ******************************************
	// Check if player have route in cache
	// ******************************************
	isCachedRoute := helpers.GetRedisState(helpers.Player)
	if isCachedRoute != "" {
		Invoke(Routes[isCachedRoute], "Handle", update)
		return
	}

	// ******************************************
	// Check if it's normal route
	// ******************************************
	isRoute, route := inRoutes(callingRoute, Routes)
	if isRoute {
		Invoke(Routes[route], "Handle", update)
		return
	}

	return
}

// inRoutes - Check if message is translated command
func inRoutes(messageRoute string, routeList map[string]interface{}) (isRoute bool, route string) {
	for route := range routeList {
		if strings.ToLower(helpers.Trans(route)) == messageRoute {
			return true, route
		}
	}

	return false, ""
}

// Invoke - Dinamicaly call method interface
func Invoke(any interface{}, name string, args ...interface{}) {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}

	reflect.ValueOf(any).MethodByName(name).Call(inputs)
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
	if message.IsCommand() {
		parsed = message.Command()
	}

	return strings.ToLower(parsed)
}

func parseCallback(callback *tgbotapi.CallbackQuery) (parsed string) {
	parsed = callback.Data
	parsed = strings.Split(parsed, ".")[0]
	return strings.ToLower(parsed)
}
