package execTests

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/aseemsethi/tctoolv2/src/globals"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net/url"
	"strings"
	"time"
)

const (
	crUser                      = iota
	crArn                       = iota
	crUserCreationTime          = iota
	crPasswordEnabled           = iota
	crPasswordLastUsed          = iota
	crPasswordLastChanged       = iota
	crPasswordNextRotation      = iota
	crMfaActive                 = iota
	crAccessKey1Active          = iota
	crAccessKey1LastRotated     = iota
	crAccessKey1LastUsedDate    = iota
	crAccessKey1LastUsedRegion  = iota
	crAccessKey1LastUsedService = iota
	crAccessKey2Active          = iota
	crAccessKey2LastRotated     = iota
	crAccessKey2LastUsedDate    = iota
	crAccessKey2LastUsedRegion  = iota
	crAccessKey2LastUsedService = iota
	crCert1Active               = iota
	crCert1LastRotated          = iota
	crCert2Active               = iota
	crCert2LastRotated          = iota
)

type CredentialReport struct {
	Name string
}

var Access_Key_1_Last_Used_Date = 10
var Access_Key_2_Last_Used_Date = 15
var iLog *logrus.Logger

func CredentialsInitialize(g *globals.TcGlobals) (bool, error) {
	firstimte := false
	iLog = globals.Globals.Log
	iLog.WithFields(logrus.Fields{
		"Test": "CIS"}).Info("CredentialReport init...")

	resp, err := globals.Globals.IamSvc.GenerateCredentialReport(&iam.GenerateCredentialReportInput{})
	if err != nil {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS"}).Info("GenerateCredentialReport Failed: ", err.Error())
	}
start:
	if *resp.State == "COMPLETE" {
		//fmt.Printf("\nCredentialReport GetCredRept..")
		resp, get_err := globals.Globals.IamSvc.GetCredentialReport(&iam.GetCredentialReportInput{})
		if get_err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case iam.ErrCodeCredentialReportNotPresentException:
					fmt.Println(iam.ErrCodeCredentialReportNotPresentException, aerr.Error())
				case iam.ErrCodeCredentialReportExpiredException:
					fmt.Println(iam.ErrCodeCredentialReportExpiredException, aerr.Error())
				case iam.ErrCodeCredentialReportNotReadyException:
					fmt.Println(iam.ErrCodeCredentialReportNotReadyException, aerr.Error())
				case iam.ErrCodeServiceFailureException:
					fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				fmt.Println(get_err.Error())
			}
		}

		//fmt.Println("\n", string(resp.Content))
		globals.Globals.Cred = string(resp.Content)
		//iLog.WithFields(logrus.Fields{
		//	"Test": "CIS"}).Info("Credential Rept generated")
		return true, nil
	} else {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS"}).Info("Credential Rept Not generated")
		if firstimte == false {
			firstimte = true
			time.Sleep(10 * time.Second)
			resp, err = globals.Globals.IamSvc.GenerateCredentialReport(&iam.GenerateCredentialReportInput{})
			if err != nil {
				iLog.WithFields(logrus.Fields{
					"Test": "CIS"}).Info("GenerateCredentialReport Failed 2nd time...exiting: : ", err.Error())
				return false, err
			} else {
				iLog.WithFields(logrus.Fields{
					"Test": "CIS"}).Info("GenerateCredentialReport Passed 2nd time")
				goto start
			}
		}
	}
	return true, nil
}

func stringToBool(input string) (output bool) {
	if strings.ToLower(input) == "true" {
		output = true
	}
	return
}

func RootAccessKeysDisabled(i *CredentialReport) {
	s := strings.Split(globals.Globals.Cred, "\n")

	for _, each := range s {
		//1.1 Avoid the use of the "root" account
		//fmt.Println("\n...", each)
		if strings.Contains(each, "<root_account>") {
			root_account := csv.NewReader(strings.NewReader(each))
			record, err := root_account.Read()
			if err != nil {
				log.Fatal(err)
				iLog.WithFields(logrus.Fields{
					"Test": "CIS", "Num": 1.12}).Info("CSV read for root cred Failed: ", err)
			}
			if record[Access_Key_1_Last_Used_Date] != "N/A" && record[Access_Key_2_Last_Used_Date] != "N/A" {
				iLog.WithFields(logrus.Fields{
					"Test": "CIS", "Num": 1.12, "Result": "Failed",
				}).Info("RootAccessKeysDisabled")
			} else {
				iLog.WithFields(logrus.Fields{
					"Test": "CIS", "Num": 1.12, "Result": "Passed",
				}).Info("RootAccessKeysDisabled")
			}
		}
		//fmt.Println(index, each)
	}
}

