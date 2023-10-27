package di

import (
	kafkaBase "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sarulabs/di"

	"github.com/Kale-Grabovski/impay/domain"
	"github.com/Kale-Grabovski/impay/kafka"
)

var ConfigService = []di.Def{
	{
		Name:  "kafka.producer",
		Scope: di.App,
		Build: func(ctx di.Container) (interface{}, error) {
			cfg := ctx.Get("config").(*domain.Config)
			logger := ctx.Get("logger").(domain.Logger)
			return kafka.NewProducer(cfg, logger)
		},
		Close: func(obj interface{}) error {
			obj.(*kafka.Producer).Close()
			return nil
		},
	},
	{
		Name:  "kafka.consumer",
		Scope: di.App,
		Build: func(ctx di.Container) (interface{}, error) {
			cfg := ctx.Get("config").(*domain.Config)
			logger := ctx.Get("logger").(domain.Logger)
			return kafka.NewConsumer(cfg, logger), nil
		},
		Close: func(obj interface{}) error {
			return obj.(*kafkaBase.Consumer).Close()
		},
	},
}
