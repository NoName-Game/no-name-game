package provider

import (
	"encoding/json"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/services"
)

func FindLanguageBySlug(slug string) (nnsdk.Language, error) {
	var language nnsdk.Language

	resp, err := services.NnSDK.MakeRequest("search/language?slug="+slug, nil).Get()
	if err != nil {
		return language, err
	}

	err = json.Unmarshal(resp.Data, &language)
	if err != nil {
		return language, err
	}

	return language, nil
}
