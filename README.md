# wkhtmltopdf-perf

performance prof for wkhtmltopdf.

[wkhtmltopdf](https://github.com/wkhtmltopdf/wkhtmltopdf) converts HTML to PDF using Webkit (QtWebKit), which is the
browser engine that is used to render HTML and javascript - Chrome uses that engine too.

wkhtmltopdf 是一个开源的，使用 Qt WebKit 渲染引擎，把 html 转换为 pdf 文件的命令行工具。
wkhtmltopdf 还有一个双胞胎兄弟 wkhtmltoimage，顾名思义，它可以把 html 转换为 image 图片。

简单的讲，wkhtmltopdf 用于把网页转换成 pdf 文件。

## 不同实现版本之间的差异对比

```sh
$ wkhtmltopdf -V
wkhtmltopdf 0.12.6 (with patched qt)
```

版本 | 进程预热 | 进程重用 | html参数 | PDF 落盘 | wkhtmltopdf | 压测1 TPS(c100) | 压测2 TPS(c25)
---|---|----|----|----|----|----|----
-| - | - | 网址 | Y |  `wk --quiet http://a.b.c/a.html a.pdf`|51.931|16.370
0| - | - | 网址 | N（stdout) | `wk --quiet http://a.b.c/a.html - \| cat` | 51.272| 16.121
1| - | - | 内容 (stdin) | N（stdout) | `wk --quiet - - \| cat` | 80.113|53.793
1p| 预热 | -| 内容 (stdin) | N（stdout) | `wk --quiet - - \| cat` | 71.630|54.223
2| 预热 | 重用 | 网址 | Y |  `wk --read-args-from-stdin` | 209.262|26.457
2p| 预热 | 重用 | 网址 | N (fuse) | `wk --read-args-from-stdin`|225.668|26.314

1. 压测1: `gobench -l ":9337?url=http://127.0.0.1:9337/b.html&v={v}"`, v2/v2p 池大小100
2. 1/1p时， 页面内链接未处理，对于有页面内css/js时，此项指标无意义

## Install

1. Get the latest downloads from [here](https://wkhtmltopdf.org/downloads.html).
1. [Linux install](https://github.com/adrg/go-wkhtmltopdf/wiki/Install-on-Linux)
    - centos: `yum localinstall wkhtmltox.rpm`
    - fedora: `dnf localinstall wkhtmltox.rpm`
    - Dibian: `sudo dpkg -i wkhtmltox.deb; sudo ldconfig`
1. 安装 `wk`
   ```sh
   $ make linux
   $ bssh scp ~/go/bin/linux_amd64/wk r:/usr/local/bin/ -H 126.72
   $ bssh -H 126.72 wk -v                                             
   Select Server :zzdev1
   Run Command   :wk -v
   wk (a go wrapper for wkhtmltopdf), v1.0.0 released at 2021-07-28 09:39:07
   ```
1. API
   `GET http://192.168.126.72:9339?url={HtmlPageURL}&extra={ExtraWkArgs}&saveFile=y`

## 5 minutes to start

1. quiet mode : `wkhtmltopdf -q a.html a.pdf`
1. stdout redirect : `wkhtmltopdf -q a.html - > a.pdf`
1. stdin/stdout:
    - [You need to specify at least one input file, and exactly one output file, Use - for stdin or stdout](https://github.com/wkhtmltopdf/wkhtmltopdf/blob/master/src/pdf/pdfcommandlineparser.cc)
    - `cat a.html | wkhtmltopdf -q - - > a.pdf`
    - `echo '<p>Hello</p>' | wkhtmltopdf -q - - > hello.pdf`
1. [wkhtmltopdf Go bindings and high level interface for HTML to PDF conversion](https://github.com/adrg/go-wkhtmltopdf)

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

## Resouces

1. [wkhtmltox to provide high performance access to wkhtmltopdf and wkhtmltoimage from node.js](https://github.com/tcort/wkhtmltox)
1. [blob One Year With Wkhtmltopdf: One Thousand Problems, One Thousand Solutions](https://blog.theodo.com/2016/12/wkhtmltopdf/)
   > Wkhtmltopdf has dependencies. On Linux, I had to install zlib, fontconfig, freetype, and X11 libs
1. 从标准输入获取参数
   > 如果需要对许多页面进行批量处理，并且感觉 `wkhtmltopdf` 开启比较慢，可以尝试使用 `--read-args-from-stdin` 参数。
   > wkhtmltopdf 命令会为 `--read-args-from-stdin` 参数发送过来的每一行进行一次单独命令调用。
   > 也就是说此参数每读取一行都会执行一次 wkhtmltopdf 命令。而最终执行的命令中的参数是命令行中参数与此参数读取的标准输入流中参数的结合。
   > 下面的代码段是一个例子:
   ```sh
   echo "https://baike.baidu.com/item/2020%E5%B9%B4%E4%B8%9C%E4%BA%AC%E5%A5%A5%E8%BF%90%E4%BC%9A#hotspotmining a.pdf" >> cmds
   echo "cover baidu.com https://baike.baidu.com/item/%E5%8F%B0%E9%A3%8E%E7%83%9F%E8%8A%B1/58020097 b.pdf" >> cmds
   wkhtmltopdf --read-args-from-stdin < cmds
   ```

