package backupUtils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ALiwoto/RestorerRobot/src/core/wotoConfig"
	"github.com/ALiwoto/RestorerRobot/src/core/wotoStyle"
	"github.com/ALiwoto/ssg/ssg"
	fErrors "github.com/go-faster/errors"
)

// BackupDatabase backups a database using its url to the specified filename
// and the type.
// for seeing which types are currently supported, refer to DatabaseBackupType.IsInvalidType
// method.
func BackupDatabase(url, filename string, bType wotoConfig.DatabaseBackupType) error {
	if bType.IsInvalidType() {
		return errors.New("unsupported backup type")
	}

	backupCommand := wotoConfig.GetPgDumpCommand() + " -d " + url + " -x --no-owner "
	// #TODO: convert this ugly if-else statement to a cute switch in future
	if bType == wotoConfig.BackupTypeSQL {
		backupCommand += ">> " + filename
	} else if bType == wotoConfig.BackupTypeDump {
		backupCommand += "> " + filename
	}

	result := ssg.RunCommand(backupCommand)
	if result.Error != nil {
		return fErrors.Wrap(result.Error, result.Stderr+result.Stdout)
	}

	return nil
}

// ZipSource transfers a source file/directory to the destination.
func ZipSource(source, target string) error {
    // Get file extension
    ext := filepath.Ext(source)
    
    // Ensure target has the same extension
    if filepath.Ext(target) != ext {
        target = target + ext
    }

    // Create destination directory if needed
    if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
        return err
    }

    return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Open source file
        sourceFile, err := os.Open(path)
        if err != nil {
            return err
        }
        defer sourceFile.Close()

        // Create target file
        targetFile, err := os.Create(target)
        if err != nil {
            return err
        }
        defer targetFile.Close()

        // Direct copy without compression
        _, err = io.Copy(targetFile, sourceFile)
        if err != nil {
            return err
        }

        // Copy permissions
        return os.Chmod(target, info.Mode())
    })
}

// GenerateCaption generates caption for the backup using the specified options.
func GenerateCaption(opts *GenerateCaptionOptions) wotoStyle.WStyle {
	md := wotoStyle.GetBold("Config name: ").Mono(opts.ConfigName)
	md.Bold("\nType: ").Mono(opts.BackupInitType)
	md.Bold("\nInitiated by: ").Mono(opts.InitiatedBy)
	if opts.UserId != 0 {
		md.Bold("\nID: ").Mono(ssg.ToBase10(opts.UserId))
	}

	if !opts.DateTime.IsZero() {
		// format should be like: Wed-01-06-2022 11:39 AM
		md.Bold("\nDate Time: ").Mono(opts.DateTime.Format("Mon-01-02-2006 03:04 PM"))
	}

	if opts.FileSize != "" {
		startingTitle := "File"
		if opts.BackupFormat != "" {
			startingTitle = opts.BackupFormat
		}

		md.Bold("\n" + startingTitle + " size: ").Mono(opts.FileSize)
	}

	return md
}

func FormatFileSize(size int64) string {
	var sizeSuffix string
	var sizeValue float64

	if size > 1024*1024*1024 {
		sizeSuffix = "GB"
		sizeValue = float64(size) / 1024 / 1024 / 1024
	} else if size > 1024*1024 {
		sizeSuffix = "MB"
		sizeValue = float64(size) / 1024 / 1024
	} else if size > 1024 {
		sizeSuffix = "KB"
		sizeValue = float64(size) / 1024
	} else {
		sizeSuffix = "B"
		sizeValue = float64(size)
	}

	return fmt.Sprintf("%.4f", sizeValue) + " " + sizeSuffix
}

// GenerateFileNameFromOrigin creates a filename from the origin,
// in "VALUE-backup-2022-5-16--15-30-59" format.
func GenerateFileNameFromValue(value string) string {
	return value + "-backup-" + time.Now().Format("2006-01-02--15-04-05")
}
