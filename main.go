package main

import (
	"log"

	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap"
)

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

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func () {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		log.Println("* " + m.Name)
	}

	log.Println("******")

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Select INBOX")

	// Select INBOX
	mbox, err := c.Select("[]INBOX", false)

	if err != nil {
		log.Fatal(err)
	}

	c.Search()

	log.Println("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 5 {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - 5
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	log.Println("mbox.Messages:", mbox.Messages)

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	log.Println("Last 4 messages:")
	for msg := range messages {
		log.Println("* " + msg.Envelope.Subject)
		log.Println("from: ", msg.Envelope.From[0])
		log.Println("to: ", msg.Envelope.To[0])
		log.Println("ReplyTo: ", msg.Envelope.ReplyTo[0])
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}