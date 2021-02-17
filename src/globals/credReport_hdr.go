package globals

import (
	"time"
)

type CredentialReportItem struct {
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

type credentialReport []CredentialReportItem
