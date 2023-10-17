package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"

	"github.com/Kale-Grabovski/impay/api"
	"github.com/Kale-Grabovski/impay/domain"
)

var statsCmd = &cobra.Command{
	Use: "stats",
	Run: func(cmd *cobra.Command, args []string) {
		runStatsApi()
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

func runStatsApi() {
	e := echo.New()
	statsApi := diContainer.Get("api.stats").(*api.StatsAction)
	e.GET("/stats/wallets", statsApi.Get)

	go func() {
		cfg := diContainer.Get("config").(*domain.Config)
		err := e.Start(":" + cfg.StatsPort)
		if err != nil {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	diContainer.DeleteWithSubContainers()
}
