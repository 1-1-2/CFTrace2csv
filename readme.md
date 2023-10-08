# Go程序 - IP地址处理器

本程序用于批量获取 Cloudflare IP 地址的 trace 信息，访问 http://IP/cdn-cgi/trace，然后将结果写入一个CSV文件中。

## 使用说明

### 安装Go

在运行这个程序之前，您需要安装Go编程语言。您可以从官方网站下载并安装Go：https://golang.org/doc/install

### 下载和编译程序

使用以下命令下载和编译程序：

```shell
git clone https://github.com/1-1-2/CFTrace2csv.git
cd CFTrace2csv
# go build Cloudflare_trace.go -o Cloudflare_trace.exe
go build Cloudflare_trace.go
```

### 运行程序

程序接受以下命令行参数：

- `-i`: IP文件路径（包含IP清单的文本文件，一行一个IP）。
- `-o`: 报告CSV路径（可选，默认为 `cf_trace.log`）。
- `-t`: 最大线程数（可选，默认为 `5`）。
- `-h`: 显示帮助文档。

示例运行程序的命令：

```shell
./Cloudflare_trace -i input.txt -o output.csv -t 10
```

### 示例

假设您有一个名为 `input.txt` 的IP文件，其中包含以下IP地址清单：

```
192.168.1.1
192.168.1.2
192.168.1.3
```

您可以运行程序来处理这些IP地址：

```shell
./Cloudflare_trace -i input.txt -o output.csv
```

程序将访问每个IP地址的特定URL，并将结果写入 `output.csv` 文件中。

## todo

1. 修正CSV格式（目前还没驯化出来，搞个python算了）



## 声明

本项目包含 AI 辅助创作

