/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Sergio Schvezov: sergio.schvezov@cannical.com
 *
 * This file is part of telepathy.
 *
 * mms is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.
 *
 * mms is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/ubports/nuntium/mms"
	"launchpad.net/go-xdg/v0"
)

const SUBPATH = "nuntium/store"

// Creates an .db file in storage with message state stored.
// Returns an empty state and not nil error if message not stored successfully.
func Create(modemId string, mNotificationInd *mms.MNotificationInd) (MMSState, error) {
	state := MMSState{
		Id:               mNotificationInd.TransactionId,
		State:            NOTIFICATION,
		ContentLocation:  mNotificationInd.ContentLocation,
		ModemId:          modemId,
		MNotificationInd: mNotificationInd,
	}
	storePath, err := xdg.Data.Ensure(path.Join(SUBPATH, mNotificationInd.UUID+".db"))
	if err != nil {
		return MMSState{}, err
	}
	if err := writeState(state, storePath); err != nil {
		return MMSState{}, err
	}
	return state, nil
}

// Removes message with UUID from storage.
// Returns a not nil error if any/more of the stored files are failed to remove.
// The returned error (if not nil) is always an Multierror type.
func Destroy(uuid string) (err error) {
	errs := Multierror{}

	if path, err := xdg.Data.Find(path.Join(SUBPATH, uuid+".db")); err == nil {
		if err := os.Remove(path); err != nil {
			errs = append(errs, ErrorRemovingFile{path, err})
		}
	} else {
		errs = append(errs, err)
	}

	if path, err := GetMMS(uuid); err == nil {
		if err := os.Remove(path); err != nil {
			errs = append(errs, ErrorRemovingFile{path, err})
		}
	}

	if path, err := xdg.Cache.Find(path.Join(SUBPATH, uuid+".m-notifyresp.ind")); err == nil {
		if err := os.Remove(path); err != nil {
			errs = append(errs, ErrorRemovingFile{path, err})
		}
	}

	if path, err := xdg.Cache.Find(path.Join(SUBPATH, uuid+".m-send.req")); err == nil {
		if err := os.Remove(path); err != nil {
			errs = append(errs, ErrorRemovingFile{path, err})
		}
	}

	return errs.Result()
}

// Creates an empty .m-notifyresp.ind file in storage for message with provided uuid.
// Returns a nil file descriptor and a non nil error if no message stored uuid or file creation failed.
// On success returns an open file descriptor and nil error.
func CreateResponseFile(uuid string) (*os.File, error) {
	_, err := GetMMSState(uuid)
	if err != nil {
		return nil, fmt.Errorf("error retrieving message state: %w", err)
	}

	filePath, err := xdg.Cache.Ensure(path.Join(SUBPATH, uuid+".m-notifyresp.ind"))
	if err != nil {
		return nil, err
	}
	return os.Create(filePath)
}

// Updates MNotificationInd field in stored MMSState.
// Returns the stored message state and a nil error on success.
// If message not in storage or other fail it returns empty or previous state and a non nil error.
func UpdateMNotificationInd(mNotificationInd *mms.MNotificationInd) (MMSState, error) {
	oldState, err := GetMMSState(mNotificationInd.UUID)
	if err != nil {
		return oldState, fmt.Errorf("error retrieving message state: %w", err)
	}

	newState := oldState
	newState.MNotificationInd = mNotificationInd

	storePath, err := xdg.Data.Find(path.Join(SUBPATH, mNotificationInd.UUID+".db"))
	if err != nil {
		return oldState, err
	}
	if err := writeState(newState, storePath); err != nil {
		return oldState, err
	}

	return newState, nil
}

