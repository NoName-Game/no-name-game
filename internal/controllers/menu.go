package controllers

import (
	"fmt"
	"math"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// MenuController
// Modulo dedicato alla gestine e visualizazione della keyboard di telegram
// in base a dove si trova il player verranno mostrati tasti e action differenti.
// ====================================
type MenuController struct {
	Controller
	SafePlanet  bool // Flag per verificare se il player si trova su un pianeta sicuro
	TitanPlanet bool // Flag per verificare se il player si trova su un pianeta titano
	TitanAlive  bool
	Payload     interface{}
}

// ====================================
// Handle
// ====================================
func (c *MenuController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error
	c.Player = player
	c.Update = update

	// Carico controller data
	if err = c.LoadControllerData(); err != nil {
		c.Logger.Panic(err)
	}

	// Recupero messaggio principale
	recap := c.GetRecap()

	msg := helpers.NewMessage(c.Player.ChatID, recap)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard:       c.GetKeyboard(),
		ResizeKeyboard: true,
	}

	// Send recap message
	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}

	// Se il player è morto non può fare altro che riposare o azioni che richiedono riposo
	if c.Data.PlayerStats.GetDead() {
		restsController := new(ShipRestsController)
		restsController.Handle(c.Player, c.Update)
	}
}

func (c *MenuController) Validator() bool {
	return false
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
func (c *MenuController) GetRecap() (message string) {
	var err error

	// Recupero posizione player
	var planet *pb.Planet
	planet = c.GetPlayerPosition()

	// Costruisco messaggio di racap in base a dove si trova il player
	if c.SafePlanet {
		message = helpers.Trans(c.Player.Language.Slug, "menu.safeplanet",
			planet.GetName(),
			c.GetPlayerTasks(),
		)
	} else if c.TitanPlanet {
		// Recupero titano pianeta corrente
		var rGetTitanByPlanetID *pb.GetTitanByPlanetIDResponse
		if rGetTitanByPlanetID, err = config.App.Server.Connection.GetTitanByPlanetID(helpers.NewContext(1), &pb.GetTitanByPlanetIDRequest{
			PlanetID: planet.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		message = helpers.Trans(c.Player.Language.Slug, "menu.titanplanet", planet.GetName(), rGetTitanByPlanetID.GetTitan().GetName())

		// Verifico se il titano è vivo o morto per arricchire il messaggio
		if rGetTitanByPlanetID.GetTitan().GetLifePoint() <= 0 {
			message += helpers.Trans(c.Player.Language.Slug, "menu.titanplanet.titan_dead")
		} else {
			c.TitanAlive = true // Flag usato per nascondere/ostrare pulsante keyboard
			message += helpers.Trans(c.Player.Language.Slug, "menu.titanplanet.titan_alive")
		}
	} else {
		// Recupero status vitale del player
		life := c.GetPlayerLife()

		// Recupero conquistatore attuale
		var currentConqueror string
		var rGetCurrentConquerorByPlanetID *pb.GetCurrentConquerorByPlanetIDResponse
		if rGetCurrentConquerorByPlanetID, err = config.App.Server.Connection.GetCurrentConquerorByPlanetID(helpers.NewContext(1), &pb.GetCurrentConquerorByPlanetIDRequest{
			PlanetID: planet.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		if rGetCurrentConquerorByPlanetID.GetPlayer().GetID() > 0 {
			currentConqueror = fmt.Sprintf("🚩 %s", rGetCurrentConquerorByPlanetID.GetPlayer().GetUsername())
		} else {
			currentConqueror = helpers.Trans(c.Player.Language.Slug, "conqueror.planet_free")
		}

		message = helpers.Trans(c.Player.Language.Slug, "menu",
			planet.GetName(),
			currentConqueror,
			life,
			c.GetPlayerTasks(),
		)
	}

	return
}

// GetPlayerPosition
// Metodo didicato allo visualizione del nome del pianeta
func (c *MenuController) GetPlayerPosition() (result *pb.Planet) {
	var err error

	// Recupero ultima posizione del player, dando per scontato che sia
	// la posizione del pianeta e quindi della mappa corrente che si vuole recuperare
	var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
	if rGetPlayerCurrentPlanet, err = config.App.Server.Connection.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Verifico se il player si trova su un pianeta valido
	if rGetPlayerCurrentPlanet.GetPlanet() == nil {
		c.Logger.Panic(err)
	}

	// Verifico se il player si trova su un pianeta sicuro
	c.SafePlanet = rGetPlayerCurrentPlanet.GetPlanet().GetSafe()

	// Verifico se il pianeta corrente è occupato da un titano
	var rGetTitanByPlanetID *pb.GetTitanByPlanetIDResponse
	rGetTitanByPlanetID, _ = config.App.Server.Connection.GetTitanByPlanetID(helpers.NewContext(1), &pb.GetTitanByPlanetIDRequest{
		PlanetID: rGetPlayerCurrentPlanet.GetPlanet().GetID(),
	})

	c.TitanPlanet = false
	if rGetTitanByPlanetID.GetTitan().GetID() > 0 {
		c.TitanPlanet = true
	}

	return rGetPlayerCurrentPlanet.GetPlanet()
}

// GetPlayerLife
// Metodo dedicato alla rappresentazione dello stato vitale del player
func (c *MenuController) GetPlayerLife() (life string) {
	// Calcolo stato vitale del player
	status := "♥️"
	if c.Data.PlayerStats.GetDead() {
		status = "☠️"
	}

	life = fmt.Sprintf("%s️ %v/%v HP", status, c.Data.PlayerStats.GetLifePoint(), 100)

	return
}

// GetPlayerTask
// Metodo dedicato alla rappresentazione dei task attivi del player
func (c *MenuController) GetPlayerTasks() (tasks string) {
	var err error
	if len(c.Data.PlayerActiveStates) > 0 {
		tasks = helpers.Trans(c.Player.Language.Slug, "menu.tasks")

		for _, state := range c.Data.PlayerActiveStates {
			var finishAt time.Time
			if finishAt, err = helpers.GetEndTime(state.FinishAt, c.Player); err != nil {
				c.Logger.Panic(err)
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

	return
}

// GetRecap
func (c *MenuController) GetKeyboard() [][]tgbotapi.KeyboardButton {
	// Se il player sta finendo il tutorial mostro il menù con i task personalizzati
	// var inTutorial bool
	for _, state := range c.Data.PlayerActiveStates {
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
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.conqueror")),
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
	for _, state := range c.Data.PlayerActiveStates {
		var keyboardRow []tgbotapi.KeyboardButton
		if state.Controller == "route.tutorial" {
			keyboardRow = tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.tutorial.continue")),
			)
		} else {
			keyboardRow = tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, state.Controller)),
			)
		}

		keyboardRows = append(keyboardRows, keyboardRow)
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
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.laboratory")),
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
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.laboratory")),
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
	rGetAll, err := config.App.Server.Connection.GetAllNPC(helpers.NewContext(1), &pb.GetAllNPCRequest{})
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

	// Se il titano è vivo il player può affrontarlo
	if c.TitanAlive {
		keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.titanplanet.tackle")),
		})
	}

	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.planet")),
	})

	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship")),
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.player")),
	})

	return keyboardRow
}
