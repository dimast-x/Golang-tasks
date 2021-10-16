package main

import (
	"bufio"
	respStruct "client/response"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

var response respStruct.Response

func main() {

	connection, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer connection.Close()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input, err := strconv.ParseInt(scanner.Text(), 10, 64)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = json.NewEncoder(connection).Encode(input)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = json.NewDecoder(connection).Decode(&response)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%s %d\n", response.Spent, response.Result)
	}
}
