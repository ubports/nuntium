/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Sergio Schvezov: sergio.schvezov@cannical.com
 *
 * This file is part of nuntium.
 *
 * nuntium is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.
 *
 * nuntium is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Mainloop struct {
	sigchan  chan os.Signal
	termchan chan int
	Bindings map[os.Signal]func()
}

/*
Start the mainloop.

This method will block its current thread. The best spot for calling this
method is right near the bottom of your application's main() function.
*/
func (m *Mainloop) Start() {
	sigs := make([]os.Signal, len(m.Bindings))
	for s, _ := range m.Bindings {
		sigs = append(sigs, s)
	}
	signal.Notify(m.sigchan, sigs...)
L:
	for {
		select {
		case sig := <-m.sigchan:
			log.Print("Received ", sig)
			m.Bindings[sig]()
		case _ = <-m.termchan:
			break L
		}
	}
	return
}

/*
Stops the mainloop.
*/
func (m *Mainloop) Stop() {
	go func() { m.termchan <- 1 }()
	return
}

func HupHandler() {
	syscall.Exit(1)
}

func IntHandler() {
	syscall.Exit(1)
}