func ParseCredentialFile(i *CredentialReport) {
	var err error
	var credReportItem globals.CredentialReportItem

	iLog.WithFields(logrus.Fields{
		"Test": "CIS"}).Info("ParseCredentialFile")
	reader := csv.NewReader(strings.NewReader(globals.Globals.Cred))
	var readErr error
	var record []string
	//var credReportItem credentialReportItem
	for {
		record, readErr = reader.Read()
		if len(record) > 0 && record[0] == "user" && record[1] == "arn" {
			continue
		}
		if readErr == io.EOF {
			break
		}
		var userName string
		if record[crUser] == "<root_account>" {
			userName = "root"
		} else {
			userName = record[crUser]
		}
		//fmt.Println(userName)
		var (
			passwordEnabled, mfaActive, accessKey1Active, accessKey2Active, cert1Active, cert2Active bool
			userCreationTime, passwordLastUsed, passwordLastChanged, passwordNextRotation,
			accessKey1LastRotated, accessKey1LastUsedDate, accessKey2LastRotated, accessKey2LastUsedDate,
			cert1LastRotated, cert2LastRotated time.Time
		)
		userCreationTime, err = time.Parse(time.RFC3339, record[crUserCreationTime])
		if err != nil {
			// Invoking an empty time.Time struct literal will return Go's zero date.
			userCreationTime = time.Time{}
		}

		passwordEnabled = stringToBool(record[crPasswordEnabled])

		passwordLastUsed, err = time.Parse(time.RFC3339, record[crPasswordLastUsed])
		if err != nil {
			passwordLastUsed = time.Time{}
		}
		passwordLastChanged, err = time.Parse(time.RFC3339, record[crPasswordLastChanged])
		if err != nil {
			passwordLastChanged = time.Time{}
		}

		passwordNextRotation, err = time.Parse(time.RFC3339, record[crPasswordNextRotation])
		if err != nil {
			passwordNextRotation = time.Time{}
		}
		mfaActive = stringToBool(record[crMfaActive])
		accessKey1Active = stringToBool(record[crAccessKey1Active])

		accessKey1LastRotated, err = time.Parse(time.RFC3339, record[crAccessKey1LastRotated])
		if err != nil {
			accessKey1LastRotated = time.Time{}
		}
		accessKey1LastUsedDate, err = time.Parse(time.RFC3339, record[crAccessKey1LastUsedDate])
		if err != nil {
			accessKey1LastUsedDate = time.Time{}
		}
		accessKey2Active = stringToBool(record[crAccessKey2Active])

		accessKey2LastRotated, err = time.Parse(time.RFC3339, record[crAccessKey2LastRotated])
		if err != nil {
			accessKey2LastRotated = time.Time{}
		}
		accessKey2LastUsedDate, err = time.Parse(time.RFC3339, record[crAccessKey2LastUsedDate])
		if err != nil {
			accessKey2LastUsedDate = time.Time{}
		}
		cert1Active = stringToBool(record[crCert1Active])

		cert1LastRotated, err = time.Parse(time.RFC3339, record[crCert1LastRotated])
		if err != nil {
			cert1LastRotated = time.Time{}
		}
		cert2Active = stringToBool(record[crCert2Active])

		cert2LastRotated, err = time.Parse(time.RFC3339, record[crCert2LastRotated])
		if err != nil {
			cert2LastRotated = time.Time{}
			err = nil
		}

		credReportItem = globals.CredentialReportItem{
			Arn:                       record[crArn],
			User:                      userName,
			UserCreationTime:          userCreationTime,
			PasswordEnabled:           passwordEnabled,
			PasswordLastUsed:          passwordLastUsed,
			PasswordLastChanged:       passwordLastChanged,
			PasswordNextRotation:      passwordNextRotation,
			MfaActive:                 mfaActive,
			AccessKey1Active:          accessKey1Active,
			AccessKey1LastRotated:     accessKey1LastRotated,
			AccessKey1LastUsedDate:    accessKey1LastUsedDate,
			AccessKey1LastUsedRegion:  record[crAccessKey1LastUsedRegion],
			AccessKey1LastUsedService: record[crAccessKey1LastUsedService],
			AccessKey2Active:          accessKey2Active,
			AccessKey2LastRotated:     accessKey2LastRotated,
			AccessKey2LastUsedDate:    accessKey2LastUsedDate,
			AccessKey2LastUsedRegion:  record[crAccessKey2LastUsedRegion],
			AccessKey2LastUsedService: record[crAccessKey2LastUsedService],
			Cert1Active:               cert1Active,
			Cert1LastRotated:          cert1LastRotated,
			Cert2Active:               cert2Active,
			Cert2LastRotated:          cert2LastRotated,
		}
		globals.Globals.CredReport = append(globals.Globals.CredReport, credReportItem)
		//fmt.Printf("%+v", credReportItem)
	}
}

