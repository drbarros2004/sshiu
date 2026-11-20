# assets dos personagens e chapéus

Preparação dos arquivos descritos em Specifications de 00_random_ssh_trab.txt.
Esta pasta fica em assets/pets, dentro do repositório ssh-focus-app.
Os originais continuam fora do repositório, em 00_sprites e na pasta externa ssh/assets.
A atualização dos chapéus foi importada de ssh/assets, preservando nomes, IDs, ordens e PNGs.
O aplicativo carrega e valida esses catálogos ao iniciar. Os arquivos são embutidos no executável;
recompile e reinicie para ver alterações.

## Onde editar

- characters.json define personagens, ordem, variantes, caminhos e composição das camadas.
- moods.json define os IDs fixos dos seis humores e seus nomes padrão em pt e en.
- characters/XX_nome/moods.json permite sobrescrever nomes de humores daquele personagem.
- hats.json define chapéus, ordem, variantes e resolução.
- sprite-layout.json na raiz define o tamanho padrão de cada quadro, a ordem dos humores nas linhas e a quantidade de quadros por linha.
- characters/XX_nome/sprite-layout.json define hatY para cada quadro ativo.
- main-assets.json identifica as duas camadas compartilhadas de partículas.
- templates guarda os marcadores da fruta e do Claude para edição ou geração de PNGs no futuro.

Os caminhos dos sprites e dos catálogos de pets são relativos a esta pasta assets/pets.
artworkColors, em characters.json, aponta para ../artworks/colors.json, também relativo a esta pasta.
Os IDs e nomes de arquivos estão em inglês, em minúsculas.
displayName contém os nomes por idioma, por exemplo {"pt": "azul", "en": "blue"}.
Edite esses nomes em characters.json e hats.json. Os IDs identificam as opções independentemente do idioma.
Novos idiomas podem ser acrescentados como outras chaves no mesmo objeto.
O carregador exige um nome em en e usá-lo quando faltar a tradução escolhida.
O campo order define a ordem de personagens, chapéus e variantes dentro de cada personagem ou chapéu.
As ordens começam em 0 e são independentes em cada lista. A variante com order 0 é a opção padrão.
Os arrays também ficam nessa ordem para facilitar a leitura.
Cartola começa com normal, normal_detailed e black; cowboy começa com normal e normal_detailed.
As demais variantes continuam na ordem definida no catálogo externo.
Os personagens são fruit, bunny e claude. Os chapéus são top_hat, cowboy, xmas_hat,
cap, crown, little_hair, pirate e firefighter_helmet.

## Variantes dos chapéus

| ID | Nome em português | Variantes |
| --- | --- | --- |
| top_hat | cartola | 8 |
| cowboy | chapéu de cowboy | 7 |
| xmas_hat | gorro de natal | 6 |
| cap | boné | 17 |
| crown | coroa | 6 |
| little_hair | topete | 9 |
| pirate | chapéu de pirata | 7 |
| firefighter_helmet | capacete de bombeiro | 8 |

São 68 variantes. As variantes reference do boné e rgb da coroa foram mantidas como
opções porque estão explicitamente cadastradas no catálogo externo, com esses nomes.
Todos os chapéus usam uma imagem 16×6, com transparência binária (alfa 0 ou 255).

## Cor da amostra

displayColor, em cada variante de characters.json, é a cor do quadradinho mostrado ao lado
do nome da cor no perfil e dos indicadores compactos. É um hexadecimal editável, independente
de artworkColor (que referencia a cor da obra).
Os valores iniciais foram extraídos da cor opaca mais frequente de cada folha body.
Essa extração foi feita só na preparação dos arquivos; o aplicativo deve ler o valor do JSON,
sem calcular a cor predominante novamente durante a execução.

## Cor da obra

