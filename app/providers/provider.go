package providers

import (
	"encoding/json"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"github.com/pkg/errors"
)

type Provider interface{}

type BaseProvider struct {
	SDKResp nnsdk.APIResponse
}

func (bp *BaseProvider) Response(response interface{}) (err error) {
	if bp.SDKResp.Error != "" && bp.SDKResp.StatusCode == 500 {
		err = errors.Errorf("%v - %s : %s", bp.SDKResp.StatusCode, bp.SDKResp.Error, bp.SDKResp.Message)
		return err
	} else if bp.SDKResp.StatusCode == 200 || bp.SDKResp.StatusCode == 201 {
		err = json.Unmarshal(bp.SDKResp.Data, &response)
		if err != nil {
			return err
		}
	}

	return nil
}
