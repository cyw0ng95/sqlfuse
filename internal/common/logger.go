package common

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func InitLogger() {
	zerolog.TimeFieldFormat = "01/02 15:04:05"
	zerolog.LevelFieldMarshalFunc = func(l zerolog.Level) string {
		return strings.ToUpper(l.String())
	}
	Logger = zerolog.New(os.Stderr).
		With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "01/02 15:04:05",
			FormatLevel: func(i interface{}) string {
				return "[" + strings.ToUpper(i.(string)) + "]"
			},
			FormatMessage: func(i interface{}) string {
				return i.(string)
			},
			FormatFieldName:  func(i interface{}) string { return "" },
			FormatFieldValue: func(i interface{}) string { return "" },
		})
}
