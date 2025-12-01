package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ssh-messer/internal/tui"
	"ssh-messer/pkg"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/rs/zerolog"
)

const (
	MesserHomeDir      = "~/.ssh_messer"
	MesserLogsDir      = "~/.ssh_messer/logs"
	MesserLogPrefix    = "ssh-messer#"
	MesserLogExtension = ".log"
	MesserLogsKeepDays = 7
)

func cleanUpLogFiles() {
	pkg.Logger.Trace().Msg("[main::handleLogFileCleanup] Cleaning up old log files...")
	result := pkg.DeleteFilesByPatternOverDays(MesserLogsDir, MesserLogPrefix+"*"+MesserLogExtension, MesserLogsKeepDays)
	pkg.Logger.Debug().Msgf("[main::handleLogFileCleanup] Cleaned up old log files: %+v", result)
	if len(result.Errors) > 0 {
		for _, errMsg := range result.Errors {
			pkg.Logger.Warn().Str("error", errMsg).Msg("Failed to clean up old log files")
		}
	}
	if len(result.DeletedFiles) > 0 {
		for _, deletedFile := range result.DeletedFiles {
			fmt.Printf("Deleted old log file over %d days: %s \n", MesserLogsKeepDays, deletedFile)
		}
	}
}

func cleanUpTerminal() {
	fmt.Fprint(os.Stdout, "\033[2J")
	fmt.Fprint(os.Stdout, "\033[H")
	fmt.Fprint(os.Stdout, "\033[?25h")
	fmt.Fprint(os.Stdout, "\033[0m")
}

func checkCreateMesserHomeFolder() string {
	logFileName := MesserLogPrefix + time.Now().Format("060102150405") + MesserLogExtension
	logFilePath := filepath.Join(MesserLogsDir, logFileName)

	if err := pkg.CreateFolderIfNotExist(MesserHomeDir); err != nil {
		pkg.HandleTerminalError(err, "")
	}
	if err := pkg.CreateFolderIfNotExist(MesserLogsDir); err != nil {
		pkg.HandleTerminalError(err, "")
	}

	pkg.InitFileLogger(MesserLogsDir, logFileName, zerolog.TraceLevel)
	return logFilePath
}

func main() {
	logFilePath := checkCreateMesserHomeFolder()
	pkg.Logger.Trace().Msg("[main::main] Check & Create messer home & log folder")

	model := tui.New()

	p := tea.NewProgram(model)
	go model.Subscribe(p)
	defer func() {
		cleanUpLogFiles()
		model.Cleanup()
		cleanUpTerminal()
	}()

	pkg.Logger.Trace().Msg("[main::main] Init Messer with subscribers")

	if _, err := p.Run(); err != nil {
		pkg.Logger.Error().Err(err).Msg("[main::main] Messer exited with error")
		pkg.HandleTerminalError(err, logFilePath)
	}
}
