# wkhtmltopdf-perf

performance prof for wkhtmltopdf.

[wkhtmltopdf](https://github.com/wkhtmltopdf/wkhtmltopdf) converts HTML to PDF using Webkit (QtWebKit), which is the
browser engine that is used to render HTML and javascript - Chrome uses that engine too.

## Versions

版本 | 进程预热 | 进程重用 | html参数 | PDF 落盘 | wkhtmltopdf
---|---|----|----|----|----
-| - | - | 网址 | Y |  `wkhtmltopdf --quiet http://a.b.c/a.html a.pdf`
0| - | - | 网址 | N（stdout) | `wkhtmltopdf --quiet http://a.b.c/a.html - \| cat`
1| - | - | 内容 (stdin) | N（stdout) | `wkhtmltopdf --quiet - - \| cat`
1p| 预热 | -| 内容 (stdin) | N（stdout) | `wkhtmltopdf --quiet - - \| cat`
2| 预热 | 重用 | 网址 | Y |  `wkhtmltopdf --read-args-from-stdin`
2p| 预热 | 重用 | 网址 | N (fuse) | `wkhtmltopdf --read-args-from-stdin`

## Install

1. Get the latest downloads from [here](https://wkhtmltopdf.org/downloads.html).
1. Linux install
    - centos: `yum localinstall wkhtmltox.rpm`
    - fedora: `dnf localinstall wkhtmltox.rpm`
    - Dibian: `sudo dpkg -i wkhtmltox.deb; sudo ldconfig`
    - [more](https://github.com/adrg/go-wkhtmltopdf/wiki/Install-on-Linux)

## 5 minutes to start

1. quiet mode : `wkhtmltopdf -q a.html a.pdf`
1. stdout redirect : `wkhtmltopdf -q a.html - > a.pdf`
1. stdin/stdout: `cat a.html | wkhtmltopdf -q - - > aa.pdf`

## read-args-from-stdin

```sh
wkhtmltopdf --read-args-from-stdin
assets/a.html a.pdf
Loading pages (1/6)
Counting pages (2/6)                                               
Resolving links (4/6)                                                       
Loading headers and footers (5/6)                                           
Printing pages (6/6)
Done                                                                      
assets/b.html b.pdf            
Loading pages (1/6)
Counting pages (2/6)                                               
Resolving links (4/6)                                                       
Loading headers and footers (5/6)                                           
Printing pages (6/6)
Done
assets/x.html x.pdf
Loading pages (1/6)
Error: Failed to load http://assets/x.html, with network status code 3 and http status code 0 - Host assets not found
Error: Failed loading page http://assets/x.html (sometimes it will work just to ignore this error with --load-error-handling ignore)
```
