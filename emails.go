package main

import (
	"encoding/base64"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"log"
	"os"
	"time"
)

var GmailService *gmail.Service

func SetUpOAuthGmailService() {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	accessToken := os.Getenv("ACCESS_TOKEN")

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost",
	}

	token := oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: os.Getenv("REFRESH_TOKEN"),
		TokenType:    "Bearer",
		Expiry:       time.Now(),
	}

	var tokenSource = config.TokenSource(context.Background(), &token)

	srv, err := gmail.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
	}

	GmailService = srv
	if GmailService != nil {
		fmt.Println("Email service is successfully initialized")
	}
}

func SendEmailOAUTH2(to string, rate string) error {
	var message gmail.Message

	//TODO extract strings so they can be changed easily
	emailTo := "To: " + to + "\r\n"
	subject := "Subject: " + "Bitcoin to UAH exchange rate update" + "\n"
	emailBody := "Bitcoin now costs " + rate + " UAH"
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	msg := []byte(emailTo + subject + mime + "\n" + emailBody)

	message.Raw = base64.URLEncoding.EncodeToString(msg)

	_, err := GmailService.Users.Messages.Send("me", &message).Do()
	if err != nil {
		return err
	}
	return nil
}
