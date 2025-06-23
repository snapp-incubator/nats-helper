package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/snapp-incubator/nats-helper/internal/eventexporter"
	"github.com/snapp-incubator/nats-helper/internal/metricexporter"
	"github.com/snapp-incubator/nats-helper/pkg/config"
	natsclient "github.com/snapp-incubator/nats-helper/pkg/nats"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	metricsAddr string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export NATS stream and consumer events",
	Long: `Export NATS stream and consumer events to various outputs.
Currently supports exporting events as Prometheus metrics.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Create NATS client
		client, err := natsclient.New(&cfg.NATS)
		if err != nil {
			return fmt.Errorf("failed to create NATS client: %w", err)
		}
		defer client.Close()

		// Create event exporter
		eventExporter := eventexporter.New(client)
		if err := eventExporter.Start(); err != nil {
			return fmt.Errorf("failed to start event exporter: %w", err)
		}
		defer eventExporter.Stop()

		// Create metric exporter
		metricExporter := metricexporter.New()
		if err := metricExporter.Start(); err != nil {
			return fmt.Errorf("failed to start metric exporter: %w", err)
		}
		defer metricExporter.Stop()

		// Start metrics HTTP server
		http.Handle("/metrics", promhttp.Handler())
		go func() {
			fmt.Println("Running metric exporter on", metricsAddr)
			if err := http.ListenAndServe(metricsAddr, nil); err != nil {
				fmt.Printf("Error starting metrics server: %v\n", err)
			}
		}()

		// Forward events from event exporter to metric exporter
		go func() {
			fmt.Println("Forwarding events from event exporter to metric exporter")
			for event := range eventExporter.Events() {
				select {
				case metricExporter.Events() <- event:
				default:
					// Channel is full, drop event
				}
			}
		}()

		// Wait for interrupt signal
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		<-ctx.Done()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Add configuration file flag
	exportCmd.Flags().StringVar(&cfgFile, "config", "", "path to configuration file")
	viper.BindPFlag("config", exportCmd.Flags().Lookup("config"))

	// Add metrics address flag
	exportCmd.Flags().StringVar(&metricsAddr, "metrics-addr", ":9090", "address to expose metrics on")
	viper.BindPFlag("metrics-addr", exportCmd.Flags().Lookup("metrics-addr"))
}
