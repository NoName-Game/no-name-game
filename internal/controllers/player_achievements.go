package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerAchievementsController
// ====================================
type PlayerAchievementsController struct {
	Controller
	Payload struct {
		AchievementCategoryID uint32
	}
}

// ====================================
// Handle
// ====================================
func (c *PlayerAchievementsController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.achievements",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlayerController{},
				FromStage: 0,
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
	c.Completing(&c.Payload)
}

// ====================================
// Validator
// ====================================
func (c *PlayerAchievementsController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verificio quale categoria di achievement è stata passata
	// ##################################################################################################
	case 1:
		var err error
		var rGetAllAchievementCategory *pb.GetAllAchievementCategoryResponse
		if rGetAllAchievementCategory, err = config.App.Server.Connection.GetAllAchievementCategory(helpers.NewContext(1), &pb.GetAllAchievementCategoryRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		for _, category := range rGetAllAchievementCategory.GetAchievementCategories() {
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("achievement.category.%s", category.GetSlug())) {
				c.Payload.AchievementCategoryID = category.GetID()
				return false
			}
		}

		return true
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *PlayerAchievementsController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Mostro categorie achievement disponibili
	// ##################################################################################################
	case 0:
		var err error
		var rGetAllAchievementCategory *pb.GetAllAchievementCategoryResponse
		if rGetAllAchievementCategory, err = config.App.Server.Connection.GetAllAchievementCategory(helpers.NewContext(1), &pb.GetAllAchievementCategoryRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		var categoriesKeyboard [][]tgbotapi.KeyboardButton
		for _, category := range rGetAllAchievementCategory.GetAchievementCategories() {
			categoriesKeyboard = append(categoriesKeyboard, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("achievement.category.%s", category.GetSlug())),
				),
			))
		}

		categoriesKeyboard = append(categoriesKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
		))

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "player.achievement.intro"))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       categoriesKeyboard,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Avanzo di stage
		c.CurrentState.Stage = 1

	// In questo stage chiedo di indicarmi quale armatura o arma intende equipaggiare
	case 1:
		// Recuero dettagli della categoria scelta
		var rGetAchievementCategoryByID *pb.GetAchievementCategoryByIDResponse
		if rGetAchievementCategoryByID, err = config.App.Server.Connection.GetAchievementCategoryByID(helpers.NewContext(1), &pb.GetAchievementCategoryByIDRequest{
			CategoryID: c.Payload.AchievementCategoryID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero achievement della categoria
		var rGetAchievementsByCategoryID *pb.GetAchievementsByCategoryIDResponse
		if rGetAchievementsByCategoryID, err = config.App.Server.Connection.GetAchievementsByCategoryID(helpers.NewContext(1), &pb.GetAchievementsByCategoryIDRequest{
			CategoryID: c.Payload.AchievementCategoryID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		var achievementList string
		for _, achievement := range rGetAchievementsByCategoryID.GetAchievements() {
			// Verifico se il player ha portato ha termine l'achievement
			var rCheckIfPlayerHaveAchievement *pb.CheckIfPlayerHaveAchievementResponse
			if rCheckIfPlayerHaveAchievement, err = config.App.Server.Connection.CheckIfPlayerHaveAchievement(helpers.NewContext(1), &pb.CheckIfPlayerHaveAchievementRequest{
				PlayerID:      c.Player.ID,
				AchievementID: achievement.GetID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			if rCheckIfPlayerHaveAchievement.GetHaveAchievement() {
				achievementList += fmt.Sprintf("✅ %s\n",
					helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("achievement.%s", achievement.GetSlug())),
				)
				continue
			}

			// Il player non ha ancora portato a termine
			achievementList += fmt.Sprintf("❌ %s\n",
				helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("achievement.%s", achievement.GetSlug())),
			)
		}

		// Invio messaggio per conferma equipaggiamento
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "player.achievement.category_recap",
			helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("achievement.category.%s", rGetAchievementCategoryByID.GetAchievementCategory().GetSlug())),
			achievementList,
		))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}
	}

	return
}
