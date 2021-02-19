package globals

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

type TcGlobals struct {
	Name           string
	Log            *logrus.Logger // holds all test cases
	FLog           *logrus.Logger // holds all failed test cases
	AllLogsFile    string
	FailedLogsFile string

	Sess    *session.Session
	GRegion string
	GArn    string
	GConf   aws.Config

	Config                  TcConfig
	SecurityHubSvc          *securityhub.SecurityHub
	IamSvc                  iamiface.IAMAPI
	ConfigSvc               *configservice.ConfigService
	ConfigRules             []*configservice.ConfigRule
	ComplianceDetailsResult []*configservice.EvaluationResult
	Cred                    string
	CredReport              credentialReport
}

type TcConfig struct {
	Target struct {
		Region string `yaml:"region"`
		Id     string `yaml:"id"`
	} `yaml:"target"`
	Email struct {
		Sender    string `yaml:"sender"`
		Recv      string `yaml:"recv"`
		Subject   string `yaml:"subject"`
		SendEmail bool   `yaml:"sendemail"`
	}
	EnabledTests []string `yaml:"tests"`
}

var Globals = TcGlobals{Name: "Test Globals"}

type tcRun func(*TcGlobals) (bool, error)

type Tcs struct {
	Id    string
	Descr string
	Run   tcRun
}

var SevCount = map[string]int64{
	"critical": 0,
	"high":     0,
	"medium":   0,
	"low":      0,
	"info":     0,
}

func parseYaml(tcg *TcGlobals) {
	f, err := os.Open("config.yml")
	if err != nil {
		fmt.Println("Yaml file error", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&tcg.Config)
	if err != nil {
		fmt.Println("Yaml decode error", err)
	}
	tcg.Log.WithFields(logrus.Fields{
		"Test": "Globals", "Config": tcg.Config}).Info("Config:")
}

func initLogs(tcg *TcGlobals) {
	//const layout = "01-02-2006"
	const layout = "2 Jan 2006 15:04:05"
	t := time.Now()

	tcg.AllLogsFile = "logs/tctool-" + t.Format(layout) + ".log"
	tcg.Log = logrus.New()
	file, err := os.OpenFile(tcg.AllLogsFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		tcg.Log.Fatal(err)
	}
	//defer file.Close()
	tcg.Log.SetOutput(file)
	tcg.Log.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true, DisableTimestamp: true})
	tcg.Log.SetLevel(logrus.InfoLevel)

	tcg.FailedLogsFile = "logs/tctool-failed-" + t.Format(layout) + ".log"
	tcg.FLog = logrus.New()
	fileF, err := os.OpenFile(tcg.FailedLogsFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		tcg.FLog.Fatal(err)
	}
	//defer fileF.Close()
	tcg.FLog.SetOutput(fileF)
	tcg.FLog.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true, DisableTimestamp: true})
	tcg.FLog.SetLevel(logrus.InfoLevel)
}

func (tcg *TcGlobals) Initialize() bool {
	initLogs(tcg)
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

	parseYaml(tcg)

	tcg.GRegion = tcg.Config.Target.Region
	tcg.GArn = fmt.Sprintf("arn:aws:iam::%v:role/tctool", tcg.Config.Target.Id)
	tcg.GConf = aws.Config{Region: aws.String(tcg.GRegion)}
	tcg.GConf.Credentials = stscreds.NewCredentials(tcg.Sess, tcg.GArn, func(p *stscreds.AssumeRoleProvider) {})

	tcg.IamSvc = iam.New(tcg.Sess, &tcg.GConf)

	tcg.Log.WithFields(logrus.Fields{
		"Test": "Globals"}).Info("**************************Globals Initialized...")
	return true
}
