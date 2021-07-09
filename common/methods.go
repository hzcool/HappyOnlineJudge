package common

import (
	"archive/zip"
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string { //生成长度为n的包含字母的字符串
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func GetMD5OfStr(str string) string { //获取一个字符串的MD5
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

//以下为文件的一系列操作
func RemoveContents(dir string) error { //移走一个文件夹下的所有文件
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func GetFilesOfSomeExts(dir string, ext []string) []string { //获取一个文件下想要的扩展名文件路径
	d, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return nil
	}
	ret := make([]string, 0)
	for _, name := range names {
		fileExt := filepath.Ext(name)
		for _, item := range ext {
			if fileExt == item {
				ret = append(ret, filepath.Join(dir, name))
			}
		}
	}
	return ret
}

func RemoveTestDatas(dir string) error { //删除一个文件夹下的.in .out .ans
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		ext := filepath.Ext(name)
		if ext == ".in" || ext == ".out" || ext == ".ans" {
			os.Remove(filepath.Join(dir, name))
		}
	}
	return nil
}

func JsonFileToMap(path string) map[string]interface{} { //将一个是json文件转化成map
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	bt, err := ioutil.ReadAll(file)
	if err != nil {
		return nil
	}
	var mp map[string]interface{}
	if err := json.Unmarshal(bt, &mp); err != nil {
		return nil
	}
	return mp
}

func GetContent(path string) (string, error) { //读取文件内容
	if yes, _ := PathExists(path); !yes {
		return "", nil
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	bt, err := ioutil.ReadAll(file)
	if err != nil {
		return "", nil
	}
	return string(bt), err
}

func WriteToFile(path string, content string) error { //写入文件
	return ioutil.WriteFile(path, []byte(content), os.ModePerm)
}

func PathExists(path string) (bool, error) { //判断文件是否存在
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func StoreZipFile(zipFIle *zip.File, dest string) error { //保存zip文件到指定文件夹
	inFile, err := zipFIle.Open()
	if err != nil {
		return err
	}
	defer inFile.Close()
	outFile, err2 := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipFIle.Mode())
	if err2 != nil {
		return err2
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, inFile)
	if err != nil {
		return err
	}
	return nil
}
func CompressToZip(srcFiles []string, dest string) error { //将所有的原文件路径压缩
	d, _ := os.Create(dest)
	defer d.Close()
	w := zip.NewWriter(d)
	defer w.Close()
	for _, filePath := range srcFiles {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		if err2 := compress(file, w); err2 != nil {
			return err2
		}
	}
	return nil
}
func compress(file *os.File, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, file)
	file.Close()
	if err != nil {
		return err
	}
	return nil
}

func FileSize(path string) uint { //求文件大小
	fileInfo, _ := os.Stat(path)
	return uint(fileInfo.Size())
}

func MD5(path string) (string, error) { //求文件的md5(每行去掉首尾的空格)
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	var content string
	for {
		line, err := rd.ReadString('\n')
		content += strings.TrimSpace(line)
		if err != nil || io.EOF == err {
			break
		}
	}
	h := md5.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil)), nil
}

func CopyDir(srcPath string, destPath string) error { //将整个文件夹进行复制
	cmd := exec.Command("cp", "-rf", srcPath, destPath)
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}
