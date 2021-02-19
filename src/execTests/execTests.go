package execTests

import (
	//"errors"
	//"fmt"
	"github.com/aseemsethi/tctoolv2/src/globals"
	"github.com/sirupsen/logrus"
)

var iLog *logrus.Logger
var fLog *logrus.Logger

var globalTests = map[string][]globals.Tcs{
	"cis":         cisTestCases,
	"inspector":   inspectorTestCases,
	"config":      configTestCases,
	"securityHub": securityHubTestCases,
}

var cisTestCases = []globals.Tcs{
	{"CIS", "Generate Credential Report", CredentialsInitialize},
}

var inspectorTestCases = []globals.Tcs{
	{"Inspector", "Inspector Init", InitInspector},
	{"Inspector", "Inspector Run", RunInspector},
}

var configTestCases = []globals.Tcs{
	{"config", "Config Init", InitConfig},
	{"config", "Config Run", RunConfig},
}

var securityHubTestCases = []globals.Tcs{
	{"securityHub", "Init SecurityHub", InitSecurityHub},
	{"securityHub", "Run SecurityHub", RunSecurityHub},
}

func cis11(g *globals.TcGlobals) (bool, error) {
	globals.SevCount["critical"] += 1
	return true, nil // errors.New("Test Passed")
}

// Contains tells whether a contains x.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func ExecTests(globals *globals.TcGlobals) {
	iLog = globals.Log
	fLog = globals.FLog

	iLog.WithFields(logrus.Fields{
		"Test": "Exec"}).Info("execTests: started")
	for k, tests := range globalTests { // Tests in Code
		if Contains(globals.Config.EnabledTests, k) { // Tests in Config.yml
			iLog.WithFields(logrus.Fields{
				"Test": k}).Info("Starting *****************************************", k)
			for _, elem := range tests {
				if _, err := elem.Run(globals); err != nil {
					iLog.WithFields(logrus.Fields{
						"Test": elem.Id, "Descr": elem.Descr}).Info("Failed")
				} else {
					iLog.WithFields(logrus.Fields{
						"Test": elem.Id, "Descr": elem.Descr}).Info("Passed")
				}
			}
		}
	}
}
