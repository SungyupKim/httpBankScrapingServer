package main

import "C"
import (
	"encoding/json"
	"fmt"
	sqlhandler "omnidocvn_server/sqlhandler"
	vn "omnidocvn_server/vn"
	"strconv"
	"sync"
)

var (
	mapMutex sync.Mutex
)

type SearchOutput struct {
	User struct {
		Name string `json:"name"`
	} `json:"user"`
	Acclist []struct {
		Accno   string `json:"accno"`
		Acctype string `json:"acctype"`
		List    []struct {
			Balance int    `json:"balance"`
			Time    string `json:"time"`
			Remark  string `json:"remark"`
			Credit  int    `json:"credit"`
			Debit   int    `json:"debit"`
		} `json:"list"`
		Lbalance string `json:"lbalance"`
		Balance  string `json:"balance"`
	} `json:"acclist"`
}

type LoginInput struct {
	RequestId string `json:"RequestId"`
	UserName  string `json:"UserName"`
	Password  string `json:"Password"`
	BankId    string `json:"BankId"`
	Signature string `json:"Signature"`
}

func (s *LoginInput) toString() string {
	return "LoginInput"
}

type LoginOutput struct {
	RequestId       string         `json:"RequestId"`
	TrxId           string         `json:"TrxId"`
	ResponseCode    string         `json:"ResponseCode"`
	ResponseMessage string         `json:"ResponseMessage"`
	BankId          string         `json:"BankId"`
	AccountInfo     []*AccountInfo `json:"AccountInfo"`
	Currency        string         `json:"Currency"`
	Signature       string         `json:"Signature"`
}

type AccountInfo struct {
	AccountNumber string `json:"AccountNumber"`
	Balance       string `json:"Balance"`
}

func (s *LoginOutput) toString() string {
	return "LoginOutput"
}

type LoginService struct {
}

type LoginServiceTest struct {
}

func (s *LoginServiceTest) execute(input DataObject) DataObject {
	loginInput := input.(*LoginInput)

	handler := sqlhandler.TrxIdHandler{}
	handler.ConnectDB()
	trxid, err := handler.IncreaseAndGetTrxid()
	handler.CloseConnection()

	output := LoginOutput{RequestId: loginInput.RequestId, BankId: loginInput.BankId}

	if err == nil {
		mapMutex.Lock()
		sessionMap[trxid] = 1
		mapMutex.Unlock()
		info := AccountInfo{AccountNumber: "16910000828964", Balance: "67801"}
		output.AccountInfo = append(output.AccountInfo, &info)
		output.Currency = "VND"
		output.TrxId = strconv.Itoa(trxid)
		output.ResponseCode = fmt.Sprintf("%08x", 0)
		output.ResponseMessage = "success"
	} else {
		output.ResponseCode = fmt.Sprintf("%08x", 2)
		output.ResponseMessage = "getting trxid failed"
	}
	return &output
}

func (s *LoginService) execute(input DataObject) DataObject {
	log.Printf("login service")
	loginInput := input.(*LoginInput)
	log.Printf("%+v\n", loginInput)
	sessionId := vn.StartSession(loginInput.BankId)

	fmt.Printf("sessionid : %d\n", sessionId)
	result := vn.LoadCaptcha(sessionId)
	result = vn.Login(sessionId, loginInput.UserName, loginInput.Password, "-")

	searchResult := vn.Search(sessionId, "25/01/2018")

	handler := sqlhandler.TrxIdHandler{}
	handler.ConnectDB()
	trxid, err := handler.IncreaseAndGetTrxid()
	handler.CloseConnection()
	output := LoginOutput{RequestId: loginInput.RequestId, BankId: loginInput.BankId}

	if searchResult.Line == 0 {
		fmt.Printf("line is 0")
		if searchResult.Data != nil {
			fmt.Printf("data is not nil")
			searchOutput := SearchOutput{}
			data := (searchResult).GetData()
			unmarshalError := json.Unmarshal([]byte(data), &searchOutput)

			if unmarshalError == nil {
				for _, account := range searchOutput.Acclist {
					info := AccountInfo{AccountNumber: account.Accno, Balance: account.Balance}
					output.AccountInfo = append(output.AccountInfo, &info)
				}
			} else {
				log.Printf("%s\n", unmarshalError.Error())
			}
		}
	}

	if err == nil {
		mapMutex.Lock()
		sessionMap[trxid] = sessionId
		mapMutex.Unlock()
		output.TrxId = strconv.Itoa(trxid)
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
	} else {
		output.ResponseCode = fmt.Sprintf("%08x", 2)
		output.ResponseMessage = "getting trxid failed"
	}

	return &output
}
