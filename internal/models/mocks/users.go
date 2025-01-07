package mocks

import (
	"time"

	"snippetbox.gobpo2002.io/internal/models"
)

type UserModel struct{}

func (m *UserModel) Insert(name, email, password string) error {
	switch email {
	case "JC_follower@gmail.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	if email == "JC_follower@gmail.com" && password == "ILoveJesus" {
		return 1, nil
	}

	return 0, models.ErrInvalidCredentials
}

func (m *UserModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

func (m *UserModel) Get(id int) (*models.User, error) {
	switch id {
	case 1:
		return &models.User{
			ID:      1,
			Name:    "Max",
			Email:   "JCFollower@gmail.com",
			Created: time.Date(2024, 07, 14, 21, 0, 0, 0, time.UTC),
		}, nil

	default:
		return nil, models.ErrNoRecord
	}
}

func (m *UserModel) UpdatePassword(id int, currentPassword, newPassword string) error {
	if id == 1 {
		if currentPassword != "pa$$word" {
			return models.ErrInvalidCredentials
		}
		return nil
	}
	return models.ErrNoRecord
}
