package main

import (
	"fmt"
	"io/ioutil"
	"launchpad.net/nuntium/mms"
	"os"
	"path/filepath"
)

func main() {
	var targetPath string
	if len(os.Args) < 2 {
		fmt.Println("Missing filepath to MMS to decode")
		os.Exit(1)
	} else if len(os.Args) == 3 {
		targetPath = os.Args[2]
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

	retConfHdr := mms.NewMRetrieveConf(mmsFile)
	dec := mms.NewDecoder(mmsData)
	if err := dec.Decode(retConfHdr); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if targetPath != "" {
		fmt.Println("Saving to", targetPath)
		writeParts(targetPath, retConfHdr.DataParts)
	}
}

func writeParts(targetPath string, parts []mms.ContentType) {
	if fi, err := os.Stat(targetPath); err != nil {
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			fmt.Println(err)
		}
	} else if !fi.IsDir() {
		fmt.Println(targetPath, "is not a directory")
		os.Exit(1)
	}

	for i, _ := range parts {
		if parts[i].Name != "" {
			ioutil.WriteFile(filepath.Join(targetPath, parts[i].Name), parts[i].Data, 0644)
		}
		fmt.Println(parts[i].MediaType, parts[i].Name)
	}
}
