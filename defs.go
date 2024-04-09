package gosmtp

const htmlTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=UTF-8">
  </head>
  <body>
    {{.Body}}
  </body>
</html>
`

const (
	FAILURE = "FAILURE"
	DELAY   = "DELAY"
	SUCCESS = "SUCCESS"
)

type Priority uint8

const (
	Highest Priority = 1
	High    Priority = 2
	Normal  Priority = 3
	Low     Priority = 4
	Lowest  Priority = 5
)

func (p Priority) String() string {
	switch p {
	case Highest:
		return "Highest"
	case High:
		return "High"
	case Normal:
		return "Normal"
	case Low:
		return "Low"
	case Lowest:
		return "Lowest"
	}
	return "unknown"
}

type EmailAddr struct {
	Name    string // Name can be omitted
	Address string
}

type MailItem struct {
	// email adress of the account used to send email
	// (may be the same as the primary email-account)
	From EmailAddr

	To         []EmailAddr
	CC         []EmailAddr
	BCC        []EmailAddr
	Attachment []string // full path of files to attach
	Subject    string
	HTMLBody   string
	TextBody   string
	Priority   Priority
	Language   string
	UserAgent  string

	// DeliveryStatusNotification causes a status email be sent
	// to the FROM address.
	DeliveryStatusNotification []string

	// DispositionNotificationTo is a request for the SMTP server
	// to send a DSN (Selivery Status Notification) as soon as the recipient opens the email.
	DispositionNotificationTo string
}

type MailCredentials struct {
	Host     string // IP address or host-name
	PortNo   int
	UserName string // the primary user-name for the smtp account
	Password string
}
