package execTests

// Learnings from https://github.com/prabhatsharma/aws-inspector-assessment

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	//"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aseemsethi/tctoolv2/src/globals"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/sirupsen/logrus"
	"time"
)

type InspectorStruct struct {
	Name string
}

func InitInspector(g *globals.TcGlobals) (bool, error) {
	return true, nil
}

func getSpecificTagValue(key string, tags []*ec2.Tag) string {
	for _, tag := range tags {
		if *(tag.Key) == key {
			return *tag.Value
		}
	}
	return "--"
}

func RunInspector(g *globals.TcGlobals) (bool, error) {
	iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("Run...")
	sess, _ := session.NewSessionWithOptions(session.Options{
		// Specify profile to load for the session's config
		Profile: "default",

		// Provide SDK Config options, such as Region.
		//Config: aws.Config{Region: aws.String("us-east-1")},

		// Force enable Shared Config support
		// Using the NewSessionWithOptions with SharedConfigState set to SharedConfigEnable will
		// create the session as if the AWS_SDK_LOAD_CONFIG environment variable was set.
		SharedConfigState: session.SharedConfigEnable,
	})
	//_, err := sess.Config.Credentials.Get()
	//fmt.Println("err: ", err)
	svc := inspector.New(sess, &g.GConf)

	/** EC2 reading **/
	ec2Svc := ec2.New(sess, &g.GConf)
	ec2Instances, err := ec2Svc.DescribeInstances(nil)
	if err != nil {
		iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("Inspector cannot get ec2s: ", err)
		return false, err
	}
	for idx := range ec2Instances.Reservations {
		for _, inst := range ec2Instances.Reservations[idx].Instances {
			inspectorTag := getSpecificTagValue("inspector", inst.Tags)
			iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("Type", *inst.InstanceType, " ID: ", *inst.InstanceId, " State: ", *inst.State.Name, " InspectorTag: ", inspectorTag)
			if inspectorTag == "true" {
				//fmt.Println("Included in Inspector run")
				iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("EC2 included in run")
			}
		}
	}
	/**********/

	rgi := &inspector.CreateResourceGroupInput{
		ResourceGroupTags: []*inspector.ResourceGroupTag{
			{
				Key:   aws.String("inspector"),
				Value: aws.String("true"),
			},
		},
	}
	iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("Inspector ResGrp created")
	rg, rgerr := svc.CreateResourceGroup(rgi)
	if rgerr != nil {
		iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("Inspector ResGrp creation failed:", rgerr)
		return false, err
	}
	iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("Inspector Resource Group created: ", *rg.ResourceGroupArn)
	//return *rg.ResourceGroupArn

	// 2. Create assessment target
	ati := &inspector.CreateAssessmentTargetInput{
		AssessmentTargetName: aws.String("InspectorRun" + "_AssessmentTarget_" + time.Now().Format("2006-01-02_15.04.05")),
		ResourceGroupArn:     rg.ResourceGroupArn,
	}
	at, aterr := svc.CreateAssessmentTarget(ati)
	if aterr != nil {
		iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("Inspector Asessment Target ceration failed: ", aterr)
		return false, err
	}
	iLog.WithFields(logrus.Fields{"Test": "Inspector", "Target": at}).Info("Inspector Asessment Target created")
	//fmt.Println("AssessmentTarget: ", at)

	// 3. create rules package input
	rpi := &inspector.ListRulesPackagesInput{
		MaxResults: aws.Int64(100),
	}
	rp, erp := svc.ListRulesPackages(rpi)
	if erp != nil {
		//fmt.Println(erp.Error())
		iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("ListRulesPackages failed: ", erp.Error())
		return false, err
	}
	iLog.WithFields(logrus.Fields{"Test": "Inspector", "Rules": rp}).Info("ListRules")
	//fmt.Println("List Rules Pkg: ", rp) // we selct all rules, i,e, N/W, CVE etc.

	// 4. create assessment template
	atli := &inspector.CreateAssessmentTemplateInput{
		AssessmentTargetArn:    aws.String(*at.AssessmentTargetArn),
		AssessmentTemplateName: aws.String("InspectorRun" + "_AssessmentTemplate_" + time.Now().Format("2006-01-02_15.04.05")),
		DurationInSeconds:      aws.Int64(180),
		RulesPackageArns:       rp.RulesPackageArns,
		UserAttributesForFindings: []*inspector.Attribute{
			{
				Key:   aws.String("inspection-type"),
				Value: aws.String("InspectorRun"),
			},
		},
	}

	atl, atlerr := svc.CreateAssessmentTemplate(atli)
	if atlerr != nil {
		//fmt.Println(atlerr)
		iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("CreateAssessmentTemplate failed: ", atlerr)
		return false, err
	}
	//fmt.Println("Asessment Template: ", atl)
	iLog.WithFields(logrus.Fields{"Test": "Inspector", "Assess Template": atl}).Info("Asessment Template")

	// 6. start assessment template run
	ari := &inspector.StartAssessmentRunInput{
		AssessmentRunName:     aws.String("InspectorRun" + "_Run_" + time.Now().Format("2006-01-02_15.04.05")),
		AssessmentTemplateArn: aws.String(*atl.AssessmentTemplateArn),
	}

	ar, arerr := svc.StartAssessmentRun(ari)
	if arerr != nil {
		//fmt.Println(arerr.Error())
		iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("StartAssessmentRun failed: ", arerr.Error())
		return false, arerr
	}
	//fmt.Println("Asessment Run start: ", ar)
	iLog.WithFields(logrus.Fields{"Test": "Inspector", "Run": ar}).Info("Asessment Run started")
	time.Sleep(300 * time.Second)

	//fmt.Println("Asessment Run complete: Info = 0.0, Low = 3.0, Medium = 6.0, High = 9.0")
	iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("Sev: Info = 0.0, Low = 3.0, Medium = 6.0, High = 9.0")
	iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("Asessment Run completed")
	var nextToken *string
	//var list *inspector.ListFindingsOutput
	for {
		input := &inspector.ListFindingsInput{
			AssessmentRunArns: []*string{
				aws.String(*ar.AssessmentRunArn),
			},
			MaxResults: aws.Int64(123),
			NextToken:  nextToken,
		}
		list, err := svc.ListFindings(input)
		if err != nil {
			//fmt.Println(err.Error())
			iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("ListFindings failed: ", arerr.Error())
			return false, err
		}
		//fmt.Println("ListFindings: ", list)
		for _, v := range list.FindingArns {
			input := &inspector.DescribeFindingsInput{
				FindingArns: []*string{
					aws.String(*v),
				},
			}
			//fmt.Println("String: ", *v)

			result, err := svc.DescribeFindings(input)
			if err != nil {
				iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("DescribeFindings failed: ", err.Error())
				return false, err
			}
			iLog.WithFields(logrus.Fields{"Test": "Inspector"}).Info("Sev: ", *result.Findings[0].NumericSeverity, ", ", *result.Findings[0].Description)
		}
		if list.NextToken != nil {
			nextToken = list.NextToken
		} else {
			break
		}
	}
	return true, nil
}
