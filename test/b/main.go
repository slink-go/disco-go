package main

import (
	"fmt"
	disco_go "github.com/slink-go/disco-go"
	"github.com/slink-go/logging"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	os.Setenv("GO_ENV", "dev")
	os.Setenv("LOGGING_LEVEL_ROOT", "TRACE")
	os.Setenv("DISCO_URL", "http://localhost:8771")
	os.Setenv("DISCO_USER", "disco")
	os.Setenv("DISCO_PASS", "disco")

	cfg := disco_go.DefaultConfig().
		SkipSslVerify().
		WithDisco([]string{os.Getenv("DISCO_URL")}).
		//WithToken(os.Getenv("DISCO_TOKEN")).
		WithAuth(os.Getenv("DISCO_USER"), os.Getenv("DISCO_PASS")).
		WithName("test").
		WithEndpoints([]string{fmt.Sprintf("http://test:8080")}).
		WithRetry(2, 1*time.Second)
	cl, err := disco_go.NewDiscoHttpClient(cfg)
	if err != nil {
		logging.GetLogger("test").Warning("join error: %s", err.Error())
		panic(err)
	}

	quitChn := make(chan os.Signal)
	signal.Notify(quitChn, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	tm := time.NewTimer(time.Second * 10)
	//var i = 0
	for {
		select {
		case <-tm.C:
			for _, v := range cl.Registry().List() {
				fmt.Println(v)
			}
		case <-quitChn:
			//if i > 1 {
			tm.Stop()
			//cl.Leave()
			return
			//}
			//i++
		}
	}

}
