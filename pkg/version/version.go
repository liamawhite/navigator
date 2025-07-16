package version

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

var (
	version   = "dev"
	commit    = "unknown"
	date      = "unknown"
	goVersion = runtime.Version()
)

type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"goVersion"`
}

func Get() string {
	return version
}

func GetCommit() string {
	return commit
}

func GetDate() string {
	return date
}

func GetGoVersion() string {
	return goVersion
}

func GetInfo() Info {
	return Info{
		Version:   version,
		Commit:    commit,
		Date:      date,
		GoVersion: goVersion,
	}
}

func (i Info) String() string {
	var dateStr string
	if i.Date != "unknown" {
		if t, err := time.Parse(time.RFC3339, i.Date); err == nil {
			dateStr = t.Format("2006-01-02 15:04:05 UTC")
		} else {
			dateStr = i.Date
		}
	} else {
		dateStr = i.Date
	}

	return fmt.Sprintf("navigator version %s\ncommit: %s\nbuilt: %s\ngo: %s",
		i.Version, i.Commit, dateStr, i.GoVersion)
}

func (i Info) JSON() (string, error) {
	data, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
