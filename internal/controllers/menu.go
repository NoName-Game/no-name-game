package controllers

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
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
	msg.ParseMode = tgbotapi.ModeHTML
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
	message = helpers.Trans(c.Player.Language.Slug, "menu.borad_system", os.Getenv("VERSION"))

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

	// Recupero annuncio
	message += c.GetAnnuncement()

	return
}

func (c *MenuController) GetAnnuncement() string {
	randAnnuncement := helpers.Trans(c.Player.Language.Slug,
		fmt.Sprintf("menu.annuncement_%v", helpers.RandomInt(1, 3)))

	return fmt.Sprintf("🔖 %s", randAnnuncement)
}

// CheckInSafePlanet
// Verifico se il player si trova su un pianeta sicuro
func (c *MenuController) CheckInSafePlanet(position *pb.Planet) bool {
	return c.Controller.CheckInSafePlanet(position)
}

// CheckInTitanPlanet
// Verifico se il player si trova su un pianeta sicuro
func (c *MenuController) CheckInTitanPlanet(position *pb.Planet) bool {

	inPlanet, titan := c.Controller.CheckInTitanPlanet(position)
	if titan != nil {
		c.Titan = titan
	}

	return inPlanet
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
		if activity.Controller == "route.ship.travel.finding" {
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

// FormatPlayerTasks
// Metodo dedicato a fomrattare e impaginare meglio i task in corso e completati
func (c *MenuController) FormatPlayerTasks(activity *pb.PlayerActivity) (tasks string) {
	var err error

	// Format Custom
	switch activity.GetController() {
	// Verifico se si tritta di una missione
	case "route.safeplanet.coalition.mission":
		var rCheckMission *pb.CheckMissionResponse
		if rCheckMission, err = config.App.Server.Connection.CheckMission(helpers.NewContext(1), &pb.CheckMissionRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		if rCheckMission.GetCompleted() {
			tasks = helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("%s.activity.done", activity.GetController()))
		} else {
			// Recupero dettagli missione
			type MissionPaylodStruct struct {
				MissionID uint32
			}

			var missionController MissionPaylodStruct
			helpers.UnmarshalPayload(activity.GetPayload(), &missionController)

			// Recupero dettagli missione
			missionDetails := c.MissionRecap(missionController.MissionID)

			tasks = helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("%s.activity.progress", activity.GetController()), missionController.MissionID, missionDetails)
		}

		return
	case "route.tutorial":
		tasks = helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("%s.activity.progress", activity.GetController()))
		return
	}

	var finishAt time.Time
	if finishAt, err = helpers.GetEndTime(activity.FinishAt, c.Player); err != nil {
		c.Logger.Panic(err)
	}

	// Se sono delle attività non ancora concluse
	if time.Until(finishAt).Minutes() >= 0 {
		finishTime := math.Abs(math.RoundToEven(time.Since(finishAt).Minutes()))
		tasks = helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("%s.activity.progress", activity.GetController()), finishTime)
	} else if activity.GetFinished() || time.Until(finishAt).Minutes() < 0 {
		tasks = helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("%s.activity.done", activity.GetController()))
	}

	return
}

