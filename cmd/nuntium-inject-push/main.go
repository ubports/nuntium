package main

import (
	"fmt"
	"net/http"
	"os"

	flags "github.com/jessevdk/go-flags"
)

func main() {
	var args struct {
		// Sender is only used in the push notification.
		Sender string `long:"sender" short:"s" description:"the sender of the MMS" default:"0118 999 881 99 9119 7253"`
		// EndPoint is the name where nuntium listens to on the System Bus.
		EndPoint string `long:"end-point" required:"true" description:"dbus name where the nuntium agent is listening for push requests from ofono"`
		// MRetrieveConf is an alternative file to use as m-retrieve.conf, no mangling is done with it.
		MRetrieveConf string `long:"m-retrieve-conf" description:"Use a specific m-retrieve.conf to test"`
		// DenialCount is an integer, which indicates how many times will MMS content serving be denied until successfuly served.
		DenialCount int `long:"denial-count" short:"d" description:"number of serving denials until successful MMS serving" default:"0"`
	}

	parser := flags.NewParser(&args, flags.Default)
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	fmt.Println("Creating web server to serve mms")
	done := make(chan bool)
	mmsHandler, err := createSpace(args.MRetrieveConf, done)
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

	if err := push(args.EndPoint, args.Sender); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	<-done

	fmt.Println("Finished serving mms")
}
