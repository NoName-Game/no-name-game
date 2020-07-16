package providers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
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

func (npc *NpcProvider) Bank(request nnsdk.BankActionRequest) (response nnsdk.MoneyResponse, err error) {
	npc.SDKResp, err = services.NnSDK.MakeRequest("npcs/bank", request).Post()
	if err != nil {
		return response, err
	}

	err = npc.Response(&response)
	return
}

func (npc *NpcProvider) Craft(request nnsdk.CraftActionRequest) (response nnsdk.CraftResponse, err error) {
	npc.SDKResp, err = services.NnSDK.MakeRequest("npcs/craft", request).Post()
	if err != nil {
		return response, err
	}

	err = npc.Response(&response)
	return
}
