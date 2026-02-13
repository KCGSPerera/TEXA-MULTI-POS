package logger

import (
	"os"
	"sync"

	"github.com/rs/zerolog"
)

var (
	log  *zerolog.Logger
	once sync.Once
)

func Init() {
	once.Do(func() {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		l := zerolog.New(os.Stdout).With().Timestamp().Logger()
		log = &l
	})
}

func L() *zerolog.Logger {
	Init()
	return log
}
