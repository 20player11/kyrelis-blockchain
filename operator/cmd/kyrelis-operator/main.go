// KYRELIS Operator: GUI for OBSERVING REAL Cosmos SDK + CometBFT nodes.
// This is not a blockchain, validator, wallet, or replacement for consensus.
package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
 "syscall"
	"time"

	"kyrelis.dev/operator/internal/realrpc"
)

//go:embed static/*
var static embed.FS

//go:embed assets/*
var assets embed.FS

const listenAt = "127.0.0.1:18789"

type Config struct {
	RPC  string `json:"rpc"`
	REST string `json:"rest"`
}
type app struct {
	mu         sync.RWMutex
	config     Config
	configPath string
	token      string
	worker     *runner
	shutdown   func()
	startedAt  time.Time
}

type runner struct {
	mu      sync.Mutex
	task    string
	running bool
	log     string
	failure string
	cmd     *exec.Cmd
}

func (r *runner) appendLog(data []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.log += string(data)
	if len(r.log) > 48000 {
		r.log = r.log[len(r.log)-48000:]
	}
}

type safeWriter struct{ r *runner }

func (w safeWriter) Write(b []byte) (int, error) { w.r.appendLog(b); return len(b), nil }
func (r *runner) begin(task, script string) error {
	exe, err := os.Executable()
 if err != nil { return err }
 binary:=filepath.Join(filepath.Dir(exe),"kyrelisd.exe")
 home:=os.Getenv("KYRELIS_NODE_HOME")
 if home==""{home=filepath.Join(filepath.Dir(exe),"node-data")}
 var args []string
 switch task {
 case "check": args=[]string{"genesis","validate","--home",home}
 case "start": args=[]string{"start","--home",home,"--api.enable","--api.address","tcp://127.0.0.1:1317","--grpc.address","127.0.0.1:9090"}
 default:return errors.New("This release already contains the compiled KYRELIS node")
 }
 r.mu.Lock()
	defer r.mu.Unlock()
	if r.running {
		return fmt.Errorf("Jiz probiha uloha: %s", r.task)
	}
	cmd := exec.Command(binary,args...)
 cmd.SysProcAttr=&syscall.SysProcAttr{CreationFlags:syscall.CREATE_NEW_PROCESS_GROUP}
 cmd.Stdout = safeWriter{r}
	cmd.Stderr = safeWriter{r}
	r.task = task
	r.running = true
	r.failure = ""
	r.log = ""
	r.cmd = cmd
	if err := cmd.Start(); err != nil {
		r.running = false
		r.failure = err.Error()
		return err
	}
	go func() {
		err := cmd.Wait()
		r.mu.Lock()
		defer r.mu.Unlock()
		r.running = false
		r.cmd = nil
		if err != nil {
			r.failure = err.Error()
		}
	}()
	return nil
}
func (r *runner) snapshot() map[string]any {
	r.mu.Lock()
	defer r.mu.Unlock()
	return map[string]any{"task": r.task, "running": r.running, "log": r.log, "failure": r.failure, "windows": runtime.GOOS == "windows"}
}
func (r *runner) stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()
 if !r.running || r.task!="start" || r.cmd==nil || r.cmd.Process==nil {return errors.New("No managed node is running")}
 result,_,err:=syscall.NewLazyDLL("kernel32.dll").NewProc("GenerateConsoleCtrlEvent").Call(1,uintptr(r.cmd.Process.Pid))
 if result==0{return err}
	return nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
