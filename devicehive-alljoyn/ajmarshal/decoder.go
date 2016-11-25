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

package ajmarshal

import (
	"encoding/binary"
	"github.com/godbus/dbus"
	"io"
	"reflect"
	"strings"
)

type decoder struct {
	in    io.Reader
	order binary.ByteOrder
	pos   int
}

// Try to read a single type from this string. If it was successfull, err is nil
// and rem is the remaining unparsed part. Otherwise, err is a non-nil
// SignatureError and rem is "". depth is the current recursion depth which may
// not be greater than 64 and should be given as 0 on the first call.
func validSingle(s string, depth int) (err error, rem string) {
	if s == "" {
		return dbus.SignatureError{Sig: s, Reason: "empty signature"}, ""
	}
	if depth > 64 {
		return dbus.SignatureError{Sig: s, Reason: "container nesting too deep"}, ""
	}
	switch s[0] {
	case 'y', 'b', 'n', 'q', 'i', 'u', 'x', 't', 'd', 's', 'g', 'o', 'v', 'h':
		return nil, s[1:]
	case 'a':
		if len(s) > 1 && s[1] == '{' {
			i := findMatching(s[1:], '{', '}')
			if i == -1 {
				return dbus.SignatureError{Sig: s, Reason: "unmatched '{'"}, ""
			}
			i++
			rem = s[i+1:]
			s = s[2:i]
			if err, _ = validSingle(s[:1], depth+1); err != nil {
				return err, ""
			}
			err, nr := validSingle(s[1:], depth+1)
			if err != nil {
				return err, ""
			}
			if nr != "" {
				return dbus.SignatureError{Sig: s, Reason: "too many types in dict"}, ""
			}
			return nil, rem
		}
		return validSingle(s[1:], depth+1)
	case '(':
		i := findMatching(s, '(', ')')
		if i == -1 {
			return dbus.SignatureError{Sig: s, Reason: "unmatched ')'"}, ""
		}
		rem = s[i+1:]
		s = s[1:i]
		for err == nil && s != "" {
			err, s = validSingle(s, depth+1)
		}
		if err != nil {
			rem = ""
		}
		return
	}
	return dbus.SignatureError{Sig: s, Reason: "invalid type character"}, ""
}

func findMatching(s string, left, right rune) int {
	n := 0
	for i, v := range s {
		if v == left {
			n++
		} else if v == right {
			n--
		}
		if n == 0 {
			return i
		}
	}
	return -1
}

var sigToType = map[byte]reflect.Type{
	'y': byteType,
	'b': boolType,
	'n': int16Type,
	'q': uint16Type,
	'i': int32Type,
	'u': uint32Type,
	'x': int64Type,
	't': uint64Type,
	'd': float64Type,
	's': stringType,
	'g': signatureType,
	'o': objectPathType,
	'v': variantType,
	'h': unixFDIndexType,
}

// typeFor returns the type of the given signature. It ignores any left over
// characters and panics if s doesn't start with a valid type signature.
func typeFor(s string) (t reflect.Type) {
	err, _ := validSingle(s, 0)
	if err != nil {
		panic(err)
	}

	if t, ok := sigToType[s[0]]; ok {
		return t
	}
	switch s[0] {
	case 'a':
		if s[1] == '{' {
			i := strings.LastIndex(s, "}")
			t = reflect.MapOf(sigToType[s[2]], typeFor(s[3:i]))
		} else {
			t = reflect.SliceOf(typeFor(s[1:]))
		}
	case '(':
		t = interfacesType
	}
	return
}

// newDecoder returns a new decoder that reads values from in. The input is
// expected to be in the given byte order.
func NewDecoder(in io.Reader, order binary.ByteOrder) *decoder {
	dec := new(decoder)
	dec.in = in
	dec.order = order
	return dec
}

// align aligns the input to the given boundary and panics on error.
func (dec *decoder) align(n int) {
	if dec.pos%n != 0 {
		newpos := (dec.pos + n - 1) & ^(n - 1)
		empty := make([]byte, newpos-dec.pos)
		if _, err := io.ReadFull(dec.in, empty); err != nil {
			panic(err)
		}
		dec.pos = newpos
	}
}

// Calls binary.Read(dec.in, dec.order, v) and panics on read errors.
func (dec *decoder) binread(v interface{}) {
	if err := binary.Read(dec.in, dec.order, v); err != nil {
		panic(err)
	}
}

func (dec *decoder) Decode(sig dbus.Signature) (vs []interface{}, err error) {
	defer func() {
		var ok bool
		v := recover()
		if err, ok = v.(error); ok {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				err = FormatError("unexpected EOF")
			}
		}
	}()
	vs = make([]interface{}, 0)
	s := sig.String()
	for s != "" {
		err, rem := validSingle(s, 0)
		if err != nil {
			return nil, err
		}
		v := dec.decode(s[:len(s)-len(rem)], 0)
		vs = append(vs, v)
		s = rem
	}
	return vs, nil
}

