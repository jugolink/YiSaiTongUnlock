// Package main
// @Description: 解锁单个文件
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("参数错误：请提供需要解锁的文件路径（可多个）")
		return
	}
	for _, sourcePath := range os.Args[1:] {
		info, err := os.Stat(sourcePath)
		if err != nil {
			fmt.Printf("[%s] 无法获取文件信息: %v\n", sourcePath, err)
			continue
		}
		if info.IsDir() {
			fmt.Printf("[%s] 不支持对目录操作\n", sourcePath)
			continue
		}

		// 输出文件名为 "xxx.docx.temp"
		dstFilePath := sourcePath + ".temp"
		err = copyFile(sourcePath, dstFilePath)
		if err != nil {
			fmt.Printf("[%s] 文件复制失败: %v\n", sourcePath, err)
			continue
		}

		fmt.Printf("[%s] 解锁完成，生成文件: %s\n", sourcePath, dstFilePath)
	}
}

func copyFile(sourcePath, dstFilePath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer source.Close()

	if _, err := os.Stat(dstFilePath); err == nil {
		return fmt.Errorf("目标文件已存在: %s", dstFilePath)
	}

	destination, err := os.OpenFile(dstFilePath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0666)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}
	return nil
}
