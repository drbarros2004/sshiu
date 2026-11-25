# Dossiê do pitch: sshiu

> Formato: Shark Tank simulado na aula de empreendedorismo, 3 minutos no máximo, banca com investidor-anjo, CEO, founder e agilista. Pedido: investimento em troca de participação.
> Status: preparado em 25/11/2026. Ensaios ainda não realizados.

Números e fontes: [`source-checks.md`](source-checks.md). Receita: [`revenue-assumptions.md`](revenue-assumptions.md). Fala: [`script.md`](script.md).

## 1. Ficha da solução

- **Em duas frases:** o sshiu é uma sala de estudo que mora no terminal. Você digita `ssh sshiu`, diz a próxima coisa concreta que vai fazer e foca junto com a sua turma, enquanto o seu pet reflete o seu foco.
- **Exemplo:** a Ana abre o terminal, escreve "fechar a issue #106, 25 min" e a sala inteira vê que ela está focando nisso. Um colega reage com 🔥, ninguém manda mensagem, e quando ela marca "feito" o pet dela fica radiante.
- **Analogia:** uma biblioteca silenciosa para quem programa, com um tamagotchi que só cresce quando você foca.
- **O que não é:** não é chat, não é rede social e não vigia ninguém. O app só sabe o que a pessoa declara: sem câmera e sem capturar tela ou teclado.
- **Estágio:** app funcionando em Go (Wish + Bubble Tea), com salas, pet, loja com moedas, modo jardim, estatísticas e tabela de foco. Ainda não há servidor público nem usuários reais fora do time. Em termos de TRL, é um protótipo validado em ambiente de desenvolvimento.
- **Insight:** quem programa já passa o dia no terminal, e o `ssh` já vem instalado em todo computador. Os apps de foco disputam atenção no celular, onde a distração mora; o sshiu fica onde o trabalho já está.
- **Evidência:** 256 commits, 66 PRs e 152 testes em 4 semanas, com 4 pessoas (medição local). Não há evidência de demanda paga.
- **Lacunas:** nenhum usuário fora do time, conversão freemium não medida e nenhuma empresa consultada.

## 2. Dor e problema

- **Personagem:** estudante de computação ou dev júnior que precisa entregar (lista, issue, trabalho), quer focar e foca melhor acompanhado.
- **Obstáculo:** atenção picada e falta de companhia. A atenção fica em média 47 s numa tela antes de trocar (Gloria Mark, UC Irvine); 68% dos trabalhadores dizem não ter tempo de foco sem interrupção (Microsoft Work Trend Index 2023); para quem programa, só 10% das sessões interrompidas voltam ao código em menos de 1 minuto (Parnin & Rugaber, 2011), e metade dos devs perde mais de 10 h por semana com ineficiências (Atlassian 2025).
- **Vontade de focar junto:** vídeos de "study with me" tiveram mais de 520 milhões de visualizações em 2023 (YouTube), e o maior Discord de estudo tem ~1 milhão de membros.
- **Lugar errado:** os apps de foco populares (Forest e afins) são de celular, e só ter o celular por perto já pode tirar atenção (Ward et al., 2017, com replicações mistas). Quem programa passa o dia no computador: foi para esse público que fizemos o sshiu.
- **Alternativas falhas:** app de celular (o celular é a distração), Discord de estudo (chat aberto, multidão), Focusmate (câmera, desconhecido).
- **Quem paga x quem usa:** no B2C, quem usa paga (skins, sshiu+). No B2B paga o líder ou a empresa, e usa o time, por isso a regra de não vigiar é parte do produto.
- **Teste de dor:** frequente (todo dia) e reconhecida (68% dizem que falta foco; 520 mi de views de "study with me"; 1 milhão de pessoas num Discord só para estudar junto). Urgência e custo são fracos no B2C, onde é "bom de ter". No B2B a dor fica mais cara: tempo do time e reunião de status.

## 3. Mercado e oportunidade

Sem TAM inflado: o pitch não mostra tamanho de mercado em dólares, porque não há recorte oficial confiável para "app de foco no terminal". Mostra quem já paga por foco e quantos devs existem no Brasil.

- **Quem já paga por foco:** Forest (48 mi de downloads), Focusmate (Plus US$ 8–12/mês; Team a partir de US$ 50/mês), Finch (pet de autocuidado, Plus US$ 5,99–9,99/mês).
- **Público alcançável:** 10,2 mi de devs no GitHub no Brasil (Octoverse 2026, a conferir).
- **Why now:** o Brasil é um dos países onde mais cresce o número de devs; apps de terminal viraram produto (o terminal.shop esgotou café vendido por SSH em dias); e o trabalho de dev já gira em torno de issues e PRs, que viram tarefas sem formulário novo.

| Critério | Nota | Por quê |
|---|---|---|
| Tamanho | 2 | Nicho de devs; cresce com o B2B |
| Crescimento | 4 | Base de devs no Brasil crescendo rápido |
| Dor | 3 | Frequente e reconhecida; pouco urgente no B2C |
| Disposição a pagar | 3 | Há quem pague por foco e por pet; dev paga por ferramenta |
| Acesso | 4 | Turmas da Poli como piloto; o convite é uma linha de texto |
| Timing | 3 | Específico, mas não regulatório |
| Defensabilidade | 2 | Código replicável; a defesa é comunidade, arte e integração |
| Viabilidade | 4 | O app já funciona; falta servidor público |

