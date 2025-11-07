package pkg

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func InitLogger(mode ...string) {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	stdConfig := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
		NoColor:    false,
		PartsOrder: []string{"time", "level", "message"},
	}

	fileConfig := stdConfig
	fileConfig.NoColor = true

	logFileFileName := "ssh-messer#" + time.Now().Format("2006-01-02#15:04:05") + ".log"

	var logMode string
	if len(mode) > 0 {
		logMode = mode[0]
	} else {
		logMode = os.Getenv("MESSER_LOG")
	}

	switch logMode {
	case "both":
		file, err := os.OpenFile(logFileFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic("Failed to create log file: " + err.Error())
		}

		fileConfig.Out = file

		Logger = zerolog.New(zerolog.MultiLevelWriter(stdConfig, fileConfig)).With().Timestamp().Logger()

	case "file":
		file, err := os.OpenFile(logFileFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic("Failed to create log file: " + err.Error())
		}

		fileConfig.Out = file
		Logger = zerolog.New(fileConfig).With().Timestamp().Logger()

	default:
		Logger = zerolog.New(stdConfig).With().Timestamp().Logger()
	}
}
