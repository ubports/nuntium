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
	"errors"
	"os"
	"path/filepath"
	"sync"

	"log"

	"launchpad.net/go-dbus/v1"
	"launchpad.net/go-xdg/v0"
)

var preferredContextPath string = filepath.Join(filepath.Base(os.Args[0]), "preferredContext")

var contextMutex sync.Mutex

type contextSettingMap map[string]dbus.ObjectPath

func SetPreferredContext(identity string, pcObjectPath dbus.ObjectPath) error {
	contextMutex.Lock()
	defer contextMutex.Unlock()

	pcFilePath, err := xdg.Cache.Ensure(preferredContextPath)
	if err != nil {
		return err
	}
	return writeContext(identity, pcObjectPath, pcFilePath)
}

func GetPreferredContext(identity string) (pcObjectPath dbus.ObjectPath, err error) {
	contextMutex.Lock()
	defer contextMutex.Unlock()

	pcFilePath, err := xdg.Cache.Find(preferredContextPath)
	if err != nil {
		return pcObjectPath, err
	}
	cs, err := readContext(pcFilePath)
	if err != nil {
		return pcObjectPath, err
	}
	if p, ok := cs[identity]; ok {
		return p, nil
	}

	return pcObjectPath, errors.New("path for identity not found")
}

func readContext(storePath string) (cs contextSettingMap, err error) {
	file, err := os.Open(storePath)
	if err != nil {
		cs = make(contextSettingMap)
		return cs, err
	}
	jsonReader := json.NewDecoder(file)
	if err = jsonReader.Decode(&cs); err != nil {
		cs = make(contextSettingMap)
	}
	return cs, err
}

func writeContext(identity string, pc dbus.ObjectPath, storePath string) error {
	log.Println(storePath)
	cs, readErr := readContext(storePath)
	if readErr != nil {
		log.Println("Cannot read previous context state")
	}

	file, err := os.Create(storePath)
	if err != nil {
		log.Println(err)
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
	cs[identity] = pc
	jsonWriter := json.NewEncoder(w)
	if err := jsonWriter.Encode(cs); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
