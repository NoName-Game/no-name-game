package app

import (
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

	// Dirigo ad una rotta normale
	isRoute, route := inRoutes(player.Language.Slug, callingRoute, Routes)
	if isRoute {
		invoke(Routes[route], "Handle", player, update)
		return
	}

	// Verifico se in memorià è presente già una rotta e se quella richiamata non sia menu
	// userò quella come main per gestire ulteriori sottostati
	isCachedRoute, _ := helpers.GetRedisState(player)
	if isCachedRoute != "" {
		invoke(Routes[isCachedRoute], "Handle", player, update)
		return
	}

}

// inRoutes - Verifica se esiste la rotta
func inRoutes(lang string, messageRoute string, routeList map[string]reflect.Type) (isRoute bool, route string) {
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
func invoke(any reflect.Type, name string, args ...interface{}) {
	// Recupero possibili input e li trasformo come argomenti da passare al metodo
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	v := reflect.New(any)
	v.MethodByName(name).Call(inputs)
}

// Metodo per il parsing del messaggio
func parseMessage(message *tgbotapi.Message) (parsed string) {
	parsed = message.Text
	if message.IsCommand() {
		parsed = message.Command()
		// Se è un comando ed è start lo parso come tutorial
		if parsed == "start" {
			parsed = "📖 Tutorial"
		}
	}

	return strings.ToLower(parsed)
}

// Metodo per il parsing della callback
func parseCallback(callback *tgbotapi.CallbackQuery) (parsed string) {
	// Prendo la prima parte del callback che contiene la rotta
	parsed = strings.Split(callback.Data, ".")[0]

	return strings.ToLower(parsed)
}
