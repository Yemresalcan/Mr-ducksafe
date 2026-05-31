package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/yemresalcan/mr-ducksafe/internal/config"
	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/nvd"
	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/server"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
	"github.com/yemresalcan/mr-ducksafe/internal/vuln"
)

var (
	flagServePort int
	flagServeBind string
)

// NewServeCmd `mr-ducksafe serve` komutu.
func NewServeCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "serve",
		Short: "Lokal web UI başlat",
		RunE:  runServe,
	}
	c.Flags().IntVar(&flagServePort, "port", 8080, "dinlenecek port")
	c.Flags().StringVar(&flagServeBind, "bind", "127.0.0.1",
		"bağlanılacak adres (0.0.0.0 ağa açar — uyarı: auth yok)")
	return c
}

func runServe(cmd *cobra.Command, _ []string) error {
	dbPath := flagDBPath
	if dbPath == "" {
		p, err := config.DBPath()
		if err != nil {
			return err
		}
		dbPath = p
	}
	st, err := store.OpenSQLite("file:" + dbPath + "?_journal=WAL&_fk=1")
	if err != nil {
		return fmt.Errorf("DB aç: %w", err)
	}
	defer st.Close()

	// process önce çöktüyse running kalmış scan'leri toparla
	if _, err := st.CleanupStaleRunning(cmd.Context(), 10*time.Minute); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), "stale cleanup:", err)
	}

	sc, err := scanner.New()
	if err != nil {
		return err
	}
	m, err := matcher.NewDefaultMatcher()
	if err != nil {
		return err
	}
	eng := scanengine.New(sc, m, vuln.NewCachedClient(nvd.NewClient()), st)

	srv, err := server.New(eng, st)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", flagServeBind, flagServePort)
	fmt.Fprintf(cmd.OutOrStdout(), "mr-ducksafe serve → http://%s\n", addr)
	if flagServeBind == "0.0.0.0" {
		fmt.Fprintln(cmd.ErrOrStderr(),
			"UYARI: 0.0.0.0 — bu makineyi ağa açar, auth yok.")
	}
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	return httpSrv.ListenAndServe()
}
