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

var csvMutex sync.Mutex // 声明互斥锁
// 处理IP地址
func processIPAddresses(ipChannel chan string, csvWriter *csv.Writer, threadID string) {
    for ip := range ipChannel {
        traceResponse, status := getTraceResponse(ip, threadID)

        if status == "200" {
            traceResponse := strings.Trim(traceResponse, "\n")  // 删除前后多余的换行符
            traceColumns := strings.Split(traceResponse, "\n")  // 将多行文本分割成字串切片
            traceColumns = append([]string{"CDN_IP=" + ip}, traceColumns...) // 在开头插入元素

            // 使用互斥锁保护写入操作
            csvMutex.Lock()
            csvWriter.Write(traceColumns) // 写入CSV
            csvMutex.Unlock()
        } else {
            csvMutex.Lock()
            csvWriter.Write([]string{"CDN_IP=" + ip, "status=failed"})
            csvMutex.Unlock()
        }
    }
}

// 获取跟踪响应
func getTraceResponse(ip, threadID string) (string, string) {
    maxRetries := 3
    // retryDelay := 1 * time.Second

    // 如果IP地址是IPv6地址并且没有方括号，则添加方括号
    if strings.Contains(ip, ":") && !strings.Contains(ip, "[") {
        ip = "[" + ip + "]"
    }

    for retry := 0; retry <= maxRetries; retry++ {
        if retry > 0 {
            fmt.Printf("%s 重试(%d/%d) : %s \n", threadID, retry, maxRetries, ip)
        } else {
            fmt.Printf("%s 尝试访问 : %s\n", threadID, ip)
        }

        // 添加最多 4 秒的超时
        client := &http.Client{Timeout: 4 * time.Second}
        url := "http://" + ip + "/cdn-cgi/trace"
        resp, err := client.Get(url)
        if err != nil {
            fmt.Printf("%s 请求失败: %s | %v\n", threadID, ip, err)
        } else {
            defer resp.Body.Close()
            if resp.StatusCode == http.StatusOK {
                body, err := ioutil.ReadAll(resp.Body)
                if err != nil {
                    fmt.Printf("%s 读取响应失败: %s | %v\n", threadID, ip, err)
                }
                return string(body), "200"
            } else {
                fmt.Printf("%s 异常响应%d: %s\n", threadID, resp.StatusCode, ip)
            }
        }
    }

    return "", "status=failed"
}
