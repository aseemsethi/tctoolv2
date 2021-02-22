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
			//fmt.Println("Results: ", output.EvaluationResults)
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
	getConfigRules(g)
	getComplianceDetails(g)
	return true, nil
}
