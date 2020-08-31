package helpers

import (
	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

func CheckAndReturnCategorySlug(locale string, text string) (result string) {
	for _, slug := range GetAllSlugCategories() {
		if text == Trans(locale, slug) {
			return slug
		}
	}
	return ""
}

func GetAllSlugCategories() (result []string) {
	rGetAllArmorCategory, err := services.NnSDK.GetAllArmorCategory(NewContext(1), &pb.GetAllArmorCategoryRequest{})
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(rGetAllArmorCategory.GetArmorCategories()); i++ {
		result = append(result, rGetAllArmorCategory.GetArmorCategories()[i].Slug)
	}

	return
}