func (dec *decoder) decode(s string, depth int) interface{} {
	dec.align(alignment(typeFor(s)))
	switch s[0] {
	case 'y':
		var b [1]byte
		if _, err := dec.in.Read(b[:]); err != nil {
			panic(err)
		}
		dec.pos++
		return b[0]
	case 'b':
		i := dec.decode("u", depth).(uint32)
		switch {
		case i == 0:
			return false
		case i == 1:
			return true
		default:
			panic(FormatError("invalid value for boolean"))
		}
	case 'n':
		var i int16
		dec.binread(&i)
		dec.pos += 2
		return i
	case 'i':
		var i int32
		dec.binread(&i)
		dec.pos += 4
		return i
	case 'x':
		var i int64
		dec.binread(&i)
		dec.pos += 8
		return i
	case 'q':
		var i uint16
		dec.binread(&i)
		dec.pos += 2
		return i
	case 'u':
		var i uint32
		dec.binread(&i)
		dec.pos += 4
		return i
	case 't':
		var i uint64
		dec.binread(&i)
		dec.pos += 8
		return i
	case 'd':
		var f float64
		dec.binread(&f)
		dec.pos += 8
		return f
	case 's':
		length := dec.decode("u", depth).(uint32)
		b := make([]byte, int(length)+1)
		if _, err := io.ReadFull(dec.in, b); err != nil {
			panic(err)
		}
		dec.pos += int(length) + 1
		return string(b[:len(b)-1])
	case 'o':
		return dbus.ObjectPath(dec.decode("s", depth).(string))
	case 'g':
		length := dec.decode("y", depth).(byte)
		b := make([]byte, int(length)+1)
		if _, err := io.ReadFull(dec.in, b); err != nil {
			panic(err)
		}
		dec.pos += int(length) + 1
		sig, err := dbus.ParseSignature(string(b[:len(b)-1]))
		if err != nil {
			panic(err)
		}
		return sig
	case 'v':
		if depth >= 64 {
			panic(FormatError("input exceeds container depth limit"))
		}
		var variant dbus.Variant
		sig := dec.decode("g", depth).(dbus.Signature)
		if len(sig.String()) == 0 {
			panic(FormatError("variant signature is empty"))
		}
		err, rem := validSingle(sig.String(), 0)
		if err != nil {
			panic(err)
		}
		if rem != "" {
			panic(FormatError("variant signature has multiple types"))
		}
		//variant.sig = sig
		//variant.value = dec.decode(sig.String(), depth+1)
		variant = dbus.MakeVariant(dec.decode(sig.String(), depth+1))
		return variant
	case 'h':
		return dbus.UnixFDIndex(dec.decode("u", depth).(uint32))
	case 'a':
		if len(s) > 1 && s[1] == '{' {
			ksig := s[2:3]
			vsig := s[3 : len(s)-1]
			v := reflect.MakeMap(reflect.MapOf(typeFor(ksig), typeFor(vsig)))
			if depth >= 63 {
				panic(FormatError("input exceeds container depth limit"))
			}
			length := dec.decode("u", depth).(uint32)
			// Even for empty maps, the correct padding must be included
			dec.align(8)
			spos := dec.pos
			for dec.pos < spos+int(length) {
				dec.align(8)
				if !isKeyType(v.Type().Key()) {
					panic(dbus.InvalidTypeError{v.Type()})
				}
				kv := dec.decode(ksig, depth+2)
				vv := dec.decode(vsig, depth+2)
				v.SetMapIndex(reflect.ValueOf(kv), reflect.ValueOf(vv))
			}
			return v.Interface()
		}
		if depth >= 64 {
			panic(FormatError("input exceeds container depth limit"))
		}
		length := dec.decode("u", depth).(uint32)
		v := reflect.MakeSlice(reflect.SliceOf(typeFor(s[1:])), 0, int(length))
		// Even for empty arrays, the correct padding must be included
		dec.align(alignment(typeFor(s[1:])))
		spos := dec.pos
		for dec.pos < spos+int(length) {
			ev := dec.decode(s[1:], depth+1)
			v = reflect.Append(v, reflect.ValueOf(ev))
		}
		return v.Interface()
	case '(':
		if depth >= 64 {
			panic(FormatError("input exceeds container depth limit"))
		}
		dec.align(8)
		v := make([]interface{}, 0)
		s = s[1 : len(s)-1]
		for s != "" {
			err, rem := validSingle(s, 0)
			if err != nil {
				panic(err)
			}
			ev := dec.decode(s[:len(s)-len(rem)], depth+1)
			v = append(v, ev)
			s = rem
		}
		return v
	default:
		panic(dbus.SignatureError{Sig: s})
	}
}

// A FormatError is an error in the wire format.
type FormatError string

func (e FormatError) Error() string {
	return "dbus: wire format error: " + string(e)
}
