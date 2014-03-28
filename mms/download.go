/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Sergio Schvezov: sergio.schvezov@cannical.com
 *
 * This file is part of mms.
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

package mms

import (
	"fmt"
	"launchpad.net/ubuntu-download-manager/bindings/golang"
	"log"
	"time"
)

func (pdu *MNotificationInd) DownloadContent(proxyHostname string, proxyPort int32) (string, error) {
	downloadManager, err := udm.NewDownloadManager()
	if err != nil {
		return "", err
	}
	download, err := downloadManager.CreateMmsDownload(pdu.ContentLocation, proxyHostname, proxyPort)
	if err != nil {
		return "", err
	}
	f := download.Finished()
	p := download.DownloadProgress()
	e := download.Error()
	log.Print("Starting download")
	download.Start()
	for {
		select {
		case progress := <-p:
			log.Print("Progress:", progress.Total, progress.Received)
		case downloadFilePath := <-f:
			log.Print("File downloaded to ", downloadFilePath)
			return downloadFilePath, nil
		case <-time.After(3 * time.Minute):
			return "", fmt.Errorf("Download timeout exceeded while fetching %s", pdu.ContentLocation)
		case err := <-e:
			return "", err
		}
	}
}
