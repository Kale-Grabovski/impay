package di

import (
	"github.com/sarulabs/di"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Kale-Grabovski/impay/domain"
)

var ConfigCommon = []di.Def{
	{
		Name:  "logger",
		Scope: di.App,
		Build: func(ctx di.Container) (interface{}, error) {
			var conf = zap.NewProductionConfig()
			cfg := ctx.Get("config").(*domain.Config)
			err := conf.Level.UnmarshalText([]byte(cfg.LogLevel))
			if err != nil {
				return nil, err
			}
			conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			return conf.Build()
		},
		Close: func(obj interface{}) error {
			return obj.(*zap.Logger).Sync()
		},
	},
}
