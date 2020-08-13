package controllers

import (
	"fmt"
	"math"
	"time"

	"github.com/golang/protobuf/ptypes"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/services"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// MenuController
// Modulo dedicato alla gestine e visualizazione della keyboard di telegram
// in base a dove si trova il player verranno mostrati tasti e action differenti.
// ====================================
type MenuController struct {
	BaseController
	SafePlanet  bool // Flag per verificare se il player si trova su un pianeta sicuro
	TitanPlanet bool // Flag per verificare se il player si trova su un pianeta titano
	Payload     interface{}
}

// ====================================
// Handle
// ====================================
func (c *MenuController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	var err error

	// Carico controller data
	c.LoadControllerData("route.menu", player, update)

	// Recupero messaggio principale
	var recap string
	recap, err = c.GetRecap()
	if err != nil {
		panic(err)
	}

	msg := services.NewMessage(c.Player.ChatID, recap)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard:       c.GetKeyboard(),
		ResizeKeyboard: true,
	}

	// Send recap message
	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}

	// Se il player è morto non può fare altro che riposare o azioni che richiedono riposo
	if c.PlayerStats.GetDead() {
		restsController := new(ShipRestsController)
		restsController.Handle(c.Player, c.Update, true)
	}
}

func (c *MenuController) Validator() {
	//
}

func (c *MenuController) Stage() {
	//
}

// GetRecap
// BoardSystem v0.1
// 🌏 Nomepianeta
// 👨🏼‍🚀 Casteponters
// ♥️ life/max-life
// 💰 XXXX 💎 XXXX
//
// ⏱ Task in corso:
// - LIST
func (c *MenuController) GetRecap() (message string, err error) {
	// Recupero posizione player
	var planet string
	planet, err = c.GetPlayerPosition()
	if err != nil {
		return message, err
	}

	// Costruisco messaggio di racap in base a dove si trova il player

	if c.SafePlanet {
		message = helpers.Trans(c.Player.Language.Slug, "menu.safeplanet")
	} else if c.TitanPlanet {
		message = helpers.Trans(c.Player.Language.Slug, "menu.titanplanet")
	} else {
		// Recupero status vitale del player
		var life string
		life, err = c.GetPlayerLife()
		if err != nil {
			return message, err
		}

		message = helpers.Trans(c.Player.Language.Slug, "menu",
			planet,
			life,
			c.GetPlayerTasks(),
		)
	}

	return
}

// GetPlayerPosition
// Metodo didicato allo visualizione del nome del pianeta
func (c *MenuController) GetPlayerPosition() (result string, err error) {
	// Recupero ultima posizione del player, dando per scontato che sia
	// la posizione del pianeta e quindi della mappa corrente che si vuole recuperare
	var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
	rGetPlayerCurrentPlanet, err = services.NnSDK.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
		PlayerID: c.Player.GetID(),
	})
	if err != nil {
		panic(err)
	}

	// Verifico se il player si trova su un pianeta sicuro
	c.SafePlanet = rGetPlayerCurrentPlanet.GetPlanet().GetSafe()
	c.TitanPlanet = rGetPlayerCurrentPlanet.GetPlanet().GetTitan()

	// Se è un pianeta sicuro modifico il messaggio
	if c.SafePlanet {
		return fmt.Sprintf("%s 🏟", rGetPlayerCurrentPlanet.GetPlanet().GetName()), err
	}

	return rGetPlayerCurrentPlanet.GetPlanet().GetName(), err
}

// GetPlayerLife
// Metodo dedicato alla rappresentazione dello stato vitale del player
func (c *MenuController) GetPlayerLife() (life string, err error) {
	// Calcolo stato vitale del player
	status := "♥️"
	if c.PlayerStats.GetDead() {
		status = "☠️"
	}

	life = fmt.Sprintf("%s️ %v/%v HP", status, c.PlayerStats.GetLifePoint(), 100)

	return
}

