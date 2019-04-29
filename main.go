package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"./p3"
)

func say(s string) {
	for true {
		seconds := rand.Intn(6) + 5
		fmt.Println(seconds)
		randTime := time.Duration(seconds)
		time.Sleep(randTime * time.Second)

	}
}

func main() {

	router := p3.NewRouter()
	if len(os.Args) > 1 {
		log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	} else {
		log.Fatal(http.ListenAndServe(":6680", router))
	}
}

func Copy(a map[string]int32) map[string]int32 {
	copy := map[string]int32{}
	for k, v := range a {
		copy[k] = v
	}
	return copy
}
