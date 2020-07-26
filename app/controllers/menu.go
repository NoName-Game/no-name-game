package controllers

import (
	"fmt"
	"math"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
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
	SafePlanet bool // Flag per verificare se il player si trova su un pianeta sicuro
}

// ====================================
// Handle
// ====================================
func (c *MenuController) Handle(player nnsdk.Player, update tgbotapi.Update, proxy bool) {
	var err error
	var playerProvider providers.PlayerProvider

	// Il menù del player refresha sempre lo status del player
	player, err = playerProvider.FindPlayerByUsername(player.Username)
	if err != nil {
		panic(err)
	}

	// Init funzionalità
	c.Controller = "route.menu"
	c.Player = player

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
	if *c.Player.Stats.Dead {
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

	return
}

// GetPlayerPosition
// Metodo didicato allo visualizione del nome del pianeta
func (c *MenuController) GetPlayerPosition() (result string, err error) {
	var playerProvider providers.PlayerProvider
	var planetProvider providers.PlanetProvider

	// Recupero ultima posizione del player, dando per scontato che sia
	// la posizione del pianeta e quindi della mappa corrente che si vuole recuperare
	var lastPosition nnsdk.PlayerPosition
	lastPosition, err = playerProvider.GetPlayerLastPosition(c.Player)
	if err != nil {
		return result, err
	}

	// Dalla ultima posizione recupero il pianeta corrente
	var planet nnsdk.Planet
	planet, err = planetProvider.GetPlanetByCoordinate(lastPosition.X, lastPosition.Y, lastPosition.Z)
	if err != nil {
		return result, err
	}

	// Verifico se il player si trova su un pianeta sicuro
	c.SafePlanet = planet.Safe

	// Se è un pianeta sicuro modifico il messaggio
	if c.SafePlanet {
		return fmt.Sprintf("%s 🏟", planet.Name), err
	}

	return planet.Name, err
}

// GetPlayerLife
// Metodo dedicato alla rappresentazione dello stato vitale del player
func (c *MenuController) GetPlayerLife() (life string, err error) {
	// Calcolo stato vitale del player
	status := "♥️"
	if *c.Player.Stats.Dead {
		status = "☠️"
	}

	life = fmt.Sprintf("%s️ %v/%v HP", status, *c.Player.Stats.LifePoint, 100)

	return
}

// GetPlayerTask
// Metodo dedicato alla rappresentazione dei task attivi del player
func (c *MenuController) GetPlayerTasks() (tasks string) {
	if len(c.Player.States) > 0 {
		tasks = helpers.Trans(c.Player.Language.Slug, "menu.tasks")

		for _, state := range c.Player.States {
			if !*state.Completed {
				// Se sono da notificare formatto con la data
				if *state.ToNotify && time.Since(state.FinishAt).Minutes() < 0 {
					finishTime := math.Abs(math.RoundToEven(time.Since(state.FinishAt).Minutes()))
					tasks += fmt.Sprintf("- %s %v\n",
						helpers.Trans(c.Player.Language.Slug, state.Controller),
						helpers.Trans(c.Player.Language.Slug, "menu.tasks.minutes_left", finishTime),
					)
				} else {
					if state.Controller == "route.tutorial" {
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
	for _, state := range c.Player.States {
		if state.Controller == "route.tutorial" {
			return c.TutorialKeyboard()
		} else if state.Controller == "route.ship.exploration" {
			return c.ExplorationKeyboard()
		} else if state.Controller == "route.mission" {
			return c.MissionKeyboard()
		}
	}

	// Se il player non ha nessun stato attivo ma si trova in un pianeta sicuro
	// allora mostro la keyboard dedicata al pianeta sicuro
	if c.SafePlanet {
		return c.SafePlanetKeyboard()
	}

	// Si trova su un pianeta normale
	return c.MainKeyboard()
}

// MainMenu
func (c *MenuController) MainKeyboard() (keyboard [][]tgbotapi.KeyboardButton) {

	keyboard = [][]tgbotapi.KeyboardButton{
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.mission")),
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

	if c.SafePlanet {
		keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.menu.npc")),
		))
	}

	return
}

// TutorialMenu
func (c *MenuController) TutorialKeyboard() (keyboardRows [][]tgbotapi.KeyboardButton) {
	// Per il tutorial costruisco keyboard solo per gli stati attivi
	for _, state := range c.Player.States {
		if !*state.Completed {
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
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.mission")),
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
	var npcProvider providers.NpcProvider

	// Recupero gli npc attivi in questo momento
	npcs, err := npcProvider.GetAll()
	if err != nil {
		panic(err)
	}

	var keyboardRow [][]tgbotapi.KeyboardButton
	for _, npc := range npcs {
		row := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("route.safeplanet.%s", npc.Slug)),
			),
		)
		keyboardRow = append(keyboardRow, row)
	}

	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship")),
	})

	return keyboardRow
}
