package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("无法获取当前工作目录: %v", err)
	}

	// BASE_DIR：当前目录的上级目录
	baseDir := filepath.Dir(cwd)

	// RAW_DIR: 当前目录下的 Raw 文件夹
	rawDir := filepath.Join(cwd, "Raw")
	// OUTPUT_DIR: 上层目录下的 "柠泽" 文件夹
	outputDir := filepath.Join(baseDir, "柠泽")

	// 支持的文件格式（小写，不含点）
	supportedFormats := []string{"jpg", "jpeg", "png", "mp4", "mov", "heic"}

	// 检查 Raw 文件夹是否存在
	info, err := os.Stat(rawDir)
	if err != nil || !info.IsDir() {
		fmt.Printf("错误：路径 %s 不存在！\n", rawDir)
		os.Exit(1)
	}

	// 创建输出目录 "柠泽"（如果不存在）
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatalf("无法创建输出目录 %s: %v", outputDir, err)
	}

	// 用于存储重复文件的列表
	var duplicateFiles []string

	// 遍历 Raw 目录下的所有文件
	err = filepath.Walk(rawDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 获取文件扩展名（不区分大小写）
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
		supported := false
		for _, format := range supportedFormats {
			if ext == format {
				supported = true
				break
			}
		}
		if !supported {
			return nil
		}

		// 获取文件修改时间，并格式化为 "YYYY-MM-DD_HHMMSS"
		modTime := info.ModTime()
		fileDate := modTime.Format("2006-01-02_150405")

		// 计算文件的 MD5 值
		fileMD5, err := computeMD5(path)
		if err != nil {
			log.Printf("计算文件 %s 的 MD5 失败: %v", path, err)
			return nil
		}

		// 构造新的文件名：日期_文件MD5.扩展名
		newFileName := fmt.Sprintf("%s_%s.%s", fileDate, fileMD5, ext)

		// 根据日期获取年月（前 7 个字符，例如 "2023-10"）
		yearMonth := fileDate[:7]
		targetDir := filepath.Join(outputDir, yearMonth)
		// 创建目标文件夹（如果不存在）
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			log.Printf("创建目录 %s 失败: %v", targetDir, err)
			return nil
		}

		targetPath := filepath.Join(targetDir, newFileName)

		// 检查目标文件是否已存在
		if _, err := os.Stat(targetPath); err == nil {
			// 获取相对路径：移除 baseDir 前缀
			relPath, err := filepath.Rel(baseDir, targetPath)
			if err != nil {
				relPath = targetPath
			}
			duplicateFiles = append(duplicateFiles, fmt.Sprintf("%s => %s", filepath.Base(path), relPath))
		} else {
			// 移动文件
			err = os.Rename(path, targetPath)
			if err != nil {
				log.Printf("移动文件 %s 到 %s 失败: %v", path, targetPath, err)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("遍历目录时出错: %v", err)
	}

	// 输出重复文件信息
	if len(duplicateFiles) > 0 {
		fmt.Println("以下文件是重复的，未被处理：")
		for _, dup := range duplicateFiles {
			fmt.Println(dup)
		}
	} else {
		fmt.Println("没有重复文件。")
	}

	fmt.Println("整理完成！")
}

// computeMD5 计算指定文件的 MD5 值
func computeMD5(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
