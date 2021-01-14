package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

var (
	requests      uint64
	sequenceMutex sync.Mutex
)

type PrePostProcessor struct {
}

func (s *PrePostProcessor) preProcess(input DataObject, r *http.Request) DataObject {
	sequenceMutex.Lock()
	requests++
	sequenceMutex.Unlock()

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return nil
	}
	return input
}

func (s *PrePostProcessor) postProcess(output DataObject, w http.ResponseWriter) {
	json.NewEncoder(w).Encode(output)

	sequenceMutex.Lock()
	requests--
	sequenceMutex.Unlock()
}
