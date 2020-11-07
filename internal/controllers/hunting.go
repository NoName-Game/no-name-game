package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// HuntingController
// ====================================
// In questo controller il player avrà la possibilità di esplorare
// la mappa del pianeta che sta visitando, e di conseguenza affrontare mob,
// recupeare tesori e cascare in delle trappole
// ====================================

type HuntingController struct {
	Controller
	Payload struct {
		CallbackMessageID int
		MapID             uint32
		PlayerPositionX   int32
		PlayerPositionY   int32
		BodySelection     int32
	}
	Enemy *pb.Enemy
}

// ====================================
// HuntingController - Settings
// ====================================
var (
	// Parti di corpo disponibili per l'attacco
	bodyParts = [4]string{"helmet", "chest", "glove", "boots"}

	// Hunting Move Actions
	moveDown = helpers.InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "down",
	}

	moveUp = helpers.InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "up",
	}

	moveLeft = helpers.InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "left",
	}

	moveRight = helpers.InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "right",
	}

	moveAction = helpers.InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "action",
	}

	moveNoAction = helpers.InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "no_action",
	}

	// Hunting Fight Actions
	fightStart = helpers.InlineDataStruct{
		C:  "hunting",
		AT: "fight",
		A:  "start_fight",
	}

	fightUp = helpers.InlineDataStruct{
		C:  "hunting",
		AT: "fight",
		A:  "up",
	}

	fightDown = helpers.InlineDataStruct{
		C:  "hunting",
		AT: "fight",
		A:  "down",
	}

	fightReturnMap = helpers.InlineDataStruct{
		C:  "hunting",
		AT: "fight",
		A:  "return_map",
	}

	fightPlayerDie = helpers.InlineDataStruct{
		C:  "hunting",
		AT: "fight",
		A:  "player_die",
	}

	fightHit = helpers.InlineDataStruct{
		C:  "hunting",
		AT: "fight",
		A:  "hit",
	}

	fightNoAction = helpers.InlineDataStruct{
		C:  "hunting",
		AT: "fight",
		A:  "no_action",
	}

	// Keyboard inline di esplorazione
	mapKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", moveUp.GetDataString())),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", moveLeft.GetDataString()),
			tgbotapi.NewInlineKeyboardButtonData("➡️", moveRight.GetDataString()),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", moveDown.GetDataString())),
	)

	tresureKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", moveUp.GetDataString())),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", moveLeft.GetDataString()),
			tgbotapi.NewInlineKeyboardButtonData("❓️", moveAction.GetDataString()),
			tgbotapi.NewInlineKeyboardButtonData("➡️", moveRight.GetDataString()),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", moveDown.GetDataString())),
	)

	enemyKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", moveUp.GetDataString())),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", moveLeft.GetDataString()),
			tgbotapi.NewInlineKeyboardButtonData("⚔️", fightStart.GetDataString()),
			tgbotapi.NewInlineKeyboardButtonData("➡️", moveRight.GetDataString()),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", moveDown.GetDataString())),
	)

	fightKeyboard = [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔼", fightUp.GetDataString())),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏃‍♂️💨", fightReturnMap.GetDataString()),
			tgbotapi.NewInlineKeyboardButtonData("⚔️", fightHit.GetDataString()),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔽", fightDown.GetDataString())),
	}
)

// ====================================
// Handle
// ====================================
func (c *HuntingController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.hunting",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBlocked: []string{"exploration"},
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
			},
		},
	}) {
		return
	}

	// Ok! Run!
	c.Stage()

	// Verifico completamento aggiuntivo per cancellare il messaggio
	if c.CurrentState.Completed {
		// Cancello messaggio contentente la mappa
		if err := helpers.DeleteMessage(c.Player.ChatID, c.Payload.CallbackMessageID); err != nil {
			c.Logger.Panic(err)
		}
	}

	// Completo progressione
	c.Completing(&c.Payload)
}

// ====================================
// Validator
// ====================================
func (c *HuntingController) Validator() (hasErrors bool) {
	return false
}

