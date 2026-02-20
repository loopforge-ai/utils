package html

import "time"

const (
	DefaultAddr        = ":8080"
	IdleTimeout        = 120 * time.Second
	MaxHeaderBytes     = 1 << 20 // 1 MiB
	ReadHeaderTimeout  = 5 * time.Second
	ReadTimeout        = 15 * time.Second
	ShutdownTimeout    = 10 * time.Second
	StaticCacheControl = "public, max-age=3600"
	WriteTimeout       = 15 * time.Second
)
