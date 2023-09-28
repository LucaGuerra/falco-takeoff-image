package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-containerregistry/pkg/crane"
	"io"
	"log"
	"os"
	"sync"
	"path/filepath"
)

func Must(err error) {
	if err != nil {
		log.Fatalf("[-] error %s", err)
	}
}

func downloadExtractImage(pullstring, target string) error {
	err := os.MkdirAll(target, os.ModePerm)
	if err != nil {
		return err // Handle error
	}

	img, err := crane.Pull(pullstring)
	if err != nil {
		return err // Handle error
	}

	log.Printf("[+] pull ok, exporting image %s\n", pullstring)

	wg := sync.WaitGroup{}
	wg.Add(1)
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		log.Printf("[+] writing container filesystem to output dir: %s", target)
		err = crane.Export(img, w)
		if err != nil {
			// TODO: Handle this error more effectively. Right now we rely on
			// error handling in the logic to extract this export in a lower
			// line, but we should probably exit early if the export encounters
			// an error, which requires watching multiple error streams.
			log.Fatalf("unable to export and flatten container filesystem: %s", err)
		}
		wg.Done()
	}()

	log.Printf("[+] extracting container filesystem to %s", target)
	if err := untar(target, r); err != nil {
		return fmt.Errorf("failed to extract tarball: %v", err)
	}
	wg.Wait()

	log.Printf("[+] image %s exported to %s\n", pullstring, target)

	return nil
}

func locateFalcoDriver() (string, error) {
	files, err := filepath.Glob("/usr/src/falco*+driver")
	if err != nil {
		return "", fmt.Errorf("could not locate driver in /usr/src/falco*+driver")
	}

	if len(files) != 1 {
		return "", fmt.Errorf("too many files match the drivers: %v", files)
	}

	return files[0], nil
}

func build() {
	log.Printf("[*] detecting os ...")
	hostRoot := os.Getenv("HOST_ROOT")
	if hostRoot == "" {
		hostRoot = "/"
	}
	p, err := detectPlatform(hostRoot)
	if err != nil {
		log.Fatalf("[-] could not detect platform %#v", err)
	}

	pj, _ := json.Marshal(p)
	platJson := string(pj)
	log.Printf("[*] detected platform: %s", platJson)

	buildPlatform(p)
}

func buildPlatform(p *Platform) {
	driverDir, err := locateFalcoDriver()
	if err != nil {
		log.Fatalf("[-] could not locate driver %s", err)
	}

	if p.Ubuntu != nil {
		BuildUbuntu(p, driverDir)
		return
	}

	if p.Fedora != nil {
		BuildFedora(p, driverDir)
		return
	}

	if p.CentOS != nil {
		BuildCentOS(p, driverDir)
		return
	}

	log.Fatalf("[-] sorry, unsupported platform")
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("[-] not enough args")
	}

	switch os.Args[1] {
	case "build":
		build()
		return

	case "build-withplatform":
		var platform Platform
		platJson := os.Args[2]

		err := json.Unmarshal([]byte(platJson), &platform)
		if err != nil {
			log.Fatalf("[-] invalid platform %s: %s", platJson, err)
		}

		buildPlatform(&platform)
		return

	case "download-image":
		var pullstring = os.Args[2]
		var target = os.Args[3]
		err := downloadExtractImage(pullstring, target)
		if err != nil {
			log.Fatalf("could not download image %s: %s", pullstring, err)
		}
		log.Printf("[+] created additional fs in %s", target)
		return

	case "detect":
		p, err := detectPlatform("/")
		if err != nil {
			log.Fatalf("%#v", err)
		}

		j, err := json.Marshal(p)
		if err != nil {
			log.Fatalf("%#v", err)
		}

		fmt.Println(string(j))
		return
	}
}
