package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	setStr := "[\"abc.com\",\"dev.com\",\"ges.com\"]"
	var EmailAccList = make([]string, 5)
	json.Unmarshal([]byte(setStr), &EmailAccList)

	setObj := make(map[string]interface{})
	for _,data := range EmailAccList {
		setObj[data] = data
	}
	fmt.Println(setObj["ges.com"])
	d,ok := setObj["ges.com1"]
	fmt.Println(d," == ", ok)
}
