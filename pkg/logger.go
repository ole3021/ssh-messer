package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

var stdConfig = zerolog.ConsoleWriter{
	Out:        os.Stdout,
	TimeFormat: "15:04:05",
	NoColor:    true, // 禁用颜色，因为我们要用纯文本
	PartsOrder: []string{"time", "level", "message"},
	FormatLevel: func(i interface{}) string {
		if ll, ok := i.(string); ok {
			switch ll {
			case "trace":
				return "🔍 TRACE"
			case "debug":
				return "🐛 DEBUG"
			case "info":
				return "📢  INFO"
			case "warn":
				return "⚠️  WARN"
			case "error":
				return "❌ ERROR"
			case "fatal":
				return "💀 FATAL"
			case "panic":
				return "🚨 PANIC"
			default:
				return strings.ToUpper(ll)
			}
		}
		return strings.ToUpper(fmt.Sprintf("%s", i))
	},
	FormatMessage: func(i interface{}) string {
		if i == nil {
			return ""
		}
		return fmt.Sprintf("%s", i)
	},
	FormatFieldName: func(i interface{}) string {
		return fmt.Sprintf("%s=", i)
	},
	FormatFieldValue: func(i interface{}) string {
		return fmt.Sprintf("%v", i)
	},
}

func InitFileLogger(path string, fileName string, level zerolog.Level) {
	expandedPath, err := ExpandHomeDir(path)
	if err != nil {
		panic("Failed to expand home directory: " + err.Error())
	}

	// Check if directory exists
	_, err = os.Stat(expandedPath)
	if err != nil {
		panic("Failed to open directory: " + path + " - " + err.Error())
	}

	// Create log file
	logFileFileName := filepath.Join(expandedPath, fileName)
	file, err := os.OpenFile(logFileFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("Failed to create log file: " + err.Error())
	}

	zerolog.SetGlobalLevel(level)

	fileConfig := stdConfig
	fileConfig.Out = file
	Logger = zerolog.New(fileConfig).With().Timestamp().Logger()
}

func InitLogger(level zerolog.Level) {
	zerolog.SetGlobalLevel(level)
	Logger = zerolog.New(stdConfig).With().Timestamp().Logger()
}
