package main

import (
	"fmt"
	vn "omnidocvn_server/vn"
	"strconv"
)

type OtpVerifyServiceInput struct {
	RequestId string `json:"RequestId"`
	TrxId     string `json:"TrxId"`
	Otp       string `json:"Otp"`
	Signature string `json:"Signature"`
}

func (s *OtpVerifyServiceInput) toString() string {
	return "OtpVerifyServiceInput"
}

type OtpVerifyServiceOutput struct {
	RequestId       string `json:"Requestid"`
	TrxId           string `json:"Trxid"`
	ResponseCode    string `json:"ResponseCode"`
	ResponseMessage string `json:"ResponnseMessage"`
	Signature       string `json:"Signature"`
}

func (s *OtpVerifyServiceOutput) toString() string {
	return "OtpVerifyServiceOutput"
}

type OtpVerifyService struct {
}

type OtpVerifyServiceTest struct {
}

func (s *OtpVerifyServiceTest) execute(input DataObject) DataObject {
	otpVerifyInput := input.(*OtpVerifyServiceInput)

	trxid, _ := strconv.Atoi(otpVerifyInput.TrxId)

	output := OtpVerifyServiceOutput{}
	mapMutex.Lock()
	sessionId := sessionMap[trxid]
	mapMutex.Unlock()
	if sessionId < 1 {
		output.ResponseCode = fmt.Sprintf("%08x", 0)
		output.ResponseMessage = "invalid trasaction"
		return &output
	}

	output.ResponseCode = fmt.Sprintf("%08x", 0)
	output.ResponseMessage = "success"

	mapMutex.Lock()
	delete(sessionMap, trxid)
	mapMutex.Unlock()

	return &output
}

func (s *OtpVerifyService) execute(input DataObject) DataObject {
	otpVerifyInput := input.(*OtpVerifyServiceInput)

	trxid, _ := strconv.Atoi(otpVerifyInput.TrxId)

	output := OtpVerifyServiceOutput{}
	mapMutex.Lock()
	sessionId := sessionMap[trxid]
	mapMutex.Unlock()
	if sessionId < 1 {
		output.ResponseCode = fmt.Sprintf("%08x", 0)
		output.ResponseMessage = "invalid trasaction"
		return &output
	}
	result := vn.Transfer(sessionMap[trxid], otpVerifyInput.Otp)

	if result.Line > 0 {
		output.ResponseCode = fmt.Sprintf("%08x", int(result.Errcode))
		err := ERROR_TABLE[int(result.Errcode)]
		if err == nil {
			output.ResponseMessage = "Unknown Error"
		} else {
			output.ResponseMessage = err.Error()
		}
	} else {
		output.ResponseCode = fmt.Sprintf("%08x", 0)
		output.ResponseMessage = "success"
	}

	mapMutex.Lock()
	delete(sessionMap, trxid)
	mapMutex.Unlock()

	return &output
}
