package main

import (
	"errors"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/DusanKasan/parsemail"

	"github.com/alash3al/go-smtpsrv"
	"github.com/zaccone/spf"
	"gopkg.in/resty.v1"
)

func handler(req *smtpsrv.Request) error {
	log.Printf(`[handler] from:"%s" to:"%s" remote_addr:"%s"`+"\n", req.From, req.To, req.RemoteAddr)

	// validate the from data
	if *flagStrictValidation {
		spam := req.SPFResult != spf.Pass
		if spam {
			// trust mails from gmail.com
			from := req.From
			ip, _, _ := net.SplitHostPort(req.RemoteAddr)
			spfres, _, _ := spf.CheckHost(net.ParseIP(ip), `gmail.com`, from)
			spam = spfres != spf.Pass
		}
		if spam {
			log.Println(`[handler] spam detected`)
			return errors.New("Your host isn't configured correctly or you are a spammer -_-")
		} else if !req.Mailable {
			log.Println(`[handler] not mailable`)
			return errors.New("Your mail isn't valid because it cannot receive emails -_-")
		}
	}

	msg, err := parsemail.Parse(req.Message)
	if err != nil {
		log.Println(`[handler] fail to read`)
		return errors.New("Cannot read your message, it may be because of it exceeded the limits")
	}

	rq := resty.R()

	// set the url-encoded-data
	rq.SetFormData(map[string]string{
		"id":                  msg.Header.Get("Message-ID"),
		"subject":             msg.Subject,
		"body[text]":          string(msg.TextBody),
		"body[html]":          string(msg.HTMLBody),
		"addresses[mailfrom]": req.From,
		"addresses[from]":     strings.Join(extractEmails(msg.From), ","),
		"addresses[to]":       strings.Join(extractEmails(msg.To), ","),
		"addresses[cc]":       strings.Join(extractEmails(msg.Cc), ","),
		"addresses[bcc]":      strings.Join(extractEmails(msg.Bcc), ","),
	})

	// set the files "attachments"
	for i, file := range msg.Attachments {
		is := strconv.Itoa(i)
		rq.SetFileReader("file["+is+"]", file.Filename, (file.Data))
	}

	// submit the form
	resp, err := rq.Post(*flagWebhook)
	if err != nil {
		log.Println(`[handler] internal error`)
		return errors.New("Cannot accept your message due to internal error, please report that to our engineers, '" + (err.Error()) + "'")
	} else if resp.StatusCode() != 200 {
		log.Println(`[handler] backend returned error`)
		return errors.New("BACKEND: " + resp.Status())
	}

	log.Println(`[handler] successfully processed`)
	return nil
}
