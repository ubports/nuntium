/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Manuel de la Pena: manuel.delapena@canonical.com
 *
 * This file is part of ubuntu-download-manager.
 *
 * ubuntu-download-manager is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.
 *
 * ubuntu-download-manager is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

// Package udm provides a go interface to work with the ubuntu download manager
package udm

import (
	"errors"
	"launchpad.net/go-dbus/v1"
	"runtime"
)

const (
	DOWNLOAD_SERVICE           = "com.canonical.applications.Downloader"
	DOWNLOAD_INTERFACE         = "com.canonical.applications.Download"
	DOWNLOAD_MANAGER_INTERFACE = "com.canonical.applications.DownloadManager"
)

type hashType string

const (
	MD5    hashType = "md5"
	SHA1   hashType = "sha1"
	SHA224 hashType = "sha224"
	SHA256 hashType = "sha256"
	SHA384 hashType = "sha384"
	SHA512 hashType = "sha512"
)

const (
	LOCAL_PATH            = "local-path"
	OBJECT_PATH           = "objectpath"
	POST_DOWNLOAD_COMMAND = "post-download-command"
)

// Progress provides how much progress has been performed in a download that was
// already started.
type Progress struct {
	Received uint64
	Total    uint64
}

// Download is the common interface of a download. It provides all the required
// methods to interact with a download created by udm.
type Download interface {
	TotalSize() (uint64, error)
	Progress() (uint64, error)
	Metadata() (map[string]string, error)
	SetThrottle(uint64) error
	Throttle() (uint64, error)
	AllowMobileDownload(bool) error
	IsMobileDownload() (bool, error)
	Start() error
	Pause() error
	Resume() error
	Cancel() error
	Started() chan bool
	Paused() chan bool
	DownloadProgress() chan Progress
	Resumed() chan bool
	Canceled() chan bool
	Finished() chan string
	Error() chan error
}

// Manager is the single point of entry of the API. Allows to interact with the
// general setting of udm as well as to create downloads at will.
type Manager interface {
	CreateDownload(string, string, string, map[string]interface{}, map[string]string) (Download, error)
	CreateMmsDownload(string, string, int, string, string) (Download, error)
}

// FileDownload represents a single file being downloaded by udm.
type FileDownload struct {
	conn       *dbus.Connection
	proxy      *dbus.ObjectProxy
	path       dbus.ObjectPath
	started    chan bool
	started_w  *dbus.SignalWatch
	paused     chan bool
	paused_w   *dbus.SignalWatch
	resumed    chan bool
	resumed_w  *dbus.SignalWatch
	canceled   chan bool
	canceled_w *dbus.SignalWatch
	finished   chan string
	finished_w *dbus.SignalWatch
	errors     chan error
	error_w    *dbus.SignalWatch
	progress   chan Progress
	progress_w *dbus.SignalWatch
}

func connectToSignal(conn *dbus.Connection, path dbus.ObjectPath, signal string) (*dbus.SignalWatch, error) {
	w, err := conn.WatchSignal(&dbus.MatchRule{
		Type:      dbus.TypeSignal,
		Sender:    DOWNLOAD_SERVICE,
		Interface: DOWNLOAD_INTERFACE,
		Member:    signal,
		Path:      path})
	return w, err
}

func (down *FileDownload) free() {
	// cancel all watches so that goroutines are done and close the
	// channels
	down.started_w.Cancel()
	down.paused_w.Cancel()
	down.resumed_w.Cancel()
	down.canceled_w.Cancel()
	down.finished_w.Cancel()
	down.error_w.Cancel()
	down.progress_w.Cancel()
}

func cleanDownloadData(down *FileDownload) {
	down.free()
}

