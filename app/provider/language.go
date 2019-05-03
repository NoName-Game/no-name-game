package provider

import (
	"encoding/json"
	"strconv"

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

func FindLanguageBy(value string, query string) (nnsdk.Language, error) {
	var language nnsdk.Language

	resp, err := services.NnSDK.MakeRequest("search/language?"+query+"="+value, nil).Get()
	if err != nil {
		return language, err
	}

	err = json.Unmarshal(resp.Data, &language)
	if err != nil {
		return language, err
	}

	return language, nil
}

func GetLanguageByID(id uint) (nnsdk.Language, error) {
	var language nnsdk.Language

	resp, err := services.NnSDK.MakeRequest("languages/"+strconv.FormatUint(uint64(id), 10), nil).Get()
	if err != nil {
		return language, err
	}

	err = json.Unmarshal(resp.Data, &language)
	if err != nil {
		return language, err
	}

	return language, nil
}

func GetLanguages() ([]nnsdk.Language, error) {
	var languages []nnsdk.Language

	resp, err := services.NnSDK.MakeRequest("languages", nil).Get()
	if err != nil {
		return languages, err
	}

	err = json.Unmarshal(resp.Data, &languages)
	if err != nil {
		return languages, err
	}

	return languages, nil
}
