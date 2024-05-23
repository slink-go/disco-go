package disco_go

import (
	"fmt"
	"github.com/slink-go/logging"
	"os"
	"testing"
	"time"
)

func TestDisco(t *testing.T) {
	os.Setenv("DISCO_URL", "http://localhost:8762")
	os.Setenv("DISCO_USER", "disco")
	os.Setenv("DISCO_PASS", "disco")

	cfg := DefaultConfig().
		SkipSslVerify().
		WithDisco([]string{os.Getenv("DISCO_URL")}).
		//WithToken(os.Getenv("DISCO_TOKEN")).
		WithAuth(os.Getenv("DISCO_USER"), os.Getenv("DISCO_PASS")).
		WithName("test").
		WithEndpoints([]string{fmt.Sprintf("http://test:8080")}).
		WithRetry(2, 1*time.Second)
	cl, err := NewDiscoHttpClient(cfg)
	if err != nil {
		logging.GetLogger("test").Warning("join error: %s", err.Error())
		t.Fatalf("join error: %s", err.Error())
	}
	time.Sleep(5000 * time.Millisecond)
	_ = cl.Leave()
}