// Copies the provided file to storage into an .mms file and updates the stored message (identified by uuid) state to DOWNLOADED.
// Returns the stored message state and a nil error on success.
// If message not in storage or other error occurs, it returns empty or previous state and a non nil error.
// Note: Can return a forced debug error if MNotificationInd has the right ContentLocation parameters.
func UpdateDownloaded(uuid, filePath string) (MMSState, error) {
	oldState, err := GetMMSState(uuid)
	if err != nil {
		return oldState, fmt.Errorf("error retrieving message state: %w", err)
	}

	// Debug error forcing if wanted.
	if err := oldState.MNotificationInd.PopDebugError(mms.DebugErrorDownloadStorage); err != nil {
		log.Printf("Forcing debug error: %#v", err)
		UpdateMNotificationInd(oldState.MNotificationInd)
		return oldState, err
	}

	// Move downloaded file (filePath) to xdg data storage.
	mmsPath, err := xdg.Data.Ensure(path.Join(SUBPATH, uuid+".mms"))
	if err != nil {
		return oldState, err
	}
	if err := os.Rename(filePath, mmsPath); err != nil {
		if err := os.Remove(mmsPath); err != nil {
			log.Printf("Error removing file \"%s\": %s", mmsPath, err)
		}
		return oldState, err
	}

	newState := oldState
	newState.State = DOWNLOADED

	storePath, err := xdg.Data.Find(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return oldState, err
	}
	if err := writeState(newState, storePath); err != nil {
		return oldState, err
	}

	return newState, nil
}

// Updates the stored message (identified by uuid) state to RECEIVED.
// Returns the stored message state and a nil error on success.
// If message not in storage or other error occurs, it returns empty or previous state and a non nil error.
// Note: Can return a forced debug error if MNotificationInd has the right ContentLocation parameters.
func UpdateReceived(uuid string) (MMSState, error) {
	oldState, err := GetMMSState(uuid)
	if err != nil {
		return oldState, fmt.Errorf("error retrieving message state: %w", err)
	}

	// Debug error forcing if wanted.
	if err := oldState.MNotificationInd.PopDebugError(mms.DebugErrorReceiveStorage); err != nil {
		log.Printf("Forcing debug error: %#v", err)
		UpdateMNotificationInd(oldState.MNotificationInd)
		return oldState, err
	}

	newState := oldState
	newState.State = RECEIVED

	storePath, err := xdg.Data.Find(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return oldState, err
	}
	if err := writeState(newState, storePath); err != nil {
		return oldState, err
	}

	return newState, nil
}

// Updates the stored message (identified by uuid) state to RESPONDED.
// Returns the stored message state and a nil error on success.
// If message not in storage or other error occurs, it returns empty or previous state and a non nil error.
// Note: Can return a forced debug error if MNotificationInd has the right ContentLocation parameters.
func UpdateResponded(uuid string) (MMSState, error) {
	oldState, err := GetMMSState(uuid)
	if err != nil {
		return oldState, fmt.Errorf("error retrieving message state: %w", err)
	}

	// Debug error forcing if wanted.
	if err := oldState.MNotificationInd.PopDebugError(mms.DebugErrorRespondStorage); err != nil {
		log.Printf("Forcing debug error: %#v", err)
		UpdateMNotificationInd(oldState.MNotificationInd)
		return oldState, err
	}

	newState := oldState
	newState.State = RESPONDED

	storePath, err := xdg.Data.Find(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return oldState, err
	}
	if err := writeState(newState, storePath); err != nil {
		return oldState, err
	}

	return newState, nil
}

// Updates the stored message (identified by uuid) TelepathyErrorNotified to true.
// Returns the stored message state and a nil error on success.
// If message not in storage or other error occurs, it returns empty or previous state and a non nil error.
func SetTelepathyErrorNotified(uuid string) (MMSState, error) {
	oldState, err := GetMMSState(uuid)
	if err != nil {
		return oldState, fmt.Errorf("error retrieving message state: %w", err)
	}

	newState := oldState
	newState.TelepathyErrorNotified = true

	storePath, err := xdg.Data.Find(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return oldState, err
	}
	if err := writeState(newState, storePath); err != nil {
		return oldState, err
	}

	return newState, nil
}

