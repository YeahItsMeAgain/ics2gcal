package updater

import (
	"context"
	"errors"
	"fmt"
	"ics2gcal/logger"
	"os"
	"runtime"

	"github.com/creativeprojects/go-selfupdate"
)

const VERSION = "1.0.0"

func SelfUpdate() error {
	latest, found, err := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("YeahItsMeAgain/ics2gcal"))
	if err != nil {
		return fmt.Errorf("error occurred while detecting version: %w", err)
	}
	if !found {
		return fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}

	if latest.LessOrEqual(VERSION) {
		logger.Logger.Infof("Current version (%s) is the latest", VERSION)
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return errors.New("could not locate executable path")
	}
	if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
		return fmt.Errorf("error occurred while updating binary: %w", err)
	}
	logger.Logger.Infof("Successfully updated to version %s", latest.Version())
	return nil
}
