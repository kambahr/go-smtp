# SMTP for Golang
## Send multiple emails (To, Cc, and Bcc) with attachements
You can also add a user-agent, DSN (Delivery Status Notification) and/or email receipt options.

### Usage example

```go
package main

import (
	"fmt"

	gosmtp "github.com/kambahr/go-smtp"
)

func main() {

	var mc = gosmtp.MailCredentials{
		Host:     "smtp host ip addr or host name",
		PortNo:   <smtp port; commonly 587>,
		UserName: "<user name; usually an email addr>",
		Password: "<password>"}
	}

	var m = gosmtp.MailItem{
		From: gosmtp.EmailAddr{Name: "<fulll name>", Address: "<email address>"},
		To: []gosmtp.EmailAddr{
			{Name: "<fulll name>", Address: "<email address>"},
			{Name: "<fulll name>", Address: "<email address>"},
			/* ... */
		},
		CC: []gosmtp.EmailAddr{{Name: "<fulll name>", Address: "<email address>"}},
		Bcc: []gosmtp.EmailAddr{{Name: "<fulll name>", Address: "<email address>"}},

		Subject: "Test email from Go smtp client",
		/* 
			a message can contain both text and html formats 
		*/
		HTMLBody: "<h1 style='color:darkgreen'>Hello world - HTML format!</h1>",
		TextBody: "Hello world! - text format!",

		Attachment: []string{
			"<full path to a document>",
			"<full path to another document>",
			/* ... */
			/* note that the attachments limit depends on your mail server config */
		},

		/* Priority is not required; Normal is the default */
		Priority:                   gosmtp.High,

		/* Language is not required */
		Language:                   "en-US",

		/* 
		DeliveryStatusNotification causes the SMTP server send
		a notification about the delivery status of an email
		(failure, success or delay); see sendMessage() func's
		comments in email.go as to how to enable this option.
		
		*/
		// DeliveryStatusNotification: []string{gosmtp.SUCCESS},

		/* 
			DispositionNotificationTo causes a request to be added to the
		 	email so that the recipient will have the option of sending
			the receipt after reading the email (aka email receipt).
		*/
		DispositionNotificationTo:  "<the FROM or any other email  address>",

		/* 
			UserAgent is not required.
		*/
		UserAgent:                  "<name of or your application>",
	}

	err := gosmtp.SendMail(m, mc)
	if err == nil {
		fmt.Println("email sent successfully")
	} else {
		fmt.Println(err)
	}
}
```
