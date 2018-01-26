package version

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var currentTestTime = time.Now()

func newAppVersion() AppVersionInfo {
	return AppVersionInfo{
		Name:         "application",
		Version:      "1.2.3",
		Revision:     "a1b2c3d4",
		Branch:       "master",
		GOVersion:    "1.9.2",
		BuiltAt:      currentTestTime,
		OS:           "linux",
		Architecture: "amd64",
	}
}

func TestAppVersionInfo_UserAgent(t *testing.T) {
	v := newAppVersion()
	assert.Equal(t, "application 1.2.3 (master; 1.9.2; linux/amd64)", v.UserAgent())
}

func TestAppVersionInfo_Line(t *testing.T) {
	v := newAppVersion()
	assert.Equal(t, "application 1.2.3 (a1b2c3d4)", v.Line())
}

func TestAppVersionInfo_ShortLine(t *testing.T) {
	v := newAppVersion()
	assert.Equal(t, "1.2.3 (a1b2c3d4)", v.ShortLine())
}

func TestAppVersionInfo_Extended(t *testing.T) {
	v := newAppVersion()
	info := v.Extended()

	assert.Regexp(t, regexp.MustCompile("Version:\\s+1.2.3"), info)
	assert.Regexp(t, regexp.MustCompile("Git revision:\\s+a1b2c3d4"), info)
	assert.Regexp(t, regexp.MustCompile("Git branch:\\s+master"), info)
	assert.Regexp(t, regexp.MustCompile("GO version:\\s+1.9.2"), info)
	ts := strings.Replace(currentTestTime.Format(time.RFC3339), "+", ".", -1)
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("Built:\\s+%s", ts)), info)
	assert.Regexp(t, regexp.MustCompile("OS/Arch:\\s+linux/amd64"), info)
}
