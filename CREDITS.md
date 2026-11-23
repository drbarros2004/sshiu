# Créditos

Arte e animações de terceiros, ou inspiradas em terceiros, só com licença permissiva ([ADR 0005](docs/adr/0005-arte-de-terceiros-so-com-licenca-permissiva.md)).

## Wallpapers do modo jardim

Todos foram escritos do zero em Go, em [`internal/tui/garden.go`](internal/tui/garden.go) e, os de pixel art, em [`internal/tui/garden_pixels.go`](internal/tui/garden_pixels.go) (os brinquedos do brainrot, em [`internal/tui/brainrot_toys.go`](internal/tui/brainrot_toys.go)). Nenhum código ou arte foi copiado; os projetos abaixo serviram de inspiração e ficam creditados.

| Wallpaper | Inspirado em | Licença |
|---|---|---|
| céu noturno | [ascii-splash](https://github.com/reowens/ascii-splash) (NightSky) e [bruh.animate](https://github.com/ethanlchristensen/bruh.animate) (twinkle) | MIT e Apache-2.0 |
| chuva lenta | [Terminal Rain](https://github.com/rmaake1/terminal-rain-lightning) | MIT |
| lâmpada de lava | [lavat](https://github.com/AngelJumbo/lavat) | MIT |
| café, árvore, estrada à noite, farol, gato dormindo, cachorro dormindo | desenhos nossos | — |
| a prensa hidráulica, o sabão e a aura do brainrot | desenhos nossos, na ideia dos vídeos de prensa e de corte de sabão | — |

## Vídeos do wallpaper brainrot

Trechos de 1 min, em 480p e sem áudio, em [`assets/videos/`](assets/videos/), cortados com o [yt-dlp](https://github.com/yt-dlp/yt-dlp) (`-f "bv*[height<=480][ext=mp4]" --download-sections`). Os quadros do wallpaper saem deles com o ffmpeg, em `go generate ./internal/tui` ([`brainrot_gen.go`](internal/tui/brainrot_gen.go)). O crédito do Minecraft parkour aparece embaixo do vídeo, no próprio wallpaper.

| Vídeo | Fonte | Trecho | Licença |
|---|---|---|---|
| `subway.mp4` | [Subway Surfers (2026) - Gameplay [4K 16x9] No Copyright](https://www.youtube.com/watch?v=i0M4ARe9v0Y), canal oficial Subway Surfers (SYBO) | 0:00–1:00 | "free to use - no copyright", na descrição |
| `minecraft.mp4` | [Minecraft Parkour Gameplay No Copyright](https://www.youtube.com/watch?v=3wsa4LTBUxo), canal Orbital - No Copyright Gameplay | 0:06–1:06 | Creative Commons Attribution (CC BY), com crédito ao canal |
