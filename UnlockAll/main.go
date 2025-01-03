// Package main
// @Description: 根据右键菜单解锁目录文件，或单个文件
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/duke-git/lancet/v2/system"
	"github.com/schollz/progressbar/v3"
	"github.com/shirou/gopsutil/cpu"
)

// 编辑器临时目录路径，用于判断程序运行环境
const EDITOR_PATH = "C:\\Users\\pc\\AppData\\Local\\Temp\\GoLand"

// 程序运行路径
var exe_path string

// 用于同步并发操作的等待组
var wg sync.WaitGroup

// 加密文件的特征字节序列 [20, 35, 101]
var lockedByte []byte

// Progress 结构体用于异步更新进度条，减少锁竞争
type Progress struct {
	current int64                    // 当前进度
	total   int64                    // 总任务数
	bar     *progressbar.ProgressBar // 进度条实例
	updates chan int                 // 进度更新通道
}

func NewProgress(total int64) *Progress {
	p := &Progress{
		total:   total,
		bar:     progressbar.Default(total),
		updates: make(chan int, 1000),
	}

	// 单独的goroutine处理进度更新
	go func() {
		for delta := range p.updates {
			p.bar.Add(delta)
		}
	}()

	return p
}

func (p *Progress) Add(delta int) {
	p.updates <- delta
}

// 使用对象池复用内存，减少GC压力
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

func init_info() bool {
	pathTemp := filepath.Dir(os.Args[0])
	if pathTemp == EDITOR_PATH {
		exe_path, _ = os.Getwd()
	} else {
		exe_path = pathTemp
	}
	lockedByte = append(lockedByte, 20)
	lockedByte = append(lockedByte, 35)
	lockedByte = append(lockedByte, 101)
	return true
}

// ReadBlock 读取文件的指定大小块
// 使用内存池优化性能，避免频繁分配内存
func ReadBlock(filePath string, bufSize int) ([]byte, error) {
	// 从对象池获取缓冲区
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf) // 使用完归还到对象池

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	n, err := f.Read(buf[:bufSize])
	if err != nil {
		return nil, err
	}

	// 返回实际读取的数据副本
	result := make([]byte, n)
	copy(result, buf[:n])
	return result, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("参数长度有误")
		fmt.Scanln()
		return
	}
	if !init_info() {
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
		lock := sync.Mutex{}
		s := spinner.New(spinner.CharSets[59], 500*time.Millisecond)
		s.Prefix = "搜索加密文件中 "
		s.Start()

		allFile, _ := getAllFileIncludeSubFolder(pathTemp)
		needFile := getNeedUnlockFile(allFile)
		s.Stop()

		bar := progressbar.Default(int64(len(needFile)))
		unlockCount := 0
		poolTemp := gopool.NewPool("Unlock", 200, gopool.NewConfig())
		for _, filePath := range needFile {
			wg.Add(1)
			temp := filePath
			poolTemp.Go(func() {
				UnlockFile(temp)
				lock.Lock()
				unlockCount++
				bar.Add(1)
				lock.Unlock()
				wg.Done()
			})
		}
		wg.Wait()
		fmt.Println("操作完成")
		fmt.Scanln()
	} else if info.Mode().IsRegular() {
		//解密当前文件
		data, err := ReadBlock(pathTemp, 4)
		if err != nil {
			return
		}
		if !bytes.Equal(data[1:], lockedByte) {
			//log.Println("文件未加密，跳过解锁：", pathTemp)
			return
		}
		UnlockFile(pathTemp)
	} else {
		fmt.Println("文件类型不支持")
	}
}

// UnlockFile
//
//	@Description: 解密文件
//	@param pathTemp 需要解密的文件路径
//	@param unlockCfg
func UnlockFile(pathTemp string) {
	docPath := pathTemp + ".docx"
	os.Rename(pathTemp, docPath)
	unlockPath := filepath.Join(exe_path, "wps.exe")
	cmd := fmt.Sprintf(`& "%v"  "%v"`, unlockPath, docPath)
	_, _, err := system.ExecCommand(cmd)
	if err != nil {
		log.Println("Failed to run command:", cmd)
		fmt.Scanln()
	} else {
		dstFilePath := docPath + ".temp"
		os.Rename(dstFilePath, pathTemp)
	}
}

