package main

import (
	"context"
	"crypto/md5"
	ws "golang-api/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

func wrapJwt(jwt *JWTService, f func(http.ResponseWriter, *http.Request, *JWTService)) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		f(rw, r, jwt)
	}
}

func (u *UserService) AddSuperadmin() {
	email := os.Getenv("CAKE_ADMIN_EMAIL")
	passwordDigest := md5.New().Sum([]byte(os.Getenv("CAKE_ADMIN_PASSWORD")))
	newUser := User{
		Email:          email,
		PasswordDigest: string(passwordDigest),
		Role:           "superadmin",
		FavoriteCake:   "admincake",
	}
	err := u.repository.Add(email, newUser)
	if err != nil {
		panic(err)
	}
}

func (u *UserService) InitSuperadminVars() {
	os.Setenv("CAKE_ADMIN_EMAIL", "admin@cake.co")   //for test only
	os.Setenv("CAKE_ADMIN_PASSWORD", "admin12345:)") //for test only
	u.AddSuperadmin()
}

func publish(messages <-chan string, ch *amqp.Channel, q amqp.Queue) {
	for body := range messages {
		ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
	}
}

func Publish(msg string) {
	messages <- msg
}

var messages = make(chan string)

func main() {

	r := mux.NewRouter()
	users := NewInMemoryUserStorage()
	userService := UserService{repository: users}
	jwtService, err := NewJWTService("pubkey.rsa", "privkey.rsa")
	if err != nil {
		panic(err)
	}
	hub := ws.NewHub()
	go hub.Run()

	rmqconn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	defer rmqconn.Close()
	ch, err := rmqconn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	userService.InitSuperadminVars()
	r.HandleFunc("/cake", logRequest(jwtService.jwtAuth(users, getCakeHandler))).Methods(http.MethodGet)
	r.HandleFunc("/user/register", logRequest(userService.Register)).Methods(http.MethodPost)
	r.HandleFunc("/user/jwt", logRequest(wrapJwt(jwtService, userService.JWT))).Methods(http.MethodPost)
	r.HandleFunc("/user/me", logRequest(jwtService.jwtAuth(users, getInfoHandler))).Methods(http.MethodGet)
	r.HandleFunc("/user/favorite_cake", logRequest(jwtService.jwtAuth(users, userService.UpdateCake))).Methods(http.MethodPut)
	r.HandleFunc("/admin/promote", logRequest(jwtService.jwtAuthSuperuserOnly(users, userService.Promote))).Methods(http.MethodPost)
	r.HandleFunc("/admin/ban", logRequest(jwtService.jwtAdminAuth(&userService, jwtService.BanUser))).Methods(http.MethodPost)
	r.HandleFunc("/admin/unban", logRequest(jwtService.jwtAdminAuth(&userService, jwtService.UnbanUser))).Methods(http.MethodPost)
	r.HandleFunc("/admin/ws", jwtService.jwtWS(&userService, hub))

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		panic(err)
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		panic(err)
	}

	// go func() {
	// 	for {
	// 		Publish("Hello World!")
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }()

	for w := 1; w <= 6; w++ {
		go publish(messages, ch, q)
	}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			hub.Broadcast <- []byte(d.Body)
		}
	}()

	srv := http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		<-interrupt
		ctx, cancel := context.WithTimeout(context.Background(),
			5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()
	log.Println("Server started, hit Ctrl+C to stop")
	err = srv.ListenAndServe()

	if err != nil {
		log.Println("Server exited with error:", err)
	}
	log.Println("Good bye :)")
}
