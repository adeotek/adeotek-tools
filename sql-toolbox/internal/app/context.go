package app

import (
	"context"

	"github.com/spf13/viper"
)

type viperKey struct{}

func setViper(ctx context.Context, v *viper.Viper) context.Context {
	return context.WithValue(ctx, viperKey{}, v)
}

func getViper(ctx context.Context) *viper.Viper {
	v, _ := ctx.Value(viperKey{}).(*viper.Viper)
	return v
}
