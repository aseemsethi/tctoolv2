package execTests

import (
	//"encoding/json"
	"fmt"
	"github.com/aseemsethi/tctoolv2/src/globals"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	//"github.com/jmoiron/jsonq"
	"github.com/sirupsen/logrus"
	//"strings"
)

type SecurityHub struct {
	Name string
}

func InitSecurityHub(g *globals.TcGlobals) (bool, error) {
	g.SecurityHubSvc = securityhub.New(g.Sess, &g.GConf)

	input := &securityhub.EnableSecurityHubInput{}
	_, err := g.SecurityHubSvc.EnableSecurityHub(input)
	if err != nil {
		//fmt.Println("failed EnableSecurityHub: %s", err)
		iLog.WithFields(logrus.Fields{"Test": "SecurityHub"}).Info("Not Enabled: ", err)
	}
	//fmt.Println("EnableSecurityHub...")
	iLog.WithFields(logrus.Fields{"Test": "SecurityHub"}).Info("Enabled")

	return true, nil
}

func listFindings(g *globals.TcGlobals) {
	var nextToken *string
	for {
		input := &securityhub.GetFindingsInput{
			MaxResults: aws.Int64(100),
			NextToken:  nextToken,
		}
		list, err := g.SecurityHubSvc.GetFindings(input)
		if err != nil {
			iLog.WithFields(logrus.Fields{"Test": "securityhub"}).Info("ListFindings failed: ", err)
			return
		}
		iLog.WithFields(logrus.Fields{"Test": "securityhub"}).Info("ListFindings passed")
		for _, v := range list.Findings {
			// data := map[string]interface{}{}
			// dec := json.NewDecoder(strings.NewReader(v.Compliance.Status))
			// dec.Decode(&data)
			// jq := jsonq.NewQuery(data)
			// if jq.String("Findings", "Compliance", "Status") == "FAILED" {
			// 	//if data["Findings"]["Compliance"]["Status"].(string) == "FAILED" {
			// 	fmt.Println("v")
			// }
			if v == nil || v.Compliance == nil {
				// True for Inspector results which we are showing separately anyways
				continue
			}
			if *v.Compliance.Status == "FAILED" {
				fmt.Println("Failed:", *v.Compliance.Status, ", Reason:", v.Compliance.StatusReasons, *v.Description, " Id::", v.Resources)
				iLog.WithFields(logrus.Fields{"Test": "securityhub"}).Info(
					"Failed:", *v.Compliance.Status, ", Reason:", v.Compliance.StatusReasons, *v.Description, " Id::", v.Resources)
			}
			//iLog.WithFields(logrus.Fields{"Test": "securityhub", "Failed Findings": v}).Info("Findings")
			//fmt.Println("Findings: ", v)
		}
		if list.NextToken != nil {
			nextToken = list.NextToken
		} else {
			break
		}
	}
}

func RunSecurityHub(g *globals.TcGlobals) (bool, error) {
	iLog.WithFields(logrus.Fields{
		"Test": "SecurityHub"}).Info("SecurityHub Run...")
	input := &securityhub.GetEnabledStandardsInput{}
	output, err := g.SecurityHubSvc.GetEnabledStandards(input)
	if err != nil {
		iLog.WithFields(logrus.Fields{"Test": "SecurityHub"}).Info("GetEnabledStandards failed: ", err)
		return false, err
	}
	iLog.WithFields(logrus.Fields{"Test": "SecurityHub", "Subscriptions": output.StandardsSubscriptions}).Info("GetEnabledStandards")
	//fmt.Println("Enabled: ", output.StandardsSubscriptions)
	listFindings(g)
	return true, nil
}
