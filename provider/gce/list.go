package gce

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fatih/images/provider/utils"
)

type listFlags struct {
	output  utils.OutputMode
	helpMsg string
	flagSet *flag.FlagSet
}

func newListFlags() *listFlags {
	l := &listFlags{}

	flagSet := flag.NewFlagSet("copy", flag.ContinueOnError)
	flagSet.Var(utils.NewOutputValue(utils.Simplified, &l.output), "output", "Output mode")
	l.helpMsg = `Usage: images list --providers gce [options]

   List images

Options:

  -output  "json"              Output mode of images. (default: "simplified")
                               Available options: "json","table" or "simplified" 
`

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, l.helpMsg)
	}
	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	l.flagSet = flagSet
	return l
}
