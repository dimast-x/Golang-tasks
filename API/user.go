package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/mail"
	//"net/mail"
)

type User struct {
	Email          string
	PasswordDigest string
	Role           string
	FavoriteCake   string
}
type UserRepository interface {
	Add(string, User) error
	Get(string) (User, error)
	Update(string, User) error
	Delete(string) (User, error)
}
type UserService struct {
	repository UserRepository
}

type UserRegisterParams struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	FavoriteCake string `json:"favorite_cake"`
}

type BanInfo struct {
	Email  string `json:"email"`
	Reason string `json:"reason"`
}

func (u *InMemoryUserStorage) Add(usr string, data User) error {
	if len(u.storage[usr].Email) != 0 {
		return errors.New("login is already present")
	}
	u.storage[usr] = data
	return nil
}

func (u *InMemoryUserStorage) Get(usr string) (User, error) {
	return u.storage[usr], nil
}
func (u *InMemoryUserStorage) Update(usr string, data User) error {
	if len(u.storage[usr].Email) == 0 {
		return errors.New("there is no such user to update")
	}
	delete(u.storage, usr)
	u.storage[data.Email] = data
	return nil
}
func (u *InMemoryUserStorage) Delete(usr string) (User, error) {
	if len(u.storage[usr].Email) != 0 {
		return User{}, errors.New("there is no such user to delete")
	}
	user := u.storage[usr]
	delete(u.storage, usr)
	return user, nil
}

func IsAlpha(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}

func validateEmail(str string) error {
	_, err := mail.ParseAddress(str)
	if err != nil {
		return errors.New("email is incorrect")
	}
	return nil
}
func validateCake(str string) error {
	if len(str) == 0 {
		return errors.New("favorite cake should not be empty")
	}
	if !IsAlpha(str) {
		return errors.New("favorite cake should be only alphabetic")
	}
	return nil
}
func validatePass(str string) error {
	if len(str) < 8 {
		return errors.New("password should be at least 8 symbols")
	}
	return nil
}

func validateRegisterParams(p *UserRegisterParams) error {

	if err := validateEmail(p.Email); err != nil {
		return err
	}
	if err := validateCake(p.FavoriteCake); err != nil {
		return err
	}
	if err := validatePass(p.Password); err != nil {
		return err
	}
	return nil
}

func (u *UserService) Register(w http.ResponseWriter, r *http.Request) {
	params := &UserRegisterParams{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		handleError(errors.New("could not read params"), w)
		return
	}
	if err := validateRegisterParams(params); err != nil {
		handleError(err, w)
		return
	}
	passwordDigest := md5.New().Sum([]byte(params.Password))
	newUser := User{
		Email:          params.Email,
		PasswordDigest: string(passwordDigest),
		FavoriteCake:   params.FavoriteCake,
		Role:           "user",
	}
	err = u.repository.Add(params.Email, newUser)
	if err != nil {
		handleError(err, w)
		return
	}
	Publish(params.Email + " is successfully registered")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("registered"))
}

func (u *UserService) UpdateCake(w http.ResponseWriter, r *http.Request, usr User) {
	resp, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	cake := string(resp)
	if err := validateCake(cake); err != nil {
		handleError(err, w)
		return
	}
	newData := User{
		Email:          usr.Email,
		PasswordDigest: usr.PasswordDigest,
		FavoriteCake:   cake,
		Role:           usr.Role,
	}
	err = u.repository.Update(usr.Email, newData)
	if err != nil {
		handleError(err, w)
		return
	}
	Publish(usr.Email + " has successfully updated his cake")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Cake updated"))
}

func (u *UserService) UpdateEmail(w http.ResponseWriter, r *http.Request, usr User) {
	resp, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	email := string(resp)
	if err := validateEmail(email); err != nil {
		handleError(err, w)
		return
	}
	newData := User{
		Email:          email,
		PasswordDigest: usr.PasswordDigest,
		FavoriteCake:   usr.FavoriteCake,
		Role:           usr.Role,
	}
	err = u.repository.Update(usr.Email, newData)
	if err != nil {
		handleError(err, w)
		return
	}
	Publish(usr.Email + " has successfully updated his email")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Email updated"))
}

