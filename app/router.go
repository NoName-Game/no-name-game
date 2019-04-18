package app

import (
	"errors"
	"reflect"
	"strings"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Routing - Check message type and call if exist the correct function
func routing(update tgbotapi.Update) {
	if update.Message != nil {
		if player := helpers.CheckUser(update.Message); player.ID >= 1 {
			if player.Stats.LifePoint == 0 {
				msg := services.NewMessage(player.ChatID, helpers.Trans("playerDie", player.Language.Slug))
				msg.ParseMode = "HTML"
				services.SendMessage(msg)
			} else {
				callingRoute := parseMessage(update.Message)

				// ******************************************
				// Check if callingRoute it's breaker routes
				// ******************************************
				isBreakerRoute, route := inRoutes(callingRoute, breakerRoutes, player)
				if isBreakerRoute {
					_, err := Call(breakerRoutes, route, update, player)
					if err != nil {
						services.ErrorHandler("Error in call command", err)
					}
					return
				}

				// ******************************************
				// Check if player have route in cache
				// ******************************************
				isCachedRoute := helpers.GetRedisState(player)
				if isCachedRoute != "" {
					_, err := Call(routes, isCachedRoute, update, player)
					if err != nil {
						services.ErrorHandler("Error in call command", err)
					}
					return
				}

				// ******************************************
				// Check if it's normal route
				// ******************************************
				isRoute, route := inRoutes(callingRoute, routes, player)
				if isRoute {
					_, err := Call(routes, route, update, player)
					if err != nil {
						services.ErrorHandler("Error in call command", err)
					}
					return
				}
			}
		}
	}
}

// inRoutes - Check if message is translated command
func inRoutes(messageRoute string, routeList map[string]interface{}, player models.Player) (isRoute bool, route string) {
	for route := range routeList {
		if strings.ToLower(helpers.Trans(route, player.Language.Slug)) == messageRoute {
			return true, route
		}
	}

	return false, ""
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
