package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/evanoberholster/imagemeta"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"os"
	"strings"
	"time"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// 选择目录
func (a *App) OpenDirectoryDialog() string {
	path, _ := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{})
	return path
}

type requestData struct {
	CheckType string `json:"checkType"`
	PathA string `json:"pathA"`
	IsAppendTimeA bool `json:"isAppendTimeA"`
	PathB string `json:"pathB"`
	IsAppendTimeB bool `json:"isAppendTimeB"`
}

//对比结果
type ComparisonResult struct {
	PathA string
	NameA string
	SizeA int64
	PathB string
	NameB string
	SizeB int64
}

type Response struct {
	Ret  int        `json:"ret"`
	Msg  string     `json:"msg"`
	Data any `json:"data"`
}

type fileInfo struct {
	Name string
	Size int64
}
func (a *App) Comparison(data requestData) string {
	// 读取 pathB 目录中所有文件名
	filesB, err := os.ReadDir(data.PathB)
	if err != nil {
		return returnJson(0, nil, "错误：读取目录 B 失败：" + err.Error())
	}

	// 存储 pathB 目录中所有文件名的临时 map
	filesMap := make(map[string]fileInfo)
	for _, fileB := range filesB {
		fileNameB := fileB.Name()
		if fileNameB == ".DS_Store" || fileB.IsDir() {
			continue
		}

		if data.CheckType == "name" && data.IsAppendTimeB {
			appendTimeToFileName := appendTimeToFileName(data.PathB, fileB.Name())
			if appendTimeToFileName != "" {
				fileNameB = appendTimeToFileName
			}
		}

		if data.CheckType == "md5" {
			fileNameB, _ = getFileMD5(getFilePath(data.PathB, fileB.Name()))
			if fileNameB == "" {
				continue
			}
		}

		fileInfoB, _ := fileB.Info()
		filesMap[fileNameB] = fileInfo{
			Name: fileB.Name(),
			Size: fileInfoB.Size(),
		}
	}

	filesA, err := os.ReadDir(data.PathA)
	if err != nil {
		return returnJson(0, nil, "错误：读取目录 A 失败：" + err.Error())
	}

	var ComparisonResults []ComparisonResult
	for _, fileA := range filesA {
		newFilenameA := fileA.Name()
		if newFilenameA == ".DS_Store" || fileA.IsDir() {
			continue
		}

		//需要追加时间到文件名
		if data.CheckType == "name" && data.IsAppendTimeA {
			appendTimeToFileName := appendTimeToFileName(data.PathA, fileA.Name())
			if appendTimeToFileName != "" {
				newFilenameA = appendTimeToFileName
			}
		}

		if data.CheckType == "md5" {
			newFilenameA, _ = getFileMD5(getFilePath(data.PathA, fileA.Name()))
			if newFilenameA == "" {
				continue
			}
		}

		fileInfoA, _ := fileA.Info()
		if _, exists := filesMap[newFilenameA]; exists {
			ComparisonResults = append(ComparisonResults, ComparisonResult{
				PathA: data.PathA,
				NameA: fileA.Name(),
				SizeA: fileInfoA.Size(),
				PathB: data.PathB,
				NameB: filesMap[newFilenameA].Name,
				SizeB: filesMap[newFilenameA].Size,
			})
		}
	}

	jsonData, _ := json.Marshal(ComparisonResults)
	// 将 JSON 字节写入文件
	err = os.WriteFile("comparison-result.json", jsonData, 0644)
	if err != nil {
		return returnJson(0, nil, "创建对比结果失败：" + err.Error())
	}

	return returnJson(1, len(ComparisonResults), "成功")
}

func (a *App) GetComparisonResult() string {
	jsonByte, err := os.ReadFile("comparison-result.json")
	if err != nil {
		return returnJson(0, nil, "暂无对比结果")
	}

	//将结果的json字节转回结构
	var result []ComparisonResult
	err = json.Unmarshal(jsonByte, &result)
	if err != nil {
		return returnJson(0, nil, "暂无对比结果" + err.Error())
	}

	return returnJson(1, result, "成功")
}

func getFilePath(path string, fileName string) string {
	join := ""
	if !strings.HasSuffix(path, "/") {
		join = "/"
	}

	return path + join + fileName;
}

func formatTime(t time.Time) (string) {
	// 格式化时间为指定格式
	output := t.Format("20060102_150405")
	return output
}

func insertTimeAfterPrefix(filename, prefix, timeStr string) string {
	// 查找前缀在文件名中的位置
	index := strings.Index(filename, prefix)
	if index == -1 {
		// 如果找不到前缀，直接返回原始文件名
		return filename
	}

	// 前缀后的位置
	insertPosition := index + len(prefix)

	// 在前缀后插入时间字符串
	newFilename := filename[:insertPosition] + "_" + timeStr + filename[insertPosition:]

	return newFilename
}

/**
追加时间到文件名中
 */
func appendTimeToFileName(path string, fileName string) string {
	//不是 IMG 前缀，跳过本文件，不进行对比
	if !strings.HasPrefix(fileName, "IMG") {
		return ""
	}

	filePath := getFilePath(path, fileName)
	f, err := os.Open(filePath)
	if err != nil {
		//fmt.Println("错误：读取文件错误", filePath)
		return ""
	}

	e, err := imagemeta.Decode(f)
	if err != nil {
		//fmt.Println("错误：解析 EXIF 失败" + fileName, err.Error())
		return ""
	}

	formatTime := formatTime(e.DateTimeOriginal())

	// 调用方法将时间插入到文件名的指定位置，将文件名处理为 IMG_20230209_163621_0002.HEIC 格式
	return insertTimeAfterPrefix(fileName, "IMG", formatTime)
}

func getFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 创建一个新的哈希实例
	hash := md5.New()

	// 复制文件内容到哈希实例中
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// 计算最终的哈希值并格式化为16进制字符串
	hashInBytes := hash.Sum(nil)
	md5String := fmt.Sprintf("%x", hashInBytes)

	return md5String, nil
}

func returnJson(ret int, data any, msg string) string  {
	response := Response{
		Ret:  ret,
		Msg:  msg,
		Data: data,
	}
	jsonData, _ := json.Marshal(response)

	return string(jsonData)
}

func (a *App) DelComparisonResult() string {
	err := os.Remove("comparison-result.json")
	if err != nil {
		return returnJson(0, nil, "删除结果失败")
	}

	return returnJson(1, nil, "删除结果成功")
}