package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/api"
	"github.com/alexandrmotologa/feedbackloom/internal/cluster"
	"github.com/alexandrmotologa/feedbackloom/internal/config"
	"github.com/alexandrmotologa/feedbackloom/internal/crm"
	"github.com/alexandrmotologa/feedbackloom/internal/ingest"
	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/prd"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
	"github.com/alexandrmotologa/feedbackloom/web"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "serve":
		runServe(os.Args[2:])
	case "ingest":
		runIngest(os.Args[2:])
	case "sync-crm":
		runSyncCRM(os.Args[2:])
	case "cluster":
		runCluster(os.Args[2:])
	case "generate-prd":
		runGeneratePRD(os.Args[2:])
	case "export-notion":
		runExportNotion(os.Args[2:])
	case "stats", "status":
		runStats(os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`feedbackloom: Autonomous Customer Voice & ARR-Weighted PRD Synthesizer

Usage:
  feedbackloom <command> [flags]

Commands:
  serve          Start the HTTP REST API and embedded web studio
  ingest         Import feedback from CSV, JSON, or seed benchmark data
  sync-crm       Synchronize customer ARR accounts from Stripe, HubSpot, or CSV
  cluster        Run HDBSCAN semantic clustering and rank opportunities by WRS
  generate-prd   Synthesize a comprehensive PRD document for an opportunity
  export-notion  Export a generated PRD to Notion workspace
  stats          Print database statistics and ARR opportunity matrix
  help           Display this help message

Run 'feedbackloom <command> --help' for details on specific command flags.`)
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.String("port", "8080", "Port to listen on")
	dbPath := fs.String("db", "./feedbackloom.db", "Path to SQLite database")
	_ = fs.Parse(args)

	cfg := config.LoadConfig()
	if *port != "" {
		cfg.Port = *port
	}
	if *dbPath != "" {
		cfg.DBPath = *dbPath
	}

	s, err := store.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database at %s: %v", cfg.DBPath, err)
	}
	defer s.Close()

	webFS, err := web.GetFileSystem()
	if err != nil {
		log.Printf("Web assets notice: %v (serving API only)", err)
	}

	server := api.NewServer(cfg, s, webFS)

	addr := ":" + cfg.Port
	fmt.Printf("feedbackloom server running on http://localhost%s\n", addr)
	fmt.Printf("Database: %s\n", cfg.DBPath)
	fmt.Println("Press Ctrl+C to terminate.")

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func runIngest(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: feedbackloom ingest <csv|json|seed> [flags]")
		os.Exit(1)
	}

	sub := args[0]
	fs := flag.NewFlagSet("ingest "+sub, flag.ExitOnError)
	dbPath := fs.String("db", "./feedbackloom.db", "Path to SQLite database")
	filePath := fs.String("file", "", "Path to file to import")
	source := fs.String("source", "", "Source channel identifier (e.g. zendesk, intercom, gong, slack)")
	_ = fs.Parse(args[1:])

	s, err := store.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	resolver := crm.NewResolver(s)

	switch sub {
	case "seed":
		// Ensure accounts exist first
		_, _ = crm.SyncStripeCustomers(ctx, s, crm.StripeSyncOptions{APIKey: "mock"})
		count, err := ingest.SeedRealisticFeedback(ctx, s, resolver)
		if err != nil {
			log.Fatalf("Seed failed: %v", err)
		}
		fmt.Printf("Successfully seeded %d realistic multi-channel enterprise feedback items.\n", count)

	case "csv":
		if *filePath == "" {
			log.Fatalf("--file is required for csv ingestion")
		}
		if *source == "" {
			*source = models.SourceCSV
		}
		count, err := ingest.ImportFeedbackFromCSV(ctx, s, resolver, *filePath, *source)
		if err != nil {
			log.Fatalf("CSV import failed: %v", err)
		}
		fmt.Printf("Successfully imported %d feedback items from %s.\n", count, *filePath)

	case "json":
		if *filePath == "" {
			log.Fatalf("--file is required for json ingestion")
		}
		if *source == "" {
			*source = models.SourceAPI
		}
		count, err := ingest.ImportFeedbackFromJSON(ctx, s, resolver, *filePath, *source)
		if err != nil {
			log.Fatalf("JSON import failed: %v", err)
		}
		fmt.Printf("Successfully imported %d feedback items from %s.\n", count, *filePath)

	default:
		log.Fatalf("Unknown ingest format: %s. Use csv, json, or seed.", sub)
	}
}

