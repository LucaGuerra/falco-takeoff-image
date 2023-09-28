package main

import (
	"fmt"
	cp "github.com/otiai10/copy"
	"log"
	"path/filepath"
)

func BuildCentOS(p *Platform, driverDir string) {
	pullstring := fmt.Sprintf("docker.io/library/centos:%s", p.CentOS.VersionID)
	bundle := filepath.Join("container_workspaces", "centos")
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

	Must(runInContainer(bundle, "/", []string{"yum", "-y", "install", "make", "gcc"}))
	Must(runInContainer(bundle, "/", []string{"yum", "-y", "install", "kernel-devel-uname-r = " + p.KernelRelease, "--enablerepo=C*"}))
	Must(runInContainer(bundle, "/usr/src/falco-driver", []string{"make", "KERNELDIR=/usr/src/kernels/" + p.KernelRelease}))
}
