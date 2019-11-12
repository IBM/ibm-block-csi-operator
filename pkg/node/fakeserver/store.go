/**
 * Copyright 2019 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fakeserver

import (
	"sync"
)

type stringInterface interface {
	String() string
}

var store = &sync.Map{}

type responses struct {
	data  []stringInterface
	index int
}

func (m *responses) Store(res stringInterface) {
	if m.data == nil {
		m.data = []stringInterface{}
	}
	m.data = append(m.data, res)
}

func (m *responses) Load() stringInterface {
	size := len(m.data)
	if size == 0 {
		return nil
	}
	m.index = m.index % size
	res := m.data[m.index]
	m.index = m.index + 1
	return res
}

func (m *responses) Clear() {
	m.data = []stringInterface{}
}

type mapping struct {
	data *sync.Map
}

func (m *mapping) Store(req, res stringInterface) {
	if m.data == nil {
		m.data = &sync.Map{}
	}
	reqStr := req.String()
	r, ok := m.data.Load(reqStr)
	if ok {
		r.(*responses).Store(res)
	} else {
		newResponses := &responses{}
		newResponses.Store(res)
		m.data.Store(reqStr, newResponses)
	}
}

func (m *mapping) Load(req stringInterface) stringInterface {
	if m.data == nil {
		m.data = &sync.Map{}
		return nil
	}
	reqStr := req.String()
	r, ok := m.data.Load(reqStr)
	if ok {
		return r.(*responses).Load()
	} else {
		return nil
	}
}

func LoadResponse(stub string, req stringInterface) stringInterface {
	m, ok := store.Load(stub)
	if ok {
		return m.(*mapping).Load(req)
	} else {
		return nil
	}
}

func StoreResponse(stub string, req, res stringInterface) {
	m, ok := store.Load(stub)
	if ok {
		m.(*mapping).Store(req, res)
	} else {
		newMapping := &mapping{}
		newMapping.Store(req, res)
		store.Store(stub, newMapping)
	}
}

func ClearAll() {
	store = &sync.Map{}
}
