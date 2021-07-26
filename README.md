# wkhtmltopdf-perf

performance tests for wkhtmltopdf.

[wkhtmltopdf](https://github.com/wkhtmltopdf/wkhtmltopdf) converts HTML to PDF using Webkit (QtWebKit), which is the
browser engine that is used to render HTML and javascript - Chrome uses that engine too..

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
