package disco_go

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestDisco(t *testing.T) {
	cfg := DefaultConfig().
		SkipSslVerify().
		WithDisco([]string{os.Getenv("DISCO_URL")}).
		WithToken(os.Getenv("DISCO_TOKEN")).
		//WithAuth(os.Getenv("DISCO_USER"), os.Getenv("DISCO_PASS")).
		WithName("test").
		WithEndpoints([]string{fmt.Sprintf("http://test:8080")}).
		WithRetry(2, 1*time.Second)
	cl, err := NewDiscoHttpClient(cfg)
	if err != nil {
		t.Fatalf("join error: %s", err.Error())
	}
	time.Sleep(100 * time.Millisecond)
	_ = cl.Leave()
}
