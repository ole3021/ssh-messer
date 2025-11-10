package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

// CleanupResult 日志清理结果
type CleanupResult struct {
	FoundCount    int      // 找到的日志文件数量
	DeletedCount  int      // 删除的文件数量
	DeletedFiles  []string // 被删除的文件名列表
	Errors        []string // 错误信息列表
}

// cleanOldLogFiles 清理7天之前的日志文件
func cleanOldLogFiles() CleanupResult {
	result := CleanupResult{
		DeletedFiles: make([]string, 0),
		Errors:       make([]string, 0),
	}

	// 查找所有匹配的日志文件
	pattern := "ssh-messer#*.log"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		result.Errors = append(result.Errors, "无法查找日志文件: "+err.Error())
		return result
	}

	result.FoundCount = len(matches)
	cutoffTime := time.Now().AddDate(0, 0, -7) // 7天前

	for _, filePath := range matches {
		// 从文件名中提取日期
		// 文件名格式：ssh-messer#2006-01-02#15:04:05.log
		fileName := filepath.Base(filePath)
		
		// 移除前缀 "ssh-messer#" 和后缀 ".log"
		if !strings.HasPrefix(fileName, "ssh-messer#") || !strings.HasSuffix(fileName, ".log") {
			continue
		}
		
		// 提取日期部分：ssh-messer#2006-01-02#15:04:05.log -> 2006-01-02#15:04:05
		dateTimeStr := strings.TrimPrefix(fileName, "ssh-messer#")
		dateTimeStr = strings.TrimSuffix(dateTimeStr, ".log")
		
		// 解析日期时间
		// 格式：2006-01-02#15:04:05
		dateTimeStr = strings.ReplaceAll(dateTimeStr, "#", " ")
		fileTime, err := time.Parse("2006-01-02 15:04:05", dateTimeStr)
		if err != nil {
			// 如果无法解析日期，记录错误但继续处理
			result.Errors = append(result.Errors, "无法解析文件日期: "+fileName+", 错误: "+err.Error())
			continue
		}

		// 如果文件时间早于7天前，删除该文件
		if fileTime.Before(cutoffTime) {
			if err := os.Remove(filePath); err != nil {
				result.Errors = append(result.Errors, "删除文件失败: "+fileName+", 错误: "+err.Error())
			} else {
				result.DeletedCount++
				result.DeletedFiles = append(result.DeletedFiles, fileName)
			}
		}
	}

	return result
}

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

	var cleanupResult CleanupResult
	switch logMode {
	case "both":
		// 在创建新日志文件之前清理旧日志文件
		cleanupResult = cleanOldLogFiles()
		
		file, err := os.OpenFile(logFileFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic("Failed to create log file: " + err.Error())
		}

		fileConfig.Out = file

		Logger = zerolog.New(zerolog.MultiLevelWriter(stdConfig, fileConfig)).With().Timestamp().Logger()

	case "file":
		// 在创建新日志文件之前清理旧日志文件
		cleanupResult = cleanOldLogFiles()
		
		file, err := os.OpenFile(logFileFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic("Failed to create log file: " + err.Error())
		}

		fileConfig.Out = file
		Logger = zerolog.New(fileConfig).With().Timestamp().Logger()

	default:
		Logger = zerolog.New(stdConfig).With().Timestamp().Logger()
	}

	// Logger 初始化后记录清理结果
	if logMode == "both" || logMode == "file" {
		Logger.Debug().Msg("[Logger] 开始清理旧日志文件")
		Logger.Debug().Int("found_count", cleanupResult.FoundCount).Msg("[Logger] 找到的日志文件数量")

		if cleanupResult.DeletedCount > 0 {
			Logger.Info().Int("deleted_count", cleanupResult.DeletedCount).Msg("[Logger] 删除的旧日志文件数量")
			for _, deletedFile := range cleanupResult.DeletedFiles {
				Logger.Debug().Str("file", deletedFile).Msg("[Logger] 已删除旧日志文件")
			}
		}

		if len(cleanupResult.Errors) > 0 {
			for _, errMsg := range cleanupResult.Errors {
				Logger.Warn().Str("error", errMsg).Msg("[Logger] 清理日志文件时发生错误")
			}
		}
	}
}
