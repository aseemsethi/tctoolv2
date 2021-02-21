package execTests

import (
	"github.com/aseemsethi/tctoolv2/src/globals"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/sirupsen/logrus"
)

type Iam struct {
	Name string
}

func InitIam(g *globals.TcGlobals) (bool, error) {
	var params *iam.GetAccountPasswordPolicyInput
	resp, err := g.IamSvc.GetAccountPasswordPolicy(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": "1.5 - 1.11", "Result": "Failed",
		}).Info(err.Error())
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": "1.5 - 1.11", "Result": "Failed",
		}).Info("Password Policy does not exist")
		return false, err
	}

	// Pretty-print the response data.
	iLog.WithFields(logrus.Fields{
		"Test": "CIS", "Num": "1.5 - 1.11"}).Info("Password Policy dump: ", resp)
	g.PwdPolicy = resp
	return true, nil
}

func PwdPolicyCheck(g *globals.TcGlobals) {
	if *g.PwdPolicy.PasswordPolicy.RequireUppercaseCharacters ||
		*g.PwdPolicy.PasswordPolicy.RequireLowercaseCharacters ||
		*g.PwdPolicy.PasswordPolicy.RequireNumbers ||
		*g.PwdPolicy.PasswordPolicy.RequireSymbols {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": "1.5 - 1.8", "Result": "Failed",
		}).Info("Password Policy doesn't require Uppercase/Lowercase Letters, Numbers and Symbols")
	} else {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": "1.5 - 1.8", "Result": "Passed",
		}).Info("Password Policy doesn't require Uppercase/Lowercase Letters, Numbers and Symbols")
	}

	if *g.PwdPolicy.PasswordPolicy.MinimumPasswordLength < 14 {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 1.9, "Result": "Failed",
		}).Info("Minimum Password length less than 14 chars")
	} else {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 1.9, "Result": "Passed",
		}).Info("Minimum Password length is more than 14 chars")
	}

	if g.PwdPolicy.PasswordPolicy.PasswordReusePrevention == nil || *g.PwdPolicy.PasswordPolicy.PasswordReusePrevention < 3 {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 1.10, "Result": "Failed",
		}).Info("Password reuse policy < 3 days or not set - CIS 1.10 failed")
	} else {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 1.10, "Result": "Passed",
		}).Info("Password reuse policy - CIS 1.10 passed")
	}

	if g.PwdPolicy.PasswordPolicy.MaxPasswordAge == nil || *g.PwdPolicy.PasswordPolicy.MaxPasswordAge < 90 {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 1.11, "Result": "Failed",
		}).Info("Passwords don't expire after at least 90 days")
	} else {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 1.11, "Result": "Passed",
		}).Info("Passwords expires after at least 90 days")
	}
}

// TBD: We need to really check MFA HArdware for root
func mfsDeviceCheck(g *globals.TcGlobals) {
	found := false
	mfaDevices, err := g.IamSvc.ListMFADevices(&iam.ListMFADevicesInput{UserName: aws.String("admin")}) // TBD: Need to check for root user only
	if err != nil {
		iLog.WithFields(logrus.Fields{
			"Test": "CIS", "Num": 1.14,
		}).Info("Failed to list mfa devices - %s", err)
	} else {
		for _, device := range mfaDevices.MFADevices {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 1.14, "Result": "Passed",
			}).Info("MFA enabled for admin with Device: ", device.SerialNumber)
			found = true
		}
		if found == false {
			iLog.WithFields(logrus.Fields{
				"Test": "CIS", "Num": 1.14, "Result": "Failed",
			}).Info("MFA not enabled for admin")
		}
	}
}

func RunIam(g *globals.TcGlobals) (bool, error) {
	iLog.WithFields(logrus.Fields{
		"Test": "CIS"}).Info("IAM Run...")
	PwdPolicyCheck(g)
	mfsDeviceCheck(g)
	return true, nil
}
