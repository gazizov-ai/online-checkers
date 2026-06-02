package repository

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrUsernameAlreadyExists = errors.New("username already exists")
var ErrEmailAlreadyExists = errors.New("email already exists")
