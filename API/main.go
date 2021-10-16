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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

var (
	TotalUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "total_users",
		Help: "The total number of users",
	})

	GivenCakes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "given_cakes",
		Help: "The total number of given cakes",
	})

	TotalRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_requests",
		Help: "The total number of requests",
	})

	Request = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests",
			Help: "Time and path of every request execution",
		},
		[]string{"time", "path"},
	)

	users    = NewInMemoryUserStorage()
	messages = make(chan string)
)

func recordMetrics() {
	go func() {
		for {
			TotalUsers.Set(float64(len(users.storage)))
			time.Sleep(time.Second)
		}
	}()
}

func Publish(msg string) {
	messages <- msg
}

func main() {
	prometheus.MustRegister(Request)
	recordMetrics()
	r := mux.NewRouter()
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

	r.Handle("/metrics", promhttp.Handler())

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
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()
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