// ====================================
// Stage Map -> Drop -> Finish
// ====================================
func (c *HuntingController) Stage() {
	switch c.CurrentState.Stage {
	// In questo stage faccio entrare il player nella mappa
	case 0:
		// Verifico se il player vuole uscire dalla caccia
		if c.Update.Message != nil {
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "hunting.leave") {
				c.CurrentState.Completed = true
				return
			}
		}

		// Avvio ufficialmente la caccia!
		c.Hunting()
	}

	return
}

// Hunting - in questo passo mi restituisco la mappa al player
func (c *HuntingController) Hunting() {
	var err error

	// Se nel payload NON è presente un ID della mappa lo
	// recupero dalla posizione del player e invio al player il messaggio
	// principale contenente la mappa e il tastierino
	if c.Update.CallbackQuery == nil && c.Update.Message != nil {
		// Se è qualsiasi messaggio diverso da hunting non lo calcolo
		// in quanto adnrebbe a generare più volte il messaggio con la stessa mappa
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.hunting") &&
			c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.tutorial") &&
			c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.tutorial.continue") {
			return
		}

		// Questo messaggio è necessario per immettere il tasto di abbandona caccia
		initHunting := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "hunting.init"))
		initHunting.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "hunting.leave"),
				),
			),
		)
		if _, err = helpers.SendMessage(initHunting); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero ultima posizione del player, dando per scontato che sia
		// la posizione del pianeta e quindi della mappa corrente che si vuole recuperare
		// Dalla ultima posizione recupero il pianeta corrente
		var playerPosition *pb.Planet
		if playerPosition, err = helpers.GetPlayerPosition(c.Player.ID); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero dettagli della mappa e per non appesantire le chiamate
		// al DB registro il tutto sula cache
		var rGetPlanetMapByID *pb.GetPlanetMapByIDResponse
		if rGetPlanetMapByID, err = config.App.Server.Connection.GetPlanetMapByID(helpers.NewContext(1), &pb.GetPlanetMapByIDRequest{
			PlanetMapID: playerPosition.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		var planetMap = rGetPlanetMapByID.GetPlanetMap()

		// Registro mappa e posizione iniziale del player
		_ = helpers.SetMapInCache(planetMap)
		c.Payload.PlayerPositionX = planetMap.GetStartPositionX()
		c.Payload.PlayerPositionY = planetMap.GetStartPositionY()

		// Trasformo la mappa in qualcosa di più leggibile su telegram
		var decodedMap string
		if decodedMap, err = helpers.DecodeMapToDisplay(planetMap, planetMap.GetStartPositionX(), planetMap.GetStartPositionY()); err != nil {
			c.Logger.Panic(err)
		}

		// Invio quindi il mesaggio contenente mappa e azioni disponibili
		msg := helpers.NewMessage(c.Player.ChatID, decodedMap)
		msg.ReplyMarkup = mapKeyboard
		msg.ParseMode = "HTML"

		var huntingMessage tgbotapi.Message
		if huntingMessage, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno lo stato e ritorno
		c.Payload.CallbackMessageID = huntingMessage.MessageID
		c.Payload.MapID = planetMap.ID

		return
	}

	// Se il messaggio è di tipo callback ed esiste una mappa associato al payload
	// potrebbe essere un messaggio lanciato da tasiterino, quindi acconsento allo spostamento
	if c.Update.CallbackQuery != nil && c.Update.Message == nil {
		var planetMap *pb.PlanetMap
		if planetMap, err = helpers.GetMapInCache(c.Payload.MapID); err != nil {
			c.Logger.Panic(err)
		}

		// Controllo tipo di callback data - move / fight
		var inlineData helpers.InlineDataStruct
		inlineData = inlineData.GetDataValue(c.Update.CallbackQuery.Data)

		// Verifica tipo di movimento e mi assicuro che non sia in combattimento
		if inlineData.AT == "move" {
			err = c.movements(inlineData, planetMap)
		} else if inlineData.AT == "fight" {
			err = c.fight(inlineData, planetMap)
		}

		if err != nil {
			c.Logger.Info(err)
		}

		// Rimuove rotella di caricamento dal bottone
		err = helpers.AnswerCallbackQuery(
			helpers.NewAnswer(c.Update.CallbackQuery.ID, "", false),
		)

		return
	}

	return
}

