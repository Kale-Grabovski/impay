package di

import (
	"github.com/sarulabs/di"

	"github.com/Kale-Grabovski/impay/api"
	"github.com/Kale-Grabovski/impay/domain"
	"github.com/Kale-Grabovski/impay/kafka"
)

var ConfigApi = []di.Def{
	{
		Name:  "api.wallet",
		Scope: di.App,
		Build: func(ctx di.Container) (interface{}, error) {
			logger := ctx.Get("logger").(domain.Logger)
			producer := ctx.Get("kafka.producer").(*kafka.Producer)
			return api.NewWalletAction(producer, logger), nil
		},
	},
	{
		Name:  "api.stats",
		Scope: di.App,
		Build: func(ctx di.Container) (interface{}, error) {
			logger := ctx.Get("logger").(domain.Logger)
			consumer := ctx.Get("kafka.consumer").(*kafka.Consumer)
			return api.NewStatsAction(consumer, logger)
		},
		Close: func(obj interface{}) error {
			obj.(*api.StatsAction).CloseConsumers()
			return nil
		},
	},
}
