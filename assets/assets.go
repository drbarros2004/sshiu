// Package assets guarda os catálogos e imagens distribuídos com o aplicativo.
package assets

import "embed"

// Pets contém somente os arquivos usados pelos catálogos; templates são auxiliares de edição.
//
//go:embed pets/*.json pets/characters pets/hats pets/main_assets artworks/colors.json coins shop
var Pets embed.FS
