package main

import (
	"fmt"
	cp "github.com/otiai10/copy"
	"log"
	"path/filepath"
)

func BuildFedora(p *Platform, driverDir string) {
	pullstring := fmt.Sprintf("docker.io/library/fedora:%s", p.Fedora.VersionID)
	bundle := filepath.Join("container_workspaces", "fedora")
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

	Must(runInContainer(bundle, "/", []string{"dnf", "-y", "install", "koji"}))
	Must(runInContainer(bundle, "/tmp", []string{"koji", "download-build", "--arch=x86_64", "--rpm", "kernel-devel-" + p.KernelRelease}))
	Must(runInContainer(bundle, "/tmp", []string{"dnf", "install", "-y", "kernel-devel-" + p.KernelRelease + ".rpm"}))
	Must(runInContainer(bundle, "/usr/src/falco-driver", []string{"make", "KERNELDIR=/usr/src/kernels/" + p.KernelRelease}))
}