func MFAEnabled(i *CredentialReport) {
	failed := false
	for _, elem := range globals.Globals.CredReport {
		//fmt.Println("Check User: ", elem.Arn)
		if elem.MfaActive == false {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": "1.2 1.13", "Result": "Failed",
			}).Info("MFA Disabled for user: ", elem.Arn)
			failed = true
		}
	}
	if failed == false {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": "1.2 1.13", "Result": "Passed",
		}).Info("MFA Enabled for all users")
	}
}

func TimeLastUsedAccessKeys(i *CredentialReport) {
	failed := false
	for _, elem := range globals.Globals.CredReport {
		// If the AccessKey is never used, it will show as N/A, and a time coversion on this will yield an error
		// At that tiem, we save null vaule in this time field
		if elem.AccessKey1LastUsedDate.IsZero() == true {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": "1.3", "Result": "Failed",
			}).Info("AccessKey credentials never used for user: ", elem.Arn)
			failed = true
		} else {
			diff := time.Now().Sub(elem.AccessKey1LastUsedDate).Hours()
			diff1 := fmt.Sprintf("%.1f", diff)
			//fmt.Println("Time elapsed for User: ", elem.Arn, " is ", diff1, " Hours")
			if diff > 90*24 {
				iLog.WithFields(logrus.Fields{
					"Test": "CIS", "Num": "1.3", "Result": "Failed",
				}).Info("TimeLastUsedAccessKeys last hrs:", diff1, " for user: ", elem.Arn)
				failed = true
			} else {
				iLog.WithFields(logrus.Fields{
					"Test": "CIS", "Num": "1.3", "Result": "Passed",
				}).Info("TimeLastUsedAccessKeys last hrs:", diff1, " for user: ", elem.Arn)
			}
		}
	}
	if failed == false {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": "1.3", "Result": "Passed",
		}).Info("AccessKey credentials used")
	}
}

func TimeLastRotatedAccessKeys(i *CredentialReport) {
	failed := false
	for _, elem := range globals.Globals.CredReport {
		// If the AccessKey is never used, it will show as N/A, and a time coversion on this will yield an error
		// At that tiem, we save null vaule in this time field
		if elem.AccessKey1LastRotated.IsZero() == true {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": "1.4", "Result": "Failed",
			}).Info("TimeLastRotatedAccessKeys more than 90 days for user: ", elem.Arn)
			failed = true
		} else {
			diff := time.Now().Sub(elem.AccessKey1LastRotated).Hours()
			diff1 := fmt.Sprintf("%.1f", diff)
			//fmt.Println("Time elapsed for User: ", elem.Arn, " is ", diff1, " Hours")
			if diff > 90*24 {
				iLog.WithFields(logrus.Fields{
					"Test": "CIS", "Num": "1.4", "Result": "Failed",
				}).Info("TimeLastRotatedAccessKeys more than 90 days, last rotated: ", diff1, " for user: ", elem.Arn)
				failed = true
			} else {
				iLog.WithFields(logrus.Fields{
					"Test": "CIS", "Num": "1.4", "Result": "Passed",
				}).Info("TimeLastRotatedAccessKeys last rotated: ", diff1, " for user: ", elem.Arn)
			}
		}
	}
	if failed == false {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": "1.4", "Result": "Passed",
		}).Info("TimeLastRotatedAccessKeys")
	}
}

