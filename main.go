package main

import (
    "log"
    "os"
    "bufio"
    "net/http"
    "io"
    "strconv"
    "flag"
    "regexp"
    "io/ioutil"
)

type Addr struct {
    ip   string
    host string
}

func (addr Addr) to_str() string {
    out := "local-zone: \"" + addr.host + "\" redirect\nlocal-data: \"" + addr.host + " A " + addr.ip + "\"\n"
    return out
}

func main() {
    log.Println("ADblock initialize...")
    filePtr := flag.String("file", "", "a string")
    flag.Parse()

    log.Println("file is: ", *filePtr)

    if *filePtr != "" {
        log.Println("Parsing and convert file...")
        err := ParseAndConvert(*filePtr, "ad.conf")
        if err != nil {
            log.Fatal(err)
        }
        log.Println("Parsinng and convert complete!")
    } else {
        UrlList := getUrlList()
        for ind,item := range UrlList {
            filename := strconv.Itoa(ind) + ".tmp"
            err := downloadList(item, filename)
            if err != nil {
                log.Fatal("[ERROR]", err)
            }
            err = ParseAndConvert(filename, filename+".conf")
            if err != nil {
                log.Fatal("[ERROR]", err)
            }
        }
        err := SplitFiles(UrlList)
        if err != nil {
            log.Fatal("[ERROR]", err)
        }
    }
    log.Println("Finaly... Copy ad.conf to /etc/unbound/unbound.conf/ad.conf")
}

func SplitFiles(files []string) error {
    conf, err := os.Create("ad.conf")
    if err != nil {
        return err
    }
    defer conf.Close()

    for ind, _ := range files {
        data, err := ioutil.ReadFile(strconv.Itoa(ind) + ".tmp.conf")
        if err != nil {
            return err
        }
        text := string(data)
        conf.WriteString(text)
    }
    return nil
}

func ParseAndConvert(filename, outfile string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    conf, err := os.Create(outfile)
    if err != nil {
        return err
    }
    defer conf.Close()

    var valid = regexp.MustCompile(`^[127,0]{0,}\.0\.0\.[0,1]{0,}[\s+]{0,}(.+)$`)
    var localhost = regexp.MustCompile(`localhost`)

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        arr := valid.FindAllStringSubmatch( line, -1 )
        if arr != nil {
            err := localhost.FindAllStringSubmatch(arr[0][1],-1)
            if err == nil {
                addr := Addr{"127.0.0.1",arr[0][1]}
                conf.WriteString(addr.to_str())
            }
        }
    }

    return nil
}

// Парсинг листов и сохранение в формат Unbound DNS
//func parseList() error {
//}

// Скачиваем файлы с АД листами
func downloadList(url, filename string) error {
    client := http.Client{}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return err
    }

    out, err := os.Create("./" + filename)
    if err != nil {
        return err
    }

    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    _, err = io.Copy(out, resp.Body)
    if err != nil {
        return err
    }

    return nil 
}

// Получаем из файла urls.txt адресса откуда брать сайты для блокировки
func getUrlList() []string {
    file, err := os.Open("urls.txt")
    if err != nil {
        log.Fatal("[ERROR]",err)
    }

    scanner := bufio.NewScanner(file)
    urlList := make([]string,0)
    for scanner.Scan() {
        url := scanner.Text()
        urlList = append(urlList, url)
    }
    return urlList
}
