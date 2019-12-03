package main
import (

	"bufio"
	"io"
	"os"
	"fmt"
)

func processBlock(line []byte) {
	//os.Stdout.Write(line)
	fmt.Println(string(line))
}

func ReadBlock(filePth string, bufSize int, hookfn func([]byte)) error {
	f, err := os.Open(filePth)
	fmt.Println("开始打开文件")
	if err != nil {
		fmt.Println("打开文件失败：", err)
		return err
	}
	defer f.Close()

	buf := make([]byte, bufSize) //一次读取多少个字节
	bfRd := bufio.NewReader(f)
	for {
		n, err := bfRd.Read(buf)
		hookfn(buf[:n]) // n 是成功读取字节数

		if err != nil { //遇到任何错误立即返回，并忽略 EOF 错误信息
			if err == io.EOF {
				return nil
			}
			return err
		}
	}

	return nil
}

func main() {
	filePathUrl := fmt.Sprintf("data%c文本模板", os.PathSeparator)
	fmt.Println("开始读取:", filePathUrl)
	ReadBlock(filePathUrl, 10000, processBlock)
}