package service

import "errors"

var ErrInvalidUsername = errors.New("invalid username")
var ErrInvalidPassword = errors.New("invalid password")
var ErrUsernameAlreadyTaken = errors.New("username already taken")
var ErrEmailAlreadyTaken = errors.New("email already taken")
var ErrInvalidCredentials = errors.New("invalid username or password")
var ErrUserNotFound = errors.New("user not found")
