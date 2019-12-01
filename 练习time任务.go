package main

import (
	"time"
	"fmt"
)

func main() {
	ticker := time.NewTicker(time.Second * 2)
	for ti := range ticker.C {
		fmt.Println(ti)
	}
}