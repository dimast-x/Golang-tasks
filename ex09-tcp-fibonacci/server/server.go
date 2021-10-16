package main

import (
	respStruct "client/response"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"time"
)

var cache = make(map[int]*big.Int)
var input int

func main() {

	client, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Server is running")

	connection, err := client.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	cache[0] = big.NewInt(0)
	cache[1] = big.NewInt(1)

	for {
		dec := json.NewDecoder(connection)
		err := dec.Decode(&input)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Received: %d \n", input)

		count := time.Now()
		pesponse := respStruct.Response{
			Result: Fibo(input),
			Spent:  time.Since(count),
		}
		enc := json.NewEncoder(connection)
		err = enc.Encode(pesponse)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Answer: %d \n", pesponse.Result)
	}
}

func Fibo(n int) *big.Int {

	if cache[n] != nil {
		return cache[n]
	}

	switch n {
	case 0:
		return cache[0]
	case 1:
		return cache[1]
	default:
		length := len(cache)
		prev := cache[length-2]
		value := cache[length-1]

		for i := length; i <= n; i++ {
			cache[i] = prev.Add(prev, value)
			prev, value = value, prev
		}
		return value
	}
}
