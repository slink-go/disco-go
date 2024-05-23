package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/joho/godotenv"
	dg "github.com/slink-go/disco-go"
	"github.com/slink-go/logging"
	"os"
	"os/signal"
	"time"
)

func main() {

	os.Setenv("GO_ENV", "dev")
	godotenv.Load(".env")

	cfg := dg.DefaultConfig().
		WithDisco([]string{"http://localhost:8762"}).
		WithAuth("disco", "disco").
		WithName("test").
		WithEndpoints([]string{fmt.Sprintf("http://test:8080")}).
		WithRetry(2, 1*time.Second)

	_, err := connect(cfg)
	if err != nil {
		panic("could not join disco")
	}

	app := fiber.New()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		_ = <-c
		time.Sleep(100 * time.Millisecond)
		logging.GetLogger("main").Info("fiber graceful shutdown...")
		_ = app.Shutdown()
	}()

	app.Use(pprof.New())
	app.Get("/", func(ctx *fiber.Ctx) error {
		err := ctx.SendString("Hello!\n")
		return err
	})
	app.Listen(":3000")

}

func connect(cfg *dg.DiscoClientConfig) (client dg.DiscoClient, err error) {
	for {
		client, err = dg.NewDiscoHttpClient(cfg)
		if err == nil {
			break
		}
		logging.GetLogger("main").Warning("%s", err)
		time.Sleep(1 * time.Second)
	}
	return
}
