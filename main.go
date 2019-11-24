package main

import (
	"log"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap"
	"time"
)

//func main() {
//	log.Println("Connecting to server...")
//
//	// Connect to server
//	c, err := client.DialTLS("imap.126.com:993", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Connected")
//
//	// Don't forget to logout
//	defer c.Logout()
//
//	// Login
//	if err := c.Login("songylwq@126.com", "371246song"); err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Logged in")
//
//	// List mailboxes
//	mailboxes := make(chan *imap.MailboxInfo, 10)
//	done := make(chan error, 1)
//	go func () {
//		done <- c.List("", "*", mailboxes)
//	}()
//
//	log.Println("Mailboxes:")
//	for m := range mailboxes {
//		log.Println("* " + m.Name)
//	}
//
//	log.Println("******")
//
//	if err := <-done; err != nil {
//		log.Fatal(err)
//	}
//
//	log.Println("Select INBOX")
//
//	// Select INBOX
//	mbox, err := c.Select("INBOX", false)
//
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Flags for INBOX:", mbox.Flags)
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
//	log.Println("mbox.Messages:", mbox.Messages)
//
//	messages := make(chan *imap.Message, 10)
//	done = make(chan error, 1)
//	go func() {
//		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
//	}()
//
//	log.Println("Last 4 messages:")
//	for msg := range messages {
//		log.Println("* " + msg.Envelope.Subject)
//		//log.Println("from: ", msg.Envelope.From[0])
//		//log.Println("to: ", msg.Envelope.To[0])
//		//log.Println("ReplyTo: ", msg.Envelope.ReplyTo[0])
//	}
//
//	if err := <-done; err != nil {
//		log.Fatal(err)
//	}
//
//	log.Println("Done!")
//}


//func main() {
//	log.Println("Connecting to server...")
//
//	// Connect to server
//	c, err := client.DialTLS("imap.126.com:993", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Connected")
//
//	// Don't forget to logout
//	defer c.Logout()
//
//	// Login
//	if err := c.Login("songylwq@126.com", "371246song"); err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Logged in")
//
//	//// List mailboxes
//	//mailboxes := make(chan *imap.MailboxInfo, 10)
//	//done := make(chan error, 1)
//	//go func () {
//	//	done <- c.List("", "*", mailboxes)
//	//	//done <- c.List("", "INBOX", mailboxes)
//	//}()
//	//
//	//log.Println("Mailboxes:")
//	//for m := range mailboxes {
//	//	log.Println("* " + m.Name)
//	//}
//	//
//	//log.Println("******")
//	//
//	//if err := <-done; err != nil {
//	//	log.Fatal(err)
//	//}
//	//
//	//log.Println("Select INBOX")
//	//
//	//// Select INBOX
//	//mbox, err := c.Select("INBOX", false)
//	//
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//	//
//	//log.Println("Flags for INBOX:", mbox.Flags)
//	//
//	////mbox.Flags = []string{"\\Deleted"}
//	////mbox.PermanentFlags = []string{"\\Deleted"}
//	////
//	////log.Println("Flags 修改后:", mbox.Flags)
//	////log.Println("PermanentFlags 修改后:", mbox.PermanentFlags)
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
//	//log.Println("mbox.Messages : ", mbox.Messages)
//	//log.Println("mbox.PermanentFlags : ", mbox.PermanentFlags)
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
//			log.Fatal(err)
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
//			log.Fatal(err)
//		}
//
//		time.Sleep(3 * time.Second)
//
//		log.Println()
//		log.Println("Last 4 messages:")
//		for msg := range messages {
//			log.Println("* " + msg.Envelope.Subject)
//			log.Println("*,Flags : ", msg.Flags)
//			//log.Println("from: ", msg.Envelope.From[0])
//			//log.Println("to: ", msg.Envelope.To[0])
//			//log.Println("ReplyTo: ", msg.Envelope.ReplyTo[0])
//			//log.Println("MessageId: ", msg.Envelope.MessageId)
//		}
//
//	}
//
//	//回复邮件
//
//
//	log.Println("Done!")
//}


//
////删除邮件
//func main() {
//	log.Println("Connecting to server...")
//
//	// Connect to server
//	c, err := client.DialTLS("imap.126.com:993", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Connected")
//
//	// Don't forget to logout
//	defer c.Logout()
//
//	// Login
//	if err := c.Login("songylwq@126.com", "371246song"); err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Logged in")
//
//	// Select INBOX
//	mbox, err := c.Select("INBOX", false)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// We will delete the last message
//	if mbox.Messages == 0 {
//		log.Fatal("No message in mailbox")
//	}
//
//	seqset := new(imap.SeqSet)
//	seqset.AddNum(mbox.Messages)
//
//	// First mark the message as deleted
//	item := imap.FormatFlagsOp(imap.AddFlags, true)
//	flags := []interface{}{imap.SeenFlag}
//	if err := c.Store(seqset, item, flags, nil); err != nil {
//		log.Fatal(err)
//	}
//
//	log.Println("删除了")
//}


//查询未读邮件
func main() {
	log.Println("Connecting to server...")
	// Connect to server
	c, err := client.DialTLS("imap.126.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")
	// Don't forget to logout
	defer c.Logout()
	// Login
	if err := c.Login("songylwq@126.com", "371246song"); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	for {
		c.Select("test", false)

		//筛选所有未读的邮件
		criteria := imap.NewSearchCriteria()
		criteria.WithoutFlags = []string{"\\Seen"}
		uids, err := c.Search(criteria)
		if err != nil {
			log.Println(err)
		}

		log.Println(uids)

		if 0 < len(uids) {
			seqset := new(imap.SeqSet)
			seqset.AddNum(uids...)

			items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}
			messages := make(chan *imap.Message, len(uids))
			err = c.Fetch(seqset, items, messages)

			log.Println("c.Fetched")

			for msg := range messages {
				log.Println("* ", msg.Envelope.Subject, "--",  msg.Flags)
			}

			//将未读的邮件标记上已读
			seqset = new(imap.SeqSet)
			seqset.AddNum(uids...)
			// First mark the message as deleted
			item := imap.FormatFlagsOp(imap.AddFlags, true)
			flags := []interface{}{imap.SeenFlag}
			if err := c.Store(seqset, item, flags, nil); err != nil {
				log.Fatal(err)
			}
			log.Println("都标记成已读了")
		}

		log.Println("暂时没有未读邮件")
		time.Sleep(time.Second * 5)
	}
}