func newFileDownload(conn *dbus.Connection, path dbus.ObjectPath) (*FileDownload, error) {
	proxy := conn.Object(DOWNLOAD_SERVICE, path)
	started_ch := make(chan bool)
	started_w, err := connectToSignal(conn, path, "started")
	if err != nil {
		return nil, err
	}

	paused_ch := make(chan bool)
	paused_w, err := connectToSignal(conn, path, "paused")
	if err != nil {
		return nil, err
	}

	resumed_ch := make(chan bool)
	resumed_w, err := connectToSignal(conn, path, "resumed")
	if err != nil {
		return nil, err
	}

	canceled_ch := make(chan bool)
	canceled_w, err := connectToSignal(conn, path, "canceled")
	if err != nil {
		return nil, err
	}

	finished_ch := make(chan string)
	finished_w, err := connectToSignal(conn, path, "finished")
	if err != nil {
		return nil, err
	}

	errors_ch := make(chan error)
	errors_w, err := connectToSignal(conn, path, "error")
	if err != nil {
		return nil, err
	}

	progress_ch := make(chan Progress)
	progress_w, err := connectToSignal(conn, path, "progress")
	if err != nil {
		return nil, err
	}

	d := FileDownload{conn, proxy, path, started_ch, started_w, paused_ch, paused_w, resumed_ch, resumed_w, canceled_ch, canceled_w, finished_ch, finished_w, errors_ch, errors_w, progress_ch, progress_w}

	// connect to the diff signals so that we have nice channels that do
	// not expose dbus watchers
	d.connectToStarted()
	d.connectToPaused()
	d.connectToResumed()
	d.connectToCanceled()
	d.connectToFinished()
	d.connectToError()
	d.connectToProgress()
	runtime.SetFinalizer(&d, cleanDownloadData)
	return &d, nil
}

// TotalSize returns the total size of the file being downloaded.
func (down *FileDownload) TotalSize() (size uint64, err error) {
	reply, err := down.proxy.Call(DOWNLOAD_INTERFACE, "totalSize")
	if err != nil || reply.Type == dbus.TypeError {
		return 0, err
	}
	if err = reply.Args(&size); err != nil {
		return 0, err
	}
	return size, nil
}

// Process returns the process so far in downloading the file.
func (down *FileDownload) Progress() (progress uint64, err error) {
	reply, err := down.proxy.Call(DOWNLOAD_INTERFACE, "progress")
	if err != nil || reply.Type == dbus.TypeError {
		return 0, err
	}
	if err = reply.Args(&progress); err != nil {
		return 0, err
	}
	return progress, nil
}

