package main

import (
	"backend/internal/config"
	"fmt"
	"log/slog"
	"os"
)

func main()  {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
	fmt.Println("Shutting down...")
	os.Exit(1)
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return  err
	}

	fmt.Printf("APP_ENV: %q\nHTTP_ADDR: %q\nDATABASE_URL: %q\nLOG_LEVEL: %q\nHTTP_READ_TIMEOUT: %q\nSHUTDOWN_TIMEOUT: %q\n", cfg.Env, cfg.Addr, cfg.DatabaseURL, cfg.LogLevel, cfg.ReadTimeout, cfg.ShutdownTimeout)

	return  nil
}
