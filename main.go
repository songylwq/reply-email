package main

import (
	"github.com/emersion/go-imap/client"
	"time"
	"github.com/go-gomail/gomail"
	"strconv"
	"github.com/emersion/go-imap"
	"os"
	"fmt"
	"bufio"
	"io"
	"math/rand"
	"crypto/tls"
	"github.com/Unknwon/goconfig"
	"strings"
	"github.com/wonderivan/logger"
	"encoding/json"
)

//邮件对象
type Email struct {
	Uid uint32
	FromAddr string
	ToAddr string
	Context string
}

//模板图片
type TempImg struct {
	//图片的base64数据
	Data string `json:"data"`
}

//文本素材切片
var TxtCont = make([]string, 1)
//文本素材长度
var TxtContNum = 0
//配置文件
var Cfg *goconfig.ConfigFile
//模板内容
var tempMap = make(map[string]string)
//模板图片内容
var TempImgList = make([]TempImg, 5)
//马甲号邮箱地址列表
var EmailAccMap = make(map[string]interface{})

func main() {
	//通过配置文件配置,初始化日志配置
	logger.SetLogger("config/log.json")
	//初始化配置文件
	initConfig()
	//初始化文本资源
	InitTxtCont()
	//初始化邮件模板
	initEmailTemp()
	//初始化模板图片
	initTempImg()
	//初始化马甲号邮箱地址列表
	initEmailAcc()

	roundTime := 1
	for {
		findUnReadMail(roundTime)
		roundTime ++
		logger.Debug("本次执行完成，30秒后重新启动")
		time.Sleep(time.Second * 30)
	}


	////清空所有的邮件，包括垃圾邮件等
	//clearAllEmail()

}

//初始化配置文件
func initConfig() {
	var err error
	Cfg, err = goconfig.LoadConfigFile("config/main.ini")
	if err != nil{
		logger.Error("读取配置文件错误：", err)
	}
	inHos,_ := Cfg.GetValue("InboxMail", "host")
	logger.Debug("读取配置文件林成功["+inHos+"]")
}


