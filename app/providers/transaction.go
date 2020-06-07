package providers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type TransactionProvider struct {
	BaseProvider
}

func (tp *TransactionProvider) CreateTransaction(request nnsdk.TransactionRequest) (response nnsdk.Transaction, err error) {
	tp.SDKResp, err = services.NnSDK.MakeRequest("transactions", request).Post()
	if err != nil {
		return response, err
	}

	err = tp.Response(&response)
	return
}
