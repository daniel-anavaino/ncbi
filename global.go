// Copyright ©2013 The bíogo.entrez Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package entrez

import (
	"code.google.com/p/biogo.entrez/global"
	"code.google.com/p/biogo.entrez/stack"
	"encoding/xml"
	"fmt"
	"io"
)

// <!--
//         This is the Current DTD for Entrez eGSearch
//         $Id: egquery.dtd 39250 2004-05-03 16:19:48Z yasmax $
// -->
// <!-- ================================================================= -->
//
// <!ELEMENT       DbName          (#PCDATA)>      <!-- .+ -->
// <!ELEMENT       MenuName        (#PCDATA)>      <!-- .+ -->
// <!ELEMENT       Count           (#PCDATA)>      <!-- \d+ -->
// <!ELEMENT       Status          (#PCDATA)>      <!-- .+ -->
// <!ELEMENT       Term            (#PCDATA)>      <!-- .+ -->
//
// <!ELEMENT       ResultItem      (
//                                      DbName,
//                                      MenuName,
//                                      Count,
//                                      Status
//                                 )>
// <!ELEMENT       eGQueryResult  (ResultItem+)>
//
// <!ELEMENT       Result         (Term, eGQueryResult)>

// A Global holds the deserialised results of an EGQuery request.
type Global struct {
	Query   string
	Results []global.Result
}

// Unmarshal fills the fields of a Global from an XML stream read from r.
func (g *Global) Unmarshal(r io.Reader) error {
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
			if t.Name.Local == "ResultItem" {
				var res global.Result
				err := res.Unmarshal(dec, st[len(st)-1:])
				g.Results = append(g.Results, res)
				if err != nil {
					return err
				}
				st = st.Drop()
				continue
			}
		case xml.CharData:
			if st.Empty() {
				continue
			}
			switch name := st.Peek(0); name {
			case "Term":
				g.Query = string(t)
			case "eGQueryResult", "Result":
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