//查询未读
func findUnReadMail(roundTime int) {
	//系统整体异常处理
	defer func() {
		if err := recover(); err != nil {
			logger.Debug("==============================================")
			logger.Debug("回复邮箱逻辑异常：", err)
			logger.Debug("==============================================")
		}
	}()
	
	logger.Debug("==================== >>>>>>>>>>>>>>>>>>>>>>> 开始第["+strconv.Itoa(roundTime)+"]轮登录邮箱操作")

	logger.Debug("开始查询新邮件")
	inHos,_ := Cfg.GetValue("InboxMail", "host")
	inPost,_ := Cfg.GetValue("InboxMail", "port")
	c, err := client.DialTLS(inHos+":"+inPost, nil)
	if err != nil {
		logger.Error(err)
	}
	logger.Debug("获取链接成功")
	defer c.Logout()

	inName,_ := Cfg.GetValue("InboxMail", "name")
	inPwd,_ := Cfg.GetValue("InboxMail", "pwd")

	if err := c.Login(inName, inPwd); err != nil {
		logger.Error(err)
	}
	logger.Debug("登录邮箱成功")

	//每5秒查询一下收件箱，读取配置
	readInBoxTimeStr,_ := Cfg.GetValue("Sys", "readInBoxTime")
	readInBoxTime, err := strconv.ParseInt(readInBoxTimeStr, 10, 64)

	//每获取多少次收件箱，就重新登录一次
	reloginNumStr,_ := Cfg.GetValue("Sys", "reloginNum")
	reloginNum, err := strconv.Atoi(reloginNumStr)


	for i:=1; i<reloginNum; i++ {
		now  := time.Now()
		//now.Format 方法格式化
		timeStr := fmt.Sprint(now.Format("2006-01-02 15:04:05"))
		logger.Debug(readInBoxTimeStr+"秒后读取收件箱["+timeStr+"]...")
		time.Sleep(time.Second * (time.Duration(readInBoxTime)))
		logger.Debug("")
		logger.Debug("======= 开始第["+strconv.Itoa(roundTime)+"]轮["+strconv.Itoa(i)+"]次读取收件箱 ======= ")

		c.Select("INBOX", false)

		logger.Debug("读取收件箱 Select 完成")

		//筛选所有未读的邮件
		criteria := imap.NewSearchCriteria()
		criteria.WithoutFlags = []string{"\\Seen"}
		uids, err := c.Search(criteria)
		if err != nil {
			logger.Debug(err)
		}
		logger.Debug("读取收件箱 Search 完成")

		if 0 < len(uids) {
			seqset := new(imap.SeqSet)
			seqset.AddNum(uids...)

			items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}
			messages := make(chan *imap.Message, len(uids))
			err = c.Fetch(seqset, items, messages)

			logger.Debug("查询收件箱完成，收件箱邮件数量：", len(uids))

			for msg := range messages {
				hostNameBlackList,_ := Cfg.GetValue("Sys", "hostNameBlackList")
				//回复马甲号地址限制，不给指定黑名单的域名发送邮件
				if -1 < strings.LastIndex(hostNameBlackList, msg.Envelope.From[0].HostName) {
					logger.Debug("XXXXXXXXXXXXX 回复邮箱域名地址黑名单，不做回复：From: ",
						"["+msg.Envelope.From[0].MailboxName + "@" +msg.Envelope.From[0].HostName+"]",
						" TO: ",
						"["+msg.Envelope.To[0].MailboxName + "@" +msg.Envelope.To[0].HostName+"]",)
					continue
				}

				hostNameWhiteList,_ := Cfg.GetValue("Sys", "hostNameWhiteList")

				//马甲号地址只在邮件白名单回信,不在白名单里面不发送
				if 0 > strings.LastIndex(hostNameWhiteList, msg.Envelope.From[0].HostName) {
					logger.Debug("XXXXXXXXXXXXX 不在回复邮箱域名白名单，不回复的地址：From: ",
						"["+msg.Envelope.From[0].MailboxName + "@" +msg.Envelope.From[0].HostName+"]",
						" TO: ",
						"["+msg.Envelope.To[0].MailboxName + "@" +msg.Envelope.To[0].HostName+"]",)
					continue
				}

				//只自动回复马甲号列表里的邮件，怎么实现再考虑
				_,isHave := EmailAccMap[msg.Envelope.From[0].HostName]
				if isHave {

				}

				logger.Debug("1秒后开始回复下一封邮件...")
				time.Sleep(time.Second)

				go func(msg *imap.Message){
					logger.Debug("开始回复邮件...")
					//参考报文 {"PersonalName":"\" 芝麻开花 \"","AtDomainList":"","MailboxName":"350956892","HostName":"qq.com"}
					//我的收件账户地址
					repEmailMyAddr := msg.Envelope.To[0].MailboxName + "@" + msg.Envelope.To[0].HostName
					//马甲号地址
					repEmailMajiaAddr := msg.Envelope.From[0].MailboxName + "@" + msg.Envelope.From[0].HostName

					logger.Debug("开始回复邮件：", repEmailMajiaAddr)
					err = SendMail(repEmailMyAddr, repEmailMajiaAddr, msg.Envelope.Subject)
					if nil != err {
						logger.Debug(err)
						logger.Error("邮件回复失败")
					}

					logger.Debug("邮件回复完成：", repEmailMajiaAddr)
				}(msg)
			}

			//将未读的邮件标记上已读
			seqset = new(imap.SeqSet)
			seqset.AddNum(uids...)
			// First mark the message as deleted
			item := imap.FormatFlagsOp(imap.AddFlags, true)
			//标记已读，并删除
			flags := []interface{}{imap.SeenFlag, imap.DeletedFlag}
			if err := c.Store(seqset, item, flags, nil); err != nil {
				logger.Error(err)
			}
			logger.Debug("都标记成已读了")
		}

		logger.Debug("暂时没有未读邮件")
	}
	logger.Debug("==================== >>>>>>>>>>>>>>>>>>>>>>> 第["+strconv.Itoa(roundTime)+"]轮登录邮箱操作结束")
	logger.Debug("")
}


