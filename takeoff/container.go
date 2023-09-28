package main

import (
	"encoding/json"
	"fmt"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

func generateConfig(cwd string, args []string) specs.Spec {
	capabilities := []string{
		"CAP_AUDIT_WRITE",
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_FSETID",
		"CAP_KILL",
		"CAP_MKNOD",
		"CAP_NET_BIND_SERVICE",
		"CAP_NET_RAW",
		"CAP_SETFCAP",
		"CAP_SETGID",
		"CAP_SETPCAP",
		"CAP_SETUID",
		"CAP_SYS_CHROOT",
		"CAP_AUDIT_WRITE",
		"CAP_KILL",
		"CAP_NET_BIND_SERVICE",
		"CAP_SYS_ADMIN",
	}
	res := specs.Spec{
		Version: "1.0.2-dev",
		Process: &specs.Process{
			Terminal: true,
			User: specs.User{
				UID: 0,
				GID: 0,
			},
			Args: args, // change this
			Env: []string{ // note this should be taken from image manifest/meta
				"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
				"TERM=xterm",
			},
			Cwd: cwd,
			Capabilities: &specs.LinuxCapabilities{
				Bounding:  capabilities,
				Effective: capabilities,
				Permitted: capabilities,
				Ambient:   capabilities,
			},
			Rlimits: []specs.POSIXRlimit{
				{
					Type: "RLIMIT_NOFILE",
					Hard: 1024,
					Soft: 1024,
				},
			},
			NoNewPrivileges: true,
		},
		Root: &specs.Root{
			Path:     "rootfs",
			Readonly: false,
		},
		Hostname: "runc",
		Mounts: []specs.Mount{
			{
				Destination: "/proc",
				Type:        "proc",
				Source:      "proc",
			},
			{
				Destination: "/dev",
				Type:        "tmpfs",
				Source:      "tmpfs",
				Options: []string{
					"nosuid",
					"strictatime",
					"mode=755",
					"size=65536k",
				},
			},
			{
				Destination: "/dev/pts",
				Type:        "devpts",
				Source:      "devpts",
				Options: []string{
					"nosuid",
					"noexec",
					"newinstance",
					"ptmxmode=0666",
					"mode=0620",
					"gid=5",
				},
			},
			{
				Destination: "/dev/shm",
				Type:        "tmpfs",
				Source:      "shm",
				Options: []string{
					"nosuid",
					"noexec",
					"nodev",
					"mode=1777",
					"size=65536k",
				},
			},
			{
				Destination: "/dev/mqueue",
				Type:        "mqueue",
				Source:      "mqueue",
				Options: []string{
					"nosuid",
					"noexec",
					"nodev",
				},
			},
			{
				Destination: "/sys/fs/cgroup",
				Type:        "cgroup",
				Source:      "cgroup",
				Options: []string{
					"nosuid",
					"noexec",
					"nodev",
					"relatime",
					"ro",
				},
			},
		},
		Linux: &specs.Linux{
			Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{
					{
						Allow:  false,
						Access: "rwm",
					},
				},
			},
			Namespaces: []specs.LinuxNamespace{
				{Type: "pid"},
				{Type: "ipc"},
				{Type: "uts"},
				{Type: "mount"},
				// {Type: "cgroup"},
			},
			MaskedPaths: []string{
				"/proc/acpi",
				"/proc/asound",
				"/proc/kcore",
				"/proc/keys",
				"/proc/latency_stats",
				"/proc/timer_list",
				"/proc/timer_stats",
				"/proc/sched_debug",
				"/sys/firmware",
				"/proc/scsi",
			},
			ReadonlyPaths: []string{
				"/proc/bus",
				"/proc/fs",
				"/proc/irq",
				"/proc/sys",
				"/proc/sysrq-trigger",
			},
		},
	}
	return res
}

func runInContainer(bundleDirectory string, cwd string, args []string) error {
	config := generateConfig(cwd, args)
	configJson, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(bundleDirectory, "config.json"), configJson, 0755)
	if err != nil {
		return err
	}

	log.Printf("[*] running in container (cwd: %s) %v", cwd, args)

	cmd := exec.Command("runc", "run", "mycontainername")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = bundleDirectory

	absoluteCwd := path.Join(bundleDirectory, "rootfs", cwd)
	err = os.MkdirAll(absoluteCwd, 0770)
	if err != nil {
		return fmt.Errorf("could not create cwd %s: %s", absoluteCwd, err)
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
