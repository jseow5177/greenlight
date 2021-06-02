package mailer

import (
	"bytes"
	"embed"
	"text/template"
	"time"

	"github.com/go-mail/mail/v2"
)

// New Go 1.16 embedded files functionality.
// Declare a new variable with the type embed.FS (embedded file system) to hold the email templates.
// This has a comment directive in the format `//go:embed <path>` directly above it.
// The directive path should be relative to the file that declares the directive.
// Below indicates to Go that we want to store the contents of the ./templates directory in the templateFS variable.

//go:embed "templates"
var templateFS embed.FS

// Define a Mailer struct which contains a mail.Dialer instance (used to connect to a SMTP server)
// and the sender information for your emails (the name and address you want the email to be from)
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	// Initializes a new mail.Dialer with the given SMTP server settings.
	dialer := mail.NewDialer(host, port, username, password)

	// Configure a 5-second timeout whenever we send an email.
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dialer: dialer,
		sender: sender,
	}

}

// Define a Send() method on the Mailer type. This takes the recipient email address as the first parameter, the name of the file containing
// the templates, and any dynamic data for the templates as an interface{} parameter.
func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	// Use ParseFS() to parse the required template file from the embedded file system.
	// The file system is rooted in the directory which contains the //go:embed directive.
	// To retrieve a file in it, we need to start with path templates/
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// Execute the named template "subject", passing in the dynamic data and store
	// the result in a bytes.Buffer variable
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Do the same thing with the "plainBody" template
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// Do the same thing with the "htmlBody" template
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// Use the mail.NewMessage() function to initialize a new mail.Message instance.
	// Then, we use the SetHeader() method to set the email recipient, sender and subject headers.
	// The SetBody() method set the plain-text body.
	// The AddAlternative() method sets the HTML body. This should always be called after SetBody().
	// It is common to send HTML emails that default to their plain text version for backward compatibility.
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// Try sending the email up to three times before aborting and returning the final error.
	// We sleep for 500 milliseconds before each attempt.
	for i := 1; i <= 3; i++ {
		// Opens a connection to the SMTP server, send the given email and closes the connection.
		// If there is a timeout, it will return a "dial tcp: i/o timeout" error.
		err = m.dialer.DialAndSend(msg)
		if err == nil {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return err
}
