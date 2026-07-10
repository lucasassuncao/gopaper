package cmd

import (
	"github.com/lucasassuncao/gopaper/internal/models"
	"github.com/lucasassuncao/yedit/editor"
	"github.com/lucasassuncao/yedit/metadata"
)

func buildGopaperHints() (editor.MetadataSource, error) {
	return metadata.New(models.Config{})
}
