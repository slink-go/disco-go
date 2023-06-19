package main

import (
	"disco-go/client"
	"disco-go/config"
	"disco-go/http"
	"fmt"
	"github.com/ws-slink/disco/common/api"
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

	threads := 1

	i := 0
	var registry client.DiscoRegistry
	for i = 0; i < threads; i++ {
		go func(idx int) {
			cfg := config.
				Default().
				WithToken(token).
				WithName("TEST").
				WithDisco([]string{"http://localhost:8080"}).
				WithEndpoints([]string{fmt.Sprintf("http://localhost:808%d", idx+1)}).
				WithBreaker(3).
				WithRetry(5, 2*time.Second).
				//WithTimeout(5 * time.Second).
				Get()
			clnt, _ := client.NewDiscoHttpClient(cfg)
			registry = clnt.Registry()
		}(i)
	}

	time.Sleep(50 * time.Millisecond)
	registry.List()

	go func() {
		t := time.NewTicker(3 * time.Second)
		for {
			select {
			case _ = <-t.C:
				cc, err := registry.Get("DISCO")
				if err != nil {
					logger.Warning("%s", err.Error())
				} else {
					ep, _ := cc.Endpoint(api.HttpEndpoint)
					logger.Notice("  > DISCO client: %s: %s", cc.ClientId(), ep)
				}
			default:
			}
		}
	}()

	//go func() {
	//	t := time.NewTicker(30 * time.Second)
	//	for {
	//		select {
	//		case _ = <-t.C:
	//			logger.Notice("     >> clients list:")
	//			for _, c := range registry.List() {
	//				logger.Notice("         %s: %s", c.ClientId(), c.State())
	//			}
	//		default:
	//		}
	//	}
	//}()

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
