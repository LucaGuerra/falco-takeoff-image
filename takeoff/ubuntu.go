package main

import (
	"fmt"
	cp "github.com/otiai10/copy"
	"log"
	"path/filepath"
)

func BuildUbuntu(p *Platform, driverDir string) {
	pullstring := fmt.Sprintf("docker.io/library/ubuntu:%s", p.Ubuntu.VersionID)
	bundle := filepath.Join("container_workspaces", "ubuntu")
	target := filepath.Join(bundle, "rootfs")

	err := downloadExtractImage(pullstring, target)
	if err != nil {
		log.Fatalf("[-] could not download and extract %s in %s: %s", pullstring, target, err)
	}

	err = PrepareRootFS(target)
	if err != nil {
		log.Fatalf("[-] could not prepare rootfs %s", err)
	}

	err = cp.Copy(driverDir, filepath.Join(target, "usr/src/falco-driver"))
	if err != nil {
		log.Fatalf("[-] could not copy %s: %s", driverDir, err)
	}

	Must(runInContainer(bundle, "/", []string{"apt-get", "update"}))
	Must(runInContainer(bundle, "/", []string{"apt-get", "-y", "install", "gcc", "make"}))
	Must(runInContainer(bundle, "/", []string{"apt-get", "-y", "install", "linux-headers-" + p.KernelRelease}))
	Must(runInContainer(bundle, "/usr/src/falco-driver", []string{"make", "KERNELDIR=/usr/src/linux-headers-" + p.KernelRelease}))
}
