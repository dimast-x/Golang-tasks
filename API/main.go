package main

import (
	"context"
	"crypto/md5"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

func getCakeHandler(w http.ResponseWriter, r *http.Request, u User) {
	w.Write([]byte(u.FavoriteCake))
}

func getInfoHandler(w http.ResponseWriter, r *http.Request, u User) {
	w.Write([]byte(u.Email + ", " + u.FavoriteCake + ", " + u.Role))
}

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

func main() {

	r := mux.NewRouter()
	users := NewInMemoryUserStorage()
	userService := UserService{repository: users}
	jwtService, err := NewJWTService("pubkey.rsa", "privkey.rsa")
	if err != nil {
		panic(err)
	}
	userService.InitSuperadminVars()
	r.HandleFunc("/cake", logRequest(jwtService.jwtAuth(users, getCakeHandler))).Methods(http.MethodGet)
	r.HandleFunc("/user/register", logRequest(userService.Register)).Methods(http.MethodPost)
	r.HandleFunc("/user/jwt", logRequest(wrapJwt(jwtService, userService.JWT))).Methods(http.MethodPost)
	r.HandleFunc("/user/me", logRequest(jwtService.jwtAuth(users, getInfoHandler))).Methods(http.MethodGet)
	r.HandleFunc("/user/favorite_cake", logRequest(jwtService.jwtAuth(users, userService.UpdateCake))).Methods(http.MethodPut)
	r.HandleFunc("/admin/promote", logRequest(jwtService.jwtAuthSuperuserOnly(users, userService.Promote))).Methods(http.MethodPost)
	r.HandleFunc("/admin/ban", logRequest(jwtService.jwtAdminAuth(&userService, jwtService.BanUser))).Methods(http.MethodPost)
	r.HandleFunc("/admin/unban", logRequest(jwtService.jwtAdminAuth(&userService, jwtService.UnbanUser))).Methods(http.MethodPost)

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