// GetPlayerTask
// Metodo dedicato alla rappresentazione dei task attivi del player
func (c *MenuController) GetPlayerTasks() (tasks string) {
	if len(c.ActiveStates) > 0 {
		tasks = helpers.Trans(c.Player.Language.Slug, "menu.tasks")

		for _, state := range c.ActiveStates {
			if !state.GetCompleted() {
				finishAt, err := ptypes.Timestamp(state.FinishAt)
				if err != nil {
					panic(err)
				}

				// Se sono da notificare formatto con la data
				if state.GetToNotify() && time.Since(finishAt).Minutes() < 0 {
					finishTime := math.Abs(math.RoundToEven(time.Since(finishAt).Minutes()))
					tasks += fmt.Sprintf("- %s %v\n",
						helpers.Trans(c.Player.Language.Slug, state.Controller),
						helpers.Trans(c.Player.Language.Slug, "menu.tasks.minutes_left", finishTime),
					)
				} else {
					if state.Controller == "route.tutorial" {
						tasks += fmt.Sprintf("- %s \n",
							helpers.Trans(c.Player.Language.Slug, state.Controller),
						)
					} else if state.Controller == "route.safeplanet.mission" {
						tasks += fmt.Sprintf("- %s \n",
							helpers.Trans(c.Player.Language.Slug, state.Controller),
						)
					} else {
						tasks += fmt.Sprintf("- %s %s\n",
							helpers.Trans(c.Player.Language.Slug, state.Controller),
							helpers.Trans(c.Player.Language.Slug, "menu.tasks.completed"),
						)
					}

				}
			}

		}
	}

	return
}

// GetRecap
func (c *MenuController) GetKeyboard() [][]tgbotapi.KeyboardButton {
	// Se il player sta finendo il tutorial mostro il menù con i task personalizzati
	// var inTutorial bool
	for _, state := range c.ActiveStates {
		if state.Controller == "route.tutorial" {
			return c.TutorialKeyboard()
		} else if state.Controller == "route.ship.travel" {
			return c.ExplorationKeyboard()
		} else if state.Controller == "route.exploration" {
			return c.MissionKeyboard()
		}
	}

	// Se il player non ha nessun stato attivo ma si trova in un pianeta sicuro
	// allora mostro la keyboard dedicata al pianeta sicuro
	if c.SafePlanet {
		return c.SafePlanetKeyboard()
	}

	// Verifico se il player si trova sul pianeta di un titano
	if c.TitanPlanet {
		return c.TitanPlanetKeyboard()
	}

	// Si trova su un pianeta normale
	return c.MainKeyboard()
}

// MainMenu
func (c *MenuController) MainKeyboard() (keyboard [][]tgbotapi.KeyboardButton) {
	return [][]tgbotapi.KeyboardButton{
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.exploration")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.hunting")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.planet")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.player")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship")),
		},
	}
}

// TutorialMenu
func (c *MenuController) TutorialKeyboard() (keyboardRows [][]tgbotapi.KeyboardButton) {
	// Per il tutorial costruisco keyboard solo per gli stati attivi
	for _, state := range c.ActiveStates {
		if !state.GetCompleted() {
			keyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, state.Controller)),
			)

			keyboardRows = append(keyboardRows, keyboardRow)
		}
	}

	return
}

// ExplorationKeyboard
func (c *MenuController) ExplorationKeyboard() [][]tgbotapi.KeyboardButton {
	return [][]tgbotapi.KeyboardButton{
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.inventory")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.crafting")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ability")),
		},
	}
}

// MissionKeyboard
func (c *MenuController) MissionKeyboard() [][]tgbotapi.KeyboardButton {
	return [][]tgbotapi.KeyboardButton{
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.exploration")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.inventory")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.crafting")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ability")),
		},
	}
}

// MainMenu
func (c *MenuController) SafePlanetKeyboard() [][]tgbotapi.KeyboardButton {
	var keyboardRow [][]tgbotapi.KeyboardButton
	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.safeplanet.coalition")),
	})

	// Recupero gli npc attivi in questo momento
	rGetAll, err := services.NnSDK.GetAllNPC(helpers.NewContext(1), &pb.GetAllNPCRequest{})
	if err != nil {
		panic(err)
	}

	for _, npc := range rGetAll.GetNPCs() {
		row := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("route.safeplanet.%s", npc.Slug)),
			),
		)
		keyboardRow = append(keyboardRow, row)
	}

	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship")),
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.player")),
	})

	return keyboardRow
}

// TitanPlanetKeboard
func (c *MenuController) TitanPlanetKeyboard() [][]tgbotapi.KeyboardButton {
	var keyboardRow [][]tgbotapi.KeyboardButton
	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.titanplanet.tackle")),
	})

	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship")),
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.player")),
	})

	return keyboardRow
}
