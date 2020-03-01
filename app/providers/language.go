package providers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type LanguageProvider struct {
	BaseProvider
}

func (lp *LanguageProvider) FindLanguageBySlug(slug string) (response nnsdk.Language, err error) {
	lp.SDKResp, err = services.NnSDK.MakeRequest("search/language?slug="+slug, nil).Get()
	if err != nil {
		return response, err
	}

	err = lp.Response(&response)
	return
}

func (lp *LanguageProvider) FindLanguageBy(value string, query string) (response nnsdk.Language, err error) {
	lp.SDKResp, err = services.NnSDK.MakeRequest("search/language?"+query+"="+value, nil).Get()
	if err != nil {
		return response, err
	}

	err = lp.Response(&response)
	return
}

func (lp *LanguageProvider) GetLanguages() (response nnsdk.Languages, err error) {
	lp.SDKResp, err = services.NnSDK.MakeRequest("languages", nil).Get()
	if err != nil {
		return response, err
	}

	err = lp.Response(&response)
	return
}
