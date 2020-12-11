package helpers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"github.com/sirupsen/logrus"
)

// GetExplorationCategories - Recupera dalla cache o dal DB le categorie di esplorazione
func GetExplorationCategories() (categories []*pb.ExplorationCategory, err error) {
	// Tento di recuperare categorie da cache
	if categories, err = GetExplorationCategoriesInCache(); err != nil {
		// Se non sono state trovare recupero dal ws
		var rGetAllExplorationCategories *pb.GetAllExplorationCategoriesResponse
		if rGetAllExplorationCategories, err = config.App.Server.Connection.GetAllExplorationCategories(NewContext(1), &pb.GetAllExplorationCategoriesRequest{}); err != nil {
			logrus.Panicf("error getting exploration categories: %s", err.Error())
		}

		categories = rGetAllExplorationCategories.GetExplorationCategories()

		// Creo cache posizione
		if err = SetExplorationCategoriesInCache(categories); err != nil {
			logrus.Errorf("error creating exploration categories in cache: %s", err.Error())
		}
	}

	return
}
