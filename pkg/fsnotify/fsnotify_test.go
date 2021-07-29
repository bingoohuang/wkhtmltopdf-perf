package fsnotify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoFsNotify(t *testing.T) {
	err := Listen(".")
	assert.Nil(t, err)

	// select {}
}

/*
➜  wkhtmltopdf-perf git:(main) ✗ echo hello > a.txt
➜  wkhtmltopdf-perf git:(main) ✗ wkhtmltopdf -q assets/a.html a.pdf

=== RUN   TestGoFsNotify
2021/07/26 10:12:14 event: "a.txt": CREATE
2021/07/26 10:12:15 event: "a.txt": CHMOD
2021/07/26 10:12:44 event: "a.pdf": CREATE
2021/07/26 10:12:44 event: "a.pdf": WRITE
2021/07/26 10:12:44 modified file: a.pdf
2021/07/26 10:12:44 event: "a.pdf": WRITE
2021/07/26 10:12:44 modified file: a.pdf
2021/07/26 10:12:44 event: "a.pdf": WRITE
2021/07/26 10:12:44 modified file: a.pdf
2021/07/26 10:12:44 event: "a.pdf": WRITE
2021/07/26 10:12:44 modified file: a.pdf

一次 PDF 多次 WRITE 事件，无法方便地判断 PDF 是否生成完毕。
*/
