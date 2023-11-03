package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/bradford-hamilton/anycache/pkg/anycache"
)

func main() {
	ac, err := anycache.New(10 * 1024 * 1024)
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i < 1000; i++ {
		ac.Set(i, strconv.Itoa(i))
	}

	fmt.Println(ac.Keys())
}

type neatStructure struct {
	Name string
	Age  int
}