//发送邮件
func SendMail(mailFrom string, mailTo string, title string) error {
	//定义邮箱服务器连接信息，如果是网易邮箱 pass填密码，qq邮箱填授权码
	//mailConn := map[string]string{
	//	"user": "songylwq@126.com",
	//	//"pass": "19861113chyylwq",
	//	"pass": "371246song",
	//	"host": "smtp.126.com",
	//	"port": "465",
	//}

	//mailConn := map[string]string{
	//	"user": "songyl",
	//	"pass": "4@2xade53",
	//	"host": "mail.wangzihan.xyz",
	//	//"host": "162.244.77.183",
	//	"port": "465",
	//}

	sendName,_ := Cfg.GetValue("SendMail", "name")
	sendPwd,_ := Cfg.GetValue("SendMail", "pwd")
	sendHost,_ := Cfg.GetValue("SendMail", "host")
	sendPort,_ := Cfg.GetValue("SendMail", "port")

	port, _ := strconv.Atoi(sendPort) //转换端口类型为int
	d := gomail.NewDialer(sendHost, port, sendName, sendPwd)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	sendCloser,err := d.Dial()

	defer sendCloser.Close()

	if err != nil {
		logger.Error("创建sendCloser异常：%s\n", err.Error())
	}
	logger.Debug(mailTo+"-创建sendCloser成功")

	randSeed := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i:=0; i<1; i++ {
		m := gomail.NewMessage()
		m.SetAddressHeader("From", mailFrom, mailFrom)
		m.SetAddressHeader("To", mailTo, mailTo)
		m.SetHeader("Subject", "Re:"+title)

		//随机使用模板
		tempNumStr,_ := Cfg.GetValue("Sys", "tempNum")
		tempNum, _ := strconv.Atoi(tempNumStr)
		randZh := rand.New(rand.NewSource(time.Now().UnixNano()))
		randTempNum := randZh.Intn(tempNum) + 1
		logger.Debug(mailTo+"-回复邮件采用模板：", randTempNum)
		bodyStr := tempMap["temp_"+strconv.Itoa(randTempNum)]
		//替换图片、文本内容
		reqEmailCont1 := GetRandromTxt(randSeed.Intn(100)+100)
		bodyStr = strings.Replace(bodyStr, "#cont1#", reqEmailCont1, 3 )
		reqEmailCont2 := GetRandromTxt(randSeed.Intn(100)+100)
		bodyStr = strings.Replace(bodyStr, "#cont2#", reqEmailCont2, 3 )
		reqEmailCont3 := GetRandromTxt(randSeed.Intn(100)+100)
		bodyStr = strings.Replace(bodyStr, "#cont3#", reqEmailCont3, 3 )

		//替换图片
		bodyStr = strings.Replace(bodyStr, "#img1#", TempImgList[randSeed.Intn(len(TempImgList))].Data, 3 )
		bodyStr = strings.Replace(bodyStr, "#img2#", TempImgList[randSeed.Intn(len(TempImgList))].Data, 3 )
		bodyStr = strings.Replace(bodyStr, "#img3#", TempImgList[randSeed.Intn(len(TempImgList))].Data, 3 )

		m.SetBody("text/html", bodyStr)
		sendCloser.Send(mailFrom, []string{mailTo}, m)
		logger.Debug("========>>>>> 发送邮件from["+mailFrom+"],to["+mailTo+"]成功")
	}



	return err
}

//初始化文本内容，读取到内存里面
func InitTxtCont(){
	filePathUrl := fmt.Sprintf("data%ctxtcontent", os.PathSeparator)
	logger.Debug("开始读取[%s]", filePathUrl)
	ReadTxtCont(filePathUrl, 10000, TxtHandle)
	TxtContNum = len(TxtCont)
	logger.Debug("读取文本素材完成,共读取[%v]条", len(TxtCont))
}

//将文本内容读取到内存里
func TxtHandle(line []byte) {
	//os.Stdout.Write(line)
	//logger.Debug(string(line))
	//将文本素材写入内存
	TxtCont = append(TxtCont, string(line))
}

