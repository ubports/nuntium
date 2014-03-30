package main

import (
	"fmt"
	"launchpad.net/nuntium/mms"
	"os"
	"io/ioutil"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Missing filepath to MMS to decode")
		os.Exit(1)
	}
	
	mmsFile := os.Args[1]
	if _, err := os.Stat(mmsFile); os.IsNotExist(err) {
		fmt.Printf("File argument %s does no exist\n", mmsFile)
		os.Exit(1)
	}

	mmsData, err := ioutil.ReadFile(mmsFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	retConfHdr := mms.NewMRetrieveConf()
	dec := mms.NewDecoder(mmsData)
	if err := dec.Decode(retConfHdr); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Decoded message: ", retConfHdr)

}