// getAllFileIncludeSubFolder 并发扫描目录获取所有文件
// 使用工作池处理子目录，提高扫描效率
func getAllFileIncludeSubFolder(folder string) ([]string, error) {
	var result []string
	var lock sync.Mutex   // 保护结果切片的互斥锁
	var wg sync.WaitGroup // 等待所有扫描任务完成

	// 创建目录扫描工作池，限制并发数为50
	pool := gopool.NewPool("Scanner", 50, gopool.NewConfig())

	// 递归扫描目录的函数
	var scanDir func(path string)
	scanDir = func(path string) {
		files, err := os.ReadDir(path)
		if err != nil {
			return
		}

		for _, file := range files {
			if file.IsDir() {
				// 对子目录启动新的扫描任务
				subPath := filepath.Join(path, file.Name())
				wg.Add(1)
				pool.Go(func() {
					defer wg.Done()
					scanDir(subPath)
				})
			} else {
				// 将文件路径添加到结果集
				lock.Lock()
				result = append(result, filepath.Join(path, file.Name()))
				lock.Unlock()
			}
		}
	}

	scanDir(folder)
	wg.Wait() // 等待所有扫描任务完成

	return result, nil
}

// ResourceLimiter 资源限制器结构体
type ResourceLimiter struct {
	maxCPUPercent float64       // CPU使用率上限
	interval      time.Duration // 检查间隔
	poolSize      int32         // 修改为int32类型
	minPoolSize   int32         // 修改为int32类型
	maxPoolSize   int32         // 修改为int32类型
	pool          gopool.Pool   // 修改为接口类型，不使用指针
	stopChan      chan struct{}
}

// NewResourceLimiter 创建新的资源限制器
func NewResourceLimiter(maxCPUPercent float64) *ResourceLimiter {
	cpuNum := int32(runtime.NumCPU()) // 转换为int32
	minSize := cpuNum * 2
	maxSize := cpuNum * 10

	return &ResourceLimiter{
		maxCPUPercent: maxCPUPercent,
		interval:      time.Second,
		poolSize:      minSize,
		minPoolSize:   minSize,
		maxPoolSize:   maxSize,
		pool:          gopool.NewPool("Unlock", minSize, gopool.NewConfig()),
		stopChan:      make(chan struct{}),
	}
}

// Start 启动资源监控
func (r *ResourceLimiter) Start() {
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.adjustPoolSize()
			case <-r.stopChan:
				return
			}
		}
	}()
}

// Stop 停止资源监控
func (r *ResourceLimiter) Stop() {
	close(r.stopChan)
}

// adjustPoolSize 根据CPU使用率调整池大小
func (r *ResourceLimiter) adjustPoolSize() {
	percent, err := getCPUUsage()
	if err != nil {
		return
	}

	// CPU使用率超过限制时减小池大小
	if percent > r.maxCPUPercent {
		newSize := int32(float64(r.poolSize) * 0.8)
		if newSize >= r.minPoolSize {
			r.poolSize = newSize
			r.pool = gopool.NewPool("Unlock", newSize, gopool.NewConfig())
		}
	} else if percent < r.maxCPUPercent*0.8 {
		newSize := int32(float64(r.poolSize) * 1.2)
		if newSize <= r.maxPoolSize {
			r.poolSize = newSize
			r.pool = gopool.NewPool("Unlock", newSize, gopool.NewConfig())
		}
	}
}

// getCPUUsage 获取当前CPU使用率
func getCPUUsage() (float64, error) {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}
	return percent[0], nil
}

// getNeedUnlockFile 并发筛选需要解密的文件
// 返回所有需要解密的文件路径列表
func getNeedUnlockFile(allFiles []string) []string {
	var result []string
	var lock sync.Mutex

	// 创建资源限制器，设置CPU使用率上限为70%
	limiter := NewResourceLimiter(70.0)
	limiter.Start()
	defer limiter.Stop()

	for _, pathTemp := range allFiles {
		filePath := pathTemp
		wg.Add(1)
		// 使用受限的工作池
		limiter.pool.Go(func() {
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

// fileIsLocked 检查文件是否为加密文件
// 通过检查文件头部特征字节判断
func fileIsLocked(filePath string) bool {
	data, err := ReadBlock(filePath, 4)
	if err != nil {
		return false
	}
	// 检查文件头是否匹配加密标记
	return bytes.Equal(data[1:], lockedByte)
}

// UnlockFiles 批量处理文件，减少WPS启动次数
// batchSize: 每批处理的文件数量
func UnlockFiles(files []string, batchSize int) {
	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]
		// 构建批处理命令
		var cmds []string
		for _, file := range batch {
			docPath := file + ".docx"
			os.Rename(file, docPath)
			cmds = append(cmds, fmt.Sprintf(`"%v"`, docPath))
		}

		// 一次性处理多个文件
		unlockPath := filepath.Join(exe_path, "wps.exe")
		cmd := fmt.Sprintf(`& "%v" %s`, unlockPath, strings.Join(cmds, " "))
		system.ExecCommand(cmd)

		// 处理解密后的文件
		for _, file := range batch {
			docPath := file + ".docx"
			dstFilePath := docPath + ".temp"
			os.Rename(dstFilePath, file)
		}
	}
}
