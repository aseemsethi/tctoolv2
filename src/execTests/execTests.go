package execTests

import (
	//"fmt"
	"github.com/aseemsethi/tctoolv2/src/globals"
	"github.com/sirupsen/logrus"
)

var mLog *logrus.Logger

func ExecTests(globals *globals.TcGlobals) {
	mLog = globals.Log
	mLog.WithFields(logrus.Fields{
		"Test": "Exec"}).Info("execTests: started")
}