func (u *UserService) UpdatePassword(w http.ResponseWriter, r *http.Request, usr User) {
	resp, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	pass := string(resp)
	if err := validatePass(pass); err != nil {
		handleError(err, w)
		return
	}
	passwordDigest := md5.New().Sum([]byte(pass))
	newData := User{
		Email:          usr.Email,
		PasswordDigest: string(passwordDigest),
		FavoriteCake:   usr.FavoriteCake,
		Role:           usr.Role,
	}
	err = u.repository.Update(usr.Email, newData)
	if err != nil {
		handleError(err, w)
		return
	}
	Publish(usr.Email + " has successfully updated his password")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Password updated"))
}

func (u *UserService) Promote(w http.ResponseWriter, r *http.Request, admin User) {
	resp, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	email := string(resp)
	if err := validateEmail(email); err != nil {
		handleError(err, w)
		return
	}
	user, err := u.repository.Get(email)
	if err != nil {
		panic(err)
	}
	if user.Role != "user" {
		panic("user is already admin!")
	}

	newData := User{
		Email:          user.Email,
		PasswordDigest: user.PasswordDigest,
		FavoriteCake:   user.FavoriteCake,
		Role:           "admin",
	}
	err = u.repository.Update(user.Email, newData)
	if err != nil {
		handleError(err, w)
		return
	}
	Publish(admin.Email + " has successfully promoted " + user.Email)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User is successfully promoted"))
}

func (u *UserService) Fire(w http.ResponseWriter, r *http.Request, admin User) {
	resp, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	email := string(resp)
	if err := validateEmail(email); err != nil {
		handleError(err, w)
		return
	}
	user, err := u.repository.Get(email)
	if err != nil {
		panic(err)
	}
	if user.Role != "admin" {
		panic("user is not an admin!")
	}

	newData := User{
		Email:          user.Email,
		PasswordDigest: user.PasswordDigest,
		FavoriteCake:   user.FavoriteCake,
		Role:           "user",
	}
	err = u.repository.Update(user.Email, newData)
	if err != nil {
		handleError(err, w)
		return
	}
	Publish(admin.Email + " has successfully fired " + user.Email)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User is successfully fired"))
}

func (j *JWTService) BanUser(w http.ResponseWriter, r *http.Request, admin User, u *UserService) {
	params := &BanInfo{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		handleError(errors.New("could not read params"), w)
		return
	}
	userToBan, err := u.repository.Get(params.Email)
	if err != nil {
		handleError(errors.New("could not read params"), w)
		return
	}
	if admin.Role == "superadmin" {
		j.blacklist[params.Email] = "banned. reason: " + params.Reason
	} else if admin.Role == "admin" && userToBan.Role == "user" {
		j.blacklist[params.Email] = "banned. reason: " + params.Reason
	} else {
		handleError(errors.New("user has not enough rights"), w)
		return
	}
	Publish(admin.Email + " has successfully banned " + userToBan.Email)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User is successfully banned"))
}

func (j *JWTService) UnbanUser(w http.ResponseWriter, r *http.Request, admin User, u *UserService) {
	params := &BanInfo{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		handleError(errors.New("could not read params"), w)
		return
	}
	userToUnban, err := u.repository.Get(params.Email)
	if err != nil {
		handleError(errors.New("could not read params"), w)
		return
	}
	if admin.Role == "superadmin" {
		delete(j.blacklist, params.Email)
	} else if admin.Role == "admin" && userToUnban.Role == "user" {
		delete(j.blacklist, params.Email)
	} else {
		handleError(errors.New("user has not enough rights"), w)
		return
	}
	Publish(admin.Email + " has successfully unbanned " + userToUnban.Email)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User is successfully unbanned"))
}

func getCakeHandler(w http.ResponseWriter, r *http.Request, u User) {
	w.Write([]byte(u.FavoriteCake))
	Publish(u.Email + " discovered that his favorite cake is " + u.FavoriteCake)
}

func getInfoHandler(w http.ResponseWriter, r *http.Request, u User) {
	w.Write([]byte(u.Email + ", " + u.FavoriteCake + ", " + u.Role))
	Publish(u.Email + " discovered some info about him: " + u.FavoriteCake + ", " + u.Role)
}

func handleError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	w.Write([]byte(err.Error()))
}
