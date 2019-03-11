package namer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"bitbucket.org/no-name-game/no-name/services"

	"github.com/mb-14/gomarkov"
)

func buildModel(order int, path string, resource string) *gomarkov.Chain {
	chain := gomarkov.NewChain(order)
	for _, data := range getDataset(path + resource) {
		chain.Add(split(data))
	}

	return chain
}

func split(str string) []string {
	return strings.Split(str, "")
}

func getDataset(fileName string) []string {
	file, _ := os.Open(fileName)
	scanner := bufio.NewScanner(file)
	var list []string
	for scanner.Scan() {
		list = append(list, scanner.Text())
	}

	return list
}

func loadModel(resource string) (*gomarkov.Chain, error) {
	var chain gomarkov.Chain
	data, err := ioutil.ReadFile(resource)
	if err != nil {
		return &chain, err
	}
	err = json.Unmarshal(data, &chain)
	if err != nil {
		return &chain, err
	}

	return &chain, nil
}

func saveModel(chain *gomarkov.Chain, path string) {
	jsonObj, _ := json.Marshal(chain)
	err := ioutil.WriteFile(path+"/model.json", jsonObj, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

// TrainName - train name
func TrainName(path string, resource string) {
	chain := buildModel(3, path, resource)
	saveModel(chain, path)
}

// GenerateName - generate name
func GenerateName(model string) string {
	chain, err := loadModel(model)
	if err != nil {
		services.ErrorHandler("Error load name models", err)
	}

	order := chain.Order
	tokens := make([]string, 0)
	for i := 0; i < order; i++ {
		tokens = append(tokens, gomarkov.StartToken)
	}
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		next, _ := chain.Generate(tokens[(len(tokens) - order):])
		tokens = append(tokens, next)
	}

	return strings.Join(tokens[order:len(tokens)-1], "")
}