// Metadata returns the metadata that was provided at creating time to the download.
func (down *FileDownload) Metadata() (metadata map[string]string, err error) {
	reply, err := down.proxy.Call(DOWNLOAD_INTERFACE, "metadata")
	if err != nil || reply.Type == dbus.TypeError {
		return nil, err
	}
	if err = reply.Args(&metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// SetThrottle sets the network throttle to be used in the download.
func (down *FileDownload) SetThrottle(throttle uint64) (err error) {
	reply, err := down.proxy.Call(DOWNLOAD_INTERFACE, "setThrottle", throttle)
	if err != nil || reply.Type == dbus.TypeError {
		return err
	}
	return nil
}

// Throttle returns the network throttle that is currently used in the download.
func (down *FileDownload) Throttle() (throttle uint64, err error) {
	reply, err := down.proxy.Call(DOWNLOAD_INTERFACE, "throttle")
	if err != nil || reply.Type == dbus.TypeError {
		return 0, err
	}
	if err = reply.Args(&throttle); err != nil {
		return 0, err
	}
	return throttle, nil
}

// AllowMobileDownload returns if the download is allow to use the mobile connect
// connection.
func (down *FileDownload) AllowMobileDownload(allowed bool) (err error) {
	reply, err := down.proxy.Call(DOWNLOAD_INTERFACE, "allowGSMDownload", allowed)
	if err != nil || reply.Type == dbus.TypeError {
		return err
	}
	return nil
}

// IsMobileDownload returns if the download will be performed over the mobile data.
func (down *FileDownload) IsMobileDownload() (allowed bool, err error) {
	reply, err := down.proxy.Call(DOWNLOAD_INTERFACE, "isGSMDownloadAllowed", allowed)
	if err != nil || reply.Type == dbus.TypeError {
		return false, err
	}
	if err = reply.Args(&allowed); err != nil {
		return false, err
	}
	return allowed, nil
}

// Start tells udm that the download is ready to be peformed and that the client is
// ready to recieve signals. The following is a commong pattern to be used when
// creating downloads in udm.
//
//     man, err := udm.NewDownloadManager()
//     if err != nil {
//     }
//
//     // variables used to create the download
//
//     url := "http://www.python.org/ftp/python/3.3.3/Python-3.3.3.tar.xz"
//     hash := "8af44d33ea3a1528fc56b3a362924500"
//     hashAlgo := MD5
//     var metadata map[string]interface{}
//     var headers map[string]string
//
//     // create the download BUT do not start downloading just yet
//     down, err := man.CreateDownload(url, hash, hashAlgo, metadata, headers)
//
//     // connect routines to the download channels so that we can get the
//     // information of the download the channel will not get any data until the
//     // Start is called.
//
//     started_signal := down.Started()
//     go func() {
//         <-started_signal
//         fmt.Println("Download started")
//     }()
//     progress_signal := down.DownloadProgress()
//     go func() {
//         for progress := range p {
//             fmt.Printf("Recieved %d out of %d\n", progress.Received, progress.Total)
//         }
//     }()
//
//     finished_signal := down.Finished()
//
//     // start download
//     down.Start()
//
//     // block until we are finished downloading
//     <- finished_signal
func (down *FileDownload) Start() (err error) {
	reply, err := down.proxy.Call(DOWNLOAD_INTERFACE, "start")
	if err != nil || reply.Type == dbus.TypeError {
		return err
	}
	return nil
}

// Pause pauses a download that was started and if not nothing is done.
func (down *FileDownload) Pause() (err error) {
	reply, err := down.proxy.Call(DOWNLOAD_INTERFACE, "pause")
	if err != nil || reply.Type == dbus.TypeError {
		return err
	}
	return nil
}

// Resumes a download that was paused or does nothing otherwise.
func (down *FileDownload) Resume() (err error) {
	reply, err := down.proxy.Call(DOWNLOAD_INTERFACE, "resume")
	if err != nil || reply.Type == dbus.TypeError {
		return err
	}
	return nil
}

// Cancel cancels a download that was in process and deletes any local files
// that were created.
func (down *FileDownload) Cancel() (err error) {
	reply, err := down.proxy.Call(DOWNLOAD_INTERFACE, "cancel")
	if err != nil || reply.Type == dbus.TypeError {
		return err
	}
	return nil
}

func (down *FileDownload) connectToStarted() {
	go func() {
		for msg := range down.started_w.C {
			var started bool
			msg.Args(&started)
			down.started <- started
		}
		close(down.started)
	}()
}

// Started returns a channel that will be used to communicate the started signals.
func (down *FileDownload) Started() chan bool {
	return down.started
}

func (down *FileDownload) connectToPaused() {
	go func() {
		for msg := range down.paused_w.C {
			var paused bool
			msg.Args(&paused)
			down.paused <- paused
		}
		close(down.paused)
	}()
}

// Paused returns a channel that will be used to communicate the paused signals.
func (down *FileDownload) Paused() chan bool {
	return down.paused
}

func (down *FileDownload) connectToProgress() {
	go func() {
		for msg := range down.progress_w.C {
			var received uint64
			var total uint64
			msg.Args(&received, &total)
			down.progress <- Progress{received, total}
		}
		close(down.progress)
	}()
}

// DownloadProgress returns a channel that will be used to communicate the progress
// signals.
func (down *FileDownload) DownloadProgress() chan Progress {
	return down.progress
}

func (down *FileDownload) connectToResumed() {
	go func() {
		for msg := range down.resumed_w.C {
			var resumed bool
			msg.Args(&resumed)
			down.resumed <- resumed
		}
		close(down.resumed)
	}()
}

// Resumed returns a channel that will be used to communicate the paused signals.
func (down *FileDownload) Resumed() chan bool {
	return down.resumed
}

func (down *FileDownload) connectToCanceled() {
	go func() {
		for msg := range down.canceled_w.C {
			var canceled bool
			msg.Args(&canceled)
			down.canceled <- canceled
		}
		close(down.canceled)
	}()
}

// Canceled returns a channel that will be used to communicate the canceled signals.
func (down *FileDownload) Canceled() chan bool {
	return down.canceled
}

func (down *FileDownload) connectToFinished() {
	go func() {
		for msg := range down.finished_w.C {
			var path string
			msg.Args(&path)
			down.finished <- path
		}
		close(down.finished)
	}()
}

// Finished returns a channel that will ne used to communicate the finished signals.
func (down *FileDownload) Finished() chan string {
	return down.finished
}

func (down *FileDownload) connectToError() {
	go func() {
		for msg := range down.error_w.C {
			var reason string
			msg.Args(&reason)
			down.errors <- errors.New(reason)
		}
		close(down.errors)
	}()
}

// Error returns the channel that will be used to communicate the error signals.
func (down *FileDownload) Error() chan error {
	return down.errors
}

type DownloadManager struct {
	conn  *dbus.Connection
	proxy *dbus.ObjectProxy
}

// NewDownloadManager creates a new manager that can be used to create download in the
// udm daemon.
func NewDownloadManager() (*DownloadManager, error) {
	conn, err := dbus.Connect(dbus.SessionBus)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	proxy := conn.Object(DOWNLOAD_SERVICE, "/")
	d := DownloadManager{conn, proxy}
	return &d, nil
}

// CreateDownload creates a new download in the udm daemon that can be used to get
// a remote resource. Udm allows to pass a hash signature and method that will be
// check once the download has been complited.
//
// The download hash can be one of the the following constants:
//
//   MD5
//   SHA1
//   SHA224
//   SHA256
//   SHA384
//   SHA512
//
// The metadata attribute can be used to pass extra information to the udm daemon
// that will just be considered if the caller is not a apparmor confined application.
//
//     LOCAL_PATH => allows to provide the local path for the download.
//     OBJECT_PATH => allows to provide the object path to be used in the dbus daemon.
//     POST_DOWNLOAD_COMMAND => allows to provide a command that will be executed on the
//         download
//
// The headers attribute allows to provide extra headers to be used in the request used
// to perform the download.
func (man *DownloadManager) CreateDownload(url string, hash string, algo hashType, metadata map[string]interface{}, headers map[string]string) (down Download, err error) {
	var t map[string]*dbus.Variant
	for key, value := range metadata {
		t[key] = &dbus.Variant{Value: value}
	}
	s := struct {
		U  string
		H  string
		A  string
		M  map[string]*dbus.Variant
		HD map[string]string
	}{url, hash, string(algo), t, headers}
	var path dbus.ObjectPath
	reply, err := man.proxy.Call(DOWNLOAD_MANAGER_INTERFACE, "createDownload", s)
	if err != nil || reply.Type == dbus.TypeError {
		return nil, err
	}
	if err = reply.Args(&path); err != nil {
		return nil, err
	}
	down, err = newFileDownload(man.conn, path)
	return down, err
}

// CreateMmsDownload creates an mms download that will be performed right away. An
// mms download only uses mobile that and an apn proxy to download a multime media
// message.
func (man *DownloadManager) CreateMmsDownload(url string, hostname string, port int32) (down Download, err error) {
	var path dbus.ObjectPath
	reply, err := man.proxy.Call(DOWNLOAD_MANAGER_INTERFACE, "createMmsDownload", url, hostname, port)
	if err != nil || reply.Type == dbus.TypeError {
		return nil, err
	}
	if err = reply.Args(&path); err != nil {
		return nil, err
	}
	down, err = newFileDownload(man.conn, path)
	return down, err
}
