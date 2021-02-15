package globals

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"os"
)

type TcGlobals struct {
	Name    string
	Log     *logrus.Logger
	Sess    *session.Session
	GRegion string
	GArn    string
	GConf   aws.Config
}

type TcConfig struct {
	Target struct {
		Region string `yaml:"region"`
		Id     string `yaml:"id"`
	} `yaml:"target"`
	Email struct {
		Id string `yaml:"id"`
	}
	Database struct {
		Username string `yaml:"user"`
		Password string `yaml:"pass"`
	} `yaml:"database"`
}

var Globals = TcGlobals{Name: "Test Globals"}
var Config TcConfig

func parseYaml(tcg *TcGlobals) {
	f, err := os.Open("config.yml")
	if err != nil {
		fmt.Println("Yaml file error", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&Config)
	if err != nil {
		fmt.Println("Yaml decode error", err)
	}
	fmt.Println("Config: ", Config)
	tcg.Log.WithFields(logrus.Fields{
		"Test": "Globals", "Config": Config}).Info("Config:")
}

func (tcg *TcGlobals) Initialize() bool {
	tcg.Log = logrus.New()
	file, err := os.OpenFile("logs/tctool.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		tcg.Log.Fatal(err)
	}
	//defer file.Close()
	tcg.Log.SetOutput(file)
	tcg.Log.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true, DisableTimestamp: true})
	tcg.Log.SetLevel(logrus.InfoLevel)

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

	tcg.GRegion = Config.Target.Region
	tcg.GArn = fmt.Sprintf("arn:aws:iam::%v:role/KVAccess", Config.Target.Id)
	tcg.GConf = aws.Config{Region: aws.String(tcg.GRegion)}
	tcg.GConf.Credentials = stscreds.NewCredentials(tcg.Sess, tcg.GArn, func(p *stscreds.AssumeRoleProvider) {})

	tcg.Log.WithFields(logrus.Fields{
		"Test": "Globals"}).Info("**************************Globals Initialized...")
	return true
}
