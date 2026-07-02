package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"nutricionist/auth"
	"nutricionist/db"
	"nutricionist/handlers"
	"nutricionist/middleware"
	"nutricionist/updater"
)

// Version is injected at build time: -ldflags "-X main.Version=1.0.0"
var Version = "dev"

//go:embed frontend/dist
var frontendFS embed.FS

func main() {
	// ── Data directory ──────────────────────────────────────────────────────
	dataDir := dataDirectory()
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		log.Fatalf("no se pudo crear directorio de datos: %v", err)
	}

	// ── JWT secret (generate once, persist) ─────────────────────────────────
	if err := auth.Init(filepath.Join(dataDir, "secret.key")); err != nil {
		log.Fatalf("error inicializando auth: %v", err)
	}

	// ── SQLite database ──────────────────────────────────────────────────────
	database, err := db.Init(filepath.Join(dataDir, "nutricionist.db"))
	if err != nil {
		log.Fatalf("error abriendo base de datos: %v", err)
	}
	defer database.Close()

	// ── Router ───────────────────────────────────────────────────────────────
	r := chi.NewRouter()

	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.Logger)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.MaxBodySize)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         300,
	}))

	// ── Image cache directory ─────────────────────────────────────────────────
	if err := os.MkdirAll(filepath.Join(dataDir, "imagenes"), 0700); err != nil {
		log.Fatalf("no se pudo crear directorio de imagenes: %v", err)
	}

	// ── API routes ───────────────────────────────────────────────────────────
	h := handlers.New(database, dataDir)

	r.Route("/api", func(r chi.Router) {
		// Public
		r.Post("/auth/login", h.Login)
		r.Post("/auth/registro", h.Registro)
		r.Post("/auth/refresh", h.RefreshToken)

		// Protected
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth)

			r.Get("/me", h.GetMe)
			r.Patch("/perfil", h.UpdatePerfil)
			r.Post("/perfil/password", h.ChangePassword)

			r.Get("/pacientes", h.ListPacientes)
			r.Post("/pacientes", h.CreatePaciente)
			r.Get("/pacientes/{id}", h.GetPaciente)
			r.Patch("/pacientes/{id}", h.UpdatePaciente)
			r.Delete("/pacientes/{id}", h.DeletePaciente)

			r.Get("/pacientes/{id}/consultas", h.ListConsultas)
			r.Post("/pacientes/{id}/consultas", h.CreateConsulta)

			r.Get("/planes", h.ListPlanes)
			r.Post("/planes", h.CreatePlan)
			r.Get("/planes/{id}", h.GetPlan)
			r.Patch("/planes/{id}", h.UpdatePlan)
			r.Delete("/planes/{id}", h.DeletePlan)
			r.Post("/planes/{id}/alimentos", h.AddAlimentoPlan)
			r.Patch("/planes/{id}/alimentos/{alimentoId}", h.UpdateAlimentoPlan)
			r.Delete("/planes/{id}/alimentos/{alimentoId}", h.RemoveAlimentoPlan)

			r.Get("/alimentos", h.ListAlimentos)
			r.Get("/platillos", h.ListPlatillos)
			r.Get("/platillos/compatibles/{pacienteId}", h.ListPlatillosCompatibles)
			r.Get("/platillos/{id}/variantes", h.ListVariantes)
			r.Get("/imagenes/{id}", h.ServeImagen)

			r.Get("/pacientes/{id}/restricciones", h.GetRestriccionesPaciente)
			r.Post("/pacientes/{id}/restricciones", h.AddRestriccionPaciente)
			r.Delete("/pacientes/{id}/restricciones/{restriccionId}", h.DeleteRestriccionPaciente)

			r.Get("/pacientes/{id}/seguimiento", h.ListSeguimientos)
			r.Post("/pacientes/{id}/seguimiento", h.AddSeguimiento)
			r.Delete("/pacientes/{id}/seguimiento/{sid}", h.DeleteSeguimiento)

			r.Post("/ia/generar", h.GenerarPlanIA)
			r.Get("/ia/config", h.GetDeepseekConfig)
			r.Post("/ia/config", h.SaveDeepseekConfig)
		})

		// Version endpoint (no auth needed — used by updater check on client)
		r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"version": Version})
		})
	})

	// ── Frontend (embedded static files) ─────────────────────────────────────
	distFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatalf("error cargando frontend: %v", err)
	}
	fileServer := http.FileServer(http.FS(distFS))
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// SPA fallback: serve index.html for unknown routes
		if _, err := fs.Stat(distFS, r.URL.Path[1:]); os.IsNotExist(err) {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})

	// ── Find available port ───────────────────────────────────────────────────
	port := findPort(3847)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	// ── Auto-updater (background) ─────────────────────────────────────────────
	updater.StartBackgroundChecker(Version, func(newVer string) {
		log.Printf("[updater] Actualizado a v%s — reiniciando...", newVer)
		time.Sleep(500 * time.Millisecond)
		restartSelf()
	})

	// ── Start server ──────────────────────────────────────────────────────────
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Nutricionist v%s iniciando en http://%s", Version, addr)

	// Open browser after short delay
	go func() {
		time.Sleep(800 * time.Millisecond)
		openBrowser(fmt.Sprintf("http://%s", addr))
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error servidor: %v", err)
	}
}

// dataDirectory returns the app data path for storing DB and config.
func dataDirectory() string {
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "Nutricionist")
		}
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "Library", "Application Support", "Nutricionist")
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".nutricionist")
	}
	return "."
}

// findPort finds the first available port starting from preferred.
func findPort(preferred int) int {
	for port := preferred; port < preferred+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			ln.Close()
			return port
		}
	}
	return preferred
}

// openBrowser opens the default browser at the given URL.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

// restartSelf re-launches the current executable.
func restartSelf() {
	exe, err := os.Executable()
	if err != nil {
		os.Exit(0)
		return
	}
	cmd := exec.Command(exe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Start()
	os.Exit(0)
}

// Ensure context key types don't conflict.
var _ context.Context = context.Background()
