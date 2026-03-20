package main

import (
	"context"
	"fmt"
	"os"

	"github.com/lucasassuncao/gopaper/internal/cmd"
	"github.com/lucasassuncao/gopaper/internal/models"

	"github.com/spf13/viper"
)

// version is set at build time via ldflags: -X main.version=<value>
var version = "dev"

func main() {
	m := &models.Gopaper{
		Viper:      viper.GetViper(),
		Logger:     nil,
		Categories: make([]*models.Categories, 0),
	}

	root := cmd.RootCmd(m, version)

	err := root.ExecuteContext(context.Background())
	if err != nil {
		fmt.Printf("failed to run the app. %v\n", err)
		os.Exit(1)
	}
}
