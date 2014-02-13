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
	"launchpad.net/ubuntu-download-manager/bindings/golang"
	"log"
)

func (pdu *MNotificationInd) Download(proxyHostname string, proxyPort int32, username, password string) (string, error) {
	downloadManager, err := udm.NewDownloadManager()
	if err != nil {
		return "", err
	}
	download, err := downloadManager.CreateMmsDownload(pdu.ContentLocation, proxyHostname, proxyPort, username, password)
	if err != nil {
		return "", err
	}
	f := download.Finished()
	p := download.DownloadProgress()
	e := download.Error()
	go func() {
		for progress := range p {
			log.Print(progress.Total, progress.Received)
		}
	}()
	download.Start()
	<- e
	downloadFilePath := <-f
	log.Print("File downloaded to", downloadFilePath)
	return downloadFilePath, nil
}