func (c *MenuController) MissionRecap(missionID uint32) (missionRecap string) {
	var err error

	// Recupero dettagli missione
	var rGetMission *pb.GetMissionResponse
	if rGetMission, err = config.App.Server.Connection.GetMission(helpers.NewContext(1), &pb.GetMissionRequest{
		MissionID: missionID,
	}); err != nil {
		c.Logger.Panic(err)
	}

	mission := rGetMission.GetMission()

	switch mission.GetMissionCategory().GetSlug() {
	// Trovare le risorse
	case "resources_finding":
		var missionPayload *pb.MissionResourcesFinding
		helpers.UnmarshalPayload(mission.GetPayload(), &missionPayload)

		// Recupero enitità risorsa da cercare
		var rGetResourceByID *pb.GetResourceByIDResponse
		if rGetResourceByID, err = config.App.Server.Connection.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
			ID: missionPayload.GetResourceID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		missionRecap += helpers.Trans(c.Player.Language.Slug,
			"safeplanet.mission.type.resources_finding.description.small",
			missionPayload.GetResourceQty(),
			helpers.GetResourceCategoryIcons(rGetResourceByID.GetResource().GetResourceCategoryID()),
			rGetResourceByID.GetResource().GetName(),
			rGetResourceByID.GetResource().GetRarity().GetSlug(),
			helpers.GetResourceBaseIcons(rGetResourceByID.GetResource().GetBase()),
		)

	// Trova le risorse
	case "planet_finding":
		var missionPayload *pb.MissionPlanetFinding
		helpers.UnmarshalPayload(mission.GetPayload(), &missionPayload)

		// Recupero pianeta da trovare
		var rGetPlanetByID *pb.GetPlanetByIDResponse
		if rGetPlanetByID, err = config.App.Server.Connection.GetPlanetByID(helpers.NewContext(1), &pb.GetPlanetByIDRequest{
			PlanetID: missionPayload.GetPlanetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		missionRecap += helpers.Trans(c.Player.Language.Slug,
			"safeplanet.mission.type.planet_finding.description.small",
			rGetPlanetByID.GetPlanet().GetName(),
			rGetPlanetByID.GetPlanet().GetPlanetSystem().GetName(),
			rGetPlanetByID.GetPlanet().GetPlanetSystem().GetID(),
		)

	// Uccidi il nemico
	case "kill_mob":
		var missionPayload *pb.MissionKillMob
		helpers.UnmarshalPayload(mission.GetPayload(), &missionPayload)

		// Recupero enemy da Uccidere
		var rGetEnemyByID *pb.GetEnemyByIDResponse
		if rGetEnemyByID, err = config.App.Server.Connection.GetEnemyByID(helpers.NewContext(1), &pb.GetEnemyByIDRequest{
			EnemyID: missionPayload.GetEnemyID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero pianeta di dove si trova il mob
		var rGetPlanetByMapID *pb.GetPlanetByMapIDResponse
		if rGetPlanetByMapID, err = config.App.Server.Connection.GetPlanetByMapID(helpers.NewContext(1), &pb.GetPlanetByMapIDRequest{
			MapID: rGetEnemyByID.GetEnemy().GetPlanetMapID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		missionRecap += helpers.Trans(c.Player.Language.Slug,
			"safeplanet.mission.type.kill_mob.description.small",
			rGetEnemyByID.GetEnemy().GetName(),
			rGetPlanetByMapID.GetPlanet().GetName(),
			rGetPlanetByMapID.GetPlanet().GetPlanetSystem().GetName(),
			rGetPlanetByMapID.GetPlanet().GetPlanetSystem().GetID(),
		)
	}

	return
}

// GetPlayerTask
// Metodo dedicato alla rappresentazione dei task attivi del player
func (c *MenuController) GetPlayerTasks() (tasks string) {
	if len(c.Data.PlayerActiveStates) > 0 {
		tasks = helpers.Trans(c.Player.Language.Slug, "menu.tasks")
		for _, state := range c.Data.PlayerActiveStates {
			tasks += fmt.Sprintf("- %s \n", c.FormatPlayerTasks(state))
		}
	}

	tasks += "\n"
	return
}

// GetRecap
func (c *MenuController) GetKeyboard(currentPosition *pb.Planet) [][]tgbotapi.KeyboardButton {
	// Se il player sta finendo il tutorial mostro il menù con i task personalizzati
	// var inTutorial bool
	for _, state := range c.Data.PlayerActiveStates {
		if state.Controller == "route.tutorial" {
			return c.TutorialKeyboard()
		} else if state.Controller == "route.ship.travel.finding" {
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
	return c.MainKeyboard(currentPosition)
}

// MainMenu
func (c *MenuController) MainKeyboard(position *pb.Planet) [][]tgbotapi.KeyboardButton {
	var keyboardRow [][]tgbotapi.KeyboardButton

	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.exploration")),
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.hunting")),
	})

	// Verifico se il pianeta corrente è quello del mercante oscuro
	var rGetDarkMerchant *pb.GetDarkMerchantResponse
	rGetDarkMerchant, _ = config.App.Server.Connection.GetDarkMerchant(helpers.NewContext(1), &pb.GetDarkMerchantRequest{})
	if rGetDarkMerchant != nil {
		if rGetDarkMerchant.PlanetID == position.ID {
			keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.darkmerchant")),
			})
		}
	}

	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.planet")),
	})

	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.player")),
	})

	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship")),
	})

	return keyboardRow
}

// TutorialMenu
func (c *MenuController) TutorialKeyboard() (keyboardRows [][]tgbotapi.KeyboardButton) {
	// Per il tutorial costruisco keyboard solo per gli stati attivi
	for _, state := range c.Data.PlayerActiveStates {
		var keyboardRow []tgbotapi.KeyboardButton
		if state.GetController() == "route.tutorial" {
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

	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.safeplanet.market")),
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.safeplanet.bank")),
	})

	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.safeplanet.crafter")),
	})

	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.safeplanet.relax")),
	})

	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.safeplanet.hangar")),
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship")),
	})

	keyboardRow = append(keyboardRow, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.safeplanet.accademy")),
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
