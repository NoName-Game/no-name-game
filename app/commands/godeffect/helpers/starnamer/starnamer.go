package starnamer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mb-14/gomarkov"
)

// TrainStarName - train star name
func TrainStarName() {
	chain := buildModel(3)
	saveModel(chain)
}

func buildModel(order int) *gomarkov.Chain {
	chain := gomarkov.NewChain(order)
	for _, data := range getDataset("resources/stars/names.txt") {
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

// StarNameModel -
func StarNameModel() (*gomarkov.Chain, error) {
	var chain gomarkov.Chain
	data, err := ioutil.ReadFile("resources/stars/model.json")
	if err != nil {
		return &chain, err
	}
	err = json.Unmarshal(data, &chain)
	if err != nil {
		return &chain, err
	}
	return &chain, nil
}

func saveModel(chain *gomarkov.Chain) {
	jsonObj, _ := json.Marshal(chain)
	err := ioutil.WriteFile("resources/stars/model.json", jsonObj, 0644)
	if err != nil {
		fmt.Println(err)
	}
}
