package main

import (
    "bufio"
    "encoding/csv"
    "flag"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "strings"
    "sync"
    "time"
)

func main() {
    ipFilePath, csvFilePath, maxThreads, showHelp := parseCommandLineFlags()

    if *showHelp {
        flag.Usage()
        return
    }

    if *ipFilePath == "" {
        fmt.Println("请提供IP文件路径，使用 -h 查看帮助文档")
        return
    }

    ipFile, csvFile := openFiles(ipFilePath, csvFilePath)
    defer ipFile.Close()
    defer csvFile.Close()

    csvWriter := csv.NewWriter(csvFile)
    defer csvWriter.Flush()

    ipChannel := make(chan string)
    var wg sync.WaitGroup

    startWorkerThreads(maxThreads, &wg, ipChannel, csvWriter)

    processIPAddressesFromFile(ipFile, ipChannel)
    close(ipChannel)
    wg.Wait()
}

// 解析命令行标志
func parseCommandLineFlags() (*string, *string, *int, *bool) {
    ipFilePath := flag.String("i", "", "输入目标IP列表路径")
    csvFilePath := flag.String("o", "cf_trace.csv", "输出报告CSV路径")
    maxThreads := flag.Int("t", 16, "最大线程数")
    showHelp := flag.Bool("h", false, "显示帮助文档")
    flag.Parse()
    return ipFilePath, csvFilePath, maxThreads, showHelp
}

// 打开文件
func openFiles(ipFilePath, csvFilePath *string) (*os.File, *os.File) {
    ipFile, err := os.Open(*ipFilePath)
    if err != nil {
        fmt.Println("无法打开IP文件:", err)
        os.Exit(1)
    }

    csvFile, err := os.Create(*csvFilePath)
    if err != nil {
        fmt.Println("无法创建CSV文件:", err)
        os.Exit(1)
    }

    return ipFile, csvFile
}

// 启动工作者线程
func startWorkerThreads(maxThreads *int, wg *sync.WaitGroup, ipChannel chan string, csvWriter *csv.Writer) {
    maxThreadsDigits := len(fmt.Sprint(*maxThreads))
    for i := 0; i < *maxThreads; i++ {
        wg.Add(1)
        threadID := fmt.Sprintf("[%0*d]", maxThreadsDigits, i+1) // 创建线程前缀，ID补零操作
        go func(id string) {
            defer wg.Done()
            processIPAddresses(ipChannel, csvWriter, id)
        }(threadID)
    }
}

// 从文件中读取IP地址并发送到通道
func processIPAddressesFromFile(ipFile *os.File, ipChannel chan string) {
    ipScanner := bufio.NewScanner(ipFile)
    for ipScanner.Scan() {
        ip := ipScanner.Text()
        ipChannel <- ip
    }
}


// 处理IP地址
func processIPAddresses(ipChannel chan string, csvWriter *csv.Writer, threadID string) {
    for ip := range ipChannel {
        fmt.Printf("%s 正在处理 IP: %s\n", threadID, ip)

        var traceResponse string
        var status string
        for i := 0; i < 3; i++ {
            traceResponse, status = getTraceResponse(ip)
            if status == "200" {
                break
            }
            time.Sleep(time.Second)
        }

        if status == "200" {
            traceColumns := strings.Split(traceResponse, "\n") // 将多行文本分割成字串切片
            traceColumns = append([]string{"CDN_IP=" + ip}, traceColumns...) // 在开头插入元素
            csvWriter.Write(traceColumns)                        // 写入CSV
        } else {
            csvWriter.Write([]string{"CDN_IP=" + ip, "status:failed"})
        }
    }
}

// 获取跟踪响应
func getTraceResponse(ip string) (string, string) {
    url := "http://" + ip + "/cdn-cgi/trace"
    resp, err := http.Get(url)
    if err != nil {
        return "", err.Error()
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            return "", err.Error()
        }
        return string(body), "200"
    }

    return "", fmt.Sprintf("status:%d", resp.StatusCode)
}
