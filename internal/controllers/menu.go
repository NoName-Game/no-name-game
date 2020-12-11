package controllers

import (
	"encoding/json"
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
	Titan        *pb.Titan
	TravelPlanet *pb.Planet
	Payload      interface{}
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

	// Recupero posizione player
	var currentPosition *pb.Planet
	if currentPosition, err = helpers.GetPlayerPosition(c.Player.ID); err != nil {
		c.Logger.Panic(err)
	}

	// Recupero messaggio principale
	recap := c.GetRecap(currentPosition)

	msg := helpers.NewMessage(c.Player.ChatID, recap)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard:       c.GetKeyboard(currentPosition),
		ResizeKeyboard: true,
	}

	// Send recap message
	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}

	// Se il player è morto non può fare altro che riposare o azioni che richiedono riposo
	if c.Player.GetDead() {
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

// GetRecap - Recap principale
func (c *MenuController) GetRecap(currentPosition *pb.Planet) (message string) {
	// Appendo board system
	message = helpers.Trans(c.Player.Language.Slug, "menu.borad_system")

	// Menu se il player si trova su un pianeta sicuro
	if c.CheckInSafePlanet(currentPosition) {
		message += helpers.Trans(c.Player.Language.Slug, "menu.safeplanet",
			currentPosition.GetName(),
		)

		// Menu se il player si trova su il pianeta di un titano
	} else if c.CheckInTitanPlanet(currentPosition) {
		message += helpers.Trans(c.Player.Language.Slug, "menu.titanplanet",
			currentPosition.GetName(),
			c.Titan.GetName(),
		)

		// Verifico se il titano è vivo o morto per arricchire il messaggio
		if c.Titan.GetLifePoint() <= 0 {
			message += helpers.Trans(c.Player.Language.Slug, "menu.titanplanet.titan_dead")
		} else {
			message += helpers.Trans(c.Player.Language.Slug, "menu.titanplanet.titan_alive")
		}

		// Menu se il player si trova in viaggio
	} else if c.CheckInTravel() {
		message += helpers.Trans(c.Player.Language.Slug, "menu.travel",
			currentPosition.GetName(),
			c.TravelPlanet.GetName(),
		)

		// Menu normale
	} else {
		message += helpers.Trans(c.Player.Language.Slug, "menu.general",
			currentPosition.GetName(),
			c.GetPlayerLife(),
		)
	}

	// Appendo Task lista task
	message += c.GetPlayerTasks()

	return
}

// CheckInSafePlanet
// Verifico se il player si trova su un pianeta sicuro
func (c *MenuController) CheckInSafePlanet(position *pb.Planet) bool {
	return position.GetSafe()
}

// CheckInTitanPlanet
// Verifico se il player si trova su un pianeta sicuro
func (c *MenuController) CheckInTitanPlanet(position *pb.Planet) bool {
	// Verifico se il pianeta corrente è occupato da un titano
	var rGetTitanByPlanetID *pb.GetTitanByPlanetIDResponse
	rGetTitanByPlanetID, _ = config.App.Server.Connection.GetTitanByPlanetID(helpers.NewContext(1), &pb.GetTitanByPlanetIDRequest{
		PlanetID: position.GetID(),
	})

	if rGetTitanByPlanetID.GetTitan().GetID() > 0 {
		c.Titan = rGetTitanByPlanetID.GetTitan()
		return true
	}

	return false
}

// CheckInTravel
// Verifico se il player sta effettuando un viaggio
func (c *MenuController) CheckInTravel() bool {
	type travelDataStruct struct {
		PlanetID uint32
	}

	// Verifico se il player si trova in viaggio
	var travelData travelDataStruct
	for _, activity := range c.Data.PlayerActiveStates {
		if activity.Controller == "route.ship.travel" {
			_ = json.Unmarshal([]byte(activity.Payload), &travelData)

			// Recupero pianeta che si vuole raggiungere
			var rGetPlanetByID *pb.GetPlanetByIDResponse
			rGetPlanetByID, _ = config.App.Server.Connection.GetPlanetByID(helpers.NewContext(1), &pb.GetPlanetByIDRequest{
				PlanetID: travelData.PlanetID,
			})

			c.TravelPlanet = rGetPlanetByID.GetPlanet()
			return true
		}
	}

	return false
}

// GetPlayerLife
// Metodo dedicato alla rappresentazione dello stato vitale del player
func (c *MenuController) GetPlayerLife() (life string) {
	// Calcolo stato vitale del player
	status := "♥️"
	if c.Player.GetDead() {
		status = "☠️"
	}

	life = fmt.Sprintf("%s️ %v/%v HP", status, c.Player.GetLifePoint(), c.Player.GetLevel().GetPlayerMaxLife())

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
				} else if state.Controller == "route.safeplanet.coalition.mission" {
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
func (c *MenuController) GetKeyboard(currentPosition *pb.Planet) [][]tgbotapi.KeyboardButton {
	// Se il player sta finendo il tutorial mostro il menù con i task personalizzati
	// var inTutorial bool
	for _, state := range c.Data.PlayerActiveStates {
		if state.Controller == "route.tutorial" {
			return c.TutorialKeyboard()
		} else if state.Controller == "route.ship.travel" {
			return c.TravelKeyboard()
		} else if state.Controller == "route.exploration" {
			return c.ExplorationKeyboard()
		}
	}

	// Se il player non ha nessun stato attivo ma si trova in un pianeta sicuro
	// allora mostro la keyboard dedicata al pianeta sicuro
	if c.CheckInSafePlanet(currentPosition) {
		return c.SafePlanetKeyboard()
	}

	// Verifico se il player si trova sul pianeta di un titano
	if c.CheckInTitanPlanet(currentPosition) {
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

// TravelKeyboard
func (c *MenuController) TravelKeyboard() [][]tgbotapi.KeyboardButton {
	return [][]tgbotapi.KeyboardButton{
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.player")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship")),
		},
	}
}

// ExplorationKeyboard
func (c *MenuController) ExplorationKeyboard() [][]tgbotapi.KeyboardButton {
	return [][]tgbotapi.KeyboardButton{
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.exploration")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.planet")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.conqueror")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.player")),
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
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.safeplanet.accademy")),
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.safeplanet.hangar")),
	})

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
	if c.Titan.GetLifePoint() > 0 {
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
