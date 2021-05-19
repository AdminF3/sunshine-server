package config

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
)

var Path = GetPath()

type General struct {
	Name           string   `toml:"name"`
	URL            string   `toml:"url"`
	Logo           string   `toml:"logo"`
	AllowedOrigins []string `toml:"allowed_origins"`
	Port           int      `toml:"port"`
	SentryDSN      string   `toml:"sentry"`
}

type Paths struct {
	Migrations string `toml:"migrations"`
	LaTeX      string `toml:"latex"`
	Uploads    string `toml:"uploads"`
}

type DB struct {
	Host     string `toml:"host"`
	Port     uint16 `toml:"port"`
	Name     string `toml:"name"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type Session struct {
	Auth     string        `toml:"auth"`
	Encr     string        `toml:"encr"`
	Path     string        `toml:"path"`
	HTTPOnly bool          `toml:"http_only"`
	SameSite http.SameSite `toml:"samesite"`
	Secure   bool          `toml:"secure"`
}

type Mail struct {
	Backend  string `toml:"backend"`
	From     string `toml:"from"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type Config struct {
	General General `toml:"general"`
	Paths   Paths   `toml:"paths"`
	DB      DB      `toml:"psql"`
	Session Session `toml:"session"`
	Mail    Mail    `toml:"mail"`
}

// Dependency stores an ID and Kind of an entity.
//
// NOTE: Find appropriate place for this one (#61).
type Dependency struct {
	ID   uuid.UUID
	Kind string
}

// ConfigFile returns full path of currently used config file,
// by reading the environment variable `SUNSHINE_CONFIG`.
func ConfigFile() string {
	configPath, ok := os.LookupEnv("SUNSHINE_CONFIG")
	if !ok {
		env, ok := os.LookupEnv("SUNSHINE_ENV")
		if !ok {
			env = "dev"
		}
		configPath = path.Join(Path, "config", env+".toml")
	}

	return configPath
}

// Load the config file or fail completely.
func Load() Config {
	var cfg Config

	if _, err := toml.DecodeFile(ConfigFile(), &cfg); err != nil {
		log.Fatal("Error loading cfg:", err)
	}

	if err := ensureFolder(&cfg.Paths.Uploads, "uploads"); err != nil {
		log.Fatalf("upload path(%s): %v", cfg.Paths.Uploads, err)
	}

	if err := ensureFolder(&cfg.Paths.LaTeX, "contract/tex"); err != nil {
		log.Fatalf("upload path(%s): %v", cfg.Paths.LaTeX, err)
	}

	return cfg
}

// ensureFolder target exist and return absolute path to it. Otherwise deduces
// it from given source and creates the folder with correct permissions.
func ensureFolder(target *string, source string) error {
	if target == nil {
		return errors.New("nil target")
	}

	if len(*target) == 0 {
		*target = filepath.Join(GetPath(), source)
	}
	if !filepath.IsAbs(*target) {
		*target = filepath.Join(GetPath(), *target)
	}

	return os.MkdirAll(*target, 0755)
}

// GetPath deduces root project path.
func GetPath() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "..")
}
