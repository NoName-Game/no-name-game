package controllers

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"

	"bitbucket.org/no-name-game/no-name/app/commands"
	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func Hunting(update tgbotapi.Update, player models.Player) {

	message := update.Message
	routeName := "route.hunting"
	state := helpers.StartAndCreatePlayerState(routeName, player)

	type payloadHunting struct {
		MobID uint
		Score int //Number of enemy defeated
	}

	var payload payloadHunting
	helpers.UnmarshalPayload(state.Payload, &payload)
	mob := models.GetEnemyByID(payload.MobID)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage", player.Language.Slug)
	switch state.Stage {
	case 1:
		if state.FinishAt.Before(time.Now()) {
			validationFlag = true
		} else {
			validationMessage = helpers.Trans("wait", player.Language.Slug, state.FinishAt.Format("15:04:05"))
		}
	case 2:
		if message.Text == helpers.Trans("continue", player.Language.Slug) && mob.LifePoint > 0 {
			validationFlag = true
		} else if mob.LifePoint == 0 {
			validationFlag = true
			mob.Delete()    // Delete the enemy from table
			state.Stage = 4 //Drop
			state.Update()
		}
	case 3:
		if strings.Contains(message.Text, helpers.Trans("combat.attack_with", player.Language.Slug)) {
			validationFlag = true
		}
	case 4:
		if message.Text == helpers.Trans("continue", player.Language.Slug) {
			validationFlag = true
			state.Stage = 0
		} else if message.Text == helpers.Trans("nope", player.Language.Slug) {
			validationFlag = true
			state.Stage = 5
		}
	}

	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(player.ChatID, validationMessage)
			services.SendMessage(validatorMsg)
		}
	}

	//====================================
	// LOGIC FLUX:
	// Searching -> Finding -> Fight [Loop] -> Drop
	//====================================

	// FIGHT SYSTEM : Enemy Card / Choose Attack -> Calculate Damage -> Apply Damage

	//====================================
	// Stage
	//====================================
	switch state.Stage {
	case 0:
		// Set timer
		state.FinishAt = commands.GetEndTime(0, int(5*(payload.Score/3)), 0)
		state.ToNotify = true
		state.Stage = 1
		mob = helpers.NewEnemy()
		payload.MobID = mob.ID
		payload.Score = 1
		payloadUpdated, _ := json.Marshal(payload)
		state.Payload = string(payloadUpdated)
		state.Update()

		services.SendMessage(services.NewMessage(player.ChatID, helpers.Trans("hunting.searching", player.Language.Slug, state.FinishAt.Format("15:04:05"))))
	case 1:
		if validationFlag {
			// Enemy found
			state.Stage = 2
			state.Update()
			msg := services.NewMessage(player.ChatID, helpers.Trans("hunting.enemy.found", player.Language.Slug, mob.Name))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("continue", player.Language.Slug))))
			services.SendMessage(msg)
		}
	case 2:
		if validationFlag {
			state.Stage = 3
			state.Update()
			msg := services.NewMessage(player.ChatID, helpers.Trans("hunting.enemy.card", player.Language.Slug, mob.Name, mob.LifePoint, player.Stats.LifePoint))
			msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
				ResizeKeyboard: true,
				Keyboard:       helpers.GenerateWeaponKeyboard(player),
			}
			services.SendMessage(msg)
		}
	case 3:
		if validationFlag {
			// Calculating damage

			weaponName := strings.SplitN(message.Text, " ", 3)[2]

			weapon := models.GetWeaponByName(weaponName)

			var playerDamage uint

			switch weapon.WeaponCategory.Slug {
			case "knife":
				// Knife damage
				playerDamage = uint(rand.Int31n(6)+1) + (weapon.RawDamage + ((player.Stats.Strength + player.Stats.Dexterity) / 2))
			default:
				playerDamage = uint(rand.Int31n(6)+1) + (weapon.RawDamage + ((player.Stats.Intelligence + player.Stats.Dexterity) / 2))
			}

			mob.LifePoint -= playerDamage
			mob.Update()
			var text string
			if mob.LifePoint == 0 {
				text = helpers.Trans("combat.last_hit", player.Language.Slug)
			} else {
				mobDamage := uint(rand.Int31n(17) + 1)
				helpers.DecrementLife(mobDamage, player)
				text = helpers.Trans("combat.damage", player.Language.Slug, playerDamage, mobDamage)
			}

			state.Stage = 2
			state.Update()

			msg := services.NewMessage(player.ChatID, text)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("continue", player.Language.Slug))))
			services.SendMessage(msg)
		}
	case 4:
		if validationFlag {
			player.Stats.Experience++
			player.Update()
			msg := services.NewMessage(player.ChatID, helpers.Trans("hunting.experience_earned", player.Language.Slug, 1))
			services.SendMessage(msg)
			msg = services.NewMessage(player.ChatID, helpers.Trans("hunting.continue", player.Language.Slug))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("continue", player.Language.Slug)), tgbotapi.NewKeyboardButton(helpers.Trans("nope", player.Language.Slug))))
			services.SendMessage(msg)
		}
	case 5:
		if validationFlag {
			//====================================
			// IMPORTANT!
			//====================================
			helpers.FinishAndCompleteState(state, player)
			//====================================

			msg := services.NewMessage(message.Chat.ID, helpers.Trans("hunting.complete", player.Language.Slug))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("back"),
				),
			)
			services.SendMessage(msg)
		}
	}
}