func errorJSON(w http.ResponseWriter, code int, s string) {
	writeJSON(w, code, map[string]string{"error": s})
}
func (a *app) client() (*realrpc.Client, error) {
	a.mu.RLock()
	cfg := a.config
	a.mu.RUnlock()
	return realrpc.New(cfg.RPC, cfg.REST)
}
func (a *app) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		sendStatic(w, "static/index.html", "text/html; charset=utf-8")
	})
	mux.HandleFunc("GET /style.css", func(w http.ResponseWriter, r *http.Request) {
		sendStatic(w, "static/style.css", "text/css; charset=utf-8")
	})
	mux.HandleFunc("GET /app.js", func(w http.ResponseWriter, r *http.Request) {
		sendStatic(w, "static/app.js", "text/javascript; charset=utf-8")
	})
	mux.HandleFunc("GET /api/session", func(w http.ResponseWriter, r *http.Request) {
		a.mu.RLock()
		cfg := a.config
		a.mu.RUnlock()
		writeJSON(w, 200, map[string]any{"token": a.token, "config": cfg, "expected_chain_id": realrpc.RequiredChainID})
	})
	api := func(fn http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && origin != "http://"+listenAt {
				errorJSON(w, 403, "Origin neni povolen")
				return
			}
			if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Kyrelis-Token")), []byte(a.token)) != 1 {
				errorJSON(w, 403, "Neplatna session")
				return
			}
			fn(w, r)
		}
	}
	mux.HandleFunc("GET /api/status", api(func(w http.ResponseWriter, r *http.Request) {
		c, err := a.client()
		if err != nil {
			errorJSON(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, c.Status(r.Context()))
	}))
	mux.HandleFunc("GET /api/host", api(func(w http.ResponseWriter, r *http.Request) {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		started := a.startedAt
		if started.IsZero() {
			started = time.Now()
		}
		uptime := time.Since(started)
		writeJSON(w, 200, map[string]any{
			"heap_mib":   float64(mem.Alloc) / 1048576.0,
			"goroutines": runtime.NumGoroutine(),
			"platform":   runtime.GOOS,
			"uptime":     fmt.Sprintf("%dh %02dm", int(uptime.Hours()), int(uptime.Minutes())%60),
		})
	}))
	mux.HandleFunc("POST /api/broadcast", api(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
		defer r.Body.Close()
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		var data struct {
			TxBytes string `json:"tx_bytes"`
		}
		if err := dec.Decode(&data); err != nil {
			errorJSON(w, 400, "Malformed transaction JSON")
			return
		}
		if dec.Decode(&struct{}{}) != io.EOF {
			errorJSON(w, 400, "Trailing data")
			return
		}
		c, err := a.client()
		if err != nil {
			errorJSON(w, 400, err.Error())
			return
		}
		resp, err := c.BroadcastTx(r.Context(), data.TxBytes)
		if err != nil {
			errorJSON(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}))
	mux.HandleFunc("GET /api/validators", api(func(w http.ResponseWriter, r *http.Request) {
		c, err := a.client()
		if err != nil {
			errorJSON(w, 400, err.Error())
			return
		}
		vs, err := c.Validators(r.Context())
		if err != nil {
			errorJSON(w, 503, err.Error())
			return
		}
		writeJSON(w, 200, map[string]any{"validators": vs})
	}))
	mux.HandleFunc("GET /api/block", api(func(w http.ResponseWriter, r *http.Request) {
		height, err := strconv.ParseInt(r.URL.Query().Get("height"), 10, 64)
		if err != nil || height < 1 {
			errorJSON(w, 400, "Neplatna vyska bloku")
			return
		}
		c, err := a.client()
		if err != nil {
			errorJSON(w, 400, err.Error())
			return
		}
		b, err := c.Block(r.Context(), height)
		if err != nil {
			errorJSON(w, 503, err.Error())
			return
		}
		writeJSON(w, 200, b)
	}))
	mux.HandleFunc("GET /api/supply", api(func(w http.ResponseWriter, r *http.Request) {
		c, err := a.client()
		if err != nil {
			errorJSON(w, 400, err.Error())
			return
		}
		s, err := c.Supply(r.Context())
		if err != nil {
			errorJSON(w, 503, err.Error())
			return
		}
		writeJSON(w, 200, map[string]string{"amount_ukyrl": s})
	}))
	mux.HandleFunc("GET /api/balance", api(func(w http.ResponseWriter, r *http.Request) {
		c, err := a.client()
		if err != nil {
			errorJSON(w, 400, err.Error())
			return
		}
		s, err := c.Balance(r.Context(), r.URL.Query().Get("address"))
		if err != nil {
			errorJSON(w, 503, err.Error())
			return
		}
		writeJSON(w, 200, map[string]string{"balance_ukyrl": s})
	}))
	mux.HandleFunc("POST /api/config", api(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		defer r.Body.Close()
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		var cfg Config
		if err := dec.Decode(&cfg); err != nil {
			errorJSON(w, 400, "Neplatne nastaveni")
			return
		}
		if dec.Decode(&struct{}{}) != io.EOF {
			errorJSON(w, 400, "Nadbytecna data")
			return
		}
		if _, err := realrpc.New(cfg.RPC, cfg.REST); err != nil {
			errorJSON(w, 400, err.Error())
			return
		}
		raw, _ := json.MarshalIndent(cfg, "", "  ")
		tmp := a.configPath + ".tmp"
		if err := os.WriteFile(tmp, raw, 0600); err != nil {
			errorJSON(w, 500, "Ulozeni selhalo")
			return
		}
		if err := os.Rename(tmp, a.configPath); err != nil {
			errorJSON(w, 500, "Ulozeni selhalo")
			return
		}
		a.mu.Lock()
		a.config = cfg
		a.mu.Unlock()
		writeJSON(w, 200, map[string]any{"ok": true})
	}))
	mux.HandleFunc("GET /api/runner", api(func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, a.worker.snapshot()) }))
	mux.HandleFunc("POST /api/runner/check", api(func(w http.ResponseWriter, r *http.Request) {
		b, _ := assets.ReadFile("assets/check.sh")
		if err := a.worker.begin("check", string(b)); err != nil {
			errorJSON(w, 409, err.Error())
			return
		}
		writeJSON(w, 202, map[string]bool{"started": true})
	}))
	mux.HandleFunc("POST /api/runner/create", api(func(w http.ResponseWriter, r *http.Request) {
		script, _ := assets.ReadFile("assets/create.sh")
		cfg, _ := assets.ReadFile("assets/config.yml")
		body := string(script)
		body = strings.Replace(body, "# Replaced by the operator's bundled shell/config payload.", "cat > config.yml <<'KYRELIS_YAML'\n"+string(cfg)+"\nKYRELIS_YAML", 1)
		if err := a.worker.begin("create", body); err != nil {
			errorJSON(w, 409, err.Error())
			return
		}
		writeJSON(w, 202, map[string]bool{"started": true})
	}))
	mux.HandleFunc("POST /api/runner/rebuild", api(func(w http.ResponseWriter, r *http.Request) {
		b, _ := assets.ReadFile("assets/rebuild.sh")
		if err := a.worker.begin("rebuild", string(b)); err != nil {
			errorJSON(w, 409, err.Error())
			return
		}
		writeJSON(w, 202, map[string]bool{"started": true})
	}))
	mux.HandleFunc("POST /api/runner/start", api(func(w http.ResponseWriter, r *http.Request) {
		b, _ := assets.ReadFile("assets/start.sh")
		if err := a.worker.begin("start", string(b)); err != nil {
			errorJSON(w, 409, err.Error())
			return
		}
		writeJSON(w, 202, map[string]bool{"started": true})
	}))
	mux.HandleFunc("POST /api/runner/stop", api(func(w http.ResponseWriter, r *http.Request) {
		if err := a.worker.stop(); err != nil {
			errorJSON(w, 409, err.Error())
			return
		}
		writeJSON(w, 200, map[string]bool{"stopping": true})
	}))

	mux.HandleFunc("POST /api/exit", api(func(w http.ResponseWriter, r *http.Request) {
		a.worker.mu.Lock()
		running := a.worker.running
		a.worker.mu.Unlock()
		if running {
			errorJSON(w, 409, "Nejdriv zastav nebo dokonci bezici WSL ulohu")
			return
		}
		writeJSON(w, 200, map[string]bool{"closing": true})
		if a.shutdown != nil {
			go a.shutdown()
		}
	}))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != listenAt {
			errorJSON(w, 403, "Invalid Host")
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		mux.ServeHTTP(w, r)
	})
}
func sendStatic(w http.ResponseWriter, path, contentType string) {
	b, err := static.ReadFile(path)
	if err != nil {
		http.Error(w, "internal asset missing", 500)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(b)
}
func openBrowser(u string) {
	if os.Getenv("KYRELIS_NO_BROWSER") == "1" {
		return
	}
	switch runtime.GOOS {
	case "windows":
		// Edge app-mode removes the browser URL bar and presents a desktop-like window.
		// It does not impersonate a signed native Tauri/WebView2 desktop application.
		candidates := []string{
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Edge", "Application", "msedge.exe"),
		}
		for _, candidate := range candidates {
			if candidate == "" {
				continue
			}
			if _, err := os.Stat(candidate); err == nil {
				if exec.Command(candidate, "--app="+u, "--window-size=1440,780").Start() == nil {
					return
				}
			}
		}
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
	case "darwin":
		_ = exec.Command("open", u).Start()
	default:
		_ = exec.Command("xdg-open", u).Start()
	}
}
func main() {
	root, e := os.UserConfigDir()
	if e != nil {
		log.Fatal(e)
	}
	dir := filepath.Join(root, "Kyrelis", "operator-v0.3")
	if e = os.MkdirAll(dir, 0700); e != nil {
		log.Fatal(e)
	}
	cfgpath := filepath.Join(dir, "settings.json")
	cfg := Config{RPC: "http://127.0.0.1:26657", REST: "http://127.0.0.1:1317"}
	if raw, e := os.ReadFile(cfgpath); e == nil {
		var loaded Config
		if json.Unmarshal(raw, &loaded) == nil {
			if _, e = realrpc.New(loaded.RPC, loaded.REST); e == nil {
				cfg = loaded
			}
		}
	}
	key := make([]byte, 32)
	if _, e = rand.Read(key); e != nil {
		log.Fatal(e)
	}
	a := &app{config: cfg, configPath: cfgpath, token: hex.EncodeToString(key), worker: &runner{}, startedAt: time.Now()}
	l, e := net.Listen("tcp", listenAt)
	if e != nil {
		log.Fatalf("Port %s uz pouziva jina aplikace: %v", listenAt, e)
	}
	defer l.Close()
	server := &http.Server{Handler: a.handler(), ReadTimeout: 8 * time.Second, WriteTimeout: 12 * time.Second, IdleTimeout: 10 * time.Second, ReadHeaderTimeout: 5 * time.Second}
	a.shutdown = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}
	url := "http://" + listenAt
	fmt.Println("KYRELIS Console v0.3 — verified CometBFT RPC; GUI:", url)
	openBrowser(url)
	if e = server.Serve(l); e != nil && !errors.Is(e, http.ErrServerClosed) {
		log.Fatal(e)
	}
}

var _ = context.Background
var _ = strings.TrimSpace
