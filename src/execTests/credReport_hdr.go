package execTests

import (
	"strings"
	"time"
)

func stringToBool(input string) (output bool) {
	if strings.ToLower(input) == "true" {
		output = true
	}
	return
}

type credentialReportItem struct {
	User                      string
	Arn                       string
	UserCreationTime          time.Time
	PasswordEnabled           bool
	PasswordLastUsed          time.Time
	PasswordLastChanged       time.Time
	PasswordNextRotation      time.Time
	MfaActive                 bool
	AccessKey1Active          bool
	AccessKey1LastRotated     time.Time
	AccessKey1LastUsedDate    time.Time
	AccessKey1LastUsedRegion  string
	AccessKey1LastUsedService string
	AccessKey2Active          bool
	AccessKey2LastRotated     time.Time
	AccessKey2LastUsedDate    time.Time
	AccessKey2LastUsedRegion  string
	AccessKey2LastUsedService string
	Cert1Active               bool
	Cert1LastRotated          time.Time
	Cert2Active               bool
	Cert2LastRotated          time.Time
}
type credentialReport []credentialReportItem

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
