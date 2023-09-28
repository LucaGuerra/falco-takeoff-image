package main

import (
	"fmt"
	"github.com/docker/docker/pkg/fileutils"
	"os"
	"path/filepath"
)

func PrepareRootFS(rootfs string) error {
	var err error
	tmpDir := filepath.Join(rootfs, "tmp")
	err = os.MkdirAll(tmpDir, 0777)
	if err != nil {
		return fmt.Errorf("could not create directory %s: %s", tmpDir, err)
	}

	err = os.Chmod(tmpDir, 0777)
	if err != nil {
		return fmt.Errorf("could not chmod dir to 777 %s: %s", tmpDir, err)
	}

	_, err = fileutils.CopyFile(filepath.Join("etc", "resolv.conf"), filepath.Join(rootfs, "etc", "resolv.conf"))
	if err != nil {
		return fmt.Errorf("could not copy /etc/resolv.conf: %s", err)
	}

	return err
}
