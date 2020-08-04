package helpers

import (
	"strconv"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

// ListRecipe - Metodo che aiuta a recuperare la lista di risore necessarie
// al crafting di un determianto item
func ListRecipe(needed map[uint32]int32) (result string, err error) {
	for resourceID, value := range needed {
		rGetResourceByID, err := services.NnSDK.GetResourceByID(NewContext(1), &pb.GetResourceByIDRequest{
			ID: resourceID,
		})
		if err != nil {
			return result, err
		}

		result += rGetResourceByID.GetResource().GetName() + " x" + strconv.Itoa(int(value)) + "\n"
	}

	return result, err
}

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

	rGetAllWeaponCategory, err := services.NnSDK.GetAllWeaponCategory(NewContext(1), &pb.GetAllWeaponCategoryRequest{})
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(rGetAllArmorCategory.GetArmorCategories()); i++ {
		result = append(result, rGetAllArmorCategory.GetArmorCategories()[i].Slug)
	}

	for i := 0; i < len(rGetAllWeaponCategory.GetWeaponCategories()); i++ {
		result = append(result, rGetAllWeaponCategory.GetWeaponCategories()[i].Slug)
	}
	return
}
