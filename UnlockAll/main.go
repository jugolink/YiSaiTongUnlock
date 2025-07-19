// Package main
// @Description: 根据右键菜单解锁目录文件，或单个文件
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/schollz/progressbar/v3"
)

var exe_path string

// 加密文件标志
var lockedByte []byte

func init_info() bool {
	exe_path = filepath.Dir(os.Args[0])
	lockedByte = []byte{20, 35, 101}
	return true
}

func ReadBlock(filePth string, bufSize int) ([]byte, error) {
	f, err := os.Open(filePth)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := make([]byte, bufSize)
	bfRd := bufio.NewReader(f)
	n, err := bfRd.Read(buf)
	if err != nil && n == 0 {
		return nil, err
	}
	return buf[:n], nil
}

const batchSize = 10 // 每批处理10个文件，可根据实际情况调整

func main() {
	if !CheckEncryptEnv() {
		fmt.Println("[环境检测] 未检测到加密环境（CDGRegedit.exe），无法进行转换操作。请在加密环境下运行！")
		fmt.Println("按回车键退出...")
		fmt.Scanln()
		return
	}

	if len(os.Args) != 2 {
		fmt.Println("参数长度有误")
		fmt.Scanln()
		return
	}
	if !init_info() {
		fmt.Println("初始化失败")
		return
	}

	pathTemp := os.Args[1]
	info, err := os.Stat(pathTemp)
	if err != nil {
		fmt.Println("无法获取文件或目录信息：", err)
		fmt.Scanln()
		return
	}
	if info.IsDir() {
		var lock sync.Mutex
		s := spinner.New(spinner.CharSets[59], 500*time.Millisecond)
		s.Prefix = "搜索加密文件中 "
		s.Start()

		allFile, err := getAllFileIncludeSubFolder(pathTemp)
		if err != nil {
			fmt.Println("遍历目录失败：", err)
			s.Stop()
			return
		}
		needFile := getNeedUnlockFile(allFile)
		s.Stop()

		bar := progressbar.Default(int64(len(needFile)))
		var successCount, failCount int
		var wg sync.WaitGroup

		// 批量分组
		for i := 0; i < len(needFile); i += batchSize {
			end := i + batchSize
			if end > len(needFile) {
				end = len(needFile)
			}
			batch := needFile[i:end]
			wg.Add(1)
			go func(files []string) {
				defer wg.Done()
				err := UnlockFilesBatch(files)
				lock.Lock()
				if err != nil {
					failCount += len(files)
				} else {
					successCount += len(files)
				}
				bar.Add(len(files))
				lock.Unlock()
			}(batch)
		}
		wg.Wait()
		fmt.Printf("操作完成，成功：%d，失败：%d\n", successCount, failCount)
		fmt.Scanln()
	} else if info.Mode().IsRegular() {
		data, err := ReadBlock(pathTemp, 4)
		if err != nil {
			fmt.Println("读取文件失败：", err)
			return
		}
		if len(data) < 4 || !bytes.Equal(data[1:], lockedByte) {
			fmt.Println("文件未加密，跳过解锁：", pathTemp)
			return
		}
		err = UnlockFile(pathTemp)
		if err != nil {
			fmt.Println("解锁失败：", err)
		} else {
			fmt.Println("解锁成功")
		}
	} else {
		fmt.Println("文件类型不支持")
	}
}

// UnlockFile
//
//	@Description: 解密文件
//	@param pathTemp 需要解密的文件路径
func UnlockFile(pathTemp string) error {
	// 1. 保存原始时间
	fileInfo, err := os.Stat(pathTemp)
	if err != nil {
		log.Println("无法获取文件信息:", err)
		return err
	}
	originalModTime := fileInfo.ModTime()

	// 2. 原有解密流程
	docPath := pathTemp + ".docx"
	if err := os.Rename(pathTemp, docPath); err != nil {
		log.Println("重命名失败:", err)
		return err
	}
	// 回滚函数
	rollback := func() {
		if _, err := os.Stat(docPath); err == nil {
			os.Rename(docPath, pathTemp)
		}
	}

	unlockPath := filepath.Join(exe_path, "wps.exe")
	if _, err := os.Stat(unlockPath); err != nil {
		log.Println("解锁程序不存在:", unlockPath)
		rollback()
		return fmt.Errorf("解锁程序不存在: %s", unlockPath)
	}
	cmd := exec.Command(unlockPath, docPath)
	if err := cmd.Run(); err != nil {
		log.Println("解密失败:", err)
		rollback()
		return err
	}

	// 3. 恢复文件并重置时间
	dstFilePath := docPath + ".temp"
	if err := os.Rename(dstFilePath, pathTemp); err != nil {
		log.Println("恢复文件名失败:", err)
		rollback()
		return err
	}
	if err := os.Chtimes(pathTemp, originalModTime, originalModTime); err != nil {
		log.Println("恢复时间失败:", err)
		// 不回滚，时间失败不影响文件内容
	}
	return nil
}

func getAllFileIncludeSubFolder(folder string) ([]string, error) {
	var result []string
	err := filepath.Walk(folder, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Println(err.Error())
			return err
		}
		if !info.IsDir() {
			result = append(result, path)
		}
		return nil
	})
	return result, err
}

func getNeedUnlockFile(allFiles []string) []string {
	var result []string
	var lock sync.Mutex
	var wg sync.WaitGroup
	poolTemp := gopool.NewPool("Unlock", 50, gopool.NewConfig())
	for _, pathTemp := range allFiles {
		filePath := pathTemp
		wg.Add(1)
		poolTemp.Go(func() {
			defer wg.Done()
			isLocked := fileIsLocked(filePath)
			if isLocked {
				lock.Lock()
				result = append(result, filePath)
				lock.Unlock()
			}
		})
	}
	wg.Wait()
	return result
}

func fileIsLocked(filePath string) bool {
	data, err := ReadBlock(filePath, 4)
	if err != nil {
		return false
	}
	if len(data) < 4 || !bytes.Equal(data[1:], lockedByte) {
		return false
	}
	return true
}

// 新增批量解锁函数
func UnlockFilesBatch(files []string) error {
	unlockPath := filepath.Join(exe_path, "wps.exe")
	args := append([]string{}, files...)
	log.Printf("调用解锁程序: %s %v", unlockPath, args)
	cmd := exec.Command(unlockPath, args...)
	output, err := cmd.CombinedOutput()
	log.Printf("解锁程序输出: %s", string(output))
	if err != nil {
		log.Printf("批量解锁失败: %v", err)
	}

	for _, file := range files {
		docPath := file + ".docx"
		tempPath := docPath + ".temp"
		log.Printf("检查文件: %s", tempPath)
		if _, err := os.Stat(tempPath); err == nil {
			log.Printf("重命名 %s -> %s", tempPath, file)
			_ = os.Rename(tempPath, file)
			_ = os.Remove(docPath)
		} else {
			if _, err := os.Stat(docPath); err == nil {
				log.Printf("回滚 %s -> %s", docPath, file)
				_ = os.Rename(docPath, file)
			} else {
				log.Printf("既没有 %s 也没有 %s，文件可能丢失", tempPath, docPath)
			}
		}
	}
	return nil
}

// 检查加密环境（CDGRegedit.exe 进程是否存在）
func CheckEncryptEnv() bool {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist")
	} else {
		cmd = exec.Command("ps", "-A")
	}
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("[环境检测] 检查加密环境失败：", err)
		return false
	}
	return strings.Contains(strings.ToLower(string(output)), "cdgregedit.exe")
}
