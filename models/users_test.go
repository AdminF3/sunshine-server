package models

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

var validUser = User{
	Name:     "John Doe",
	Email:    "john@doe.com",
	Password: "foo",
}

func TestSetPassword(t *testing.T) {

	password := "foo"
	u := User{}
	u.SetPassword(password)

	if u.Password == password {
		t.Fatalf("SetPassword should produce hash; got raw value")
	}

	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		t.Errorf("Expected nil bcrypt.CompareHashAndPassword; got %s", err)
	}

	password += "a"
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil {
		t.Errorf("Got false-positive from bcrypt.CompareHashAndPassword")
	}
}
