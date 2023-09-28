package main

import (
	"bufio"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path"
	"regexp"
	"strings"
)

type Platform struct {
	KernelRelease string `json:"kernel_release"` // uname -r

	Ubuntu       *UbuntuPlatform       `json:"ubuntu,omitempty""`
	Fedora       *FedoraPlatform       `json:"fedora,omitempty"`
	CentOS       *CentOSPlatform       `json:"centos,omitempty"`
	AmazonLinux2 *AmazonLinux2Platform `json:"amazonlinux2,omitempty"`
	Bottlerocket *BottlerocketPlatform `json:"bottlerocket,omitempty"`
}

type UbuntuPlatform struct {
	VersionID string `json:"version_id"` // like "20.04"
}

type FedoraPlatform struct {
	VersionID string `json:"version_id"` // like "36"
}

type CentOSPlatform struct {
	VersionID string `json:"version_id"` // like "7"
}

type AmazonLinux2Platform struct {
}

// not implemented yet
type BottlerocketPlatform struct {
	Version string
	Variant string
	Arch    string
}

var simpleOSRvalue = regexp.MustCompile(`^([A-Za-z_]+)=([A-Za-z0-9_\.-]+)$`)
var dqOSRvalue = regexp.MustCompile(`^([A-Za-z_]+)="(.*)"$`)
var sqOSRvalue = regexp.MustCompile(`^([A-Za-z_]+)='(.*)'$`)

func parseOSRelease(osReleaseContent string) map[string]string {
	res := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(osReleaseContent))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		sm := simpleOSRvalue.FindStringSubmatch(line)
		if len(sm) > 0 {
			res[sm[1]] = sm[2]
			continue
		}

		sm = sqOSRvalue.FindStringSubmatch(line)
		if len(sm) > 0 {
			res[sm[1]] = sm[2]
			continue
		}

		sm = dqOSRvalue.FindStringSubmatch(line)
		if len(sm) > 0 {
			res[sm[1]] = sm[2]
			continue
		}
	}

	return res
}

func intToString(a [65]byte) string {
	var (
		tmp [65]byte
		i   int
	)
	for i = 0; a[i] != 0; i++ {
		tmp[i] = byte(a[i])
	}
	return string(tmp[:i])
}

// kernel crawler supports:
// AlmaLinux, AmazonLinux, AmazonLinux2, AmazonLinux2022, CentOS, Fedora, Oracle6, Oracle7, Oracle8, PhotonOS, RockyLinux, OpenSUSE, Debian, Ubuntu, Flatcar, Minikube, ArchLinux

// falco prebuilt support:
//amazonlinux
//amazonlinux2
//amazonlinux2022
//centos
//debian
//minikube
//photon
//ubuntu-aws  ubuntu-aws-5.0  ubuntu-aws-5.11  ubuntu-aws-5.13  ubuntu-aws-5.15  ubuntu-aws-5.3  ubuntu-aws-5.4  ubuntu-aws-5.8
//ubuntu-azure  ubuntu-azure-4.15  ubuntu-azure-5.11  ubuntu-azure-5.13  ubuntu-azure-5.15  ubuntu-azure-5.3  ubuntu-azure-5.4  ubuntu-azure-5.8  ubuntu-azure-cvm
//ubuntu-gcp  ubuntu-gcp-4.15  ubuntu-gcp-5.11  ubuntu-gcp-5.13  ubuntu-gcp-5.15  ubuntu-gcp-5.3  ubuntu-gcp-5.4  ubuntu-gcp-5.8
//ubuntu-generic  ubuntu-generic-4.15  ubuntu-generic-5.15  ubuntu-generic-5.4
//ubuntu-genericop  ubuntu-genericop-5.4
//ubuntu-gke
//ubuntu-hwe  ubuntu-hwe-5.0  ubuntu-hwe-5.11  ubuntu-hwe-5.13  ubuntu-hwe-5.15  ubuntu-hwe-5.4  ubuntu-hwe-5.8
//ubuntu-ibm
//ubuntu-intel-5.13  ubuntu-intel-iotg  ubuntu-intel-iotg-5.15
//ubuntu-kvm
//ubuntu-lowlatency
//ubuntu-nvidia
//ubuntu-oem  ubuntu-oem-5.10  ubuntu-oem-5.13  ubuntu-oem-5.14  ubuntu-oem-5.17  ubuntu-oem-5.6
//ubuntu-oracle  ubuntu-oracle-5.0  ubuntu-oracle-5.11  ubuntu-oracle-5.13  ubuntu-oracle-5.15  ubuntu-oracle-5.3  ubuntu-oracle-5.4  ubuntu-oracle-5.8

func detectUbuntu(osRelease map[string]string) (*UbuntuPlatform, error) {
	version := osRelease["VERSION_ID"]
	if version == "" {
		return nil, fmt.Errorf("could not locate VERSION_ID in os-release file for Ubuntu")
	}

	res := &UbuntuPlatform{VersionID: version}
	return res, nil
}

func detectFedora(osRelease map[string]string) (*FedoraPlatform, error) {
	version := osRelease["VERSION_ID"]
	if version == "" {
		return nil, fmt.Errorf("could not locate VERSION_ID in os-release file for Fedora")
	}

	res := &FedoraPlatform{VersionID: version}
	return res, nil
}

func detectAmazonLinux(osRelease map[string]string) (*AmazonLinux2Platform, error) {
	version := osRelease["VERSION_ID"]
	if version != "2" {
		return nil, fmt.Errorf("only Amazon Linux 2 is supported for now")
	}

	res := &AmazonLinux2Platform{}
	return res, nil
}

func detectPlatform(hostRoot string) (*Platform, error) {
	ret := &Platform{}

	var name unix.Utsname
	err := unix.Uname(&name)
	if err != nil {
		return nil, err
	}

	ret.KernelRelease = intToString(name.Release)

	osReleasePath := path.Join(hostRoot, "etc", "os-release")
	buf, err := os.ReadFile(osReleasePath)
	if err != nil {
		return nil, fmt.Errorf("could not read %s", osReleasePath)
	}

	s := string(buf)
	osRelease := parseOSRelease(s)

	id := strings.ToLower(osRelease["ID"])

	switch id {
	case "ubuntu":
		ub, err := detectUbuntu(osRelease)
		if err != nil {
			return nil, err
		}
		ret.Ubuntu = ub
	case "fedora":
		fe, err := detectFedora(osRelease)
		if err != nil {
			return nil, err
		}
		ret.Fedora = fe
	case "amzn":
		amzn2, err := detectAmazonLinux(osRelease)
		if err != nil {
			return nil, err
		}
		ret.AmazonLinux2 = amzn2
	}
	// 2. read HOST_ROOT/etc/os-release
	// ^ if that is not found it could be a distro that doesn't have it or something ancient like RHEL 6
	// 3. call relevant platform-specific code

	return ret, nil
}
