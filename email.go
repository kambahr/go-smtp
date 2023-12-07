package gosmtp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
)

type EmailAddr struct {
	Name    string // Name can be omitted
	Address string
}

type MailItem struct {
	// email adress of the account used to send email
	// (may be the same as the primary email-account)
	From EmailAddr

	To       []EmailAddr
	CC       []EmailAddr
	BCC      []EmailAddr
	Subject  string
	HTMLBody string
}

type MailCredentials struct {
	Host     string // IP address or host-name
	PortNo   int
	UserName string // the primary user-name for the smtp account
	Password string
}

func emailValid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// SendMail send multiple emails (to,cc, and bcc) via an smtp host.
func SendMail(m MailItem, mc MailCredentials) error {

	var headerTo []string
	var headerCC []string

	if mc.Host == "" || mc.PortNo < 1 ||
		mc.UserName == "" || len(m.To) == 0 || !emailValid(m.From.Address) {
		return errors.New("invalid settings")
	}

	fromHeader := fmt.Sprintf(`"%s" <%s>`, m.From.Name, m.From.Address)

	// prepare the headers

	// To
	for i := 0; i < len(m.To); i++ {
		s := fmt.Sprintf(`"%s" <%s>`, m.To[i].Name, m.To[i].Address)
		headerTo = append(headerTo, s)
	}

	// CC
	for i := 0; i < len(m.CC); i++ {
		s := fmt.Sprintf(`"%s" <%s>`, m.CC[i].Name, m.CC[i].Address)
		headerCC = append(headerCC, s)
	}

	// smtp message format:
	// Each header goes on a separate line; and the body of the body
	// of the email is placed at the last line:
	//
	//   From: "<name goes here>" <<email goes here>>
	//   To:   "<name goes here>" <<email goes here>>
	//     for multiple recipients, add another string separated by comma (on the same line)
	//     e.g.
	//         "<name_1>" <<email_1>>,"<name_2>" <<email_2>>,...
	//
	//   Cc:   "<name goes here>" <<email goes here>>
	//     for multiple recipients: same as the above
	//
	//   Subject: <subject goes here>
	//   Content-Type: text/html; charset=utf-8
	//   <all of the html body follows here on a new line>

	// concat lines into a single string
	msgHeaders := ""
	msgHeaders = fmt.Sprintf("%s%s:%s\r\n", msgHeaders, "From", fromHeader)
	msgHeaders = fmt.Sprintf("%s%s:%s\r\n", msgHeaders, "To", strings.Join(headerTo, ","))
	msgHeaders = fmt.Sprintf("%s%s:%s\r\n", msgHeaders, "Cc", strings.Join(headerCC, ","))
	msgHeaders = fmt.Sprintf("%s%s:%s\r\n", msgHeaders, "Subject", m.Subject)
	msgHeaders = fmt.Sprintf("%sContent-Type:text/html; charset=utf-8\r\n", msgHeaders)
	msgHeaders = fmt.Sprintf("%s\r\n%s", msgHeaders, m.HTMLBody)

	servername := fmt.Sprintf("%s:%d", mc.Host, mc.PortNo)
	auth := smtp.PlainAuth("", mc.UserName, mc.Password, mc.Host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         mc.Host}

	// create a connection
	c, err := smtp.Dial(servername)
	if err != nil {
		return err
	}

	c.StartTLS(tlsconfig)

	// Autheticate
	if err = c.Auth(auth); err != nil {
		return err
	}

	// From (sender's email address)
	if err = c.Mail(m.From.Address); err != nil {
		return err
	}

	// Add all recipients using the same connection

	// To
	for i := 0; i < len(m.To); i++ {
		s := fmt.Sprintf(`"%s" <%s>`, m.To[i].Name, m.To[i].Address)
		if err = c.Rcpt(s); err != nil {
			return err
		}
	}

	// CC
	for i := 0; i < len(m.CC); i++ {
		s := fmt.Sprintf(`"%s" <%s>`, m.CC[i].Name, m.CC[i].Address)
		if err = c.Rcpt(s); err != nil {
			return err
		}
	}

	// BCC (not included in the header)
	for i := 0; i < len(m.BCC); i++ {
		s := fmt.Sprintf(`"%s" <%s>`, m.BCC[i].Name, m.BCC[i].Address)
		if err = c.Rcpt(s); err != nil {
			return err
		}
	}

	// Add the Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	// write the bytes of the headers to the data writer.
	if _, err = w.Write([]byte(msgHeaders)); err != nil {
		return err
	}

	// close the write (flush)
	if err = w.Close(); err != nil {
		return err

	}

	// close the connection
	if err = c.Quit(); err != nil {
		return err
	}

	return nil
}
