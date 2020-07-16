package api_client

import (
	"fmt"
	"golang.org/x/mod/semver"
	"strconv"
	"strings"
)

type SystemVersion struct {
	Major, Minor, Patch int64
	Build               string
}

func (r SystemVersion) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d.%d.%d.%s", r.Major, r.Minor, r.Patch, r.Build)), nil
}

func (r *SystemVersion) UnmarshalJSON(data []byte) error {
	value := string(data)

	// Remove surrounding quotes if present
	if value[0] == '"' {
		value = value[1:]
	}
	if i := len(value) - 1; value[i] == '"' {
		value = value[:i]
	}

	parts := strings.Split(value, ".")

	var major, minor, patch int64
	var build string

	major, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return err
	}
	minor, err = strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		return err
	}
	if len(parts) > 2 {
		patch, err = strconv.ParseInt(parts[2], 10, 32)
		if err != nil {
			return err
		}
		if len(parts) > 3 {
			build = parts[3]
		} else {
			build = "0"
		}
	} else {
		patch = 0
		build = "0"
	}

	r.Major = major
	r.Minor = minor
	r.Patch = patch
	r.Build = build

	return nil
}

func (r *SystemVersion) MajorMinorPatch() string {
	return fmt.Sprintf("%d.%d.%d", r.Major, r.Minor, r.Patch)
}

func (r *SystemVersion) SemVer() string {
	return fmt.Sprintf("v%s-%s", r.MajorMinorPatch(), r.Build)
}

func (r *SystemVersion) Compare(w interface{}) int {
	var output int
	switch v := w.(type) {
	case *SystemVersion:
		output = semver.Compare(r.SemVer(), v.SemVer())
	case string:
		output = semver.Compare(r.SemVer(), v)
	default:
		output = 1
	}
	return output
}
