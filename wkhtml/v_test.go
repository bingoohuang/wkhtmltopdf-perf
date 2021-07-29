package wkhtml

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExec(t *testing.T) {
	options := ExecOptions{Timeout: 10 * time.Second}
	a, _ := os.ReadFile("testdata/a.html")
	result, err := options.Exec(a, wkhtmltopdf, "--quiet", "-", "-")
	assert.Nil(t, err)
	assert.True(t, len(result) >= 1000)
}
