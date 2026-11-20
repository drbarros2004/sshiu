# cores das obras

colors.json é o catálogo compartilhado das cores das peças das obras.
Cada entrada tem id, displayName por idioma, order e color em hexadecimal.
artworkColor, nas variantes de ../pets/characters.json, referencia um desses IDs.
As ordens começam em 0 e preservam a ordem atual das cores do aplicativo.

color preserva a cor de referência usada hoje pelo aplicativo para obter o matiz da peça.
Não é a cor final de todos os pixels: a recoloração mantém o brilho e a saturação dos
marcadores do desenho. O terminal ainda pode converter o resultado para 256 cores.
Esses valores foram mantidos como estão para preservar a aparência na integração futura.
As cores fixas de chão, detalhes e peças compartilhadas continuam definidas no aplicativo.

Na orquestra, a cor também escolhe o instrumento conforme a ordem atual do catálogo:
blue: violino, green: violão, yellow: trompete, coral: tambor,
purple: violoncelo, turquoise: harpa, orange: saxofone, pink: piano.
O carregador futuro deve preservar essa correspondência ao substituir a seleção atual.

Esta pasta prepara a configuração; o aplicativo ainda usa internal/pet/palettes.json,
internal/pet/artwork_colors.go e internal/tui/obra.go para renderizar as obras.