func runSyncCRM(args []string) {
	fs := flag.NewFlagSet("sync-crm", flag.ExitOnError)
	dbPath := fs.String("db", "./feedbackloom.db", "Path to SQLite database")
	provider := fs.String("provider", "stripe", "CRM provider: stripe, hubspot, or csv")
	stripeKey := fs.String("stripe-key", "", "Stripe Secret API Key (or 'mock' for realistic seed)")
	hubspotKey := fs.String("hubspot-key", "", "HubSpot Private App Access Token")
	csvFile := fs.String("csv", "", "Path to customer accounts CSV file")
	_ = fs.Parse(args)

	s, err := store.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	var count int

	switch strings.ToLower(*provider) {
	case "stripe":
		key := *stripeKey
		if key == "" {
			key = os.Getenv("STRIPE_KEY")
		}
		if key == "" {
			key = "mock"
		}
		count, err = crm.SyncStripeCustomers(ctx, s, crm.StripeSyncOptions{APIKey: key})
		if err != nil {
			log.Fatalf("Stripe sync error: %v", err)
		}
		fmt.Printf("Successfully synced %d customer accounts from Stripe.\n", count)

	case "hubspot":
		key := *hubspotKey
		if key == "" {
			key = os.Getenv("HUBSPOT_KEY")
		}
		if key == "" {
			key = "mock"
		}
		count, err = crm.SyncHubSpotCompanies(ctx, s, crm.HubSpotSyncOptions{AccessToken: key})
		if err != nil {
			log.Fatalf("HubSpot sync error: %v", err)
		}
		fmt.Printf("Successfully synced %d customer accounts from HubSpot.\n", count)

	case "csv":
		if *csvFile == "" {
			log.Fatalf("--csv file path is required when provider is csv")
		}
		count, err = crm.ImportAccountsFromCSV(ctx, s, *csvFile)
		if err != nil {
			log.Fatalf("CSV account import error: %v", err)
		}
		fmt.Printf("Successfully imported %d customer accounts from CSV %s.\n", count, *csvFile)

	default:
		log.Fatalf("Unknown CRM provider %s. Use stripe, hubspot, or csv.", *provider)
	}
}

func runCluster(args []string) {
	fs := flag.NewFlagSet("cluster", flag.ExitOnError)
	dbPath := fs.String("db", "./feedbackloom.db", "Path to SQLite database")
	minClusterSize := fs.Int("min-cluster-size", 2, "Minimum cluster size")
	minPoints := fs.Int("min-points", 2, "Minimum points for core distance")
	_ = fs.Parse(args)

	s, err := store.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	resolver := crm.NewResolver(s)
	vec := cluster.NewLocalSemanticVectorizer(128)
	ext := cluster.NewFeedbackExtractor()
	pipeline := cluster.NewPipeline(s, vec, ext, resolver)

	opts := cluster.PipelineOptions{
		MinClusterSize: *minClusterSize,
		MinPoints:      *minPoints,
	}

	summary, err := pipeline.Run(ctx, opts)
	if err != nil {
		log.Fatalf("Clustering failed: %v", err)
	}

	fmt.Println("\n=== HDBSCAN Semantic Opportunity Clustering Completed ===")
	fmt.Printf("Total Feedback Items Processed : %d\n", summary.TotalProcessed)
	fmt.Printf("Synthesized Opportunities     : %d\n", summary.ClustersCreated)
	fmt.Printf("Unclustered Noise / Outliers   : %d\n", summary.NoiseItemCount)
	fmt.Printf("Total Demanded ARR at Stake    : $%.2f USD\n", summary.TotalDemandedARR)
	fmt.Printf("Execution Duration             : %d ms\n\n", summary.DurationMS)

	opps, err := s.ListOpportunities(ctx)
	if err == nil && len(opps) > 0 {
		fmt.Println("Top Ranked Opportunities by Weighted Revenue Score (WRS):")
		fmt.Println("---------------------------------------------------------------------------------------------------------")
		fmt.Printf("%-14s | %-12s | %-14s | %-12s | %s\n", "ID", "TOTAL ARR", "WEIGHTED WRS", "ACCOUNTS", "TITLE")
		fmt.Println("---------------------------------------------------------------------------------------------------------")
		for _, o := range opps {
			fmt.Printf("%-14s | $%-11.0f | %-14.2f | %-12d | %s\n", o.ID, o.TotalARRUSD, o.WeightedScore, o.AccountCount, o.Title)
		}
		fmt.Println("---------------------------------------------------------------------------------------------------------")
	}
}

