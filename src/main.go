package main

import (
	"disco-go/client"
	"disco-go/config"
	"disco-go/http"
	"fmt"
	"github.com/ws-slink/disco/common/util/logger"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const baseUrl = "http://localhost:8080"

func getToken(tenant string) (string, error) {
	client := http.
		ClientBuilder("").
		WithTimeout(time.Second).
		WithBreaker(3).
		WithRetry(10, 1*time.Second).
		Build()
	data, _, err := client.Get(fmt.Sprintf("%s/api/token/%s", baseUrl, tenant), nil)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
func testGetToken(wg *sync.WaitGroup) {
	v, err := getToken("test")
	if err != nil {
		logger.Warning("could not get token (%T): %s", err, err.Error())
	} else {
		logger.Notice("got token: %s", v)
	}
	if wg != nil {
		wg.Done()
	}
}
func testWrappers() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go testGetToken(&wg)
	go testGetToken(&wg)
	wg.Wait()
}

func main() {

	token, err := getToken("ta")
	if err != nil {
		panic(err)
	}

	//client := http.
	//	ClientBuilder(token).
	//	WithTimeout(time.Second).
	//	WithBreaker(3).
	//	WithRetry(10, 1*time.Second).
	//	Build()
	//data, err := client.Get(fmt.Sprintf("%s/api/list", baseUrl), nil)
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "%s", err.Error())
	//} else {
	//	fmt.Printf("%s\n", data)
	//}

	//waitSec := 25
	threads := 10

	i := 0
	for i = 0; i < threads; i++ {
		go func(idx int) {
			logger.Notice(">>> config")
			cfg := config.
				Default().
				WithToken(token).
				WithName("DISCO").
				WithDisco([]string{"http://localhost:8080"}).
				WithEndpoints([]string{fmt.Sprintf("http://localhost:808%d", idx+1)}).
				WithBreaker(3).
				WithRetry(5, 2*time.Second).
				//WithTimeout(5 * time.Second).
				Get()
			logger.Notice(">>> init")
			_, _ = client.NewDiscoHttpClient(cfg)
			//if err != nil {
			//	panic(err)
			//}
			//time.Sleep(time.Duration(waitSec) * time.Second)
			//logger.Notice(">>> leave")
			//err = clnt.Leave()
			//if err != nil {
			//	panic(err)
			//}
		}(i)
	}

	handleSignals(syscall.SIGINT, syscall.SIGTERM)
	time.Sleep(time.Second)
}

func handleSignals(signals ...os.Signal) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, signals...)
	done := make(chan bool, 1)
	go func() {
		sig := <-sigs
		logger.Debug("[main][signal] received %s signal", sig)
		done <- true
	}()
	<-done
}
