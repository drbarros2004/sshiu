# Roteiro cronometrado

> Status: roteiro preparado; **ainda não cronometrado em voz alta**. Só o ensaio diz se cabe em 3 min.
> Fonte única: as linhas `Fala:` dos cards do deck ([`deck/index.html`](deck/index.html)). Mudou lá, gere este arquivo de novo.

Orçamento: 177 s de slides (2:57) para um limite de 3:00, com 3 s de folga. Falado: 359 palavras, a ~140 palavras/min ≈ 154 s, incluindo a narração por cima do vídeo.

| # | Bloco | Início | Tempo | Palavras |
|---|---|---|---|---|
| 1 | Capa | 0:00 | 12 s | 24 |
| 2 | O problema | 0:12 | 24 s | 57 |
| 3 | A solução | 0:36 | 18 s | 42 |
| 4 | Demo | 0:54 | 45 s | 63 |
| 5 | Alternativas | 1:39 | 18 s | 40 |
| 6 | Modelo de negócio | 1:57 | 27 s | 58 |
| 7 | Onde estamos | 2:24 | 18 s | 38 |
| 8 | O pedido | 2:42 | 15 s | 37 |
| | **Total** | | **177 s** | **359** |

## 1. Capa · 0:00 · 12 s

> Quanto do seu dia foi foco de verdade? Para quem programa, a resposta costuma ser: bem menos do que parece. Nós somos o sshiu.

- Pausa: olhar para a banca antes de começar.
- Direção: a capa mostra os pets do time subindo a ladeira juntos; os mesmos quatro voltam radiantes no último slide.
- Gancho: 'E tem um número que explica por quê.'

## 2. O problema · 0:12 · 24 s

> Na tela, a atenção dura em média 47 segundos antes de trocar de janela.

> 68% das pessoas dizem que falta tempo de foco, e muita gente busca companhia: study with me teve 520 milhões de views.

> Só que os apps de foco são de celular, e quem programa vive no computador. Fizemos o sshiu para esse público.

- Conferido: 47 s = tempo médio numa tela antes de trocar, nas medições de Gloria Mark (UC Irvine, 2004–2020; livro Attention Span, 2023). Mede troca de tela, não 'capacidade de atenção'. 68% = Microsoft Work Trend Index 2023, 31 mil pessoas em 31 países. 520 mi = blog do YouTube, views em 2023 de vídeos com 'study with me' no título.
- Pergunta: 'e para dev, especificamente?' → só 10% das sessões de programação interrompidas voltam ao código em menos de 1 min (Parnin & Rugaber, 10 mil sessões), e metade dos devs perde mais de 10 h por semana com ineficiências (Atlassian 2025, 3.500 devs).
- Pergunta: 'mas o celular atrapalha mesmo?' → num estudo de 2017 (Ward et al.), só ter o celular na mesa já reduziu a memória de trabalho; as replicações são mistas, então usamos como indício, não como prova.
- Gancho: 'Então levamos a sala de estudo para onde o dev já está: o terminal.'

## 3. A solução · 0:36 · 18 s

> O sshiu é uma sala de estudo que mora no terminal. Você digita ssh sshiu, diz a próxima coisa concreta que vai fazer e foca junto com a sua turma.

> Sem chat, só reações. E o seu pet reflete o seu foco.

- Pergunta: 'por que sem chat?' → é uma biblioteca: o chat vira a distração que a gente quer tirar (ADR 0004). As reações dizem 'tô vendo' sem ninguém precisar falar.
- Gancho: 'Melhor do que explicar é mostrar.'

## 4. Demo · 0:54 · 45 s

> Esse é o app de verdade. Eu digito a próxima coisa que vou fazer, e a sala inteira já está focando junto.

> Enquanto eu foco, a tela descansa: só o relógio, a tarefa e uma vista.

> Terminando tarefas eu ganho moedas e visto o meu pet. E o foco da turma constrói a obra da sala.

- Direção: o vídeo começa sozinho e já tem legenda em cada corte. Fale pouco, uma frase por cena, e deixe o gancho (ssh sshiu) correr calado.
- Atenção: testar o volume na sala antes: a música fica bem abaixo da voz (volume do notebook em ~30%).
- Conferido: tomadas gravadas em 25 e 26/11/2026 no app real (commit 854000a), em modo demo, com bots e servidor local. Ainda não há servidor público. Montagem em trailer/ (Remotion, versão trailer2-flow), trilha 8-bit gerada com o earshot.
- Gancho: 'E é aí que entra o dinheiro, mas antes: quem mais está nesse jogo.'

