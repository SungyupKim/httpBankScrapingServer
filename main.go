package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	vn "omnidocvn_server/vn"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kenshaw/sdhook"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	log          *logrus.Entry
	handledCount int
	sessionMap   = make(map[int]int)
	done         = make(chan bool, 1)
	server       *http.Server

	prePostProcessor = PrePostProcessor{}
	loginService     = &LoginServiceTest{}
	//loginService = &LoginService{}
	preTransferService = &PreTransferServiceTest{}
	//preTransferService = &PreTransferService{}
	otpVerifyService = &OtpVerifyServiceTest{}
	//otpVerifyService = &OtpVerifyService{}
)

func handleTermSignal() {

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM)
		log.Infof("in signal checking\n")
		_ = <-sigs
		log.Infof("got signal\n")
		ctxShutDown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer func() {
			cancel()
		}()

		err := server.Shutdown(ctxShutDown)
		if err != nil {
			log.Printf("%s\n", err.Error())
		}

		done <- true
	}()
}

func login(w http.ResponseWriter, r *http.Request) {
	log.Infof("got login request")
	log.Infof("headers : %+v\n", r.Header)
	input := LoginInput{}
	prePostProcessor.preProcess(&input, r)
	log.Infof("%+v\n", input)
	output := (loginService).execute(&input)

	prePostProcessor.postProcess(output, w)
}

func pretransfer(w http.ResponseWriter, r *http.Request) {
	log.Infof("got pretransfer request")
	if CheckHasCookieHeader(r.Header) == false {
		output := PreTransferOutput{ResponseMessage: "you should send cookie"}
		marshalOutput, _ := json.Marshal(output)
		w.Write(marshalOutput)
		return
	}
	input := PreTransferInput{}
	prePostProcessor.preProcess(&input, r)
	log.Infof("%+v\n", input)
	output := (preTransferService).execute(&input)
	prePostProcessor.postProcess(output, w)
}

func otpverify(w http.ResponseWriter, r *http.Request) {
	log.Printf("got otpverify request")

	if CheckHasCookieHeader(r.Header) == false {
		output := OtpVerifyServiceOutput{ResponseMessage: "you should send cookie"}
		marshalOutput, _ := json.Marshal(output)
		w.Write(marshalOutput)
		return
	}

	input := OtpVerifyServiceInput{}
	prePostProcessor.preProcess(&input, r)
	log.Printf("%+v\n", input)
	output := (otpVerifyService).execute(&input)
	prePostProcessor.postProcess(output, w)
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))
}

func metrics(w http.ResponseWriter, r *http.Request) {
	sequenceMutex.Lock()
	fmt.Fprintf(w, "bankbe_requests{} %d", requests)
	sequenceMutex.Unlock()
}

func initLogger() {
	h, err := sdhook.New(
		sdhook.GoogleServiceAccountCredentialsFile("./BankBe-61e363d54339.json"),
		sdhook.LogName("bankbe"),
	)

	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}

	_log := &logrus.Logger{
		Out: io.MultiWriter(os.Stdout, &lumberjack.Logger{
			Filename:  "./bankbe.log",
			MaxSize:   100, // MB
			LocalTime: true,
			MaxAge:    365, // Days
			Compress:  true,
		}),
		Formatter: &prefixed.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05.000000",
			FullTimestamp:   true,
			ForceFormatting: true,
			ForceColors:     false,
		},
		Hooks: make(logrus.LevelHooks),
		Level: logrus.DebugLevel,
	}

	_log.AddHook(h)

	log = _log.WithFields(logrus.Fields{
		"pod": os.Getenv("HOSTNAME"),
	})
}
func init() {

	requests = 0
	handledCount = 0
	initLogger()
	ret := vn.Initialize(".", 4)
	if ret < 0 {
		log.Fatalf("initialize failed")
	}

	log.Infof("initialize success")

}

func main() {
	server = &http.Server{Addr: ":3000", Handler: http.DefaultServeMux}
	http.HandleFunc("/", healthcheck)
	http.HandleFunc("/login", login)
	http.HandleFunc("/metrics", metrics)
	http.HandleFunc("/pretransfer", pretransfer)
	http.HandleFunc("/otpverify", otpverify)
	handleTermSignal()
	server.ListenAndServe()

	_ = <-done
}
