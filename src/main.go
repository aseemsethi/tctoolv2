package main

import (
	"fmt"
	"github.com/aseemsethi/tctoolv2/src/execTests"
	"github.com/aseemsethi/tctoolv2/src/globals"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

var mLog *logrus.Logger

func sendEmail(globals *globals.TcGlobals) {
	config := globals.Config
	CharSet := "UTF-8"
	//TextBody := "Hi, this is email from Aseem Sethi"
	b, err := ioutil.ReadFile("logs/tctool.log")
	if err != nil {
		fmt.Print(err)
		return
	}
	TextBody := string(b)
	svc := ses.New(globals.Sess, aws.NewConfig().WithRegion("us-east-2"))
	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(config.Email.Recv),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				// Html: &ses.Content{
				// 	Charset: aws.String(CharSet),
				// 	Data:    aws.String(HtmlBody),
				// },
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(TextBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(config.Email.Subject),
			},
		},
		Source: aws.String(config.Email.Sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	//fmt.Println(input)
	// Attempt to send the email.
	_, err = svc.SendEmail(input) // result, err :=

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
			mLog.WithFields(logrus.Fields{
				"Test": "Init"}).Info("Email error:", aerr.Code(), err.Error())
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
			mLog.WithFields(logrus.Fields{
				"Test": "Init"}).Info("Email error:", err.Error())
		}
		return
	}

	//fmt.Println("Email Sent to address: " + config.Email.Recv)
	mLog.WithFields(logrus.Fields{
		"Test": "Init"}).Info("Email sent:", config.Email.Recv)
	//fmt.Println(result)
}

// Call with tctool <region> <accountid>
func main() {
	fmt.Printf("\nTest Compliance Tool Starting..")

	globals.Globals.Initialize()
	mLog = globals.Globals.Log
	mLog.WithFields(logrus.Fields{
		"Test": "Init"}).Info("Security Tests Starting:  *****************************************")
	execTests.ExecTests(&globals.Globals)
	if globals.Globals.Config.Email.SendEmail == true {
		sendEmail(&globals.Globals)
	}
	mLog.WithFields(logrus.Fields{
		"Test": "Tests", "Summary": globals.SevCount}).Info("Summary")
}
