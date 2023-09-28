package main

import (
	"reflect"
	"testing"
)

func Test_parseOSRelease(t *testing.T) {
	tests := []struct {
		name             string
		osReleaseContent string
		want             map[string]string
	}{
		{"ubuntu basic", `NAME="Ubuntu"
VERSION="20.04.5 LTS (Focal Fossa)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 20.04.5 LTS"
VERSION_ID="20.04"
`,
			map[string]string{"NAME": "Ubuntu", "VERSION": "20.04.5 LTS (Focal Fossa)", "ID": "ubuntu", "ID_LIKE": "debian", "PRETTY_NAME": "Ubuntu 20.04.5 LTS", "VERSION_ID": "20.04"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseOSRelease(tt.osReleaseContent); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseOSRelease() = %v, want %v", got, tt.want)
			}
		})
	}
}
