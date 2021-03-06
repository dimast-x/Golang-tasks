package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	ws "golang-api/websocket"
	"net/http"
	"strings"

	"github.com/openware/rango/pkg/auth"
)

type JWTService struct {
	keys      *auth.KeyStore
	blacklist map[string]string
}

func NewJWTService(privKeyPath, pubKeyPath string) (*JWTService, error) {
	keys, err := auth.LoadOrGenerateKeys(privKeyPath, pubKeyPath)
	if err != nil {
		return nil, err
	}
	return &JWTService{
		keys:      keys,
		blacklist: make(map[string]string),
	}, nil
}
func (j *JWTService) GenearateJWT(u User) (string, error) {
	return auth.ForgeToken("empty", u.Email, "empty", 0, j.keys.PrivateKey, nil)
}
func (j *JWTService) ParseJWT(jwt string) (auth.Auth, error) {
	return auth.ParseAndValidate(jwt, j.keys.PublicKey)
}

type JWTParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *UserService) JWT(
	w http.ResponseWriter,
	r *http.Request,
	jwtService *JWTService,
) {
	params := &JWTParams{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		handleError(errors.New("could not read params"), w)
		return
	}
	passwordDigest := md5.New().Sum([]byte(params.Password))
	user, err := u.repository.Get(params.Email)
	if err != nil {
		handleError(err, w)
		return
	}
	if string(passwordDigest) != user.PasswordDigest {
		handleError(errors.New("invalid login params"), w)
		return
	}
	token, err := jwtService.GenearateJWT(user)
	if err != nil {
		handleError(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token))
}

type ProtectedHandler func(rw http.ResponseWriter, r *http.Request, u User)

func (j *JWTService) jwtAuth(users UserRepository, h ProtectedHandler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		auth, err := j.ParseJWT(token)
		if err != nil {
			rw.WriteHeader(401)
			rw.Write([]byte("unauthorized"))
			return
		}
		user, err := users.Get(auth.Email)
		if err != nil {
			rw.WriteHeader(401)
			rw.Write([]byte("unauthorized"))
			return
		}
		reason := j.blacklist[auth.Email]
		if len(reason) != 0 {
			rw.WriteHeader(401)
			rw.Write([]byte("banned. reason: " + reason))
			return
		}
		h(rw, r, user)
	}
}

type ProtectedHandlerWithUserStorage func(rw http.ResponseWriter, r *http.Request, u User, service *UserService)

func (j *JWTService) jwtAdminAuth(service *UserService, h ProtectedHandlerWithUserStorage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		auth, err := j.ParseJWT(token)
		if err != nil {
			rw.WriteHeader(401)
			rw.Write([]byte("unauthorized"))
			return
		}
		user, err := service.repository.Get(auth.Email)
		if err != nil {
			rw.WriteHeader(401)
			rw.Write([]byte("unauthorized"))
			return
		}
		reason := j.blacklist[auth.Email]
		if len(reason) != 0 {
			rw.WriteHeader(401)
			rw.Write([]byte("banned. reason: " + reason))
			return
		}
		if user.Role == "user" {
			rw.WriteHeader(401)
			rw.Write([]byte("you are not allowed to ban other users"))
			return
		}
		h(rw, r, user, service)
	}
}

func (j *JWTService) jwtAuthSuperuserOnly(
	users UserRepository,
	h ProtectedHandler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		auth, err := j.ParseJWT(token)
		if err != nil {
			rw.WriteHeader(401)
			rw.Write([]byte("unauthorized"))
			return
		}
		user, err := users.Get(auth.Email)
		if err != nil {
			rw.WriteHeader(401)
			rw.Write([]byte("unauthorized"))
			return
		}
		if user.Role != "superadmin" {
			rw.WriteHeader(401)
			rw.Write([]byte("unauthorized"))
			return
		}
		h(rw, r, user)
	}
}

func (j *JWTService) jwtWS(service *UserService, hub *ws.Hub) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		auth, err := j.ParseJWT(token)
		if err != nil {
			rw.WriteHeader(401)
			rw.Write([]byte("unauthorized"))
			return
		}
		user, err := service.repository.Get(auth.Email)
		if err != nil {
			rw.WriteHeader(401)
			rw.Write([]byte("unauthorized"))
			return
		}
		reason := j.blacklist[auth.Email]
		if len(reason) != 0 {
			rw.WriteHeader(401)
			rw.Write([]byte("banned. reason: " + reason))
			return
		}
		if user.Role == "user" {
			rw.WriteHeader(401)
			rw.Write([]byte("you are not allowed to ban other users"))
			return
		}
		ws.ServeWs(hub, rw, r)
	}
}
