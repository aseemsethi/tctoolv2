package globals

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
	"os"
)

type TcGlobals struct {
	Name string
	Log  *logrus.Logger
	Sess *session.Session
	// 	GRegion string
	// 	GArn    string
	// 	GConf   aws.Config
}

var Globals = TcGlobals{Name: "Test Globals"}

func (tcg *TcGlobals) Initialize() bool {
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile:           "default",
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		fmt.Println("Error creating new session")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	tcg.Sess = sess

	// 	tcg.GRegion = region
	// 	tcg.GArn = fmt.Sprintf("arn:aws:iam::%v:role/KVAccess", account)
	// 	tcg.GConf = aws.Config{Region: aws.String(tcg.GRegion)}
	// 	tcg.GConf.Credentials = stscreds.NewCredentials(tcg.Sess, tcg.GArn, func(p *stscreds.AssumeRoleProvider) {})

	tcg.Log = logrus.New()
	file, err := os.OpenFile("logs/tctool.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		tcg.Log.Fatal(err)
	}
	//defer file.Close()
	tcg.Log.SetOutput(file)
	tcg.Log.SetFormatter(&logrus.JSONFormatter{})
	tcg.Log.SetLevel(logrus.InfoLevel)
	tcg.Log.WithFields(logrus.Fields{
		"Test": "Globals"}).Info("**************************Globals Initialized...")
	return true
}
