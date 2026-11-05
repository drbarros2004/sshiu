# Paletas do pet por parte

As cores do pet eram calculadas a partir de um tom e variavam conforme o humor. Isso dificultava
reproduzir desenhos feitos no editor de sprites e escolher, por exemplo, a cor das bochechas
independentemente do corpo.

## Decisão

O PNG passa a ser um template: cada cor opaca representa uma parte nomeada. Um JSON associa
esses marcadores a nomes e define uma cor padrão para cada parte em `defaults`. Cada paleta lista apenas suas diferenças
em `colors`.
As cores são fixas entre humores; energia e foco continuam escolhendo o desenho e a animação.

Um marcador desconhecido ou uma parte sem cor padrão é erro ao iniciar, com indicação do que
falta cadastrar. Não criamos grupos automaticamente. Partes distintas podem terminar com a
mesma cor sem perder sua identidade no template.

HEX é o formato principal e `xterm:N` permite indicar explicitamente um índice de 0 a 255.
Os IDs e a ordem das oito escolhas existentes são preservados, sem migração de banco.
A obra mantém sua recoloração anterior, separada das paletas do pet.

## Consequências

Editar cores ou acrescentar uma parte não exige escrever Go: o trabalho fica no template e
no JSON. Tamanho e organização dos quadros continuam definidos no renderizador.
As oito paletas têm referências visuais exportadas do editor. As cores do PNG são preservadas
no JSON, incluindo as três aproximações conhecidas em terminais de 256 cores.
O passo a passo está em [pets-paletas.md](../pets-paletas.md).
