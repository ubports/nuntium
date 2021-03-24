package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"launchpad.net/go-xdg/v0"
)

type mainFlags struct {
	// Sender affects only the MMS payload (not notification).
	Sender string `long:"sender" short:"s" description:"The sender of the multimedia message (if not set or empty, it defaults to: 01189998819991197253)"`
	// SenderNotification is only used in the push notification.
	SenderNotification string `long:"sender-notification" description:"The sender of the message push notification (if not set or empty, it defaults to: +543515924906)"`
	// EndPoint is the name where nuntium listens to on the System Bus.
	EndPoint string `long:"end-point" description:"Dbus name where the nuntium agent is listening for push requests from ofono (if not set or empty, it tries to retrieve the end-point from logs)"`
	// MRetrieveConf is an alternative file to use as m-retrieve.conf, no mangling is done with it.
	MRetrieveConf string `long:"m-retrieve-conf" description:"Use a specific m-retrieve.conf to test (the --sender flag will not be considered)"`
	// DenialCount is an integer, which indicates how many times will MMS content serving be denied until successfuly served.
	DenialCount int `long:"denial-count" short:"d" description:"Number of serving denials until successful message serving" default:"0"`
	// TransactionId is an string, which uniqely identifies an message.
	TransactionId string `long:"transaction-id" short:"t" description:"Unique identifier for the push notification. If empty, TransactionId will not be filled"`
	// ErrorActivateContext indicates how many nuntium ErrorActivateContext errors will be thrown (and communicated to telepathy), before message download attempt starts.
	ErrorActivateContext uint64 `long:"error-activate-context" description:"Number of activate context errors before message download attempt starts"`
	// ErrorGetProxy indicates how many nuntium ErrorGetProxy errors will be thrown (and communicated to telepathy), before message download attempt starts.
	ErrorGetProxy uint64 `long:"error-get-proxy" description:"Number of get proxy errors before message download attempt starts"`
	// ErrorDownloadStorage indicates how many nuntium ErrorDownloadStorage errors will be thrown (and communicated to telepathy), before message download attempt starts.
	ErrorDownloadStorage uint64 `long:"error-storage" description:"Number of storage errors after successful message download"`
}

func main() {
	var args mainFlags

	parser := flags.NewParser(&args, flags.Default)
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	if args.EndPoint == "" {
		cacheDir := filepath.Join(xdg.Cache.Dirs()[0], "upstart")
		cmd := "find " + cacheDir + " -name 'nuntium*' | xargs zgrep 'Registering' | sed 's/^[^:]*://' | sort | tail -n1 | cut -d: -f4"
		fmt.Printf("No end-point flag provided, trying to retrieve end-point from nuntium logs in: %s\n", cacheDir)
		out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
		if err != nil {
			fmt.Printf("Command \"%s\" execution error: %v\n", cmd, err)
			os.Exit(1)
		}
		if len(out) == 0 {
			fmt.Printf("No end-point found\n")
			os.Exit(1)
		}
		args.EndPoint = ":" + strings.Split(strings.Split(string(out), "\n")[0], "\r")[0]
		fmt.Printf("Using endpoint: \"%s\"\n", args.EndPoint)
	}

	fmt.Println("Creating web server to serve mms")
	done := make(chan bool)
	mmsHandler, err := createSpace(args, done)
	if err != nil {
		fmt.Println("Issues while creating mms local server instance:", err)
		os.Exit(1)
	}

	fmt.Println("Denial count:", args.DenialCount)
	http.HandleFunc("/mms", func(w http.ResponseWriter, r *http.Request) {
		if args.DenialCount > 0 {
			fmt.Println("Serving MMS content denied")
			args.DenialCount -= 1
			http.Error(w, "Intentional denial", http.StatusInternalServerError)
			return
		}
		fmt.Print("Serving MMS content...")
		mmsHandler(w, r)
		fmt.Println(" Done")
	})

	go http.ListenAndServe("localhost:9191", nil) //http.FileServer(http.Dir(servedDir)))

	if err := push(args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	<-done

	fmt.Println("Finished serving mms")
}