func runGeneratePRD(args []string) {
	fs := flag.NewFlagSet("generate-prd", flag.ExitOnError)
	dbPath := fs.String("db", "./feedbackloom.db", "Path to SQLite database")
	oppID := fs.String("opportunity-id", "", "Target Opportunity ID (e.g. opp_87a2f1)")
	output := fs.String("output", "", "Optional output path to write Markdown PRD")
	_ = fs.Parse(args)

	s, err := store.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer s.Close()

	ctx := context.Background()

	// If no ID provided, pick the highest weighted opportunity
	targetID := *oppID
	if targetID == "" {
		opps, err := s.ListOpportunities(ctx)
		if err != nil || len(opps) == 0 {
			log.Fatalf("No opportunities found. Run 'feedbackloom cluster' first.")
		}
		targetID = opps[0].ID
		fmt.Printf("No opportunity-id specified; selecting top ARR opportunity: %s (%s)\n\n", targetID, opps[0].Title)
	}

	opp, err := s.GetOpportunityByID(ctx, targetID)
	if err != nil {
		log.Fatalf("Failed to find opportunity %s: %v", targetID, err)
	}

	quotes, err := s.ListEnrichedFeedbackForOpportunity(ctx, targetID)
	if err != nil {
		log.Fatalf("Failed to load feedback quotes: %v", err)
	}

	gen := prd.NewPRDGenerator()
	md, notionJSON, err := gen.GeneratePRD(ctx, opp, quotes)
	if err != nil {
		log.Fatalf("PRD generation error: %v", err)
	}

	_ = s.UpdateOpportunityPRD(ctx, targetID, md, notionJSON)

	if *output != "" {
		if err := os.WriteFile(*output, []byte(md), 0644); err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		fmt.Printf("Successfully generated and saved PRD to %s\n", *output)
	} else {
		fmt.Println(md)
	}
}

func runExportNotion(args []string) {
	fs := flag.NewFlagSet("export-notion", flag.ExitOnError)
	dbPath := fs.String("db", "./feedbackloom.db", "Path to SQLite database")
	oppID := fs.String("opportunity-id", "", "Target Opportunity ID")
	notionKey := fs.String("notion-key", "", "Notion Integration Token")
	pageID := fs.String("page-id", "page_default", "Parent Notion Page or Database ID")
	_ = fs.Parse(args)

	if *oppID == "" {
		log.Fatalf("--opportunity-id is required")
	}

	s, err := store.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	opp, err := s.GetOpportunityByID(ctx, *oppID)
	if err != nil {
		log.Fatalf("Opportunity not found: %v", err)
	}

	if opp.GeneratedPRDNotionJSON == "" {
		// Generate if not yet created
		quotes, _ := s.ListEnrichedFeedbackForOpportunity(ctx, *oppID)
		gen := prd.NewPRDGenerator()
		md, notionJSON, _ := gen.GeneratePRD(ctx, opp, quotes)
		_ = s.UpdateOpportunityPRD(ctx, *oppID, md, notionJSON)
		opp.GeneratedPRDNotionJSON = notionJSON
	}

	key := *notionKey
	if key == "" {
		key = os.Getenv("NOTION_KEY")
	}
	if key == "" {
		key = "mock"
	}

	url, err := prd.ExportToNotionAPI(ctx, opp.Title, opp.GeneratedPRDNotionJSON, key, *pageID)
	if err != nil {
		log.Fatalf("Notion export error: %v", err)
	}

	fmt.Printf("Successfully published PRD to Notion:\n%s\n", url)
}

func runStats(args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	dbPath := fs.String("db", "./feedbackloom.db", "Path to SQLite database")
	_ = fs.Parse(args)

	s, err := store.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	stats, err := s.GetDashboardStats(ctx)
	if err != nil {
		log.Fatalf("Failed to get stats: %v", err)
	}

	fmt.Println("\n=== feedbackloom System & Revenue Status ===")
	fmt.Printf("Registered Customer Accounts : %v\n", stats["total_accounts"])
	fmt.Printf("Total Portfolio ARR          : $%.2f USD\n", stats["portfolio_arr_usd"])
	fmt.Printf("Ingested Feedback Items      : %v\n", stats["total_feedback"])
	fmt.Printf("Synthesized Opportunities    : %v\n", stats["total_opportunities"])
	fmt.Printf("Total Demanded ARR at Stake  : $%.2f USD\n\n", stats["total_demanded_arr_usd"])
}
