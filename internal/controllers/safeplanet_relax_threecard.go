package controllers

import (
	"fmt"
	"sort"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetRelaxThreeCardController
// ====================================
type SafePlanetRelaxThreeCardController struct {
	Controller
	Payload struct {
		CardChoiced int32
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetRelaxThreeCardController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.relax.threecard",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetRelaxController{},
				FromStage: 1,
			},
			PlanetType: []string{"safe"},
		},
	}) {
		return
	}

	// Validate
	if c.Validator() {
		c.Validate()
		return
	}

	// Ok! Run!
	c.Stage()

	// Completo progressione
	c.Completing(nil)
}

// ====================================
// Validator
// ====================================
func (c *SafePlanetRelaxThreeCardController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico avvio missione
	// ##################################################################################################
	case 0:
		if helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play") == c.Update.Message.Text {
			c.CurrentState.Stage = 1
		}

	// ##################################################################################################
	// In questo stage andremo a verificare lo stato della missione
	// ##################################################################################################
	case 2:
		// Verifico subito se il player ha deciso di terminare la partita
		if helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.give_up") == c.Update.Message.Text {
			c.CurrentState.Stage = 4
			return false
		}

		// Verifico se è stato passanto un numero
		verifyCards := map[int32]string{1: "1️⃣", 2: "2️⃣", 3: "3️⃣"}
		for i, card := range verifyCards {
			if card == c.Update.Message.Text {
				c.Payload.CardChoiced = i
				return false
			}
		}

		return true
	case 3:
		if helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.continue") != c.Update.Message.Text {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetRelaxThreeCardController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Illustro gioco e chiedo se vuole giocare
	// ##################################################################################################
	case 0:
		var keyboardRows [][]tgbotapi.KeyboardButton
		keyboardRows = append(keyboardRows, []tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play")),
		})

		// Aggiungo anche abbandona
		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.menu"),
			),
		))

		// Invio messaggi con il tipo di missioni come tastierino
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.intro"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRows,
			ResizeKeyboard: true,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

	// ##################################################################################################
	// Avvio partita e consegno prime carte
	// ##################################################################################################
	case 1:
		_, err = config.App.Server.Connection.ThreeCardGamePlay(helpers.NewContext(1), &pb.ThreeCardGamePlayRequest{
			PlayerID: c.Player.GetID(),
		})

		// Verifico se il player ha abbastanza soldi per giocare
		if err != nil && strings.Contains(err.Error(), "player dont have enough money") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di monete
			errorMsg := helpers.NewMessage(c.ChatID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.no_money"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		} else if err != nil {
			c.Logger.Panic(err)
		}

		// Se non è esploso nulla allora posso dare le prime carte
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.start"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1️⃣"),
				tgbotapi.NewKeyboardButton("2️⃣"),
				tgbotapi.NewKeyboardButton("3️⃣"),
			),
		)
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Avanzo di stato
		c.CurrentState.Stage = 2

	// ##################################################################################################
	// Verifico carta scelta e premi
	// ##################################################################################################
	case 2:
		// Effettuo chiamata per verificare la giocata
		var rThreeCardGameCheckResponse *pb.ThreeCardGameCheckResponse
		if rThreeCardGameCheckResponse, err = config.App.Server.Connection.ThreeCardGameCheck(helpers.NewContext(1), &pb.ThreeCardGameCheckRequest{
			PlayerID:     c.Player.GetID(),
			PlayerChoice: c.Payload.CardChoiced,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Ordino risorse droppate
		winningResources := c.winningResourcesList(rThreeCardGameCheckResponse.GetResources())

		// Recupero possibile bottino
		var recapList string
		if len(winningResources) > 0 {
			recapList = helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.recap")
			for _, resource := range winningResources {
				recapList += fmt.Sprintf(
					"- %s %v x %s (<b>%s</b>) %s\n",
					helpers.GetResourceCategoryIcons(resource.Resource.GetResourceCategoryID()),
					resource.Quantity,
					resource.Resource.Name,
					resource.Resource.Rarity.Slug,
					helpers.GetResourceBaseIcons(resource.Resource.GetBase()),
				)
			}
		}

		// ***********************************
		// Lose
		// ***********************************
		if !rThreeCardGameCheckResponse.GetWin() {
			loseMessage := helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.round_lose")

			// Invio messggio con ancora le 3 scelte di carte
			msg := helpers.NewMessage(c.Player.ChatID, fmt.Sprintf("%s \n%s", loseMessage, recapList))
			msg.ParseMode = tgbotapi.ModeHTML
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.continue"),
					),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
					),
				),
			)

			if _, err = helpers.SendMessage(msg); err != nil {
				c.Logger.Panic(err)
			}

			// Avanzo di stato
			c.CurrentState.Stage = 3
			return
		}

		resourceFound := fmt.Sprintf(
			"%s %s (<b>%s</b>) %s\n",
			helpers.GetResourceCategoryIcons(rThreeCardGameCheckResponse.GetResource().GetResourceCategoryID()),
			rThreeCardGameCheckResponse.GetResource().Name,
			rThreeCardGameCheckResponse.GetResource().Rarity.Slug,
			helpers.GetResourceBaseIcons(rThreeCardGameCheckResponse.GetResource().GetBase()),
		)

		// Recupero risorsa vinta in questo turno
		winMessage := helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.round_win", resourceFound)

		// Invio messggio con ancora le 3 scelte di carte
		msg := helpers.NewMessage(c.Player.ChatID, fmt.Sprintf("%s \n%s", winMessage, recapList))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1️⃣"),
				tgbotapi.NewKeyboardButton("2️⃣"),
				tgbotapi.NewKeyboardButton("3️⃣"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.give_up"),
				),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

	// ##################################################################################################
	// Verifica se il player vuole usare i diamanti per continuare
	// ##################################################################################################
	case 3:
		// Se entro qui vuol dire che ha deciso di usare i diamanti
		var rThreeCardGameRecoverPlay *pb.ThreeCardGameRecoverPlayResponse
		rThreeCardGameRecoverPlay, err = config.App.Server.Connection.ThreeCardGameRecoverPlay(helpers.NewContext(1), &pb.ThreeCardGameRecoverPlayRequest{
			PlayerID: c.Player.GetID(),
		})

		// Verifico se il player ha abbastanza soldi per giocare
		if err != nil && strings.Contains(err.Error(), "player dont have enough diamond") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di monete
			errorMsg := helpers.NewMessage(c.ChatID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.no_money"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		} else if err != nil {
			c.Logger.Panic(err)
		}

		recoverMessage := helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.recoverd")

		// Ordino risorse droppate
		winningResources := c.winningResourcesList(rThreeCardGameRecoverPlay.GetResources())

		// Recupero possibile bottino
		var recapList string
		if len(winningResources) > 0 {
			recapList = helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.recap")
			for _, resource := range winningResources {
				recapList += fmt.Sprintf(
					"- %s %v x %s (<b>%s</b>) %s\n",
					helpers.GetResourceCategoryIcons(resource.Resource.GetResourceCategoryID()),
					resource.Quantity,
					resource.Resource.Name,
					resource.Resource.Rarity.Slug,
					helpers.GetResourceBaseIcons(resource.Resource.GetBase()),
				)
			}
		}

		// Invio messggio con ancora le 3 scelte di carte
		msg := helpers.NewMessage(c.Player.ChatID, fmt.Sprintf("%s \n%s", recoverMessage, recapList))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1️⃣"),
				tgbotapi.NewKeyboardButton("2️⃣"),
				tgbotapi.NewKeyboardButton("3️⃣"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.give_up"),
				),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Se il player decide di usare il diamante allora torna allo stato 2
		c.CurrentState.Stage = 2

	// ##################################################################################################
	// Il player ha deciso di terminare e portarsi a casa il bottino
	// ##################################################################################################
	case 4:
		// Richiamo ws per conclusione e mando messaggio se tutto ok
		if _, err = config.App.Server.Connection.ThreeCardGameEndGame(helpers.NewContext(1), &pb.ThreeCardGameEndGameRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.end_game"))
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}

// Struttura per riepilogo risorse vinte
type WinningResourcesDropped struct {
	ResourceID uint32
	Resource   *pb.Resource
	Quantity   int32
}

func (c *SafePlanetRelaxThreeCardController) winningResourcesList(winResults []*pb.Resource) (results []WinningResourcesDropped) {
	for _, drop := range winResults {
		var found bool
		for i, resource := range results {
			if drop.ID == resource.ResourceID {
				results[i].Quantity++
				found = true
			}
		}

		// Se non è stato mai recuperata appendo
		if !found {
			results = append(results, WinningResourcesDropped{
				ResourceID: drop.ID,
				Resource:   drop,
				Quantity:   1,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Quantity > results[j].Quantity
	})

	return
}
