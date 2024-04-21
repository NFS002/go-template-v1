package utils

import (
	"github.com/rs/zerolog/log"
)

func InfoLog(message string) {
	log.Info().Msg(message)
}

func ErrorLog(message string, err error) {
	log.Error().AnErr("error", err).Msg(message)
}

func PanicLog(message string, err error) {
	log.Panic().AnErr("error", err).Msg(message)
}