// ====================================
// Movements
// ====================================
func (c *HuntingController) movements(inlineData helpers.InlineDataStruct, planetMap *pb.PlanetMap) (err error) {
	// Refresh della mappa
	var cellGrid [][]bool
	if err = json.Unmarshal([]byte(planetMap.CellGrid), &cellGrid); err != nil {
		c.Logger.Panic(err)
	}

	// Eseguo azione
	switch inlineData.A {
	case "up":
		if c.Payload.PlayerPositionX > 0 && !cellGrid[c.Payload.PlayerPositionX-1][c.Payload.PlayerPositionY] {
			c.Payload.PlayerPositionX--
			break
		}

		return
	case "down":
		// Se non è un muro posso proseguire
		if c.Payload.PlayerPositionX < int32(len(cellGrid)-1) && !cellGrid[c.Payload.PlayerPositionX+1][c.Payload.PlayerPositionY] {
			c.Payload.PlayerPositionX++
			break
		}

		return
	case "left":
		if c.Payload.PlayerPositionY > 0 && !cellGrid[c.Payload.PlayerPositionX][c.Payload.PlayerPositionY-1] {
			c.Payload.PlayerPositionY--
			break
		}

		return
	case "right":
		if c.Payload.PlayerPositionY < int32(len(cellGrid)-1) && !cellGrid[c.Payload.PlayerPositionX][c.Payload.PlayerPositionY+1] {
			c.Payload.PlayerPositionY++
			break
		}

		return
	case "action":
		c.action(planetMap)
		return
	case "no_action":
		// No action
	default:
		return errors.New("action not recognized")
	}

	// Trasformo la mappa in qualcosa di più leggibile su telegram
	var decodedMap string
	if decodedMap, err = helpers.DecodeMapToDisplay(planetMap, c.Payload.PlayerPositionX, c.Payload.PlayerPositionY); err != nil {
		c.Logger.Panic(err)
	}

	// Se l'azione è valida e completa aggiorno risultato
	msg := helpers.NewEditMessage(c.Player.ChatID, c.Update.CallbackQuery.Message.MessageID, decodedMap)
	msg.ReplyMarkup = &mapKeyboard

	// Se un player si trova sulla stessa posizione un mob o di un tesoro effettuo il controllo
	if _, nearMob := helpers.CheckForMob(planetMap, c.Payload.PlayerPositionX, c.Payload.PlayerPositionY); nearMob {
		msg.ReplyMarkup = &enemyKeyboard
	}

	if _, nearTresure := helpers.CheckForTresure(planetMap, c.Payload.PlayerPositionX, c.Payload.PlayerPositionY); nearTresure {
		msg.ReplyMarkup = &tresureKeyboard
	}

	msg.ParseMode = "HTML"
	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Info(err)
	}

	return
}

