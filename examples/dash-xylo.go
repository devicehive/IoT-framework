/*
  DeviceHive IoT-Framework business logic

  Copyright (C) 2016 DataArt

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at
 
      http://www.apache.org/licenses/LICENSE-2.0
 
  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/ 

package main

import (
	"log"
	"strconv"
	"time"
)

type melody struct {
	w *dbusWrapper
}

func NewMelody(w *dbusWrapper) *melody {
	m := &melody{w: w}

	return m
}

func (m *melody) Play(mac string, s string, bpm int) (err error) {
	notes := map[rune]string{'c': "060000",
		'd': "0603e8",
		'e': "0607d0",
		'f': "060bb8",
		'g': "060fa0",
		'a': "061388",
		'b': "061770",
		'C': "061b58"}

	noteLen := 1
	noteLenStr := ""
	prev := '\x00'
	m.w.GattWriteNoResp(mac, "af230002-879d-6186-1f49-deca0e85d9c1", "c804")
	for _, c := range s {
		log.Printf("c = %d", c)
		note, ok := notes[c]
		switch {
		case '0' <= c && c <= '9':
			if '0' <= prev && prev <= '9' {
				noteLenStr = noteLenStr + string(c)
			} else {
				noteLenStr = string(c)
			}
			prev = c
			continue
		case (ok || (c == '*')):
			if '0' <= prev && prev <= '9' {
				l, _ := strconv.ParseInt(noteLenStr, 0, 32)
				noteLen = int(l)
				log.Printf("Note length: %d", noteLen)
			}
			if ok {
				log.Printf("Playing %s", note)
				m.w.GattWriteNoResp(mac, "af230002-879d-6186-1f49-deca0e85d9c1", note)
				m.w.GattWriteNoResp(mac, "af230002-879d-6186-1f49-deca0e85d9c1", "10")
			}
		}
		prev = c
		wait := ((bpm * 1000 / 60) * 4 / noteLen)
		log.Printf("Wait: %d", wait)
		time.Sleep(time.Duration(wait) * time.Millisecond)
	}

	return nil
}

func main() {
	m := new(melody)
	m.Play("xxx", "1c2dd4eeee8ffffffff16gggggggggggggggg", 60)
}
