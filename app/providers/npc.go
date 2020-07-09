package providers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
	"fmt"
)

type NpcProvider struct {
	BaseProvider
}

func (npc *NpcProvider) GetAll() (response nnsdk.Npcs, err error) {
	npc.SDKResp, err = services.NnSDK.MakeRequest("npcs", nil).Get()
	if err != nil {
		return response, err
	}

	err = npc.Response(&response)
	return
}

func (npc *NpcProvider) Deposit(playerID uint, amount int32) (response nnsdk.MoneyResponse, err error) {
	npc.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("npcs/bank/deposit/%v/%v", playerID, amount), nil).Post()
	if err != nil {
		return response, err
	}

	err = npc.Response(&response)
	return
}
func (npc *NpcProvider) Withdraw(playerID uint, amount int32) (response nnsdk.MoneyResponse, err error) {
	npc.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("npcs/bank/withdraw/%v/%v", playerID, amount), nil).Post()
	if err != nil {
		return response, err
	}

	err = npc.Response(&response)
	return
}