// ====================================
// Aciton
// ====================================
func (c *HuntingController) action(planetMap *pb.PlanetMap) {
	var err error

	// Verifico se si trova sopra un tesoro o una trappola
	if tresure, nearTresure := helpers.CheckForTresure(planetMap, c.Payload.PlayerPositionX, c.Payload.PlayerPositionY); nearTresure {
		// Chiamo WS e recupero tesoro
		var rDropTresure *pb.DropTresureResponse
		if rDropTresure, err = config.App.Server.Connection.DropTresure(helpers.NewContext(1), &pb.DropTresureRequest{
			TresureID: tresure.ID,
			PlayerID:  c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Verifico cosa è tornato e rispondo
		var tresureMessage tgbotapi.EditMessageTextConfig
		if rDropTresure.GetResource().GetID() > 0 {
			tresureMessage = helpers.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "tresure.found.resource", rDropTresure.GetResource().GetName()),
			)
		} else if rDropTresure.GetItem().GetID() > 0 {
			itemFound := helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("items.%s", rDropTresure.GetItem().GetSlug()))
			tresureMessage = helpers.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "tresure.found.item", itemFound),
			)
		} else if rDropTresure.GetTransaction().GetID() > 0 {
			tresureMessage = helpers.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "tresure.found.transaction", rDropTresure.GetTransaction().GetValue()),
			)
		} else if rDropTresure.GetTrap().GetID() > 0 {
			// Se è una trappola e il player è morto
			if rDropTresure.GetTrap().GetPlayerDie() {
				// Aggiorno messaggio notificando al player che è morto
				tresureMessage = helpers.NewEditMessage(
					c.Player.ChatID,
					c.Update.CallbackQuery.Message.MessageID,
					helpers.Trans(c.Player.Language.Slug, "combat.player_killed"),
				)

				var ok = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(
							helpers.Trans(c.Player.Language.Slug, "continue"), fightPlayerDie.GetDataString(),
						),
					),
				)

				tresureMessage.ReplyMarkup = &ok

				// Invio messaggio
				if _, err = helpers.SendMessage(tresureMessage); err != nil {
					c.Logger.Panic(err)
				}

				return
			}

			// Player sopravvive...
			tresureMessage = helpers.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "tresure.found.trap", rDropTresure.GetTrap().GetDamage()),
			)
		} else {
			// Non hai trovato nulla
			tresureMessage = helpers.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "tresure.found.nothing"),
			)
		}

		// Refresh della mappa per rimuovere il tosoro dalla memoria
		c.RefreshMap(planetMap.ID)

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Ok!", moveNoAction.GetDataString()),
			),
		)

		tresureMessage.ReplyMarkup = &keyboard
		tresureMessage.ParseMode = "markdown"
		if _, err = helpers.SendMessage(tresureMessage); err != nil {
			c.Logger.Panic(err)
		}

		return
	}
}

// ====================================
// Fight
// ====================================
func (c *HuntingController) fight(inlineData helpers.InlineDataStruct, planetMap *pb.PlanetMap) (err error) {
	// Verifico immediatamente se il player vuole tornare alla mappa o è morto
	switch inlineData.A {
	case "return_map":
		c.ReturnToMap(planetMap)
		return
	case "player_die":
		c.CurrentState.Completed = true
		return
	}

	// Recupero dettagli aggiornati enemy
	var enemy *pb.Enemy
	enemy, _ = helpers.CheckForMob(planetMap, c.Payload.PlayerPositionX, c.Payload.PlayerPositionY)
	if &enemy != nil {
		var rGetEnemyByID *pb.GetEnemyByIDResponse
		if rGetEnemyByID, err = config.App.Server.Connection.GetEnemyByID(helpers.NewContext(1), &pb.GetEnemyByIDRequest{
			EnemyID: enemy.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}
		enemy = rGetEnemyByID.GetEnemy()
	}

	switch inlineData.A {
	// Avvio di un nuovo combattimento
	case "start_fight":
		// In questo metodo non c'è niente da fare procedo con il stampare la card del combattimento
	case "up":
		// Setto nuova parte del corpo da colpire
		if c.Payload.BodySelection > 0 {
			c.Payload.BodySelection--
		} else {
			c.Payload.BodySelection = 3
		}
	case "down":
		// Setto nuova parte del corpo da colpire
		if c.Payload.BodySelection < 3 {
			c.Payload.BodySelection++
		} else {
			c.Payload.BodySelection = 0
		}
	case "no_action":
		//
	case "hit":
		c.Hit(enemy, planetMap, inlineData)
		return
	case "use":
		c.UseItem(inlineData)
		return
	}

	// Recupero arma equipaggiata
	var rGetPlayerWeaponEquipped *pb.GetPlayerWeaponEquippedResponse
	rGetPlayerWeaponEquipped, _ = config.App.Server.Connection.GetPlayerWeaponEquipped(helpers.NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
		PlayerID: c.Player.ID,
	})

	weaponEquipped := helpers.Trans(c.Player.Language.Slug, "combat.no_weapon")
	var weaponDurability int32
	if rGetPlayerWeaponEquipped.GetWeapon().GetID() > 0 {
		weaponEquipped = rGetPlayerWeaponEquipped.GetWeapon().GetName()
		weaponDurability = rGetPlayerWeaponEquipped.GetWeapon().GetDurability()
	}

	combactStatusMessage := helpers.NewEditMessage(
		c.Player.ChatID,
		c.Update.CallbackQuery.Message.MessageID,
		helpers.Trans(c.Player.Language.Slug, "combat.card",
			enemy.Name, strings.ToUpper(enemy.Rarity.Slug),
			enemy.LifePoint,
			enemy.LifeMax,
			c.Player.Username,
			c.Player.GetLifePoint(),
			c.Player.GetLevel().GetPlayerMaxLife(),
			helpers.Trans(c.Player.Language.Slug, bodyParts[c.Payload.BodySelection]), // Parte del corpo selezionata
			weaponEquipped, weaponDurability, // Arma equipaggiata e durabilità
		),
	)

	combactStatusMessage.ParseMode = "markdown"
	combactStatusMessage.ReplyMarkup = c.PlayerFightKeyboard()
	if _, err = helpers.SendMessage(combactStatusMessage); err != nil {
		c.Logger.Panic(err)
	}

	return
}

