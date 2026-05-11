package exchanges

import "errors"

var ErrNoExchangesFound = errors.New("user not found")
var ErrExchangeAlreadyExists = errors.New("exchange already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
