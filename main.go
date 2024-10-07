package main

import (
	"context"
	"log"
	"net/http"
	"os"

	_ "embed"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

var (
	version string
	commit  string
	date    string
	builtBy string

	//go:embed VERSION
	fallbackVersion string
)

func main() {
	var (
		ctx                 = context.Background()
		app                 = kingpin.New("postfix_exporter", "Prometheus metrics exporter for postfix")
		versionFlag         = app.Flag("version", "Print version information").Bool()
		toolkitFlags        = kingpinflag.AddFlags(app, ":9154")
		metricsPath         = app.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		postfixShowqPath    = app.Flag("postfix.showq_path", "Path at which Postfix places its showq socket.").Default("/var/spool/postfix/public/showq").String()
		logUnsupportedLines = app.Flag("log.unsupported", "Log all unsupported lines.").Bool()
	)

	InitLogSourceFactories(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if version == "" {
		version = fallbackVersion
	}

	if *versionFlag {
		os.Stdout.WriteString(version)
		os.Exit(0)
	}
	versionString := "postfix_exporter " + version
	if commit != "" {
		versionString += " (" + commit + ")"
	}
	if date != "" {
		versionString += " built on " + date
	}
	if builtBy != "" {
		versionString += " by: " + builtBy
	}
	log.Print(versionString)

	logSrc, err := NewLogSourceFromFactories(ctx)
	if err != nil {
		log.Fatalf("Error opening log source: %s", err)
	}
	defer logSrc.Close()

	exporter, err := NewPostfixExporter(
		*postfixShowqPath,
		logSrc,
		*logUnsupportedLines,
	)
	if err != nil {
		log.Fatalf("Failed to create PostfixExporter: %s", err)
	}
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	lc := web.LandingConfig{
		Name:        "Postfix Exporter",
		Description: "Prometheus exporter for postfix metrics",
		Version:     versionString,
		Links: []web.LandingLinks{
			{
				Address: *metricsPath,
				Text:    "Metrics",
			},
		},
	}
	lp, err := web.NewLandingPage(lc)
	if err != nil {
		log.Fatalf("Failed to create landing page: %s", err)
	}
	http.Handle("/", lp)

	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()
	go exporter.StartMetricCollection(ctx)

	server := &http.Server{}
	logger := promslog.New(&promslog.Config{})
	log.Fatal(web.ListenAndServe(server, toolkitFlags, logger))
}