func (c *HuntingController) PlayerFightKeyboard() *tgbotapi.InlineKeyboardMarkup {
	var err error
	newfightKeyboard := new(tgbotapi.InlineKeyboardMarkup)

	// #######################
	// Usabili: recupero quali item possono essere usati in combattimento
	// #######################
	// Ciclo pozioni per ID item
	for _, itemID := range []uint32{1, 2, 3} {
		var rGetItemByID *pb.GetItemByIDResponse
		if rGetItemByID, err = config.App.Server.Connection.GetItemByID(helpers.NewContext(1), &pb.GetItemByIDRequest{
			ItemID: itemID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		var rGetPlayerItemByID *pb.GetPlayerItemByIDResponse
		if rGetPlayerItemByID, err = config.App.Server.Connection.GetPlayerItemByID(helpers.NewContext(1), &pb.GetPlayerItemByIDRequest{
			PlayerID: c.Player.ID,
			ItemID:   itemID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiunto tasto solo se la quantità del player è > 0
		if rGetPlayerItemByID.GetPlayerInventory().GetQuantity() > 0 {
			var potionStruct = helpers.InlineDataStruct{C: "hunting", AT: "fight", A: "use", D: rGetItemByID.GetItem().GetID()}
			newfightKeyboard.InlineKeyboard = append(newfightKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("%s (%v)",
						helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("items.%s", rGetItemByID.GetItem().GetSlug())),
						rGetPlayerItemByID.GetPlayerInventory().GetQuantity(),
					),
					potionStruct.GetDataString(),
				),
			))
		}
	}

	// #######################
	// Keyboard Selezione, attacco e fuga
	// #######################
	newfightKeyboard.InlineKeyboard = append(newfightKeyboard.InlineKeyboard, fightKeyboard...)

	// #######################
	// Abilità
	// #######################
	// Verifico se il player possiede abilità di comattimento o difesa
	var rCheckIfPlayerHaveAbility *pb.CheckIfPlayerHaveAbilityResponse
	if rCheckIfPlayerHaveAbility, err = config.App.Server.Connection.CheckIfPlayerHaveAbility(helpers.NewContext(1), &pb.CheckIfPlayerHaveAbilityRequest{
		PlayerID:  c.Player.ID,
		AbilityID: 7, // Attacco pesante
	}); err != nil {
		c.Logger.Panic(err)
	}

	if rCheckIfPlayerHaveAbility.GetHaveAbility() {
		// Appendo abilità player
		var dataAbilityStruct = helpers.InlineDataStruct{C: "hunting", AT: "fight", A: "hit", SA: "ability", D: rCheckIfPlayerHaveAbility.GetAbility().GetID()}
		newfightKeyboard.InlineKeyboard = append(newfightKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("safeplanet.accademy.ability.%s", rCheckIfPlayerHaveAbility.GetAbility().GetSlug())),
				dataAbilityStruct.GetDataString(),
			),
		))
	}

	return newfightKeyboard
}

