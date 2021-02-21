package globals

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/s3"
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

	Config         TcConfig
	SecurityHubSvc *securityhub.SecurityHub
	IamSvc         iamiface.IAMAPI
	PwdPolicy      *iam.GetAccountPasswordPolicyOutput

	// Config
	ConfigSvc               *configservice.ConfigService
	ConfigRules             []*configservice.ConfigRule
	ComplianceDetailsResult []*configservice.EvaluationResult
	Cred                    string
	CredReport              credentialReport

	// CloudTrail
	CloudTrailSvc *cloudtrail.CloudTrail
	S3Svc         *s3.S3

	// CloudWatch
	CloudWatchSvc *cloudwatchlogs.CloudWatchLogs
	LogGroups     *cloudwatchlogs.DescribeLogGroupsOutput
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

// from https://github.com/aws/aws-sdk-go-v2/issues/225
type Value string
type Policy struct {
	// 2012-10-17 or 2008-10-17 old policies, do NOT use this for new policies
	Version    string       `json:"Version"`
	Id         string       `json:"Id,omitempty"`
	Statements []Statement1 `json:"Statement"`
}

type Statement1 struct {
	Sid          string           `json:"Sid,omitempty"`          // statement ID, service specific
	Effect       string           `json:"Effect"`                 // Allow or Deny
	Principal    map[string]Value `json:"Principal,omitempty"`    // principal that is allowed or denied
	NotPrincipal map[string]Value `json:"NotPrincipal,omitempty"` // exception to a list of principals
	Action       Value            `json:"Action"`                 // allowed or denied action
	NotAction    Value            `json:"NotAction,omitempty"`    // matches everything except
	Resource     Value            `json:"Resource,omitempty"`     // object or objects that the statement covers
	NotResource  Value            `json:"NotResource,omitempty"`  // matches everything except
	Condition    json.RawMessage  `json:"Condition,omitempty"`    // conditions for when a policy is in effect
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

// TBD: Does not check for Principal: *, need to check S3 Policies manually
// User JSON decoder in policyDecoder going forwaard
// str is Jaon Policy formatted *string
func CheckPolicyForAllowAll(str *string) bool {
	var p Policy
	var jsonData = []byte(*str)

	//fmt.Println("Called with string: ", *str)
	err := json.Unmarshal(jsonData, &p)
	if err != nil {
		//fmt.Println("CheckPolicyForAllowAll: unexpected error parsing policy", err)
		Globals.Log.WithFields(logrus.Fields{
			"Test": "Globals"}).Info("CheckPolicyForAllowAll: unexpected error parsing policy: ", err)
		return false
	}
	//fmt.Printf("%+v", p)
	for _, val := range p.Statements {
		//fmt.Println("\nEffect/Allow: ", val.Effect, val.Principal)
		if val.Effect == "Allow" && val.Principal["AWS"] == "*" {
			return true
		}
	}
	return false
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
