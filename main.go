package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/bradford-hamilton/anycache/pkg/anycache"
)

// incomparable is a zero-width, non-comparable type. Adding it to a struct
// makes that struct also non-comparable, and generally doesn't add
// any size (as long as it's first).
//
// From: https://github.com/shogo82148/go/blob/3839447ac39b1c49cb14833f0832e5f934e5bf6b/src/net/http/http.go#L22
type incomparable [0]func()

type incmpUser struct {
	_ incomparable

	name string
	age  int
}

type cmpUser struct {
	name string
	age  int
}

func main() {
	ac, err := anycache.New(10 * 1024 * 1024)
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i < 1000; i++ {
		ac.Set(i, strconv.Itoa(i))
	}

	ac.Set(cmpUser{name: "bradford", age: 34}, "valid")

	item, ok := ac.Get(cmpUser{name: "bradford", age: 34})

	fmt.Println(ok)
	fmt.Println(item)

	// ac.Set(incmpUser{name: "bradford", age: 34}, "panic") // this panics as the keys aren't comparable

	m1 := make(map[cmpUser]int)
	// m2 := make(map[incmpUser]int) // doesn't compile - invalid map key type incmpUser compiler(IncomparableMapKey)

	// IncomparableMapKey occurs when a map key type does not support the == and
	// != operators.
	//
	// Per the spec:
	//  "The comparison operators == and != must be fully defined for operands of
	//  the key type; thus the key type must not be a function, map, or slice."
	//
	// Example:
	//  var x map[T]int
	//
	//  type T []int
	fmt.Println(m1, incmpUser{name: "", age: 0})
}
