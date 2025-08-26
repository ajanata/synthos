package controller

import (
	"errors"
)

var ErrInvalidToken = errors.New("invalid token")
var ErrUnableToStartSynth = errors.New("unable to start synth")
