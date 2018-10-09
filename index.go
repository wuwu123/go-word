package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	"io/ioutil"
	"sort"
	"time"
)

// 字符转码
func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

func GbkToUtf(src string) string {
	return ConvertToString(src, "gbk", "utf-8")
}

// 写入文件
func WriteFile(title string, outputString string) {
	outputFile, outputError := os.OpenFile("./text/"+title+".text", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if outputError != nil {
		fmt.Printf("An error occurred with file opening or creation\n")
		return
	}
	defer outputFile.Close()
	outputWriter := bufio.NewWriter(outputFile)
	// re, _ := regexp.Compile(`全本推荐.\n`)
	// outputString = re.ReplaceAllString(outputString, "\n")
	outputWriter.WriteString(strings.Replace(outputString, "聽聽聽聽", "  ", -1))
	outputWriter.Flush()
}

// 获取文件内容
func GetFileContent(fileName string) string {
	filePath := "./text/"+fileName+".text"
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	defer f.Close()
	err = os.RemoveAll(filePath)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		contentByte, _ := ioutil.ReadAll(f)
		return string(contentByte)
	}
	return ""
}
func WriteIndex(allTest map[int]string) {
	outputFile, outputError := os.OpenFile("./text/index.md", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if outputError != nil {
		fmt.Printf("An error occurred with file opening or creation\n")
		return
	}
	defer outputFile.Close()
	outputWriter := bufio.NewWriter(outputFile)
	keys := []int{}
	for key := range allTest {
		keys = append(keys , key)
	}
	sort.Sort(sort.IntSlice(keys))
	for _, key := range keys {
		outputWriter.WriteString(GetFileContent(allTest[key]))
	}
	outputWriter.Flush()
}

// 获取章节内容
func GetBody(title string, url string, ch chan<- string) {
	res, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprintf("URL获取数据 %s", err.Error())
		return
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		ch <- fmt.Sprintf("demo错误 %s", err.Error())
		return
	}
	WriteFile(title, "##"+title+"\r\n"+GbkToUtf(doc.Find("#content").Text()))
	ch <- title
}

func GetMenu() {
	baseUrl := "http://www.126shu.com"
	res, err := http.Get(baseUrl + "/1761/")
	if err != nil {
		fmt.Println("获取数据", err.Error())
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("demo错误", err.Error())
		return
	}
	ch := make(chan string)
	menuArray := make(map[int]string)
	allLine := 0
	j := 1
	doc.Find("#list dl dd").Each(func(i int, s *goquery.Selection) {
		allLine++
		title := GbkToUtf(s.Find("a").Text())
		menuArray[int(i + 1)] = title
		href, _ := s.Find("a").Attr("href")
		go GetBody(title, baseUrl+GbkToUtf(href), ch)
		if i%100 == 0 {
			for j := 1; j <= allLine; j++ {
				fmt.Println(<-ch)
			}
			allLine = 0
			j = 1
		}

	})
	for j := 1; j <= allLine; j++ {
		fmt.Println(<-ch)
	}
	WriteIndex(menuArray)
}

func main() {
	begin := time.Now().Unix()
	GetMenu()
	end := time.Now().Unix()
	fmt.Println(end-begin)
}