// Saves an message with DRAFT state to storage and creates an empty .m-send.req file in storage for message with provided uuid.
// Returns a nil file descriptor and a non nil error if message store error or send file creation failed.
// On success returns an open file descriptor to the send file and nil error.
// Note: If there is an message stored under uuid, the message is rewritten.
func CreateSendFile(uuid string) (*os.File, error) {
	state := MMSState{
		State: DRAFT,
	}
	storePath, err := xdg.Data.Ensure(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return nil, err
	}
	if err := writeState(state, storePath); err != nil {
		os.Remove(storePath)
		return nil, err
	}
	filePath, err := xdg.Cache.Ensure(path.Join(SUBPATH, uuid+".m-send.req"))
	if err != nil {
		return nil, err
	}
	return os.Create(filePath)
}

// Returns .mms file path to message identified by uuid.
// If file doesn't exists, a non nil error is returned.
func GetMMS(uuid string) (string, error) {
	return xdg.Data.Find(path.Join(SUBPATH, uuid+".mms"))
}

// Gets message state from storage stored under uuid.
// Returns empty state and a non nil error if message not stored or load failed.
func GetMMSState(uuid string) (MMSState, error) {
	storePath, err := xdg.Data.Find(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return MMSState{}, err
	}

	f, err := os.Open(storePath)
	if err != nil {
		return MMSState{}, err
	}
	defer f.Close()

	mmsState := MMSState{}
	jsonReader := json.NewDecoder(f)
	if err := jsonReader.Decode(&mmsState); err != nil {
		return MMSState{}, err
	}

	return mmsState, nil
}

// Returns stored MNotificationInd for message identified by uuid.
// If message not in storage or message state is not NOTIFICATION, nil is returned.
func GetMNotificationInd(uuid string) *mms.MNotificationInd {
	mmsState, err := GetMMSState(uuid)
	if err != nil {
		log.Print("MMS state retrieving error:", err)
		return nil
	}

	if mmsState.State != NOTIFICATION {
		log.Print("MMS was already downloaded")
		return nil
	}

	return mmsState.MNotificationInd
}

func writeState(state MMSState, storePath string) error {
	file, err := os.Create(storePath)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		if err != nil {
			os.Remove(storePath)
		}
	}()
	w := bufio.NewWriter(file)
	defer w.Flush()
	jsonWriter := json.NewEncoder(w)
	if err := jsonWriter.Encode(state); err != nil {
		return err
	}
	return nil
}

// Returns list of UUID strings stored in storage, sorted by creation date ascending.
// Note: If creation date is not supported by filesystem, UUIDs are sorted by modificatin date.
func GetStoredUUIDs() []string {
	// Search for all *.db files in xdg data directory in SUBPATH subfolder and extract UUID from filenames.

	storeDir, err := xdg.Data.Find(SUBPATH)
	if err != nil {
		log.Printf("Storage directory %s not found in xdg data directories", SUBPATH)
		return nil
	}

	uuidsWithTime := make([]struct {
		uuid  string
		ctime time.Time
	}, 0) // For sorting.
	err = filepath.Walk(storeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match("*.db", filepath.Base(path)); err != nil {
			return err
		} else if matched {
			ctime := info.ModTime()
			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
				ctime = time.Unix(stat.Ctim.Unix())
			}
			uuidsWithTime = append(uuidsWithTime, struct {
				uuid  string
				ctime time.Time
			}{
				strings.TrimSuffix(filepath.Base(path), ".db"),
				ctime,
			})
		}
		return nil
	})
	if err != nil {
		return nil
	}

	// Sort uuids by cdates ascending.
	sort.SliceStable(uuidsWithTime, func(i, j int) bool {
		return uuidsWithTime[i].ctime.Before(uuidsWithTime[j].ctime)
	})
	uuids := make([]string, len(uuidsWithTime))
	for i, uuidWithTime := range uuidsWithTime {
		uuids[i] = uuidWithTime.uuid
	}
	return uuids
}