func (c *HuntingController) ReturnToMap(planetMap *pb.PlanetMap) {
	var err error

	// Trasformo la mappa in qualcosa di più leggibile su telegram
	var decodedMap string
	if decodedMap, err = helpers.DecodeMapToDisplay(planetMap, c.Payload.PlayerPositionX, c.Payload.PlayerPositionY); err != nil {
		c.Logger.Panic(err)
	}

	// Forzo invio messaggio contenente la mappa
	returnMessage := helpers.NewEditMessage(
		c.Player.ChatID,
		c.Update.CallbackQuery.Message.MessageID,
		decodedMap,
	)

	returnMessage.ParseMode = "HTML"
	returnMessage.ReplyMarkup = &mapKeyboard
	if _, err = helpers.SendMessage(returnMessage); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *HuntingController) UseItem(inlineData helpers.InlineDataStruct) {
	var err error

	// Recupero dettagli item che si vuole usare
	var rGetItemByID *pb.GetItemByIDResponse
	if rGetItemByID, err = config.App.Server.Connection.GetItemByID(helpers.NewContext(1), &pb.GetItemByIDRequest{
		ItemID: inlineData.D,
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Richiamo il ws per usare l'item selezionato
	if _, err = config.App.Server.Connection.UseItem(helpers.NewContext(1), &pb.UseItemRequest{
		PlayerID: c.Player.ID,
		ItemID:   rGetItemByID.GetItem().GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(helpers.Trans(c.Player.GetLanguage().GetSlug(), "continue"), fightNoAction.GetDataString()),
		),
	)

	var combactMessage tgbotapi.EditMessageTextConfig
	combactMessage = helpers.NewEditMessage(
		c.Player.ChatID,
		c.Update.CallbackQuery.Message.MessageID,
		helpers.Trans(c.Player.Language.Slug, "combat.use_item",
			helpers.Trans(c.Player.GetLanguage().GetSlug(), fmt.Sprintf("items.%s", rGetItemByID.GetItem().GetSlug())),
			rGetItemByID.GetItem().GetValue(),
		),
	)
	combactMessage.ReplyMarkup = &keyboard
	combactMessage.ParseMode = "markdown"
	if _, err = helpers.SendMessage(combactMessage); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *HuntingController) Hit(enemy *pb.Enemy, planetMap *pb.PlanetMap, inlineData helpers.InlineDataStruct) {
	var err error

	// Verifico se il player vuole usare un'abilità
	var abilityID uint32
	if inlineData.SA == "ability" {
		abilityID = inlineData.D
	}

	// Effettuo chiamata al ws e recupero response dell'attacco
	var rHitEnemy *pb.HitEnemyResponse
	if rHitEnemy, err = config.App.Server.Connection.HitEnemy(helpers.NewContext(1), &pb.HitEnemyRequest{
		EnemyID:       enemy.GetID(),
		PlayerID:      c.Player.ID,
		BodySelection: c.Payload.BodySelection,
		AbilityID:     abilityID,
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Verifico se il MOB è morto
	if rHitEnemy.GetEnemyDie() {
		c.EnemyDie(rHitEnemy, planetMap)
		return
	}

	// Verifico se il PLAYER è morto
	if rHitEnemy.GetPlayerDie() {
		c.PlayerDie()
		return
	}

	var combactMessage tgbotapi.EditMessageTextConfig
	if rHitEnemy.GetEnemyDodge() {
		combactMessage = helpers.NewEditMessage(
			c.Player.ChatID,
			c.Update.CallbackQuery.Message.MessageID,
			helpers.Trans(c.Player.Language.Slug, "combat.enemy_dodge", rHitEnemy.GetEnemyDamage()),
		)
	} else if rHitEnemy.GetPlayerDodge() {
		combactMessage = helpers.NewEditMessage(
			c.Player.ChatID,
			c.Update.CallbackQuery.Message.MessageID,
			helpers.Trans(c.Player.Language.Slug, "combat.player_dodge", rHitEnemy.GetPlayerDamage()),
		)
	} else {
		combactMessage = helpers.NewEditMessage(
			c.Player.ChatID,
			c.Update.CallbackQuery.Message.MessageID,
			helpers.Trans(c.Player.Language.Slug, "combat.damage", rHitEnemy.GetPlayerDamage(), rHitEnemy.GetEnemyDamage()),
		)
	}

	// Aggiungo dettagli abilità
	if abilityID == 7 {
		combactMessage.Text += "\n A causa della tua abilità hai perso ulteriri 5HP"
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Ok!", fightNoAction.GetDataString()),
		),
	)

	combactMessage.ReplyMarkup = &keyboard
	combactMessage.ParseMode = "markdown"
	if _, err = helpers.SendMessage(combactMessage); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *HuntingController) PlayerDie() {
	// Aggiorno messaggio notificando al player che è morto
	playerDieMessage := helpers.NewEditMessage(
		c.Player.ChatID,
		c.Update.CallbackQuery.Message.MessageID,
		helpers.Trans(c.Player.Language.Slug, "combat.player_killed"),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				helpers.Trans(c.Player.Language.Slug, "continue"), fightPlayerDie.GetDataString(),
			),
		),
	)

	playerDieMessage.ReplyMarkup = &keyboard
	playerDieMessage.ParseMode = "markdown"
	if _, err := helpers.SendMessage(playerDieMessage); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *HuntingController) EnemyDie(rHitEnemy *pb.HitEnemyResponse, planetMap *pb.PlanetMap) {
	// Costruisco messaggio di recap del drop
	var dropRecap string

	// Aggiungo risorse o item trovati
	if rHitEnemy.GetEnemyDrop().GetResource() != nil {
		dropRecap += helpers.Trans(c.Player.Language.Slug, "combat.found.resource", rHitEnemy.GetEnemyDrop().GetResource().GetName())
	} else if rHitEnemy.GetEnemyDrop().GetItem() != nil {
		itemFound := helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("items.%s", rHitEnemy.GetEnemyDrop().GetItem().GetSlug()))
		dropRecap += helpers.Trans(c.Player.Language.Slug, "combat.found.item", itemFound)
	} else {
		dropRecap += helpers.Trans(c.Player.Language.Slug, "combat.found.nothing")
	}

	// Aggiungo dettaglio monete recuperate
	if rHitEnemy.GetEnemyDrop().GetTransaction() != nil {
		dropRecap += helpers.Trans(c.Player.Language.Slug, "combat.found.transaction", rHitEnemy.GetEnemyDrop().GetTransaction().GetValue())
	}

	// Aggiungo anche esperinza recuperata
	dropRecap += fmt.Sprintf("\n\n%s", helpers.Trans(c.Player.Language.Slug, "combat.experience", rHitEnemy.GetPlayerExperience()))

	// Aggiorno modifica del messaggio
	enemyDieMessage := helpers.NewEditMessage(
		c.Player.ChatID,
		c.Update.CallbackQuery.Message.MessageID,
		helpers.Trans(c.Player.Language.Slug, "combat.mob_killed", dropRecap),
	)

	var keyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				helpers.Trans(c.Player.Language.Slug, "continue"), fightReturnMap.GetDataString(),
			),
		),
	)

	enemyDieMessage.ParseMode = "markdown"
	enemyDieMessage.ReplyMarkup = &keyboard
	if _, err := helpers.SendMessage(enemyDieMessage); err != nil {
		c.Logger.Panic(err)
	}

	// Se il mob è morto è necessario aggiornare la mappa
	c.RefreshMap(planetMap.ID)
}

// RefreshMap - Necessario per refreshare la mappa in caso
// di sconfitta di mob o apertura di tesori.
func (c *HuntingController) RefreshMap(MapID uint32) {
	var err error

	// Un mob è stato scofinto riaggiorno mappa e riaggiorno record cache
	var rGetPlanetMapByID *pb.GetPlanetMapByIDResponse
	if rGetPlanetMapByID, err = config.App.Server.Connection.GetPlanetMapByID(helpers.NewContext(1), &pb.GetPlanetMapByIDRequest{
		PlanetMapID: MapID,
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Registro mappa e posizione iniziale del player
	if err = helpers.SetMapInCache(rGetPlanetMapByID.GetPlanetMap()); err != nil {
		c.Logger.Panic(err)
	}
	return
}
