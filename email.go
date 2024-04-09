package gosmtp

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
)

// emailValid validates an email.
func emailValid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// getRandString returns a random string.
func getRandString() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x%x%x", b[0:4], b[4:6], b[6:8])
}

// addAttachments creates headers and 64based text;
// and adds it to the msg headers.
func addAttachments(bndry string, attch []string, msgHeaders string) string {

	for i := 0; i < len(attch); i++ {
		fPath := attch[i]
		bFile, err := os.ReadFile(fPath)
		fileName := filepath.Base(fPath)
		if err != nil {
			continue
		}

		// convert bytes to base64 string
		b64Str := base64.StdEncoding.EncodeToString(bFile)

		// this is the separator for each section
		msgHeaders = fmt.Sprintf("%s--%s\r\n", msgHeaders, bndry)

		msgHeaders = fmt.Sprintf("%sContent-Type: application/octet-stream; name=\"%s\"\r\n", msgHeaders, fileName)
		msgHeaders = fmt.Sprintf("%sContent-Disposition: attachment; filename=\"%s\"\r\n", msgHeaders, fileName)
		msgHeaders = fmt.Sprintf("%sContent-Transfer-Encoding: base64\r\n", msgHeaders)

		msgHeaders = fmt.Sprintf("%s\r\n%s\r\n", msgHeaders, b64Str)
	}

	return msgHeaders
}

// prepareHeaders creates message headers.
// smtp message format:
//
//	 From: "<name goes here>" <<email goes here>>
//	 To:   "<name goes here>" <<email goes here>>
//	   for multiple recipients, just add another string separated by comma (on the same line)
//	   e.g.
//	       "<name_1>" <<email_1>>,"<name_2>" <<email_2>>,...
//
//	 Cc:   "<name goes here>" <<email goes here>>
//	   for multiple recipients: same as the above
//
//	 Subject: <subject goes here>
//
//	for multipart (body text + attachments) see : https://www.w3.org/Protocols/rfc1341/7_2_Multipart.html
func prepareHeaders(m MailItem) string {

	var headerTo []string
	var headerCC []string

	fromHeader := fmt.Sprintf(`"%s" <%s>`, m.From.Name, m.From.Address)

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

	bndry := getRandString()
	msgHeaders := ""

	// conact lines into a single string
	msgHeaders = fmt.Sprintf("%s%s:%s\r\n", msgHeaders, "From", fromHeader)
	msgHeaders = fmt.Sprintf("%s%s:%s\r\n", msgHeaders, "To", strings.Join(headerTo, ","))
	msgHeaders = fmt.Sprintf("%s%s:%s\r\n", msgHeaders, "Cc", strings.Join(headerCC, ","))
	msgHeaders = fmt.Sprintf("%s%s:%s\r\n", msgHeaders, "Subject", m.Subject)

	// add the recipient's receipt
	if m.DispositionNotificationTo != "" {
		msgHeaders = fmt.Sprintf("%sDisposition-Notification-To: \"%s\" <%s>\r\n", msgHeaders, m.DispositionNotificationTo, m.DispositionNotificationTo)
	}

	if m.Priority > 5 || m.Priority < 1 {
		m.Priority = Normal
	}
	msgHeaders = fmt.Sprintf("%sX-Priority: %d (%s)\r\n", msgHeaders, m.Priority, m.Priority.String())

	if m.Language != "" {
		msgHeaders = fmt.Sprintf("%sContent-Language: %s\r\n", msgHeaders, m.Language)
	}

	// User-Agent
	if m.UserAgent != "" {
		msgHeaders = fmt.Sprintf("%sUser-Agent: %s\r\n", msgHeaders, m.UserAgent)
	}

	// add multipart/mixed and set the boundary, even if there is no attachments
	msgHeaders = fmt.Sprintf("%sContent-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", msgHeaders, bndry)

	// add body text
	if m.TextBody != "" {
		msgHeaders = fmt.Sprintf("%s--%s\r\n", msgHeaders, bndry)
		msgHeaders = fmt.Sprintf("%sContent-Type: text/plain; charset=utf-8\r\n\r\n", msgHeaders)
		msgHeaders = fmt.Sprintf("%s%s\r\n\r\n", msgHeaders, m.TextBody)
	}

	// add body html
	if m.HTMLBody != "" {

		bodyHTML := strings.ReplaceAll(htmlTemplate, "{{.Body}}", m.HTMLBody)

		msgHeaders = fmt.Sprintf("%s--%s\r\n", msgHeaders, bndry)
		msgHeaders = fmt.Sprintf("%sContent-Type: text/html; charset=utf-8\r\n\r\n", msgHeaders)
		msgHeaders = fmt.Sprintf("%s%s\r\n\r\n", msgHeaders, bodyHTML)
	}

	// attachemnets if any
	msgHeaders = addAttachments(bndry, m.Attachment, msgHeaders)

	return msgHeaders
}

// sendMessage sends an smtp message.
// Go's smtp package does not have an option to send
// delivery status notification as of version 1.22.
// To enable this option, rename github.com.kambahr.go-smtp.smtp_notify.go.txt
// to github.com.kambahr.go-smtp.smtp_notify.go and copy it to .../go/src/net/smtp
// (i.e. /usr/local/go/src/net/smtp).
// Note that the SMTP server must support this option to begin with.
func sendMessage(m MailItem, mc MailCredentials, msgHeaders string) error {

	// Connect to the SMTP Server: <hostname>:<portno>
	servername := fmt.Sprintf("%s:%d", mc.Host, mc.PortNo)
	auth := smtp.PlainAuth("", mc.UserName, mc.Password, mc.Host)

	// create a connection first
	c, err := smtp.Dial(servername)
	if err != nil {
		return err
	}

	if ok, _ := c.Extension("STARTTLS"); ok {
		// TLS config
		tlsconfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         mc.Host}

		c.StartTLS(tlsconfig)
	}

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

		// if err = c.RcptNotify(s, m.From.Address, m.DeliveryStatusNotification); err != nil {
		// 	return err
		// }

		// To enable DSN (Delivery Status Notification), see above comments...
		// comment out the following func, and uncomment the above func;
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

	// BCC (not included in the header, which makes it not visible to recipients)
	for i := 0; i < len(m.BCC); i++ {
		s := fmt.Sprintf(`"%s" <%s>`, m.BCC[i].Name, m.BCC[i].Address)
		if err = c.Rcpt(s); err != nil {
			return err
		}
	}

	// get a writer for the Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	// write the bytes of the headers to the data writer.
	if _, err = w.Write([]byte(msgHeaders)); err != nil {
		return err
	}

	// close the writer (flush); this causes the email to be sent.
	if err = w.Close(); err != nil {
		return err

	}

	// close the connection
	if err = c.Quit(); err != nil {
		return err
	}

	return nil
}

func validate(m MailItem, mc MailCredentials) error {

	if mc.Host == "" || mc.PortNo < 1 || mc.PortNo > 65535 || mc.UserName == "" ||
		len(m.To) == 0 || !emailValid(m.From.Address) {
		return errors.New("invalid settings")
	}

	if len(m.Language) > 5 {
		return errors.New("invalid language")
	}

	if m.Priority > 0 && (m.Priority > 5) {
		return errors.New("invalid priority")
	}

	return nil
}

// SendMail send multiple emails (to,cc, and bcc) to an smtp host.
func SendMail(m MailItem, mc MailCredentials) error {

	if err := validate(m, mc); err != nil {
		return err
	}

	// prepare the headers
	msgHeaders := prepareHeaders(m)

	// send the message
	return sendMessage(m, mc, msgHeaders)
}
