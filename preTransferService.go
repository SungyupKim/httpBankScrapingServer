package main

import (
	"fmt"
	vn "omnidocvn_server/vn"
	"strconv"
)

type PreTransferInput struct {
	TrxId      string `json:"TrxId"`
	RequestId  string `json:"RequestId"`
	SrcBankId  string `json:"SrcbankId"`
	SrcAccount string `json:"SrcAccount"`
	Amount     int    `json:"Amount"`
	BenBankId  string `json:"BenBankid"`
	BenAccount string `json:"BenAccount"`
	Signature  string `json:"Signature"`
}

func (s *PreTransferInput) toString() string {
	return "PreTransferInput"
}

type PreTransferOutput struct {
	RequestId       string `json:"RequestId"`
	TrxId           string `json:"TrxId"`
	ResponseCode    string `json:"ResponseCode"`
	ResponseMessage string `json:"ResponseMessage"`
	Signature       string `json:"Signature"`
}

func (s *PreTransferOutput) toString() string {
	return "PreTransferOutput"
}

type PreTransferService struct {
}

type PreTransferServiceTest struct {
}

func (s *PreTransferServiceTest) execute(input DataObject) DataObject {
	log.Printf("pretransfer service test")
	preTransferInput := input.(*PreTransferInput)
	log.Printf("%+v\n", preTransferInput)

	preTransferOutput := PreTransferOutput{}
	preTransferOutput.RequestId = preTransferInput.RequestId
	numTrxId, _ := strconv.Atoi(preTransferInput.TrxId)
	mapMutex.Lock()
	fmt.Printf("%+v\n", sessionMap)
	sessionid := sessionMap[numTrxId]
	mapMutex.Unlock()

	if sessionid > 0 {
		log.Printf("has session")
		preTransferOutput.TrxId = preTransferInput.TrxId
		preTransferOutput.ResponseCode = fmt.Sprintf("%08x", 0)
		preTransferOutput.ResponseMessage = "success"
	} else {
		preTransferOutput.ResponseCode = "0x00000001"
		preTransferOutput.ResponseMessage = "you should login first"
	}

	return &preTransferOutput
}

func (s *PreTransferService) execute(input DataObject) DataObject {
	log.Printf("pretransfer service")
	preTransferInput := input.(*PreTransferInput)
	log.Printf("%+v\n", preTransferInput)

	preTransferOutput := PreTransferOutput{}
	preTransferOutput.RequestId = preTransferInput.RequestId
	numTrxId, _ := strconv.Atoi(preTransferInput.TrxId)
	mapMutex.Lock()
	sessionid := sessionMap[numTrxId]
	mapMutex.Unlock()

	if sessionid > 0 {
		log.Printf("has session")
		preTransferOutput.TrxId = preTransferInput.TrxId
		res := vn.SelectAccount(sessionid, preTransferInput.BenAccount)

		if res < 0 {
			preTransferOutput.ResponseMessage = "account select failed"
		}

		result := vn.PreTransfer(sessionid, preTransferInput.BenAccount, "", preTransferInput.BenBankId, preTransferInput.Amount)

		if result.Line > 0 {
			preTransferOutput.ResponseCode = fmt.Sprintf("%08x", int(result.Errcode))
			err := ERROR_TABLE[int(result.Errcode)]
			if err == nil {
				preTransferOutput.ResponseMessage = "Unknown Error"
			} else {
				preTransferOutput.ResponseMessage = err.Error()
			}
		} else {
			preTransferOutput.ResponseCode = fmt.Sprintf("%08x", 0)
			preTransferOutput.ResponseMessage = "success"
		}

	} else {
		preTransferOutput.ResponseCode = "0x00000001"
		preTransferOutput.ResponseMessage = "you should login first"
	}
	return &preTransferOutput
}
