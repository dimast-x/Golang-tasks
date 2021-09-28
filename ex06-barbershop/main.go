package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Customer struct {
	Id int
}

type Barber struct {
	busy bool
}

func (c Customer) String() string {
	return fmt.Sprintf("%d", c.Id)
}

func ClientGenerator(customers chan *Customer) {
	ID := 1
	for {
		rnd := rand.Intn(5)
		time.Sleep(time.Duration(rnd) * time.Second)
		customers <- &Customer{ID}
		ID += 1
	}
}

func Haircut(barber *Barber, customer *Customer, finished chan *Barber) {
	time.Sleep(2 * time.Second)
	fmt.Printf("Done with the client %s.\n", customer)
	finished <- barber
}

func BarberShop(customers <-chan *Customer) {
	barber := &Barber{busy: false}
	lobby := []*Customer{}
	syncBarberChan := make(chan *Barber)

	for {
		select {
		case customer := <-customers:
			if barber.busy {
				if len(lobby) < 3 {
					lobby = append(lobby, customer)
					fmt.Printf("Customer %s seated on a free seat.\n", customer)
				} else {
					fmt.Printf("No free seat, the client %s goes out.\n", customer)
				}
			} else {
				fmt.Printf("Customer %s goes to the barber.\n", customer)
				barber.busy = true
				go Haircut(barber, customer, syncBarberChan)
			}
			fmt.Printf("Lobby: %+v\n", lobby)
		case barber := <-syncBarberChan:
			if len(lobby) > 0 {
				customer := lobby[0]
				lobby = lobby[1:]
				fmt.Printf("Client %s goes to the barber.\n", customer)
				go Haircut(barber, customer, syncBarberChan)
			} else {
				barber.busy = false
				fmt.Printf("Barber goes sleep.\n")
			}
			fmt.Printf("Lobby: %v\n", lobby)
		}
	}
}

func main() {
	customers := make(chan *Customer)
	go ClientGenerator(customers)
	go BarberShop(customers)
	time.Sleep(1000 * time.Second)
}
