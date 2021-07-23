package wkhtml

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestExec(t *testing.T) {
	options := ExecOptions{Timeout: 10 * time.Second}
	a, _ := os.ReadFile("testdata/a.html")
	result, err := options.Exec(a, "wkhtmltopdf", []string{"--quiet", "-", "-"}...)
	assert.Nil(t, err)
	assert.True(t, len(result) >= 1000)
}
