package main
import (

	"bufio"
	"io"
	"os"
	"fmt"
	"github.com/Unknwon/goconfig"
	"github.com/wonderivan/logger"
	"strings"
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

	Cfg, err := goconfig.LoadConfigFile("config/main.ini")
	if err != nil{
		logger.Error("读取配置文件错误：", err)
	}
	inHos,_ := Cfg.GetValue("InboxMail", "host")
	logger.Debug("读取配置文件林成功["+inHos+"]")

	testStr,_ := Cfg.GetValue("Sys", "hostNameBlackList")
	logger.Debug(testStr)
	logger.Debug(strings.LastIndex(testStr, "abc.com"))

}