package tools

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"saurfang/internal/config"
	"saurfang/internal/models/datasource"
)

func UploadToOss(target int) (path, source string, err error) {
	var ossInfo datasource.SaurfangDatasources
	if err := config.DB.Debug().Raw("select * from saurfang_datasources where id = ?;", target).Scan(&ossInfo).Error; err != nil {
		return "", "", err
	}
	args := []string{
		"sync",
		"--no-update-modtime",
		"--no-update-dir-modtime",
		"--metadata",
		os.Getenv("SERVER_PACKAGE_DEST_PATH"),
		fmt.Sprintf("%s:%s%s", ossInfo.Profile, ossInfo.Bucket, ossInfo.Path),
	}
	fmt.Println(">>>", args)
	var stdErr bytes.Buffer
	cmd := exec.Command("rclone", args...)
	cmd.Stderr = &stdErr
	err = cmd.Run()
	if err != nil {
		return "", "", fmt.Errorf(stdErr.String())
	}
	return ossInfo.Path, ossInfo.Label, nil
}
