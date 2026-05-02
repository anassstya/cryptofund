package auth

import "errors"

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserAlreadyExists = errors.New("user with this email already exists")
var ErrUserNotFound = errors.New("user not found")
