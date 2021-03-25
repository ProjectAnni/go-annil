package storage

import (
	"os"
	"testing"
	"time"
)

func TestUser(t *testing.T) {
	err := Init()
	if err != nil {
		t.Errorf("failed to init: %v", err)
		t.FailNow()
	}
	if !UserExists("Admin") {
		t.Errorf("no default user")
		t.Fail()
	}
	if !CheckPassword("Admin", "12345") {
		t.Errorf("default admin password incorrect")
		t.Fail()
	}

	err = Register("TestUser2", "123456")
	if err != nil {
		t.Errorf("failed to create user: %v", err)
		t.FailNow()
	}
	if !CheckPassword("TestUser2", "123456") {
		t.Errorf("failed to check password")
		t.Fail()
	}
	if CheckPassword("TestUser2", "random") {
		t.Errorf("failed to check password")
		t.Fail()
	}
	err = ChangePassword("TestUser2", "1234567")
	if err != nil {
		t.Errorf("failed to change password")
	}
	if !CheckPassword("TestUser2", "1234567") {
		t.Errorf("failed to check password")
		t.Fail()
	}
	if CheckPassword("TestUser2", "123456") {
		t.Errorf("failed to check password")
		t.Fail()
	}

	date, err := RegisterDate("TestUser2")
	if err != nil {
		t.Errorf("failed to read register date: %v", err)
		t.Fail()
	}

	if time.Now().Unix()-date.Unix() > 1000 || time.Now().Unix()-date.Unix() < 0 {
		t.Error("wrong register date")
		t.Fail()
	}

	if AllowShare("TestUser2") {
		t.Errorf("wrong default ")
		t.Fail()
	}
	err = SetAllowShare("TestUser2", true)
	if err != nil {
		t.Errorf("failed to allow share")
		t.Fail()
	}
	if !AllowShare("TestUser2") {
		t.Errorf("wrong allow share")
		t.Fail()
	}
	err = SetAllowShare("TestUser2", false)
	if err != nil {
		t.Errorf("failed to allow share")
		t.Fail()
	}

	if IsAdmin("TestUser2") {
		t.Errorf("wrong default admin")
		t.Fail()
	}

	err = SetAdmin("TestUser2", true)
	if err != nil {
		t.Errorf("failed to set admin: %v", err)
		t.Fail()
	}
	if !IsAdmin("TestUser2") {
		t.Error("wrong admin")
		t.Fail()
	}
	err = SetAdmin("TestUser2", false)
	if err != nil {
		t.Errorf("failed to set admin: %v", err)
		t.Fail()
	}
	if IsAdmin("TestUser2") {
		t.Error("wrong admin")
		t.Fail()
	}
	err = RevokeUser("TestUser2")
	if err != nil {
		t.Errorf("failed tr revoke user: %v", err)
		t.Fail()
	}
	if UserExists("TestUser2") {
		t.Errorf("wrong user exist")
		t.Fail()
	}
	_ = os.Remove("data.db")
}
