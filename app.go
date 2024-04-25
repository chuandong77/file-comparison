package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/evanoberholster/imagemeta"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"github.com/disintegration/imaging"
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

// 选择目录
func (a *App) OpenFileDialog(pathA string, pathB string) error {
	exec.Command("open", pathA).Start()
	exec.Command("open", pathB).Start()

	return nil
}


type delFileRequest struct {
	PathA string
	PathB string
}
func (a *App) DelFile(params []delFileRequest, dirType string) error {
	for _,path :=  range params {
		if dirType == "A" {
			os.Remove(path.PathA)
		}

		if dirType == "B" {
			os.Remove(path.PathB)
		}
	}

	return nil
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
	Base64A string

	PathB string
	NameB string
	SizeB int64
	Base64B string
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
	err = os.WriteFile(getComparisonResultFileName(), jsonData, 0777)
	if err != nil {
		return returnJson(0, nil, "创建对比结果失败：" + err.Error())
	}

	return returnJson(1, len(ComparisonResults), "成功")
}

func (a *App) GetComparisonResult(page int, pageSize int) string {
	jsonByte, err := os.ReadFile(getComparisonResultFileName())
	if err != nil {
		return returnJson(0, nil, "暂无对比结果")
	}

	//将结果的json字节转回结构
	var result []ComparisonResult
	err = json.Unmarshal(jsonByte, &result)
	if err != nil {
		return returnJson(0, nil, "暂无对比结果" + err.Error())
	}

	//总条数
	total := len(result)

	//切片分页
	result = pagination(result, page, pageSize)

	var wg = sync.WaitGroup{}
	wg.Add(len(result) * 2)

	for key,item := range result {
		key := key
		item := item
		prefix := []string {"A", "B"}

		for _, px := range prefix {
			px := px
			go func() {
				defer wg.Done()

				if px == "A" {
					result[key].Base64A = getCacheImageToBase64(getFilePath(item.PathA, item.NameA), item.NameA, "A")
				}

				if px == "B" {
					result[key].Base64B = getCacheImageToBase64(getFilePath(item.PathB, item.NameB), item.NameB, "B")
				}
			}()
		}
	}

	wg.Wait()

	type ret struct {
		List []ComparisonResult
		Total int
	}

	data := ret {
		List: result,
		Total: total,
	}

	return returnJson(1, data, "成功")
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
	err := os.Remove(getComparisonResultFileName())
	if err != nil {
		return returnJson(0, nil, "删除结果失败")
	}

	return returnJson(1, nil, "删除结果成功")
}

// GetCurrentDirectory 返回当前程序运行的目录路径
func getCurrentDirectory() (string) {
	// 获取可执行文件的路径
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}

	// 获取可执行文件所在的目录
	currentDir := filepath.Dir(exePath)

	return currentDir
}

func getComparisonResultFileName() string  {
	// 调用 getCurrentDirectory 函数获取当前程序运行的目录
	return getCurrentDirectory() + "/comparison-result.json"
}

// AdjustImage 调整图片大小并转换格式
func adjustImage(inputPath, outputPath string) error {
	// 打开图片文件
	src, err := imaging.Open(inputPath)
	if err != nil {
		fmt.Println("打开失败" + err.Error())
		return err
	}

	ext := filepath.Ext(inputPath)

	// 如果图片格式为 HEIC，则转换为 JPEG
	if ext == ".HEIC" {
		// 调整图片大小为宽度 200px，高度按比例调整
		dst := imaging.Resize(src, 200, 0, imaging.Lanczos)

		// 保存为 JPEG 格式
		err = imaging.Save(dst, outputPath)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		fmt.Println("图片格式为 HEIC，已转换为 JPEG，并调整大小完成")
	} else {
		// 调整图片大小为宽度 200px，高度按比例调整
		dst := imaging.Resize(src, 200, 0, imaging.Lanczos)

		// 保存为原始格式
		err = imaging.Save(dst, outputPath)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		fmt.Println("图片格式为非 HEIC，已调整大小完成")
	}

	return nil
}

// imageToBase64 将本地图片转换为 base64 编码
func imageToBase64(imagePath string) (string) {
	// 读取本地图片文件
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		fmt.Println("转换Base64失败" + err.Error())
		return ""
	}

	// 将图片内容进行 base64 编码
	base64Encoded := base64.StdEncoding.EncodeToString(imageData)

	return base64Encoded
}

func getCacheImageToBase64(file string, fileName string, prefix string) string {
	return ""
	cacheFilePath := ""

	if !isImage(file) {
		return ""
	}

	if isHEIC(file) {
		cacheFilePath = getCacheFileName(prefix, replaceHEICExt(fileName))
		//heic文件转为jpg
		HeicToJpg(file, cacheFilePath)

		src, _ := imaging.Open(cacheFilePath)
		// 调整图片大小为宽度 200px，高度按比例调整
		dst := imaging.Resize(src, 200, 0, imaging.Lanczos)
		imaging.Save(dst, cacheFilePath)
	} else {
		cacheFilePath = getCacheFileName(prefix, fileName)
		src, _ := imaging.Open(file)
		// 调整图片大小为宽度 200px，高度按比例调整
		dst := imaging.Resize(src, 200, 0, imaging.Lanczos)
		imaging.Save(dst, cacheFilePath)
	}

	result := imageToBase64(cacheFilePath)
	if result != "" {
		defer os.Remove(cacheFilePath)
	}

	//返回base64
	return result
}

func getCacheFileName(prefix string, fileName string) string {
	err := cacheDirExists(getCurrentDirectory() + "/image_cache")
	if err != nil {
		return ""
	}

	return getCurrentDirectory() + "/image_cache/" + prefix + "_" + fileName
}

// IsImage 判断文件是否为图片
func isImage(filePath string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".heic"}
	ext := strings.ToLower(filepath.Ext(filePath))
	for _, imageExt := range imageExts {
		if ext == imageExt {
			return true
		}
	}

	return false
}

func isHEIC(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".heic"
}

// replaceHEICExt 替换文件路径中的 HEIC 扩展名为 JPG
func replaceHEICExt(filePath string) string {
	ext := filepath.Ext(filePath)
	if ext == ".heic" || ext == ".HEIC" {
		return strings.TrimSuffix(filePath, ext) + ".jpg"
	}
	return filePath
}


// cacheDirExists 检查目录是否存在，不存在则创建目录
func cacheDirExists(dirPath string) error {
	// 检查目录是否存在
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		// 目录不存在，创建目录
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			return err
		}
	} else if err != nil {
		// 其他错误，返回错误信息
		return err
	}

	return nil
}

func pagination(slice []ComparisonResult, page int, pageSize int) []ComparisonResult {
	start := (page - 1) * pageSize
	end := start + pageSize

	// 确保 start 和 end 在合法范围内
	if start < 0 {
		start = 0
	}
	if end > len(slice) {
		end = len(slice)
	}

	return slice[start:end]
}

