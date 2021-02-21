package execTests

import (
	"bytes"
	"encoding/json"
	//"fmt"
	"github.com/aseemsethi/tctoolv2/src/globals"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

func InitCloudTrail(g *globals.TcGlobals) (bool, error) {

	// Create a CloudTrail service client.
	g.CloudTrailSvc = cloudtrail.New(g.Sess, &g.GConf)

	// Create S3 service client
	g.S3Svc = s3.New(g.Sess, &g.GConf)

	return true, nil
}

func checkS3(g *globals.TcGlobals, bucketName *string) {
	// Get the bucket name configured for CloudTrail
	//fmt.Println("Search Bucket: ", *bucketName)
	iLog.WithFields(logrus.Fields{
		"Test": "CIS", "Num": 2.3,
	}).Info("Search S3 Bucket for CloudTrail: ", *bucketName)
	in := &s3.HeadBucketInput{
		Bucket: aws.String(*bucketName),
	}
	_, err := g.S3Svc.HeadBucket(in)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 2.3, "Result": "Failed",
			}).Info("S3 Bucket not found: ", aerr.Code())
		} else {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 2.3, "Result": "Failed",
			}).Info("S3 Bucket not found..: ", err.Error())
		}
		return
	}
	iLog.WithFields(logrus.Fields{
		"Test": "CIS", "Num": 2.3, "Result": "Passed",
	}).Info("S3 Bucket found..: ")

	// Ensure the policy does not contain a Statement having an Effect set to
	// Allow and a Principal set to "*" or {"AWS" : "*"}
	// Call S3 to retrieve the JSON formatted policy for the selected bucket.
	result, err := g.S3Svc.GetBucketPolicy(&s3.GetBucketPolicyInput{
		Bucket: aws.String(*bucketName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 2.3, "Result": "Failed",
			}).Info("S3 Bucket Policy not found: ", aerr.Code())
		} else {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 2.3, "Result": "Failed",
			}).Info("S3 Bucket Policy not found..: ", err.Error())
		}
		return
	}

	out := bytes.Buffer{}
	policyStr := aws.StringValue(result.Policy)
	if err := json.Indent(&out, []byte(policyStr), "", "  "); err != nil {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 2.3, "Result": "Failed",
		}).Info("Failed to pretty the S3 Policy: ", err)
	}
	//fmt.Printf("Bucket Policy:\n")
	//fmt.Println(out.String())
	iLog.WithFields(logrus.Fields{
		"Test": "CIS", "Num": 2.3, "Result": "Failed",
	}).Info("S3 Bucket Policy: ", out.String())
	allow := globals.CheckPolicyForAllowAll(result.Policy)
	if allow == true {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 2.3, "Result": "Failed",
		}).Info("S3 Policy allows Public access: ", err)
	} else {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 2.3, "Result": "Passed",
		}).Info("S3 Policy does not allows Public access: ")
	}

	logResult, err := g.S3Svc.GetBucketLogging(&s3.GetBucketLoggingInput{
		Bucket: aws.String(*bucketName)})
	if err != nil {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 2.6,
		}).Info("Cloudtrail S3 logging retrieval error", err)
	} else {
		//fmt.Println("logResult.LoggingEnabled:", logResult.LoggingEnabled)
		if logResult.LoggingEnabled == nil {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 2.6, "Result": "Failed",
			}).Info("Cloudtrail S3 logging disabled")
		} else {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 2.6, "Result": "Passed",
			}).Info("Cloudtrail S3 logging enabled: ", logResult.LoggingEnabled)
		}
	}
}

func checkTrailProperties(g *globals.TcGlobals, trail *cloudtrail.Trail) {
	if trail.CloudWatchLogsLogGroupArn == nil {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 2.4, "Result": "Failed",
		}).Info("CloudTrail CloudWatch integration not done")
	} else {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 2.4, "Result": "Passed",
		}).Info("CloudTrail CloudWatch integration done: ", *trail.CloudWatchLogsLogGroupArn)
	}
	//fmt.Println("IsMultiRegionTrail: ", *trail.IsMultiRegionTrail)
	if *trail.IsMultiRegionTrail == false {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 2.5, "Result": "Failed",
		}).Info("CloudTrail IsMultiRegionTrail is false")
	} else {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 2.5, "Result": "Passed",
		}).Info("CloudTrail IsMultiRegionTrail done")
	}
	//fmt.Println("KMS:", trail.KmsKeyId)
	if trail.KmsKeyId == nil {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 2.7, "Result": "Failed",
		}).Info("CloudTrail KmsKeyId is not set")
	} else {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 2.7, "Result": "Passed",
		}).Info("CloudTrail KmsKeyId is set")
	}
}

/* AWS CloudTrail is now enabled by default for ALL CUSTOMERS and will provide visibility
 * into the past seven days of account activity without the need for you to configure a
 * trail in the service to get started
 * We thus check if any trail is configured.
 */
func checkIfEnabled(g *globals.TcGlobals) {
	resp, err := g.CloudTrailSvc.DescribeTrails(&cloudtrail.DescribeTrailsInput{TrailNameList: nil})
	if err != nil {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS"}).Info("Error getting trail: ", err.Error())
	}

	iLog.WithFields(logrus.Fields{
		"Test": "CIS"}).Info("Found trail len: ", len(resp.TrailList))
	if len(resp.TrailList) == 0 {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 2.1, "Result": "Failed",
		}).Info("CloudTrail is disabled")
	} else {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 2.1, "Result": "Passed",
		}).Info("CloudTrail is enabled")
		for _, trail := range resp.TrailList {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS"}).Info("Found Trail: ", *trail.Name, " Bucket: ", *trail.S3BucketName)
			if trail.LogFileValidationEnabled == nil || *trail.LogFileValidationEnabled == false {
				iLog.WithFields(logrus.Fields{
					"Test": "CIS", "Num": 2.2, "Result": "Failed",
				}).Info("CloudTrail LogFileValidationEnabled is disabled for trail: ", *trail.Name)
			} else {
				iLog.WithFields(logrus.Fields{
					"Test": "CIS", "Num": 2.2, "Result": "Passed",
				}).Info("CloudTrail LogFileValidationEnabled is enabled for trail: ", *trail.Name)
			}
			// For each Trail, check the following configurations
			checkS3(g, trail.S3BucketName)
			checkTrailProperties(g, trail)
		}
	}
}

func chckifFlowLogsEnabled(g *globals.TcGlobals) {
	svc := ec2.New(g.Sess, &g.GConf)
	input := &ec2.DescribeFlowLogsInput{}

	result, err := svc.DescribeFlowLogs(input)
	if err != nil {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 2.9,
		}).Info("CloudTrail Flowlogs retrieval error", err)
	} else {
		//fmt.Println("Len:", len(result.FlowLogs))
		if len(result.FlowLogs) == 0 {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 2.9, "Result": "Failed",
			}).Info("CloudTrail Flowlogs disabled: ", result)
		} else {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 2.9, "Result": "Passed",
			}).Info("CloudTrail Flowlogs enabled: ", result)
		}
	}
}

func RunCloudTrail(g *globals.TcGlobals) (bool, error) {
	iLog.WithFields(logrus.Fields{
		"Test": "CIS"}).Info("CloudTrail Run...")
	checkIfEnabled(g)
	chckifFlowLogsEnabled(g)
	return true, nil
}