## 5. Alternativas · 1:39 · 18 s

> As pessoas já pagam por foco: o Forest passou de 48 milhões de downloads, o Focusmate cobra até 12 dólares por mês.

> O Brasil já tem 10 milhões de desenvolvedores no GitHub, e nenhuma dessas alternativas está no terminal deles.

- Conferido: Forest 48 mi de downloads no Android (AppBrain, 2026); Focusmate Plus US$ 8–12/mês e Team a partir de US$ 50/mês (focusmate.com/business, 2026); Brasil 10,2 mi de devs no GitHub (Octoverse 2026, set/2026).
- A confirmar: abrir o Octoverse 2026 e a página de preços do Focusmate; os dois números vieram de resumos (ver pitch/source-checks.md).
- Pergunta: 'e se o Forest copiar?' → o Forest é celular e individual; a nossa vantagem é a sala da turma dentro do terminal, onde o celular não chega.
- Gancho: 'Como o sshiu ganha dinheiro.'

## 6. Modelo de negócio · 1:57 · 27 s

> Três fontes de receita. Uma: skins. Personagens, cores e chapéus exclusivos, pagos, sem nunca vender foco nem energia.

> Duas: freemium. O básico é grátis; o sshiu+, a 9,90 por mês, abre salas maiores, histórico completo e exportação.

> Três: empresas. Uma sala para o time, cobrada por assento, e no futuro a tarefa vem direto da issue do GitHub.

- Conferido: preços e cenários são hipóteses (pitch/revenue-assumptions.md). Base: 2.000 pessoas ativas, 4% no sshiu+, 5% comprando R$ 10/mês em skins e 5 empresas com 20 assentos = R$ 3.292/mês, uns R$ 40 mil/ano, sem impostos nem taxas.
- Pergunta: 'R$ 40 mil por ano é pouco' → é o cenário base do primeiro ano de piloto; o que escala é o B2B, por assento.
- Pergunta: 'empresa vai usar isso para vigiar funcionário?' → não: o sshiu só sabe o que a pessoa declara. Sem captura de tela, sem monitorar teclado; a tabela de foco só aparece se o dono da sala abrir.
- Pergunta: 'como isso entra no ágil?' → a tarefa já é a próxima coisa concreta, como quebrar a história em tarefas. Com as issues, a sala vira o quadro do sprint ao vivo, sem reunião de status.
- Gancho: 'E isso não é slide: o app já existe.'

## 7. Onde estamos · 2:24 · 18 s

> Em quatro semanas, quatro pessoas fizeram 273 commits, 70 pull requests e 174 testes. O app funciona hoje.

> Próximo passo: ssh sshiu no ar e um piloto com turmas. Depois, login pelo GitHub e as issues virando tarefas.

- Conferido: medido em 27/11/2026 no master (commit 06b18c4): git rev-list --count master (273 commits desde 28/10); 70 PRs mesclados (gh pr list --state merged); 174 funções de teste em Go. Se entrar commit antes da apresentação, o número sobe.
- Pergunta: 'e o custo para escalar?' → a sala manda ~90 KB/s por pessoa (medido em 30/10/2026); um servidor pequeno aguenta uma turma, e a passada de performance já está planejada.
- Gancho: 'O que a gente pede.'

## 8. O pedido · 2:42 · 15 s

> Pedimos R$ 30 mil por 10% para, em dois meses, colocar o sshiu no ar, pagar artistas para as skins e fazer o piloto com turmas e empresas.

> Foco de verdade, junto, sem sair do terminal. Obrigado.

- Conferido: R$ 300 mil pós-investimento, ~7× a receita base de 12 meses (hipótese). Uso em 2 meses: ~R$ 12 mil artistas, ~R$ 10 mil ajuda de custo, ~R$ 6 mil piloto, ~R$ 2 mil servidor e domínio.
- Pergunta: 'como chegaram nesse valuation?' → 'o dinheiro cobre exatamente os 2 meses até o piloto; 300 mil é umas 7 vezes a receita base de um ano, e o app já existe.'
- Pergunta: 'por que só 2 meses?' → 'o app já funciona; falta servidor público, skins e o piloto. Em 2 meses a gente volta com usuários reais.'
- Direção: os quatro pets da capa voltam radiantes: o foco em grupo deu certo.