Cada variante em characters.json tem artworkColor com o ID da cor da sua peça na obra.
Os IDs disponíveis e seus nomes em pt e en ficam em ../artworks/colors.json.
Por exemplo, a variante normal do Claude usa "artworkColor": "coral".
As demais variantes usam o ID da própria cor; fruta e coelho usam os mesmos oito IDs.
Mudar displayName não muda essa ligação, pois ela usa IDs, não nomes de apresentação.
Um artworkColor ausente ou que não exista no catálogo gera erro ao iniciar.
A obra recebe a cor de referência desse catálogo, sem depender do nome da variante do personagem.

## Nomes dos humores

moodOrder, em sprite-layout.json, usa os IDs fixos de moods.json para identificar as linhas.
Para mudar um nome para todos os personagens, edite displayName do humor em moods.json.
Para mudar somente um personagem, abra o moods.json referenciado por moods dentro daquele personagem em characters.json.
Todos esses arquivos usam uma lista moods com entradas id e displayName.
O moods.json da raiz define os nomes padrão. O de cada personagem lista somente suas exceções.
Uma lista vazia, {"moods": []}, usa todos os nomes padrão.
Por exemplo, para o Bunny chamar radiant de saltitante em português e bouncy em inglês, use em
characters/01_bunny/moods.json:

```json
{
  "moods": [
    {
      "id": "radiant",
      "displayName": {
        "pt": "saltitante",
        "en": "bouncy"
      }
    }
  ]
}
```

Você pode preencher só pt ou só en. Remover a tradução do objeto faz o idioma voltar ao padrão.
Cada id deve existir no catálogo padrão e aparecer no máximo uma vez na lista do personagem.
Fruta e coelho usam os nomes padrão. Claude mantém as exceções já versionadas em
characters/02_claude/moods.json: bugado/bugged, não funciona/nonfunctional e funcional/functional.
Essas exceções foram preservadas na importação, mesmo com a lista vazia na pasta externa.
O nome em inglês de sleeping é asleep, seguindo a tradução atual do aplicativo.
A escolha procura o nome do personagem no idioma selecionado e então
o padrão desse idioma. Se ambos faltarem, deve procurar en do personagem e depois en do catálogo.
Os IDs continuam iguais quando se muda o texto de apresentação.

## Folhas e encaixe

As folhas atuais medem 64×96. Cada quadro mede 16×16.
As seis linhas são wilted, tired, normal, good, radiant e sleeping.
frameCounts é [2, 2, 2, 2, 2, 4]. Quadros vazios não precisam ser detectados.
Um sprite-layout.json pode sobrescrever frameWidth, frameHeight, moodOrder, frameCounts e frameDurationMs.
frameDurationMs informa a duração de cada quadro em milissegundos, na mesma ordem de moodOrder.
Os valores iniciais são [2000, 1500, 1000, 1000, 500, 1500], preservando as velocidades anteriores.
Quanto menor o valor, mais rápida a animação. O mínimo é 500 ms, que corresponde ao tick da tela. O arquivo local sprite-layout.json pode sobrescrever
essa lista, assim como os outros campos; uma lista local substitui a lista padrão inteira.
Neste conjunto, todos usam os mesmos valores e cada arquivo local contém somente hatY.
O campo spriteLayout na raiz de characters.json aponta para o layout geral.
O campo spriteLayout dentro de cada personagem aponta para seu layout específico.
Campos ausentes no personagem herdam o valor geral.
Campos presentes nos dois devem usar o valor do personagem, e campos locais como hatY devem ser acrescentados.
Uma lista local substitui a lista geral inteira. Campos desconhecidos devem gerar erro.

hatY conta de baixo para cima, de 1 a 16. Ele posiciona a linha inferior da caixa inteira do chapéu,
inclusive quando há pixels transparentes nessa linha. Cada linha de hatY tem um valor por quadro ativo.
Para um quadro de 16 pixels, a linha inferior do chapéu em coordenadas da imagem é 16 - hatY.
Como o chapéu tem 6 pixels, seu topo é 16 - hatY - 5. Um topo negativo indica que ele ultrapassa o quadro.
O alinhamento horizontal usa as 16 colunas completas do quadro e do chapéu.
Os traços dos TXT foram removidos do JSON, porque representam quadros inativos.

