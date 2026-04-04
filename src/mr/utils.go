package mr

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

func DelFileByMapID(targetNumber int, path string) interface{} {
	pattern := fmt.Sprintf(`^mr-out-%d-\d+$`, targetNumber)
	regex, error := regexp.Compile(pattern)
	if error != nil {
		return error
	}

	files, error := ioutil.ReadDir(path)
	if error != nil {
		return error
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		if regex.MatchString(fileName) {
			filePath := filepath.Join(path, fileName)
			err := os.Remove(filePath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func DelFileByReduceId(targetNumber int, path string) error {
	// 创建正则表达式，X 是可变的指定数字
	pattern := fmt.Sprintf(`^mr-out-\d+-%d$`, targetNumber)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	// 读取当前目录中的文件
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	// 遍历文件，查找匹配的文件
	for _, file := range files {
		if file.IsDir() {
			continue // 跳过目录
		}
		fileName := file.Name()
		if regex.MatchString(fileName) {
			// 匹配到了文件，删除它
			filePath := filepath.Join(path, file.Name())
			err := os.Remove(filePath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ReadSpecificFile(targetNumber int, path string) (fileList []*os.File, err error) {
	pattern := fmt.Sprintf(`^mr-out-\d+-%d$`, targetNumber)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if regex.MatchString(entry.Name()) {
			filePath := filepath.Join(path, entry.Name())
			file, err := os.Open(filePath)
			if err != nil {
				log.Fatalf("cannot open %v", filePath)
				for _, oFile := range fileList {
					oFile.Close()
				}
				return nil, err
			}
			fileList = append(fileList, file)
		}
	}
	return fileList, nil
}
