package app

import (
	"errors"
	"reflect"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
)

// Routing - Effetua check sul tipo di messagio ed esegue un routing
func routing(player nnsdk.Player, update tgbotapi.Update) {
	// Verifica il tipo di messaggio
	var callingRoute string
	if update.Message != nil {
		callingRoute = parseMessage(update.Message)
	} else if update.CallbackQuery != nil {
		callingRoute = parseCallback(update.CallbackQuery)
	}

	// Verifico se è una rotta di chiusura
	// Questa tipologia di rotta implica un blocco immediato dell'azione in corso
	isBreakerRoute, route := inRoutes(player.Language.Slug, callingRoute, BreakerRoutes)
	if isBreakerRoute {
		invoke(BreakerRoutes[route], "Handle", player, update)
		return
	}

	// Verifico se in memorià è presente già una rotta
	// userò quella come main per gestire ulteriori sottostati
	isCachedRoute, _ := helpers.GetRedisState(player)
	if isCachedRoute != "" {
		invoke(Routes[isCachedRoute], "Handle", player, update)
		return
	}

	// Dirigo ad una rotta normale
	isRoute, route := inRoutes(player.Language.Slug, callingRoute, Routes)
	if isRoute {
		invoke(Routes[route], "Handle", player, update)
		return
	}

	return
}

// inRoutes - Verifica se esiste la rotta
func inRoutes(lang string, messageRoute string, routeList map[string]interface{}) (isRoute bool, route string) {
	// Ciclo lista di rotte
	for route := range routeList {
		// Traduco le rotte in base alla lingua del player per trovare corrispondenza
		if strings.ToLower(helpers.Trans(lang, route)) == messageRoute {
			return true, route
		}
	}

	return false, ""
}

// invoke - Invoco dinamicamente un metodo di un controller
func invoke(any interface{}, name string, args ...interface{}) {
	// Recupero possibili input e li trasformo come argomenti da passare al metodo
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}

	reflect.ValueOf(any).MethodByName(name).Call(inputs)
}

// call - Metodo dedicato al richiamare dinamicamente una specifca funzione
func call(m map[string]interface{}, name string, params ...interface{}) (result []reflect.Value, err error) {
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

// Metodo per il parsing del messaggio
func parseMessage(message *tgbotapi.Message) (parsed string) {
	parsed = message.Text
	if message.IsCommand() {
		parsed = message.Command()
	}

	return strings.ToLower(parsed)
}

// Metodo per il parsing della callback
func parseCallback(callback *tgbotapi.CallbackQuery) (parsed string) {
	parsed = callback.Data
	// TODO: spiegare perchè facciamo così
	parsed = strings.Split(parsed, ".")[0]
	return strings.ToLower(parsed)
}
