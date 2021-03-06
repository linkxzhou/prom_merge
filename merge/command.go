package cmd

import (
	"fmt"
	"net/http"
	httpurl "net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var config *Config

func NewRootCommand() *cobra.Command {
	app := new(App)

	cmd := &cobra.Command{
		Use:   "prom_merge",
		Short: "Merge Prometheus metrics from multiple sources",
		Run:   app.run,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if app.viper.GetBool("verbose") {
				log.SetLevel(log.DebugLevel)
			} else {
				log.SetLevel(log.InfoLevel)
			}
		},
	}

	app.Bind(cmd)

	return cmd
}

type App struct {
	viper *viper.Viper
}

func (app *App) Bind(cmd *cobra.Command) {
	app.viper = viper.New()
	app.viper.SetEnvPrefix("MERGER")
	app.viper.AutomaticEnv()

	configPath := cmd.PersistentFlags().StringP(
		"config-path", "c", "",
		"Path to the configuration file.")

	cobra.OnInitialize(func() {
		var err error
		if configPath != nil && *configPath != "" {
			config, err = ReadConfig(*configPath)
			if err != nil {
				log.WithField("error", err).Errorf("failed to load config file '%s'", *configPath)
				os.Exit(1)
				return
			}
			urls := []string{}
			for _, e := range config.Exporters {
				urls = append(urls, e.URL)
			}
			app.viper.SetDefault("urls", strings.Join(urls, " "))
		}
	})

	cmd.PersistentFlags().Int(
		"listen-port", 8080,
		"Listen port for the HTTP server. (ENV:MERGER_PORT)")
	app.viper.BindPFlag("port", cmd.PersistentFlags().Lookup("listen-port"))

	cmd.PersistentFlags().Int(
		"exporters-timeout", 10,
		"HTTP client timeout for connecting to exporters. (ENV:MERGER_EXPORTERSTIMEOUT)")
	app.viper.BindPFlag("timeout", cmd.PersistentFlags().Lookup("exporters-timeout"))

	cmd.PersistentFlags().BoolP(
		"verbose", "v", false,
		"Include debug messages to output. (ENV:MERGER_VERBOSE)")
	app.viper.BindPFlag("verbose", cmd.PersistentFlags().Lookup("verbose"))

	cmd.PersistentFlags().StringSlice(
		"url", nil,
		"URL to scrape, Can be speficied multiple times. (ENV:MERGER_URLS,space-seperated)")
	app.viper.BindPFlag("urls", cmd.PersistentFlags().Lookup("url"))
}

func (app *App) run(cmd *cobra.Command, args []string) {
	hostList := make(map[string]string, 0)
	for _, url := range app.viper.GetStringSlice("urls") {
		if u, err := httpurl.Parse(url); err == nil {
			hostList[url] = u.Host
		}
	}
	hostAlias := "instance"
	if config != nil && len(config.HostAlias) > 0 {
		hostAlias = config.HostAlias
	}
	http.Handle("/metrics", Handler{
		Exporters:            app.viper.GetStringSlice("urls"),
		ExportersHTTPTimeout: app.viper.GetInt("timeout"),
		ExportersHostList:    hostList,
		ExportersHostAlias:   hostAlias,
	})

	port := app.viper.GetInt("port")
	log.Infof("Starting HTTP server on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
