// Copyright ©2013 The bíogo.entrez Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package entrez

import (
	"code.google.com/p/biogo.entrez/stack"
	"code.google.com/p/biogo.entrez/summary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"reflect"
)

// <!--
// This is the Current DTD for Entrez eSummary version 2
// $Id: eSummary_041029.dtd 49514 2004-10-29 15:52:04Z parantha $
// -->
// <!-- ================================================================= -->
//
// <!ELEMENT Id                (#PCDATA)>          <!-- \d+ -->
//
// <!ELEMENT Item              (#PCDATA|Item)*>   <!-- .+ -->
//
// <!ATTLIST Item
//     Name CDATA #REQUIRED
//     Type (Integer|Date|String|Structure|List|Flags|Qualifier|Enumerator|Unknown) #REQUIRED
// >
//
// <!ELEMENT ERROR             (#PCDATA)>  <!-- .+ -->
//
// <!ELEMENT DocSum            (Id, Item+)>
//
// <!ELEMENT eSummaryResult    (DocSum|ERROR)+>

// A Summary holds the deserialised results of an ESummary request.
type Summary struct {
	Database string
	Docs     []summary.Doc
	Err      error
}

// Unmarshal fills the fields of a Summary from an XML stream read from r.
func (s *Summary) Unmarshal(r io.Reader) error {
	dec := xml.NewDecoder(r)
	var st stack.Stack
	for {
		t, err := dec.Token()
		if err != nil {
			if err != io.EOF {
				return err
			}
			if !st.Empty() {
				return io.ErrUnexpectedEOF
			}
			break
		}
		switch t := t.(type) {
		case xml.ProcInst:
		case xml.Directive:
		case xml.StartElement:
			st = st.Push(t.Name.Local)
			if t.Name.Local == "DocSum" {
				var d summary.Doc
				err := d.Unmarshal(dec, st[len(st)-1:])
				if !(reflect.DeepEqual(d, summary.Doc{})) {
					s.Docs = append(s.Docs, d)
				}
				if err != nil {
					return err
				}
				st = st.Drop()
			}
		case xml.CharData:
			if st.Empty() {
				continue
			}
			switch name := st.Peek(0); name {
			case "ERROR":
				s.Err = errors.New(string(t))
			case "eSummaryResult":
			default:
				return fmt.Errorf("entrez: unknown name: %q", name)
			}
		case xml.EndElement:
			st, err = st.Pair(t.Name.Local)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
