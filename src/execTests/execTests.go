package execTests

import (
	"errors"
	"fmt"
	"github.com/aseemsethi/tctoolv2/src/globals"
	"github.com/sirupsen/logrus"
)

var mLog *logrus.Logger

var tcCases = []globals.Tcs{
	{"CIS1.1", "Avoid Root Account", cis11},
	{"CIS1.2", "Enable MFA for all Accounts", cis12},
}

func cis11(*globals.TcGlobals) (bool, error) {
	fmt.Print("cis11 called")
	return true, nil // errors.New("Test Passed")
}
func cis12(*globals.TcGlobals) (bool, error) {
	fmt.Print("cis12 called")
	return false, errors.New("Test Failed")
}

func ExecTests(globals *globals.TcGlobals) {
	mLog = globals.Log
	mLog.WithFields(logrus.Fields{
		"Test": "Exec"}).Info("execTests: started")
	for _, elem := range tcCases {
		if result, err := elem.Run(globals); err != nil {
			fmt.Print("Failed Test", result)
		} else {
			fmt.Print("Passed Test", result)
		}
	}
}
