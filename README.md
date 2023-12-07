# SMTP for Golang
## Send multiple emails (To, Cc, and Bcc)

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

	var to []gosmtp.EmailAddr
	var cc []gosmtp.EmailAddr
	var bcc []gosmtp.EmailAddr

	to = append(to, gosmtp.EmailAddr{Name: "<first name> <last name> or a display name", Address: "<email addr>"})
	to = append(to, gosmtp.EmailAddr{Name: "<first name> <last name> or a display name", Address: "<email addr>"})
	cc = append(cc, gosmtp.EmailAddr{Name: "<first name> <last name> or a display name", Address: "<email addr>"})
	bcc = append(bcc, gosmtp.EmailAddr{Name: "<first name> <last name> or a display name", Address: "<email addr>"})

	var m = gosmtp.MailItem{
        From:     gosmtp.EmailAddr{Name: "<first name> <last name> or a display name i.e. an org name", Address: "<email addr>"},
		To:       to,
		CC:       cc,
		BCC:      bcc,
		Subject:  "<subject>",
		HTMLBody: `html or plain text i.e <h1 style="color:dodgerblue">Hello World</h1>``,
	}

	err := gosmtp.SendMail(m, mc)
	if err == nil {
		fmt.Println("email sent successfully")
	} else {
		fmt.Println(err)
	}
}
```
