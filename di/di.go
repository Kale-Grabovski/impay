package di

import "github.com/sarulabs/di"

var Config []di.Def

func init() {
	diConfigs := [][]di.Def{
		ConfigCommon,
		ConfigService,
		ConfigApi,
	}

	for _, c := range diConfigs {
		Config = append(Config, c...)
	}
}
