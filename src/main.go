package main

import (
	"context"
	"database/sql/driver"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sqids/sqids-go"
	_ "modernc.org/sqlite"
)

//go:embed web/dist
var webFS embed.FS

const serverAddr = "127.0.0.1:10101"

var (
	sq      *sqids.Sqids
	dataDir = "data"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)

	if v := os.Getenv("DATA_DIR"); v != "" {
		dataDir = v
	}

	var err error
	sq, err = sqids.New(sqids.Options{MinLength: 8})
	if err != nil {
		log.Fatalf("[App] failed to init sqids: %v", err)
	}
}

func encodeID(id int64) string {
	s, _ := sq.Encode([]uint64{uint64(id)})
	return s
}

func decodeID(s string) int64 {
	if s == "" || s == "all" {
		return 0
	}
	nums := sq.Decode(s)
	if len(nums) == 0 {
		return 0
	}
	return int64(nums[0])
}

type Group struct {
	ID        string `json:"_id"`
	Name      string `json:"name"`
	Sort      int    `json:"sort"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type Proxy struct {
	ID        string `json:"_id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	IP        string `json:"ip"`
	Lang      string `json:"lang"`
	Timezone  string `json:"timezone"`
	Location  string `json:"location"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type FingerprintConfig struct {
	Seed                int32    `json:"seed"`
	Platform            string   `json:"platform"`
	Brand               string   `json:"brand"`
	HardwareConcurrency string   `json:"hardwareConcurrency"`
	DeviceMemory        string   `json:"deviceMemory"`
	DisableFeatures     []string `json:"disableFeatures"`
	Screen              string   `json:"screen"`
	Lang                string   `json:"lang"`
	Timezone            string   `json:"timezone"`
	IP                  string   `json:"ip"`
	Location            string   `json:"location"`
	DisableFingerprint  []string `json:"disableFingerprint"`
	RandomFingerprint   bool     `json:"randomFingerprint"`
	ProxyLang           bool     `json:"proxyLang"`
	ProxyTimezone       bool     `json:"proxyTimezone"`
	ProxyLocation       bool     `json:"proxyLocation"`
}

func (f *FingerprintConfig) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("unsupported type for FingerprintConfig: %T", src)
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, f)
}

func (f FingerprintConfig) Value() (driver.Value, error) {
	b, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type Profile struct {
	ID          string            `json:"_id"`
	Name        string            `json:"name"`
	GroupID     string            `json:"groupId"`
	Sort        int               `json:"sort"`
	Proxy       string            `json:"proxy"`
	Fingerprint FingerprintConfig `json:"fingerprint"`
	Args        string            `json:"args"`
	Notes       string            `json:"notes"`
	CreatedAt   int64             `json:"createdAt"`
	UpdatedAt   int64             `json:"updatedAt"`
}

type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// writeJSON writes a JSON-encoded response to the response writer.
func writeJSON(w http.ResponseWriter, resp any) {
	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// withMiddleware wraps the handler.
func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Panic recovery
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[Server] panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func main() {
	initDB()
	defer db.Close()

	mux := http.NewServeMux()

	// Group routes
	mux.HandleFunc("GET /get_groups", getGroups)
	mux.HandleFunc("POST /add_group", addGroup)
	mux.HandleFunc("POST /update_group", updateGroup)
	mux.HandleFunc("POST /delete_group", deleteGroup)

	// Profile routes
	mux.HandleFunc("GET /get_profiles", getProfiles)
	mux.HandleFunc("GET /get_profile", getProfile)
	mux.HandleFunc("POST /add_profile", addProfile)
	mux.HandleFunc("POST /update_profile", updateProfile)
	mux.HandleFunc("POST /delete_profile", deleteProfile)
	mux.HandleFunc("POST /launch_profile", launchProfile)
	mux.HandleFunc("POST /stop_profile", stopProfile)
	mux.HandleFunc("GET /show_profile", showProfile)
	mux.HandleFunc("GET /export_cookies", exportCookies)
	mux.HandleFunc("POST /import_cookies", importCookies)

	// Proxy routes
	mux.HandleFunc("GET /get_proxies", getProxies)
	mux.HandleFunc("GET /get_proxy", getProxy)
	mux.HandleFunc("POST /add_proxy", addProxy)
	mux.HandleFunc("POST /update_proxy", updateProxy)
	mux.HandleFunc("POST /delete_proxy", deleteProxy)

	// SSE
	mux.HandleFunc("GET /events", eventsHandler)

	// Static files
	fsSub, _ := fs.Sub(webFS, "web/dist")
	mux.Handle("/", http.FileServer(http.FS(fsSub)))

	handler := withMiddleware(mux)
	server := &http.Server{Addr: serverAddr, Handler: handler}

	go func() {
		log.Printf("[Server] listening on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Server] listen error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Printf("[Server] shutdown signal received, shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("[Server] forced shutdown: %v", err)
	}
	log.Printf("[Server] exited gracefully")
}
