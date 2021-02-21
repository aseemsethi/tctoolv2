package execTests

import (
	"github.com/aseemsethi/tctoolv2/src/globals"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/sirupsen/logrus"
	"strings"
)

func InitCloudWatch(g *globals.TcGlobals) (bool, error) {
	g.CloudWatchSvc = cloudwatchlogs.New(g.Sess, &g.GConf)
	return true, nil
}

func lookupCloudWatchLogMetricFilter(g *globals.TcGlobals, name, logGroupName string, nextToken *string, filter *string) {
	input := cloudwatchlogs.DescribeMetricFiltersInput{
		//FilterNamePrefix: aws.String(name),
		LogGroupName: aws.String(logGroupName),
		NextToken:    nextToken,
	}
	//fmt.Printf("Reading CloudWatch Log Metric Filter: %s", input)
	resp, err := g.CloudWatchSvc.DescribeMetricFilters(&input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 3.3,
			}).Info("CloudWatch Log Metric Filters not retrieved - ResourceNotFoundException: ", err)
			return
		}
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 3.3,
		}).Info("CloudWatch Log Metric Filters not retrieved: ", err)
		return
	}
	for _, mf := range resp.MetricFilters {
		//fmt.Println("\nFilterName: ", mf)
		//if strings.Contains(*mf.FilterPattern, "$.userIdentity.type = \"Root\"") {
		if strings.Contains(*mf.FilterPattern, *filter) {
			//fmt.Println("CloudWatch Log Metric Filter found:", mf)
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 3.3, "Result": "Passed",
			}).Info("CloudWatch Log Metric Filter checking found: ", mf)
			return
		}
	}

	if resp.NextToken != nil {
		lookupCloudWatchLogMetricFilter(g, name, logGroupName, resp.NextToken, filter)
		return
	}
	//fmt.Println("CloudWatch Log Metric Filter checking Not found:", *filter)
	iLog.WithFields(logrus.Fields{
		"Test": "CIS", "Num": 3.3, "Result": "Failed",
	}).Info("CloudWatch Log Metric Filter checking Not found:", *filter)
}

func GetLogGroups(svc *cloudwatchlogs.CloudWatchLogs) (result *cloudwatchlogs.DescribeLogGroupsOutput, error error) {
	input := &cloudwatchlogs.DescribeLogGroupsInput{}
	data, err := svc.DescribeLogGroups(input)
	if err != nil {
		return nil, err
	}
	token := data.NextToken
	for token != nil {
		input := &cloudwatchlogs.DescribeLogGroupsInput{
			NextToken: token,
		}
		nextResult, err := svc.DescribeLogGroups(input)
		if err != nil {
			return nil, err
		}
		data.LogGroups = append(data.LogGroups, nextResult.LogGroups...)
		token = nextResult.NextToken
	}
	return data, nil
}

func RunCloudWatch(g *globals.TcGlobals) (bool, error) {
	var err error
	iLog.WithFields(logrus.Fields{
		"Test": "CIS"}).Info("CloudWatch Run...")
	g.LogGroups, err = GetLogGroups(g.CloudWatchSvc)
	if err != nil {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS"}).Info("CloudWatch Groups retrieval error: ", err)
		return false, err
	}
	//i.LogGroups = result
	iLog.WithFields(logrus.Fields{
		"Test": "CIS"}).Info("CloudWatch Groups: ", g.LogGroups)
	//fmt.Println("LogGroups: ", i.LogGroups)
	for _, groups := range g.LogGroups.LogGroups {
		filter := "$.userIdentity.type = \"Root\""
		lookupCloudWatchLogMetricFilter(g, "userIdentity.type", *groups.LogGroupName, nil, &filter)
		filter = "($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\")"
		lookupCloudWatchLogMetricFilter(g, "userIdentity.type", *groups.LogGroupName, nil, &filter)

	}
	return true, nil
}
