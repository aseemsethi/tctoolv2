package execTests

import (
	"fmt"
	"github.com/aseemsethi/tctoolv2/src/globals"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/sirupsen/logrus"
)

type ConfigTool struct {
	Name                    string
	svc                     *configservice.ConfigService
	configRules             []*configservice.ConfigRule
	complianceDetailsResult []*configservice.EvaluationResult
}

func InitConfig(g *globals.TcGlobals) (bool, error) {
	g.ConfigSvc = configservice.New(g.Sess, &g.GConf)
	g.ConfigRules = make([]*configservice.ConfigRule, 0)
	g.ComplianceDetailsResult = make([]*configservice.EvaluationResult, 0)

	iLog.WithFields(logrus.Fields{"Test": "Config"}).Info("Enabled")
	return true, nil
}

func getComplianceDetails(g *globals.TcGlobals) {
	iLog.WithFields(logrus.Fields{"Test": "Config"}).Info("Config Evaluation Results............................................")
	for _, configRule := range g.ConfigRules {
		nextToken := ""
		for {
			output, err := g.ConfigSvc.GetComplianceDetailsByConfigRule(
				&configservice.GetComplianceDetailsByConfigRuleInput{
					ConfigRuleName: configRule.ConfigRuleName,
					NextToken:      &nextToken,
				})
			if err != nil {
				fmt.Println(err)
				continue
			}
			g.ComplianceDetailsResult = append(g.ComplianceDetailsResult, output.EvaluationResults...)
			for _, r := range output.EvaluationResults {
				if *r.ComplianceType == "NON_COMPLIANT" {
					//fmt.Println("Config Svc Failed: ", r.EvaluationResultIdentifier.EvaluationResultQualifier)
					g.FLog.WithFields(logrus.Fields{"Test": "Config",
						"Eval Results": r.EvaluationResultIdentifier.EvaluationResultQualifier}).Info("Compliance Results")
				}
			}
			iLog.WithFields(logrus.Fields{"Test": "Config", "Eval Results": output.EvaluationResults}).Info("Compliance Results")
			if output.NextToken == nil {
				break
			}
			nextToken = *output.NextToken
		}
	}
}

func getConfigRules(g *globals.TcGlobals) {
	nextToken := ""
	for {
		output, err := g.ConfigSvc.DescribeConfigRules(&configservice.DescribeConfigRulesInput{
			NextToken: &nextToken,
		})
		if err != nil {
			iLog.WithFields(logrus.Fields{"Test": "Config"}).Info("Error in getConfigRules: ", err)
		}
		//iLog.WithFields(logrus.Fields{"Test": "Config", "Rules": output.ConfigRules}).Info("getConfigRules")
		g.ConfigRules = append(g.ConfigRules, output.ConfigRules...)

		if output.NextToken == nil {
			break
		}
		nextToken = *output.NextToken
	}
}

func RunConfig(g *globals.TcGlobals) (bool, error) {
	iLog.WithFields(logrus.Fields{
		"Test": "Config"}).Info("ConfigTool Run...")
	g.FLog.WithFields(logrus.Fields{"Test": "Config"}).Info("ConfigSvc Failed Cases ***********************************")

	getConfigRules(g)
	getComplianceDetails(g)
	return true, nil
}