## Composição

composition.withoutHat e composition.withHat listam as camadas de baixo para cima.
As entradas são referências de camadas. Elas seguem os nomes das pastas e são resolvidas pelos catálogos.

| Referência | Arquivo escolhido |
| --- | --- |
| characters/body | spriteSheet da variante de cor do personagem escolhido |
| characters/other_assets/plant | otherAssets.plant da mesma variante |
| characters/other_assets/sparkles | otherAssets.sparkles da mesma variante |
| hats/selected | image da variante do chapéu escolhido em hats.json |
| main_assets/tired_particles | spriteSheet do ID tired_particles em main-assets.json |
| main_assets/sleep_particles | spriteSheet do ID sleep_particles em main-assets.json |

characters representa o personagem selecionado e sua variante, sem fixar sua pasta numérica ou sua cor.
selected representa a escolha atual de chapéu e variante. Não há uma pasta chamada selected.
body, plant e sparkles são pastas que agrupam as versões coloridas das camadas.
As partículas são IDs com caminhos de PNG explícitos em main-assets.json.
Cada referência resulta em uma imagem. A composição mantém a ordem definida mesmo ao trocar a variante.
Para acrescentar uma camada específica do personagem, cadastre seu caminho em otherAssets de suas variantes
e use characters/other_assets/ seguido de seu ID nas listas de composição em que ela deve aparecer.

- fruit sem chapéu usa body, plant, sparkles, tired_particles e sleep_particles.
- bunny sem chapéu usa body e sleep_particles.
- claude sem chapéu usa body, tired_particles e sleep_particles.
- Os três com chapéu usam body, hat e sleep_particles.

A ausência de chapéu é uma escolha do estado de personalização, não um chapéu cadastrado em hats.json.
O estado salva character, variant, hat e hatVariant. hat vazio significa sem chapéu.
Os catálogos só listam as opções; não há restrições por recompensas nesta versão.
Cada partícula já está na linha do humor correspondente, com transparência nas demais linhas.
Todas as camadas de personagem e partículas mantêm o mesmo tamanho e alinhamento.
Os PNGs de templates ficam fora dos catálogos de seleção.

## Verificações feitas

- 11 JSONs de pets e o catálogo de cores das obras lidos sem chaves duplicadas.
- IDs únicos, ordens sequenciais, nomes em pt/en e caminhos conferidos com a mesma capitalização dos arquivos.
- Os 115 PNGs do repositório correspondem byte a byte à pasta externa; 68 são chapéus.
- Dimensões conferidas: chapéus 16×6; personagens, camadas e templates 64×96. Alfa sempre 0 ou 255.
- 336 quadros ativos dos corpos têm desenho; os quadros inativos das folhas de seleção estão transparentes.
- As 42 posições de chapéu seguem frameCounts e o intervalo de 1 a 16.
- Composição, referências das camadas, humores e artworkColor resolvem os catálogos existentes.
- As definições mais recentes do repositório foram preservadas: composição com referências de pastas,
  ordem dos campos e cores de obras. A pasta externa ainda não contém essas mesmas definições.

Todos os chapéus usam cores da paleta xterm de referência. Duas cores já existentes na fruta
não pertencem exatamente à paleta: #FFE066 em characters/00_fruit/other_assets/plant/fruit_plant_green.png
e #F7F7F7 em characters/00_fruit/other_assets/sparkles/fruit_sparkles_pink.png.
Os PNGs foram preservados sem recoloração. Essas duas cores podem ser aproximadas pelo terminal.
As imagens em templates são marcadores de edição e não precisam representar uma paleta final.

O carregador em internal/pet/catalog.go verifica os catálogos, caminhos, dimensões, camadas,
quadros e posições, com erros de contexto ao iniciar. A seleção acontece no perfil (c) e no
editor de aparência da sala dev (ctrl+l).
