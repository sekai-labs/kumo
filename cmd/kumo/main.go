package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sekai-labs/kumo/internal/core/application"
	"github.com/sekai-labs/kumo/internal/core/ports"
	"github.com/sekai-labs/kumo/internal/infrastructure/cache"
	"github.com/sekai-labs/kumo/internal/infrastructure/cloudflare"
	"github.com/sekai-labs/kumo/internal/infrastructure/config"
	"github.com/sekai-labs/kumo/internal/infrastructure/logging"
	"github.com/sekai-labs/kumo/internal/presentation/tui/app"
)

var (
	version = "0.1.0"
	commit  = "dev"
	date    = "2026-09-27"
)

func main() {
	var (
		flagProfile    = flag.String("profile", "", "Configuration profile to activate")
		flagAccount    = flag.String("account", "", "Initial Cloudflare account ID to select")
		flagZone       = flag.String("zone", "", "Initial Cloudflare zone ID or name to select")
		flagLogLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		flagLogPath    = flag.String("log-path", "", "Path to log file (defaults to $XDG_STATE_HOME/kumo/kumo.log or /tmp/kumo.log)")
		flagVersion    = flag.Bool("version", false, "Print version information and exit")
		flagConfigFile = flag.String("config", "", "Custom configuration file path")
	)

	flag.Parse()

	if *flagVersion {
		fmt.Printf("kumo %s (commit: %s, built: %s)\n", version, commit, date)
		return
	}

	logPath := *flagLogPath
	if logPath == "" {
		if stateHome := os.Getenv("XDG_STATE_HOME"); stateHome != "" {
			logPath = filepath.Join(stateHome, "kumo", "kumo.log")
		} else if home, err := os.UserHomeDir(); err == nil {
			logPath = filepath.Join(home, ".local", "state", "kumo", "kumo.log")
		} else {
			logPath = filepath.Join(os.TempDir(), "kumo.log")
		}
	}

	parsedLevel := logging.ParseLevel(*flagLogLevel)
	logger, err := logging.NewLogger(logPath, parsedLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize file logger: %v\n", err)
	} else {
		defer logger.Close()
		logger.Info("Starting kumo %s (commit: %s)", version, commit)
	}

	configRepo := config.NewConfigRepository(*flagConfigFile)
	ctxLoad, cancelLoad := context.WithTimeout(context.Background(), 2*time.Second)
	cfg, err := configRepo.Load(ctxLoad)
	cancelLoad()
	if err != nil {
		if logger != nil {
			logger.Warn("Failed to load config file: %v. Using environment/defaults.", err)
		}
	}

	activeProfileName := *flagProfile
	if activeProfileName == "" {
		activeProfileName = cfg.DefaultProfile
	}
	if activeProfileName == "" {
		activeProfileName = "default"
	}

	activeProfile, hasProfile := cfg.Profiles[activeProfileName]
	if !hasProfile {
		activeProfile = ports.Profile{
			Name: activeProfileName,
		}
	}

	if *flagAccount != "" {
		activeProfile.AccountID = *flagAccount
	}
	if *flagZone != "" {
		activeProfile.ZoneID = *flagZone
	}

	var (
		accountRepo   ports.AccountRepository
		zoneRepo      ports.ZoneRepository
		dnsRepo       ports.DNSRepository
		tunnelRepo    ports.TunnelRepository
		workerRepo    ports.WorkerRepository
		pagesRepo     ports.PagesRepository
		rulesetRepo   ports.RulesetRepository
		analyticsRepo ports.AnalyticsRepository
	)
	token := activeProfile.Token
	if token == "" && activeProfile.TokenEnv != "" {
		token = os.Getenv(activeProfile.TokenEnv)
	}
	if token == "" {
		token = os.Getenv("CLOUDFLARE_API_TOKEN")
	}

	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: No Cloudflare API token provided.")
		fmt.Fprintln(os.Stderr, "Please provide a token via:")
		fmt.Fprintln(os.Stderr, "  1. Environment variable: export CLOUDFLARE_API_TOKEN=\"your_token\"")
		fmt.Fprintln(os.Stderr, "  2. Configuration profile in ~/.config/kumo/config.yaml")
		os.Exit(1)
	}

	cfClient := cloudflare.NewClient(token)

	accountRepo = cfClient
	zoneRepo = cfClient.Zones()
	dnsRepo = cfClient.DNS()
	tunnelRepo = cfClient.Tunnels()
	workerRepo = cfClient.Workers()
	pagesRepo = cfClient.Pages()
	rulesetRepo = cfClient
	analyticsRepo = cfClient

	memCache := cache.NewMemoryCache()
	repos := application.Repositories{
		Account:   accountRepo,
		Zone:      zoneRepo,
		DNS:       dnsRepo,
		Tunnel:    tunnelRepo,
		Worker:    workerRepo,
		Pages:     pagesRepo,
		Ruleset:   rulesetRepo,
		Analytics: analyticsRepo,
	}
	appService := application.NewAppService(repos, memCache)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if tokInfo, err := appService.VerifyToken(ctx, false); err != nil {
		if logger != nil {
			logger.Warn("Token verification warning: %v", err)
		}
	} else if logger != nil {
		logger.Info("Token verified: ID=%s, Status=%s", tokInfo.ID, tokInfo.Status)
	}

	appConfig := app.Config{
		Service:        appService,
		Profile:        activeProfileName,
		InitialAccount: activeProfile.AccountID,
		InitialZone:    activeProfile.ZoneID,
	}

	appModel := app.New(appConfig)

	p := tea.NewProgram(
		appModel,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		if logger != nil {
			logger.Error("Program exited with error: %v", err)
		}
		fmt.Fprintf(os.Stderr, "Error running kumo: %v\n", err)
		os.Exit(1)
	}
}