//读取文本内容
func ReadTxtCont(filePth string, bufSize int, hookfn func([]byte)) error {
	f, err := os.Open(filePth)
	if err != nil {
		return err
	}
	defer f.Close()

	bfRd := bufio.NewReader(f)
	for {
		line, err := bfRd.ReadBytes('\n')
		hookfn(line) //放在错误处理前面，即使发生错误，也会处理已经读取到的数据。
		if err != nil { //遇到任何错误立即返回，并忽略 EOF 错误信息
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}

//随机获取一个指定长度的文本
func GetRandromTxt(getCount int) string {
	//先休眠300纳秒要不然连续读取时会出现数据相同的问题
	time.Sleep(300*time.Nanosecond)

	//获取随机数
	randZh := rand.New(rand.NewSource(time.Now().UnixNano()))
	randId := randZh.Intn(TxtContNum)
	reStr := TxtCont[randId]
	for getCount > len([]rune(reStr)) {
		randZh = rand.New(rand.NewSource(time.Now().UnixNano()+int64(len([]rune(reStr)))))
		randId = randZh.Intn(TxtContNum)
		reStr += TxtCont[randId]
	}

	nameRune := []rune(reStr)

	return string(nameRune[1 : getCount])
}

//清空所有的邮件，包括垃圾邮件等
func clearAllEmail(){
	defer func() {
		if err := recover(); err != nil {
			logger.Debug("==============================================")
			logger.Debug("清空邮箱逻辑异常：", err)
			logger.Debug("==============================================")
		}
	}()

	logger.Debug("=======>>>>> 开始删除邮件的登录邮箱操作")

	logger.Debug("开始查询新邮件")
	inHos,_ := Cfg.GetValue("InboxMail", "host")
	inPost,_ := Cfg.GetValue("InboxMail", "port")
	c, err := client.DialTLS(inHos+":"+inPost, nil)
	if err != nil {
		logger.Error(err)
	}
	logger.Debug("获取链接成功")
	defer c.Logout()

	inName,_ := Cfg.GetValue("InboxMail", "name")
	inPwd,_ := Cfg.GetValue("InboxMail", "pwd")

	if err := c.Login(inName, inPwd); err != nil {
		logger.Error(err)
	}
	logger.Debug("登录邮箱成功")

	//每5秒查询一下收件箱，读取配置
	readInBoxTimeStr,_ := Cfg.GetValue("Sys", "readInBoxTime")
	readInBoxTime, err := strconv.ParseInt(readInBoxTimeStr, 10, 64)

	//每获取多少次收件箱，就重新登录一次
	reloginNumStr,_ := Cfg.GetValue("Sys", "reloginNum")
	reloginNum, err := strconv.Atoi(reloginNumStr)

	for i:=1; i<reloginNum; i++ {
		now  := time.Now()
		//now.Format 方法格式化
		timeStr := fmt.Sprint(now.Format("2006-01-02 15:04:05"))
		logger.Debug(readInBoxTimeStr+"秒后读取收件箱["+timeStr+"]...")
		time.Sleep(time.Second * (time.Duration(readInBoxTime)))
		logger.Debug("")
		logger.Debug("======= 开始第["+strconv.Itoa(i)+"]次读取收件箱 ======= ")

		c.Select("INBOX", false)

		logger.Debug("读取收件箱 Select 完成")

		//筛选所有未读的邮件
		criteria := imap.NewSearchCriteria()
		criteria.WithoutFlags = []string{imap.SeenFlag, imap.DeletedFlag}
		uidsOrd, err := c.Search(criteria)
		if err != nil {
			logger.Debug(err)
		}
		logger.Debug("读取收件箱 Search 完成 uidsOrd：", len(uidsOrd))
		uids := uidsOrd
		logger.Debug("uids:",uids)

		if 0 < len(uids) {
			logger.Debug("开始处理：", len(uids))
			seqset := new(imap.SeqSet)
			seqset.AddNum(uids...)

			items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}
			messages := make(chan *imap.Message, len(uids))
			err = c.Fetch(seqset, items, messages)

			logger.Debug("查询收件箱完成，收件箱邮件数量：", len(uids))

			//将未读的邮件标记上已读
			seqset = new(imap.SeqSet)
			seqset.AddNum(uids...)
			// First mark the message as deleted
			item := imap.FormatFlagsOp(imap.AddFlags, true)
			//标记已读，并删除
			flags := []interface{}{imap.SeenFlag, imap.DeletedFlag}
			if err := c.Store(seqset, item, flags, nil); err != nil {
				logger.Error(err)
			}
			logger.Debug("都标记成已读了并删除")
		}

		logger.Debug("暂时没有未读邮件")
	}
	logger.Debug("==================== >>>>>>>>>>>>>>>>>>>>>>> 邮箱操作结束")
	logger.Debug("")
}

//初始化模板图片
func initTempImg() error {
	filePathUrl := fmt.Sprintf("data%cimg.json", os.PathSeparator)
	logger.Debug("开始读取模板图片数据[%s]", filePathUrl)
	tempStr,err := ReadFileCont(filePathUrl)
	if nil != err {
		return err
	}
	json.Unmarshal([]byte(tempStr), &TempImgList)
	return nil
}

//初始化马甲号邮箱地址列表
func initEmailAcc() error {
	filePathUrl := fmt.Sprintf("data%cemail_acc.json", os.PathSeparator)
	logger.Debug("开始读取马甲号邮箱地址列表[%s]", filePathUrl)
	tempStr,err := ReadFileCont(filePathUrl)
	if nil != err {
		return err
	}

	tmpEmailAccList := make([]string, 5)
	json.Unmarshal([]byte(tempStr), &tmpEmailAccList)
	for _,data := range tmpEmailAccList {
		EmailAccMap[data] = data
	}
	logger.Debug("读取马甲号邮箱地址列表[%s]完成，共[%v]条", filePathUrl, len(tmpEmailAccList))
	return nil
}


//初始化邮件模板
func initEmailTemp() error{
	logger.Debug("开始读取初始化邮件模板")
	//读取模板配置
	tempNumStr,_ := Cfg.GetValue("Sys", "tempNum")
	tempNum, _ := strconv.Atoi(tempNumStr)

	for t:=1;t<=tempNum;t++ {
		filePathUrl := fmt.Sprintf("data%ctemp_%v.txt", os.PathSeparator, t)
		logger.Debug("开始读取邮件模板[%s]", filePathUrl)
		tempStr,err := ReadFileCont(filePathUrl)

		if err != nil {
			return err
		}
		tempMap["temp_"+strconv.Itoa(t)] = tempStr
	}
	return nil
}

//读取文件内容
func ReadFileCont(filePth string) (string,error) {
	f, err := os.Open(filePth)
	if err != nil {
		return "",err
	}
	defer f.Close()

	bfRd := bufio.NewReader(f)
	res := ""
	for {
		line, err := bfRd.ReadBytes('\n')
		res += string(line)

		if err != nil { //遇到任何错误立即返回，并忽略 EOF 错误信息
			if err == io.EOF {
				return res,nil
			}
			return "",err
		}
	}
}



//func main() {
//	logger.Debug("Connecting to server...")
//
//	// Connect to server
//	c, err := client.DialTLS("imap.126.com:993", nil)
//	if err != nil {
//		logger.Error(err)
//	}
//	logger.Debug("Connected")
//
//	// Don't forget to logout
//	defer c.Logout()
//
//	// Login
//	if err := c.Login("songylwq@126.com", "371246song"); err != nil {
//		logger.Error(err)
//	}
//	logger.Debug("Logged in")
//
//	// List mailboxes
//	mailboxes := make(chan *imap.MailboxInfo, 10)
//	done := make(chan error, 1)
//	go func () {
//		done <- c.List("", "*", mailboxes)
//	}()
//
//	logger.Debug("Mailboxes:")
//	for m := range mailboxes {
//		logger.Debug("* " + m.Name)
//	}
//
//	logger.Debug("******")
//
//	if err := <-done; err != nil {
//		logger.Error(err)
//	}
//
//	logger.Debug("Select INBOX")
//
//	// Select INBOX
//	mbox, err := c.Select("INBOX", false)
//
//	if err != nil {
//		logger.Error(err)
//	}
//	logger.Debug("Flags for INBOX:", mbox.Flags)
//
//	// Get the last 4 messages
//	from := uint32(1)
//	to := mbox.Messages
//	if mbox.Messages > 5 {
//		// We're using unsigned integers here, only substract if the result is > 0
//		from = mbox.Messages - 5
//	}
//	seqset := new(imap.SeqSet)
//	seqset.AddRange(from, to)
//
//	logger.Debug("mbox.Messages:", mbox.Messages)
//
//	messages := make(chan *imap.Message, 10)
//	done = make(chan error, 1)
//	go func() {
//		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
//	}()
//
//	logger.Debug("Last 4 messages:")
//	for msg := range messages {
//		logger.Debug("* " + msg.Envelope.Subject)
//		//logger.Debug("from: ", msg.Envelope.From[0])
//		//logger.Debug("to: ", msg.Envelope.To[0])
//		//logger.Debug("ReplyTo: ", msg.Envelope.ReplyTo[0])
//	}
//
//	if err := <-done; err != nil {
//		logger.Error(err)
//	}
//
//	logger.Debug("Done!")
//}


//func main() {
//	logger.Debug("Connecting to server...")
//
//	// Connect to server
//	c, err := client.DialTLS("imap.126.com:993", nil)
//	if err != nil {
//		logger.Error(err)
//	}
//	logger.Debug("Connected")
//
//	// Don't forget to logout
//	defer c.Logout()
//
//	// Login
//	if err := c.Login("songylwq@126.com", "371246song"); err != nil {
//		logger.Error(err)
//	}
//	logger.Debug("Logged in")
//
//	//// List mailboxes
//	//mailboxes := make(chan *imap.MailboxInfo, 10)
//	//done := make(chan error, 1)
//	//go func () {
//	//	done <- c.List("", "*", mailboxes)
//	//	//done <- c.List("", "INBOX", mailboxes)
//	//}()
//	//
//	//logger.Debug("Mailboxes:")
//	//for m := range mailboxes {
//	//	logger.Debug("* " + m.Name)
//	//}
//	//
//	//logger.Debug("******")
//	//
//	//if err := <-done; err != nil {
//	//	logger.Error(err)
//	//}
//	//
//	//logger.Debug("Select INBOX")
//	//
//	//// Select INBOX
//	//mbox, err := c.Select("INBOX", false)
//	//
//	//if err != nil {
//	//	logger.Error(err)
//	//}
//	//
//	//logger.Debug("Flags for INBOX:", mbox.Flags)
//	//
//	////mbox.Flags = []string{"\\Deleted"}
//	////mbox.PermanentFlags = []string{"\\Deleted"}
//	////
//	////logger.Debug("Flags 修改后:", mbox.Flags)
//	////logger.Debug("PermanentFlags 修改后:", mbox.PermanentFlags)
//	//
//	//// Get the last 4 messages
//	//from := uint32(1)
//	//to := mbox.Messages
//	//if mbox.Messages > 2 {
//	//	// We're using unsigned integers here, only substract if the result is > 0
//	//	from = mbox.Messages - 2
//	//}
//	//seqset := new(imap.SeqSet)
//	//seqset.AddRange(from, to)
//	//
//	//logger.Debug("mbox.Messages : ", mbox.Messages)
//	//logger.Debug("mbox.PermanentFlags : ", mbox.PermanentFlags)
//
//
//
//
//	//done := make(chan error, 1)
//
//	for i:=1;i<10;i++{
//		//读取服务器的收件箱邮件的数量，确定读取邮件的起止角标
//		mbox, err := c.Select("INBOX", false)
//
//		if err != nil {
//			logger.Error(err)
//		}
//
//		from := uint32(1)
//		to := mbox.Messages
//		if mbox.Messages > 2 {
//			// We're using unsigned integers here, only substract if the result is > 0
//			from = mbox.Messages - 2
//		}
//		seqset := new(imap.SeqSet)
//		seqset.AddRange(from, to)
//
//		messages := make(chan *imap.Message, 10)
//		err = c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}, messages)
//
//		if err != nil {
//			logger.Error(err)
//		}
//
//		time.Sleep(3 * time.Second)
//
//		logger.Debug()
//		logger.Debug("Last 4 messages:")
//		for msg := range messages {
//			logger.Debug("* " + msg.Envelope.Subject)
//			logger.Debug("*,Flags : ", msg.Flags)
//			//logger.Debug("from: ", msg.Envelope.From[0])
//			//logger.Debug("to: ", msg.Envelope.To[0])
//			//logger.Debug("ReplyTo: ", msg.Envelope.ReplyTo[0])
//			//logger.Debug("MessageId: ", msg.Envelope.MessageId)
//		}
//
//	}
//
//	//回复邮件
//
//
//	logger.Debug("Done!")
//}


//
////删除邮件
//func main() {
//	logger.Debug("Connecting to server...")
//
//	// Connect to server
//	c, err := client.DialTLS("imap.126.com:993", nil)
//	if err != nil {
//		logger.Error(err)
//	}
//	logger.Debug("Connected")
//
//	// Don't forget to logout
//	defer c.Logout()
//
//	// Login
//	if err := c.Login("songylwq@126.com", "371246song"); err != nil {
//		logger.Error(err)
//	}
//	logger.Debug("Logged in")
//
//	// Select INBOX
//	mbox, err := c.Select("INBOX", false)
//	if err != nil {
//		logger.Error(err)
//	}
//
//	// We will delete the last message
//	if mbox.Messages == 0 {
//		logger.Error("No message in mailbox")
//	}
//
//	seqset := new(imap.SeqSet)
//	seqset.AddNum(mbox.Messages)
//
//	// First mark the message as deleted
//	item := imap.FormatFlagsOp(imap.AddFlags, true)
//	flags := []interface{}{imap.SeenFlag}
//	if err := c.Store(seqset, item, flags, nil); err != nil {
//		logger.Error(err)
//	}
//
//	logger.Debug("删除了")
//}


////查询未读邮件
//func main() {
//	logger.Debug("Connecting to server...")
//	// Connect to server
//	c, err := client.DialTLS("imap.126.com:993", nil)
//	if err != nil {
//		logger.Error(err)
//	}
//	logger.Debug("Connected")
//	// Don't forget to logout
//	defer c.Logout()
//	// Login
//	if err := c.Login("songylwq@126.com", "371246song"); err != nil {
//		logger.Error(err)
//	}
//	logger.Debug("Logged in")
//
//	for {
//		c.Select("test", false)
//
//		//筛选所有未读的邮件
//		criteria := imap.NewSearchCriteria()
//		criteria.WithoutFlags = []string{"\\Seen"}
//		uids, err := c.Search(criteria)
//		if err != nil {
//			logger.Debug(err)
//		}
//
//		logger.Debug(uids)
//
//		if 0 < len(uids) {
//			seqset := new(imap.SeqSet)
//			seqset.AddNum(uids...)
//
//			items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}
//			messages := make(chan *imap.Message, len(uids))
//			err = c.Fetch(seqset, items, messages)
//
//			logger.Debug("c.Fetched")
//
//			for msg := range messages {
//				logger.Debug("* ", msg.Envelope.Subject, "--",  msg.Flags)
//			}
//
//			//将未读的邮件标记上已读
//			seqset = new(imap.SeqSet)
//			seqset.AddNum(uids...)
//			// First mark the message as deleted
//			item := imap.FormatFlagsOp(imap.AddFlags, true)
//			flags := []interface{}{imap.SeenFlag}
//			if err := c.Store(seqset, item, flags, nil); err != nil {
//				logger.Error(err)
//			}
//			logger.Debug("都标记成已读了")
//		}
//
//		logger.Debug("暂时没有未读邮件")
//		time.Sleep(time.Second * 5)
//	}
//}


