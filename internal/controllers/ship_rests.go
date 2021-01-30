package controllers

import (
	"time"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ShipRestsController
// ====================================
type ShipRestsController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *ShipRestsController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.ship.rests",
		},
		Configurations: ControllerConfigurations{
			ControllerBlocked: []string{"hunting", "exploration"},
			ControllerBack: ControllerBack{
				To:        &ShipController{},
				FromStage: 1,
			},
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
func (c *ShipRestsController) Validator() (hasErrors bool) {
	// NNT-140 Se il player arriva dalla caccia (quindi callback query) mando direttamente allo stage 0
	if c.Update.CallbackQuery != nil {
		c.CurrentState.Stage = 0
		return false
	}

	// Se è stato passato il comando sveglia e il player sta effettivamente dormendo lo sveglio
	if helpers.Trans(c.Player.Language.Slug, "ship.rests.wakeup") == c.Update.Message.Text {
		var rRestCheck *pb.RestCheckResponse
		if rRestCheck, _ = config.App.Server.Connection.RestCheck(helpers.NewContext(1), &pb.RestCheckRequest{
			PlayerID: c.Player.GetID(),
		}); rRestCheck.GetInRest() {
			c.CurrentState.Stage = 2
		}
	}

	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se il player necessita davvero di dormire
	// ##################################################################################################
	case 0:
		var err error
		var rGetRestsInfo *pb.GetRestsInfoResponse
		if rGetRestsInfo, err = config.App.Server.Connection.GetRestsInfo(helpers.NewContext(1), &pb.GetRestsInfoRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		if !rGetRestsInfo.GetNeedRests() {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.rests.dont_need")
			return true
		}

	// ##################################################################################################
	// Verifico se il player vuole dormire
	// ##################################################################################################
	case 1:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "ship.rests.start") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			return true
		}

	// ##################################################################################################
	// Verifico se il player vuole svegliarsi
	// ##################################################################################################
	case 2:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "ship.rests.wakeup") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.rests.validator.need_to_wakeup")
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *ShipRestsController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Dettaglio riposo
	// ##################################################################################################
	case 0:
		// Recupero informazioni per il recupero totale delle energie
		var restsInfo *pb.GetRestsInfoResponse
		if restsInfo, err = config.App.Server.Connection.GetRestsInfo(helpers.NewContext(1), &pb.GetRestsInfoRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Costruisco info per riposo
		var restsRecap string
		restsRecap = helpers.Trans(c.Player.Language.Slug, "ship.rests")
		if restsInfo.NeedRests {
			restsRecap += helpers.Trans(c.Player.Language.Slug, "ship.rests.full_rest_time", restsInfo.GetFullRestsTime())

			// Verifico quanto è necessario riposare per tornare in vita/svolgere le attività
			if restsInfo.GetPlayRestsTime() > 0 {
				restsRecap += helpers.Trans(c.Player.Language.Slug, "ship.rests.play_rest_time", restsInfo.GetPlayRestsTime())
			}
		}
		restsRecap += helpers.Trans(c.Player.Language.Slug, "ship.rests.info")

		// Aggiongo bottone start riposo
		var keyboardRow [][]tgbotapi.KeyboardButton
		if restsInfo.NeedRests {
			newKeyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "ship.rests.start"),
				),
			)
			keyboardRow = append(keyboardRow, newKeyboardRow)
		}

		// Aggiungo abbandona solo se il player non è morto e quindi obbligato a dormire
		if !c.Player.GetDead() {
			keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			))
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Player.ChatID, restsRecap)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRow,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 1

	// ##################################################################################################
	// Avvio riposo
	// ##################################################################################################
	case 1:
		var rStartPlayerRest *pb.StartPlayerRestResponse
		if rStartPlayerRest, err = config.App.Server.Connection.StartPlayerRest(helpers.NewContext(1), &pb.StartPlayerRestRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero orario fine riposo
		var finishAt time.Time
		if finishAt, err = helpers.GetEndTime(rStartPlayerRest.GetRestEndTime(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.rests.sleep", finishAt.Format("15:04:05")),
		)

		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "ship.rests.wakeup")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
	// ##################################################################################################
	// Fine riposo
	// ##################################################################################################
	case 2:
		var rEndPlayerRest *pb.EndPlayerRestResponse
		if rEndPlayerRest, err = config.App.Server.Connection.EndPlayerRest(helpers.NewContext(1), &pb.EndPlayerRestRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recap Rest
		var recapMessage = helpers.Trans(c.Player.Language.Slug, "ship.rests.need_to_rest")
		if rEndPlayerRest.GetLifeRecovered() > 0 {
			lifeRecovered := rEndPlayerRest.GetLifeRecovered()

			// Verifico overflow vita
			if lifeRecovered > c.Player.GetLevel().GetPlayerMaxLife() {
				lifeRecovered = c.Player.GetLevel().GetPlayerMaxLife()
			}

			recapMessage = helpers.Trans(c.Player.Language.Slug, "ship.rests.finish", lifeRecovered)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, recapMessage)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.more"),
				),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
