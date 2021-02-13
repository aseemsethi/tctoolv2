package main

import (
	"fmt"
	"github.com/aseemsethi/tctoolv2/src/globals"
	"github.com/sirupsen/logrus"
)

var mLog *logrus.Logger

// Call with tctool <region> <accountid>
func main() {
	fmt.Printf("\nTest Compliance Tool Starting..")

	globals.Globals.Initialize()
	mLog = globals.Globals.Log
	mLog.WithFields(logrus.Fields{
		"Test": "Init"}).Info("Security Tests Starting.....................................................")
}