## 4. Concorrência (alternativas)

| Alternativa | Como funciona | Preço | Faz bem | Onde falha para nós |
|---|---|---|---|---|
| Forest | Árvore cresce se você não mexe no celular | App + compras | Gamificação simples | Celular e individual |
| Focusmate | Sessão por vídeo com um desconhecido | US$ 8–12/mês; times US$ 50+ | Compromisso social forte | Câmera, desconhecido, fora do terminal |
| Discord de estudo | Canais de voz e chat | Grátis | Comunidade grande | Chat é distração; multidão, não a sua turma |
| Pomodoro no terminal | Timer de linha de comando | Grátis | Já está no terminal | Sozinho, sem sala nem motivo para voltar |
| Não fazer nada | Focar por conta própria | — | Zero custo | É o problema |

Atributos únicos → valor: no terminal, sem instalar (começa em segundos); sala da própria turma ou time, sem chat (companhia sem distração); pet e moedas ganhos focando (motivo para voltar amanhã); tarefa declarada (próxima coisa concreta, que no futuro vem da issue).

## 5. Posicionamento

> Para estudantes de computação e times de desenvolvimento que querem focar juntos sem abrir mais uma distração, o sshiu é uma sala de foco no terminal que mostra quem está focando e em quê, sem chat e sem câmera, com um pet que cresce com o seu foco.

Old game → new game: **focar sozinho no celular → focar junto no terminal.**

## 6. Story spine

Era uma vez uma turma de computação que precisava entregar listas e projetos. Todos os dias, cada um abria o notebook para estudar e, dez minutos depois, estava no celular. Até que um dia a turma percebeu que focava melhor quando estava junto, mas o Discord virava conversa e o app de foco estava no celular. Por causa disso, levaram a sala de estudo para o terminal, onde já trabalhavam. Por causa disso, cada um passou a dizer a próxima coisa concreta e a ver os colegas focando, com um pet que cresce junto. Até que finalmente o foco virou hábito da turma. **E desde então**, foco de verdade é algo que se faz junto, sem sair do terminal.

## 7. Banco de perguntas (dado no bolso)

| Pergunta provável | Resposta curta | Dado no bolso |
|---|---|---|
| Isso já não existe? | Foco existe; foco **junto, no terminal, sem chat** não. O par recortado é o novo. | Tabela de alternativas |
| Por que vocês? | Somos o usuário: estudantes de computação que fizeram o app para a própria turma e entregaram 256 commits em 4 semanas. | 66 PRs, 152 testes, 12 ADRs |
| Empresa vai vigiar funcionário? (CEO) | Não. O sshiu só sabe o que a pessoa declara; a tabela de foco só aparece se o dono da sala abrir. | ADR 0010 |
| Como entra no ágil? (agilista) | A tarefa já é "a próxima coisa concreta", como quebrar a história em tarefas. Com issues, a sala vira o quadro do sprint ao vivo, sem reunião de status. | Login pelo GitHub já desenhado em `IDEIAS.md` |
| Skin paga não estraga o jogo? | Só visual: as moedas continuam vindo do foco e dinheiro nunca compra energia. | `docs/regras.md`, moedas e loja |
| R$ 40 mil/ano é pouco | É o cenário base do primeiro ano; o que escala é o B2B por assento. | Empresas = 46% da receita base |
| Por que terminal e não web? | Começamos pelo público que já está lá (custo de aquisição baixo, convite é uma linha). Web é a segunda onda. | `ssh -t sshiu <código>` |
| Quanto custa escalar? | A sala manda ~90 KB/s por pessoa; um servidor pequeno aguenta uma turma, e a passada de performance já está planejada. | Medição de 30/10/2026 |
| E se o Forest copiar? | O Forest é celular e individual; copiar a sala no terminal é mudar de produto. A defesa é comunidade, arte e a integração com issues. | — |
| Propriedade da arte? | Arte própria ou com licença permissiva (ADR 0005). Personagem pago só com arte nossa. | `CREDITS.md` |
| Como chegaram no valuation? | [resposta do time] | — |
| E para dev, especificamente? | Só 10% das sessões interrompidas voltam ao código em menos de 1 min, e metade dos devs perde mais de 10 h por semana com ineficiências. | Parnin & Rugaber 2011; Atlassian 2025 |
| O celular atrapalha mesmo? | Há indício: só ter o celular na mesa reduziu a memória de trabalho num estudo de 2017, mas as replicações são mistas. O ponto principal é que dev trabalha no computador. | Ward et al. 2017 |
| Por que agora? | Devs no Brasil crescendo rápido, e dev já compra pelo terminal (terminal.shop). | Octoverse 2026; blog da Charm |

## 8. Lacunas [A CONFIRMAR]

- O pedido (valor e participação) e a justificativa do valuation.
- A data da apresentação, na capa.
- O Octoverse 2026 e o preço do Focusmate Plus na página original.
- Uma conversa com 2–3 colegas ou um tech lead antes da aula seria a primeira evidência de demanda; hoje não há nenhuma.
- O personagem "claude" na loja: tirar da venda paga ou pedir autorização.