func policyAttachedToUserCheck(i *CredentialReport) {
	found := false
	for _, cred := range globals.Globals.CredReport {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 1.16,
		}).Info("policyAttachedToUserCheck for user: ", cred.User)
		attachedPolicies, err := globals.Globals.IamSvc.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{UserName: aws.String(cred.User)})
		if err != nil {
			if cred.User == "root" {
				// A policy retrieval for root gives an error, so we skip root for this test. No username 'root' found
				continue
			}
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Result": "Failed",
			}).Info("policyAttachedToUserCheck failed to list policies for user: ", cred.User, err)
			continue
		}
		found = false
		for _, attachedPolicy := range attachedPolicies.AttachedPolicies {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 1.16,
			}).Info("policyAttachedToUserCheck found for user: ", cred.User, " Policy: ", attachedPolicy.PolicyArn)
			found = true
		}
		if found == false {
			fmt.Println("No Policy attached to user: ", cred.User)
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 1.16,
			}).Info("policyAttachedToUserCheck not found for user: ", cred.User)
		}
	}
	if found == true {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 1.16, "Result": "Failed",
		}).Info("No IAM Policy attachd to user")
	} else {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 1.16, "Result": "Passed",
		}).Info("IAM Policy attachd to user")
	}
}

func listAllPolicies(i *CredentialReport) {
	actions := []string{"*"}
	resources := []string{"*"}

	params := &iam.ListPoliciesInput{
		Scope: aws.String("Local"), // only looking at non AWS policies
	}
	resp, err := globals.Globals.IamSvc.ListPolicies(params)
	if err != nil {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 1.17}).Info("Error retrieving policies: ", err)
		return
	}
	//fmt.Println("Policy: ", resp)

	for _, val := range resp.Policies {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 1.17}).Info("Checking Policy: ", *val.Arn)
		params1 := &iam.GetPolicyVersionInput{
			PolicyArn: aws.String(*val.Arn), // Required
			VersionId: aws.String("v2"),     // Required
		}
		resp1, err := globals.Globals.IamSvc.GetPolicyVersion(params1)
		if err != nil {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 1.17}).Info("Error retrieving policy doc: ", err)
			continue
		}
		// The policy document returned in this structure is URL-encoded compliant with RFC 3986 .
		// You can use a URL decoding method to convert the policy back to plain JSON text.
		//fmt.Println(awsutil.StringValue(resp1))
		doc := globals.PolicyDocument{}
		policy, err := url.QueryUnescape(aws.StringValue(resp1.PolicyVersion.Document))
		if err != nil {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 1.17}).Info("Error decoding policy doc: ", err)
			continue
		}
		//fmt.Println("Policy:", policy)
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 1.17,
		}).Info("IAM Policy dump ******* ", policy)
		err = json.Unmarshal([]byte(policy), &doc)
		// ensure policy should not have any Statement block with "Effect":
		//"Allow" and Action set to "*" and Resource set to "*"
		for _, v := range doc.Statement {
			hasActions := v.Action.Contains(actions)
			hasResources := v.Resource.Contains(resources)
			//fmt.Println("Resource:", *v.Resource, " Actions:", *v.Action, " Checking: ", actions, resources)
			hasEffect := v.Effect
			//fmt.Println("hasActions:", hasActions, "hasRes: ", hasResources, "hasEffects:", hasEffect)
			res := hasActions && hasResources && (hasEffect == "Allow")
			if res {
				iLog.WithFields(logrus.Fields{
					"Test": "CIS", "Num": 1.17, "Result": "Failed",
				}).Info("IAM Policy allows * access to all Resources, ", *val.Arn)
			} else {
				iLog.WithFields(logrus.Fields{
					"Test": "CIS", "Num": 1.17, "Result": "Passed",
				}).Info("IAM Policy disallows * access to all Resources, ", *val.Arn)
			}
		}
	}
}

func (i *CredentialReport) Run() {
	iLog.WithFields(logrus.Fields{
		"Test": "CIS"}).Info("CredentialReport Run...")
	RootAccessKeysDisabled(i)
	ParseCredentialFile(i)
	MFAEnabled(i)
	TimeLastUsedAccessKeys(i)
	TimeLastRotatedAccessKeys(i)
	policyAttachedToUserCheck(i)
	listAllPolicies(i)
}
