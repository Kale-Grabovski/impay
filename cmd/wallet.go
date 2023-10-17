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

var walletCmd = &cobra.Command{
	Use: "wallet",
	Run: func(cmd *cobra.Command, args []string) {
		runWalletApi()
	},
}

func init() {
	rootCmd.AddCommand(walletCmd)
}

func runWalletApi() {
	e := echo.New()
	walletApi := diContainer.Get("api.wallet").(*api.WalletAction)

	e.GET("/wallets", walletApi.GetAll)
	e.GET("/wallets/:id", walletApi.GetById)
	e.POST("/wallets", walletApi.Create)
	e.PUT("/wallets/:id", walletApi.Update)
	e.DELETE("/wallets/:id", walletApi.Delete)
	e.POST("/wallets/:id/deposit", walletApi.Deposit)
	e.POST("/wallets/:id/withdraw", walletApi.Withdraw)
	e.POST("/wallets/:id/transfer", walletApi.Transfer)

	go func() {
		cfg := diContainer.Get("config").(*domain.Config)
		err := e.Start(":" + cfg.WalletPort)
		if err != nil {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	diContainer.DeleteWithSubContainers()
}
