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
	"strings"

	"github.com/ubports/nuntium/mms"
	"launchpad.net/go-xdg/v0"
)

const SUBPATH = "nuntium/store"

//TODO:version - change back to Create(uuid, contentLocation string) to increase only minor version instead of major (add ModemId, MNotificationInd after every Create)?
func Create(modemId string, mNotificationInd *mms.MNotificationInd) error {
	state := MMSState{
		Id:               mNotificationInd.TransactionId,
		State:            NOTIFICATION,
		ContentLocation:  mNotificationInd.ContentLocation,
		ModemId:          modemId,
		MNotificationInd: mNotificationInd,
	}
	storePath, err := xdg.Data.Ensure(path.Join(SUBPATH, mNotificationInd.UUID+".db"))
	if err != nil {
		return err
	}
	return writeState(state, storePath)
}

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

func UpdateMNotificationInd(mNotificationInd *mms.MNotificationInd) error {
	log.Printf("jezek - UpdateMNotificationInd(%v)", mNotificationInd)
	state, err := GetMMSState(mNotificationInd.UUID)
	if err != nil {
		return fmt.Errorf("error retrieving message state: %w", err)
	}

	state.MNotificationInd = mNotificationInd

	storePath, err := xdg.Data.Find(path.Join(SUBPATH, mNotificationInd.UUID+".db"))
	if err != nil {
		return err
	}
	return writeState(state, storePath)
}

func UpdateDownloaded(uuid, filePath string) error {
	log.Printf("jezek - UpdateDownloaded(%s, %s)", uuid, filePath)
	state, err := GetMMSState(uuid)
	if err != nil {
		return fmt.Errorf("error retrieving message state: %w", err)
	}

	// Debug error forcing if wanted.
	if err := state.MNotificationInd.PopDebugError(mms.DebugErrorDownloadStorage); err != nil {
		log.Printf("Forcing debug error: %#v", err)
		UpdateMNotificationInd(state.MNotificationInd)
		return err
	}

	// Move downloaded file (filePath) to xdg data storage.
	mmsPath, err := xdg.Data.Ensure(path.Join(SUBPATH, uuid+".mms"))
	if err != nil {
		return err
	}
	if err := os.Rename(filePath, mmsPath); err != nil {
		//TODO delete file
		return err
	}

	state.State = DOWNLOADED

	storePath, err := xdg.Data.Find(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return err
	}
	return writeState(state, storePath)
}

func UpdateReceived(uuid string) error {
	state, err := GetMMSState(uuid)
	if err != nil {
		return fmt.Errorf("error retrieving message state: %w", err)
	}

	state.State = RECEIVED
	state.TelepathyNotified = true

	storePath, err := xdg.Data.Find(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return err
	}
	return writeState(state, storePath)
}

func UpdateResponded(uuid string) error {
	state, err := GetMMSState(uuid)
	if err != nil {
		return fmt.Errorf("error retrieving message state: %w", err)
	}

	state.State = RESPONDED

	storePath, err := xdg.Data.Find(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return err
	}
	return writeState(state, storePath)
}

func UpdateTelepathyNotified(uuid string) error {
	state, err := GetMMSState(uuid)
	if err != nil {
		return fmt.Errorf("error retrieving message state: %w", err)
	}

	state.TelepathyNotified = true

	storePath, err := xdg.Data.Find(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return err
	}
	return writeState(state, storePath)
}

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

func GetMMS(uuid string) (string, error) {
	return xdg.Data.Find(path.Join(SUBPATH, uuid+".mms"))
}

// Gets MMSState from strorage stored under uuid.
func GetMMSState(uuid string) (MMSState, error) {
	mmsState := MMSState{}
	storePath, err := xdg.Data.Find(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return mmsState, err
	}

	f, err := os.Open(storePath)
	if err != nil {
		return MMSState{}, err
	}
	defer f.Close()

	jsonReader := json.NewDecoder(f)
	if err := jsonReader.Decode(&mmsState); err != nil {
		return mmsState, err
	}

	return mmsState, nil
}
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

// Returns list of UUID strings stored in storage.
func GetStoredUUIDs() []string {
	// Search for all *.db files in xdg data directory in SUBPATH subfolder and extract UUID from filenames.

	storeDir, err := xdg.Data.Find(SUBPATH)
	if err != nil {
		log.Printf("Storage directory %s not found in xdg data directories", SUBPATH)
		return nil
	}

	uuids := make([]string, 0)
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
			uuids = append(uuids, strings.TrimSuffix(filepath.Base(path), ".db"))
		}
		return nil
	})
	if err != nil {
		return nil
	}

	return uuids
}
