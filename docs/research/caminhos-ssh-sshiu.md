# Caminhos para a turma entrar com `ssh sshiu`

> **É uma sugestão, não uma decisão.** Esta pesquisa serve de ponto de partida: a decisão de onde hospedar fica para a issue [#25](https://github.com/drbarros2004/ssh-focus-app/issues/25). Preços e planos mudam rápido: quem for contratar deve conferir de novo as páginas citadas.

Pesquisa da issue [#24](https://github.com/drbarros2004/ssh-focus-app/issues/24), feita em 02/11/2026. Cada preço, limite e comportamento foi conferido na página oficial do provedor, na doc oficial ou na man page, na data acima. Quando a página não carregou ou não diz, está escrito "não conferido". Os reais usam a PTAX de fechamento do Banco Central de 30/10/2026 (02/11 é feriado): **US$ 1 = R$ 5,42** e **€ 1 = R$ 6,05** ([API PTAX do BCB](https://olinda.bcb.gov.br/olinda/servico/PTAX/versao/v1/odata/)).

O teste com outro computador explica o ponto de partida: na mesma Wi-Fi funcionou porque o IP da máquina na rede local é alcançável ali; nos dados móveis, não, porque esse IP é privado. Para o MVP, a turma decidiu ficar na rede local: o servidor roda no notebook de alguém do grupo e todo mundo entra pela mesma Wi-Fi ([a seção da rede local](#rede-local-o-caminho-do-mvp)). Os outros caminhos dão ao servidor um endereço que se alcança de qualquer rede e ficam para depois.

## Resumo

- **No MVP, na rede local:** `ssh sshiu` sai com um bloco `Host sshiu` de três linhas apontando para `sshiu.local`, colado uma vez por cada colega. Sem colar nada, o comando mais bonito é `ssh sshiu.local`, com o notebook chamado `sshiu` e o servidor na porta 22. O maior risco da demo é a rede: Wi-Fi de faculdade costuma isolar um aparelho do outro, então o mais seguro é um roteador ou o hotspot de um celular.

- **O nome `sshiu` sozinho não existe no DNS público.** Sem domínio, `ssh sshiu` só sai de um bloco `Host sshiu` no `~/.ssh/config` de cada colega (uma linha para colar, abaixo), do MagicDNS do Tailscale (só para quem está **dentro** da tailnet, e o plano grátis tem até 6 pessoas) ou do arquivo hosts. Máquina do Tailscale **compartilhada** com quem é de fora só responde pelo nome completo (`sshiu.<tailnet>.ts.net`).
- **A banda decide o resto.** 30 pessoas × 2 h por dia dão **~583 GB/mês** numa sala parada (90 KB/s) e **~2,5 TB/mês** se todo mundo ficar no brainrot (380 KB/s). Por pessoa: 0,72 Mbit/s e 3 Mbit/s de upload de quem hospeda.
- **Fora da rede local, VPS em São Paulo é o caminho mais simples para a turma:** nada para instalar além do `ssh`, porta 22, fica no ar com o PC desligado. A Oracle Always Free tem duas regiões no Brasil e 10 TB/mês de saída de graça, mas pede cartão, pode dar "out of host capacity" e recolhe instância ociosa. A Vultr tem São Paulo por US$ 5 (~R$ 27/mês) com 1 TB.
- **Tailscale** resolve CGNAT e dados móveis sem abrir porta, mas cada colega instala o app, cria conta e aceita o compartilhamento, e o PC de quem hospeda fica ligado.
- **Túneis TCP públicos** (bore.pub, playit, pinggy, ngrok) dão endereço com porta aleatória, e o grátis tem limites que não servem para a turma (ngrok: 1 GB; pinggy: 60 min).
- **PC de casa com port forwarding** esbarra no CGNAT, comum nas operadoras, e no upload doméstico.
- **Achado no código:** o convite e a mensagem de "crie uma chave" montam o endereço a partir de `s.LocalAddr()` ([main.go:146](../../main.go#L146), [main.go:70](../../main.go#L70)). Atrás de um túnel, de um redirecionamento de porta ou numa VPS com IP público por NAT (Oracle, AWS, GCP), o convite mostra um endereço interno, e não `ssh -t sshiu <código>`.
- **Chave do host:** em todos os caminhos com TCP puro, o cliente vê a chave do Wish. O risco é outro: se `.ssh/id_ed25519` some ou o servidor roda de outra pasta, o Wish **gera uma chave nova sem avisar** ([wish v2.0.5, options.go:69-78](https://github.com/charmbracelet/wish/blob/v2.0.5/options.go#L69-L78)), e aí a turma inteira vê "REMOTE HOST IDENTIFICATION HAS CHANGED".

## Comparação

| Caminho | Custo | Passos de quem hospeda | Passos de cada colega | Banda | No ar com o PC desligado? | O que dá errado |
|---|---|---|---|---|---|---|
| **Rede local** (notebook do grupo, o caminho do MVP) | R$ 0 | chamar o notebook de `sshiu`; rodar na porta 22; liberar no firewall | entrar na mesma Wi-Fi; colar o bloco `Host sshiu` (ou usar `ssh sshiu.local`) | a Wi-Fi, de sobra para a turma | não | rede que isola aparelhos (comum em faculdade); `.local` não testado com o notebook em Windows; só funciona na mesma rede; convite mostra o IP local |
| **VPS Oracle Always Free** (São Paulo ou Vinhedo) | R$ 0 | criar conta com cartão; escolher região de origem no Brasil; criar a VM; abrir a porta na security list e no firewall da imagem; systemd com a porta 22 | colar o bloco `Host sshiu` (ou usar o IP) | 10 TB/mês de saída grátis; E2.1.Micro tem 50 Mbps | sim | "out of host capacity" na região; cartão pré-pago/virtual recusado; instância ociosa por 7 dias pode ser recolhida; conta parada por 30 dias pode ser encerrada; convite mostra o IP privado |
| **VPS paga no Brasil** (Vultr São Paulo) | US$ 5/mês (~R$ 27) | criar conta; criar a VM; systemd com a porta 22 | colar o bloco `Host sshiu` (ou usar o IP) | 1 TB/mês; cobre a sala parada, não 30 pessoas no brainrot | sim | excedente de tráfego (preço não conferido); custo mensal |
| **VPS paga fora** (Hetzner, DigitalOcean) | Hetzner € 5,99/mês (~R$ 36); DO US$ 4/mês (~R$ 22) | igual à anterior | igual | Hetzner 20 TB; DO 500 GiB (+US$ 0,01/GiB) | sim | latência de pelo menos 77 a 113 ms só de distância; Hetzner sem estoque no dia da consulta |
| **Tailscale** (servidor no PC de casa) | R$ 0 (até 6 pessoas na tailnet) | instalar Tailscale; nomear a máquina `sshiu`; rodar com `-host ""`; compartilhar a máquina por link | instalar Tailscale, criar conta, aceitar o link; quem é de fora usa o nome completo | upload de casa (ou relé DERP, mais lento) | não | 30 pessoas passa de 6: o resto entra por compartilhamento, sem nome curto; PC ligado; convite mostra o IP 100.x |
| **Túnel TCP público** (bore.pub, playit, pinggy, ngrok) | R$ 0 a US$ 10/mês | rodar o cliente do túnel ao lado do servidor | colar um `Host sshiu` com host e porta do túnel | upload de casa, mais o limite do túnel | não | porta aleatória que muda ao reiniciar; ngrok grátis tem 1 GB e pede cartão; pinggy grátis cai em 60 min; bore.pub sem garantia; convite mostra 127.0.0.1 |
| **Serveo com alias** (`ssh -J`) | grátis (limites do alias não informados) | `ssh -R <alias>:22:localhost:23234 serveo.net` | `Host sshiu` com `ProxyJump serveo.net` | não informado | não | serviço de terceiro sem SLA; o alias não é controle de acesso; não testado |
| **Cloudflare Tunnel** | grátis (plano não conferido) + domínio | domínio na Cloudflare; `cloudflared` no servidor; Access | instalar `cloudflared` e usar como `ProxyCommand`, ou terminal no navegador | não conferido | não | cada colega instala programa; o terminal no navegador quebra "a chave é a identidade" |
| **PC de casa com port forwarding** (+ DDNS) | R$ 0 (IP público pode ser pago na operadora) | conferir CGNAT; redirecionar a porta no roteador; liberar no firewall do Windows; DDNS | colar o bloco `Host sshiu` | upload de casa: ~55 pessoas na sala ou ~13 no brainrot com 50 Mbps | não | CGNAT impede tudo; IP muda; PC ligado; rede da faculdade pode bloquear portas de saída incomuns |
| **AWS / Google Cloud grátis** | grátis por pouco tempo, depois caro | | | GCP: 1 GB/mês de saída; AWS sa-east-1: US$ 0,15/GB | sim | não serve: o grátis acaba ou não cobre a banda |

## A conta da banda

Medidas do [`IDEIAS.md`](../../IDEIAS.md#passada-de-performance): sala parada ~90 KB/s, chuva ~48 KB/s, brainrot ~380 KB/s por pessoa (KB = 1000 bytes, como lá: 380 KB/s ≈ 1,4 GB/h).

| Cenário (30 pessoas × 2 h/dia × 30 dias = 1800 h) | Por hora | Por mês | Upload por pessoa |
|---|---|---|---|
| sala parada, 90 KB/s | 324 MB | **~583 GB** (543 GiB) | 0,72 Mbit/s |
| chuva, 48 KB/s | 173 MB | ~311 GB | 0,38 Mbit/s |
| brainrot, 380 KB/s | 1,37 GB | **~2,46 TB** (2293 GiB) | 3,04 Mbit/s |

| Plano | Tráfego incluso | Sala parada (583 GB) | Todo mundo no brainrot (2,46 TB) |
|---|---|---|---|
| Oracle Always Free | 10 TB | cabe | cabe |
| Hetzner CX23 | 20 TB | cabe | cabe |
| Vultr US$ 5 (São Paulo) | 1 TB | cabe | passa ~1,4 TB |
| DigitalOcean US$ 4 | 500 GiB | passa ~43 GiB (~US$ 0,43) | passa ~1793 GiB (~US$ 18) |
| Lightsail US$ 5 em São Paulo | 0,5 TB | passa ~83 GB | passa ~2 TB |
| Google Cloud e2-micro grátis | 1 GB | não cabe (~3 h de uma pessoa) | não cabe |
| AWS EC2 sa-east-1 | 0 (US$ 0,15/GB) | ~US$ 87 | ~US$ 369 |

Upload de casa (com 20% de folga):

| Upload | Pessoas na sala parada | Pessoas no brainrot |
|---|---|---|
| 10 Mbit/s | ~11 | ~2 |
| 50 Mbit/s | ~55 | ~13 |
| 100 Mbit/s | ~111 | ~26 |
| 300 Mbit/s | ~333 | ~78 |

## Rede local: o caminho do MVP

O servidor roda no notebook de alguém do grupo e a turma entra pela mesma rede. Não custa nada nem depende de conta, cartão ou operadora; o preço é só funcionar quando todo mundo está na mesma Wi-Fi.

### Que comando dá para ter

Um nome de uma palavra só (`sshiu`) não tem como ser achado na rede por todos os sistemas sem configuração. O que existe na rede local é o mDNS (o Bonjour da Apple): cada máquina anuncia `<nome>.local`, e qualquer nome que termina em `.local` "MUST be sent to the mDNS IPv4 link-local multicast address" ([RFC 6762, seção 3](https://www.rfc-editor.org/rfc/rfc6762#section-3)). Sem o `.local`, a RFC deixa mandar para o mDNS só como opção, e os sistemas não fazem isso:

| Quem entra | `ssh sshiu.local` | `ssh sshiu` sem configurar | Fonte |
|---|---|---|---|
| Mac | sim | não | testado num Mac com macOS 26.3: `<nome>.local` resolve, `<nome>` não |
| Linux com `nss-mdns` | sim | não: "will always deny to resolve host names that don't end in `.local`" | [nss-mdns](https://github.com/lathiat/nss-mdns) |
| Windows 10/11 | provavelmente; **não conferido em fonte primária** (só respostas de fórum da Microsoft) | só de Windows para Windows, pelo LLMNR ou NetBIOS, que a Microsoft está desligando aos poucos em favor do mDNS (fontes secundárias: [Born City, 2022](https://borncity.com/win/2022/04/25/microsoft-lsst-netbios-namensauflsung-und-llmnr-zuknftig-auslaufen/)) | [Microsoft Tech Community](https://techcommunity.microsoft.com/t5/windows-10/mdns-windows-10-resolving-sub-domain-local/td-p/1598354) |

Então há dois comandos possíveis:

1. **`ssh sshiu.local`, sem ninguém configurar nada.** Pede o notebook chamado `sshiu` e o servidor na porta 22.
2. **`ssh sshiu`, com o bloco colado uma vez.** O bloco aponta para `sshiu.local`, então continua valendo quando o roteador der outro IP ao notebook, e a porta deixa de importar:

```
Host sshiu
    HostName sshiu.local
    RequestTTY yes
    HostKeyAlias sshiu
```

Com o `RequestTTY yes`, o convite também dispensa o `-t`: `ssh sshiu <código>`. Os comandos de uma linha para colar são os [do bloco pronto](#o-bloco-pronto), trocando o `HostName` por `sshiu.local` e tirando o `Port`. Quem não colou o bloco usa `ssh sshiu.local`.

### O que quem hospeda faz

- **Nome do notebook:**
  - Mac: Ajustes do Sistema > Geral > Compartilhamento > "Nome do host local" > Editar; o nome vira `sshiu.local` ([Apple](https://support.apple.com/guide/mac-help/change-computers-local-hostname-mac-mchlp2322/mac)). Se outro aparelho da rede já usa o nome, o macOS acrescenta um número.
  - Windows: Configurações > Sistema > Sobre > Renomear este PC, ou `Rename-Computer -NewName sshiu` ([Microsoft](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.management/rename-computer)). Se o Windows anuncia o próprio nome por mDNS, a ponto de `sshiu.local` funcionar de um Mac, **não foi conferido**: testar antes da demo.
  - Linux: `hostnamectl set-hostname sshiu`; o anúncio por mDNS depende do `avahi-daemon` estar rodando.
- **Porta 22**, para o `ssh sshiu.local` não precisar de `-p`: `-host "" -port 22`.
  - Mac: o servidor sobe na 22 sem `sudo` (testado no macOS 26.3), desde que o "Login remoto" esteja desligado, porque ele usa a 22.
  - Windows: a 22 costuma estar livre (o servidor OpenSSH do Windows não vem instalado). O firewall pede a regra de entrada ([a do PC de casa](#3-pc-de-casa-com-port-forwarding), com `-LocalPort 22`).
  - Linux: `sudo setcap cap_net_bind_service=+ep ./sshiu`, a mesma capability da [VPS](#o-wish-na-porta-22).
  - Com o bloco `Host sshiu`, a porta tanto faz: dá para manter a 23234 e pôr `Port 23234` no bloco.
- **Rodar sempre da mesma pasta**, a que tem o `.ssh/` e o `sshiu.db` ([a chave do host](#a-chave-do-host-em-cada-caminho)).

### O que dá errado

- **A rede da faculdade.** Redes de campus e eduroam costumam isolar os aparelhos entre si: cada notebook alcança a internet, mas não o vizinho, e aí nada funciona, nem com o IP. Quais redes fazem isso, não dá para conferir daqui; é testar na sala. Para a apresentação, o mais seguro é um roteador próprio ou o hotspot de um celular, com o notebook e a turma nele. O mDNS também não atravessa redes diferentes ("meaningful only on the link where they originate", [RFC 6762](https://www.rfc-editor.org/rfc/rfc6762#section-3)).
- **Hotspot de celular:** o upload não pesa, porque o tráfego vai de aparelho para aparelho sem passar pela operadora. O limite é quantos aparelhos o hotspot aceita (costuma ficar em torno de 10, varia por aparelho; não conferido).
- **Notebook dormindo:** se a tampa fecha ou o sistema suspende, o servidor some no meio da demo.
- **O convite** mostra o IP local (`ssh -t 192.168.0.12 <código>`), não `sshiu` ([o convite](#o-convite-e-o-endereço-errado)). Na rede local ele funciona, mas não é o comando bonito. A flag de nome público (`-name sshiu`) resolve isso aqui também.

## 1. VPS com IP público

O binário sobe sem dependências externas: o `modernc.org/sqlite` não usa cgo ([pkg.go.dev](https://pkg.go.dev/modernc.org/sqlite)), então `GOOS=linux GOARCH=arm64 go build` (Oracle A1, Hetzner CAX) ou `GOARCH=amd64` sai do PC de quem hospeda pronto para copiar.

| Provedor | Menor plano | Preço (02/11/2026) | Tráfego incluso | Brasil? | Fonte |
|---|---|---|---|---|---|
| **Oracle Always Free** | Ampere A1: 1500 OCPU-horas e 9000 GB-horas/mês, "equivalent to 2 OCPUs and 12 GB of memory"; ou até 2 VMs AMD E2.1.Micro (50 Mbps cada) | R$ 0 | "10 TB per month of outbound data" | São Paulo (`sa-saopaulo-1`) e Vinhedo (`sa-vinhedo-1`) | [Always Free Resources](https://docs.oracle.com/en-us/iaas/Content/FreeTier/resourceref.htm), [Regions](https://docs.oracle.com/en-us/iaas/Content/General/Concepts/regions.htm) |
| **Vultr** | `vc2-1c-1gb`: 1 vCPU, 1 GB | US$ 5/mês (~R$ 27) | 1 TB | São Paulo (`sao`); o plano de US$ 3,50 não está em São Paulo | API pública: [regions](https://api.vultr.com/v2/regions), [plans](https://api.vultr.com/v2/plans) (a página de preços deu 403) |
| **Hetzner** | CX23: 2 vCPU, 4 GB (CAX11 Ampere: € 6,49) | € 5,99/mês sem IVA, com IPv4 (~R$ 36) | 20 TB; excedente € 1 por TB | não: Alemanha, Finlândia, EUA, Singapura; o CX23 só em Nuremberg e Helsinque | [Cost-Optimized](https://www.hetzner.com/cloud/cost-optimized), [Cloud](https://www.hetzner.com/cloud/) |
| **DigitalOcean** | Basic: 1 vCPU, 512 MiB | US$ 4/mês (~R$ 22) | 500 GiB; excedente US$ 0,01/GiB | não (16 datacenters, nenhum na América do Sul) | [Droplets](https://www.digitalocean.com/pricing/droplets), [Bandwidth](https://docs.digitalocean.com/products/billing/bandwidth/), [Regional availability](https://docs.digitalocean.com/platform/regional-availability/) |
| **AWS Lightsail** | 0,5 GB, 2 vCPU, IPv4 | US$ 5/mês (IPv6-only: US$ 3,50) | 1 TB, **metade (0,5 TB) em São Paulo** | São Paulo | [Lightsail pricing](https://aws.amazon.com/lightsail/pricing/) |
| **AWS Free Tier** | créditos | "up to $200 over 6 months"; a conta grátis fecha depois de 6 meses | saída de sa-east-1 a US$ 0,15/GB (primeiros 10 TB) | São Paulo | [AWS Free](https://aws.amazon.com/free/), [Price List API sa-east-1](https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AWSDataTransfer/current/sa-east-1/index.json) |
| **Google Cloud Free Tier** | 1 e2-micro | R$ 0 | **1 GB/mês** de saída | não: só Oregon, Iowa e Carolina do Sul | [Free cloud features](https://docs.cloud.google.com/free/docs/free-cloud-features) |
| **Magalu Cloud** | | não conferido: a calculadora da [página de preços](https://magalu.cloud/precos/) não carregou | não conferido | sim (cobra em reais) | |

**Oracle, em detalhe** (tudo da [doc de Always Free](https://docs.oracle.com/en-us/iaas/Content/FreeTier/resourceref.htm) e da [FAQ](https://www.oracle.com/cloud/free/faq/)):

- A instância grátis tem de ficar na **região de origem** ("You must create the Always Free compute instances in your home region"), escolhida no cadastro. Para ter latência baixa, a região de origem tem de ser São Paulo ou Vinhedo.
- **A cota de A1 caiu:** a doc de hoje diz 2 OCPUs e 12 GB. Guias de terceiros ainda falam em 4 OCPUs e 24 GB. Para o sshiu, os dois bastam.
- **Cartão:** a FAQ pede cartão para verificar identidade e diz: "We do not accept debit cards with a PIN or virtual, single-use, or prepaid cards".
- **Capacidade:** "out of host capacity" é "a temporary lack of Always Free shapes in your home region", e a FAQ avisa que pode levar "several days". As regiões do Brasil têm 1 domínio de disponibilidade cada, então não dá para tentar outro.
- **Instância ociosa:** pode ser recolhida se, em 7 dias, o CPU no percentil 95, a rede e (no A1) a memória ficarem abaixo de 20%. Um sshiu com uma turma pequena fica abaixo disso quase sempre. Fontes secundárias dizem que passar a conta para Pay As You Go (que continua grátis dentro dos limites) evita isso: **não conferido na doc oficial**.
- **Conta parada:** "Accounts left idle for 30 days or more may be deemed abandoned".
- **Firewall duplo:** a security list padrão só libera a porta 22 de "authorized source IP addresses" ([Security Lists](https://docs.oracle.com/en-us/iaas/Content/Network/Concepts/securitylists.htm)), e as imagens da Oracle vêm com o firewall do sistema ligado. No Ubuntu da Oracle, a regra nova entra em `/etc/iptables/rules.v4` **antes** do REJECT que já existe (segundo o [blog de desenvolvedores da Oracle](https://blogs.oracle.com/developers/post/enabling-network-traffic-to-ubuntu-images-in-oracle-cloud-infrastructure), que deu 403 aqui: conferido só pelo trecho da busca).
- **O IP público fica fora da máquina:** ele é atribuído a "a private IP object linked to the instance's VNIC" ([Public IP Addresses](https://docs.oracle.com/en-us/iaas/Content/Network/Tasks/managingpublicIPs.htm)). Por isso o `LocalAddr()` da sessão é o IP privado, e o convite sai com ele (ver [o convite](#o-convite-e-o-endereço-errado)). Use um IP público **reservado**, que "Exists until you delete it", para o endereço não mudar se a VM for recriada.

**Latência.** Não foi medida (a pesquisa não foi feita do Brasil). O piso físico, com a luz a ~200 000 km/s na fibra, ida e volta, a partir de São Paulo: Ashburn/Nova York ≥ 77 ms, Frankfurt/Nuremberg ≥ 98 ms, Hillsboro ≥ 109 ms, Helsinque ≥ 113 ms. A rota real costuma ser bem mais longa. Uma VPS em São Paulo fica em poucos ms a algumas dezenas, dependendo da cidade da faculdade. Num TUI a 8 fps, 100 a 200 ms aparecem como atraso no eco das teclas.

### O Wish na porta 22

O sshd da própria VPS já está na 22. Há três saídas:

| Jeito | Como | Cuidado |
|---|---|---|
| **Mover o sshd da VPS** e dar a 22 ao sshiu | `Port 2222` no `sshd_config` ("The default is 22", [sshd_config(5)](https://man.openbsd.org/sshd_config)); depois, o sshiu na 22 com a capability abaixo | Abra a 2222 na security list e no firewall **antes** de reiniciar o sshd, senão quem hospeda fica do lado de fora. No Ubuntu, um arquivo em `/etc/ssh/sshd_config.d/` vence o principal, porque "OpenSSH uses the first value set for most directives" ([Ubuntu Server docs](https://ubuntu.com/server/docs/how-to/security/openssh-server/)). Se o sshd do Ubuntu recente vem ativado por socket (`ssh.socket`), não foi conferido |
| **Capability no systemd** | `AmbientCapabilities=CAP_NET_BIND_SERVICE` com `User=sshiu` e `WorkingDirectory=/opt/sshiu` ([systemd.exec(5)](https://www.freedesktop.org/software/systemd/man/latest/systemd.exec.html)). `CAP_NET_BIND_SERVICE`: "Bind a socket to Internet domain privileged ports (port numbers less than 1024)" ([capabilities(7)](https://man7.org/linux/man-pages/man7/capabilities.7.html)) | O `WorkingDirectory` é obrigatório: o `sshiu.db` e o `.ssh/` são lidos da pasta atual ([main.go:31](../../main.go#L31), [main.go:43](../../main.go#L43)). Sem ele, o Wish gera uma chave nova em outra pasta |
| **Redirecionar 22 → 23234** | `nft add table nat`, `nft 'add chain nat prerouting { type nat hook prerouting priority -100; }'`, `nft add rule nat prerouting tcp dport 22 redirect to 23234` ([nftables wiki](https://wiki.nftables.org/wiki-nftables/index.php/Performing_Network_Address_Translation_(NAT))) | Ainda assim o sshd da VPS tem de sair da 22 (o redirect pega todo o tráfego que chega na 22). O `LocalAddr()` passa a mostrar a porta 23234, e o convite sai com `-p 23234` |

Unit mínima (a do meio, que é a mais simples):

```ini
[Service]
User=sshiu
WorkingDirectory=/opt/sshiu
ExecStart=/opt/sshiu/sshiu -host "" -port 22
AmbientCapabilities=CAP_NET_BIND_SERVICE
Restart=always
```

## 2. Túneis

### Tailscale

| Pergunta | Resposta | Fonte |
|---|---|---|
| Plano grátis | Personal: "Up to 6 users", "Unlimited user devices", US$ 0, uso não comercial | [Pricing](https://tailscale.com/pricing) |
| Compartilhar com quem está fora | por convite ou link (de uso único ou até 1000 usos; link sem uso expira em 30 dias); quem recebe precisa de conta Tailscale e do app; a máquina compartilhada aceita conexões, mas não inicia; o link ainda está em beta | [Sharing](https://tailscale.com/kb/1084/sharing) |
| Nome curto `sshiu` | funciona para quem está **dentro** da tailnet: com o MagicDNS, `ping monitoring` e `ping monitoring.yak-bebop.ts.net` são a mesma coisa; ligado por padrão em tailnets criadas desde 20/10/2022 | [MagicDNS](https://tailscale.com/kb/1081/magicdns) |
| Nome curto para máquina compartilhada | **não**: "Shared machines can only be reached by using their fully qualified domain name" (`sshiu.<tailnet>.ts.net`) | [Sharing](https://tailscale.com/kb/1084/sharing) |
| Nomear a máquina | `tailscale set --hostname=sshiu` (também muda o nome no MagicDNS) | [CLI](https://tailscale.com/kb/1080/cli) |
| Funnel (expor para a internet) | só portas 443, 8443 e 10000; "Funnel only works over TLS-encrypted connections"; tem `--tcp` e `--tls-terminated-tcp`, mas a doc não mostra SSH. O `ssh` puro não fala TLS, então precisaria de um `ProxyCommand` com `openssl s_client`, que o Windows não traz. **Não testado**; banda com "non-configurable bandwidth limits" | [Funnel](https://tailscale.com/kb/1223/funnel), [tailscale funnel](https://tailscale.com/kb/1311/tailscale-funnel) |
| CGNAT e dados móveis | começa pelo relé DERP e tenta a conexão direta; "can't establish a direct connection if both devices are behind a hard NAT"; o DERP é "generally slower" | [Connection types](https://tailscale.com/kb/1257/connection-types) |
| Windows | instalador `.exe`, Windows 10 ou mais novo, login pelo navegador | [Install on Windows](https://tailscale.com/kb/1022/install-windows) |

Na prática: até 6 pessoas, todo mundo convidado para a tailnet de quem hospeda, e `ssh sshiu` funciona sem configurar nada. Para 30, as outras 24 recebem a máquina por compartilhamento e precisam do bloco `Host sshiu` apontando para `sshiu.<tailnet>.ts.net`. Se quem recebe a máquina compartilhada conta nas 6 vagas, a doc não diz.

### Outros túneis

Todos encaminham TCP puro, então o cliente vê a chave do Wish (exceto o terminal no navegador da Cloudflare).

| Túnel | SSH puro sem programa extra no cliente? | Porta e endereço | Plano grátis | Fonte |
|---|---|---|---|---|
| **ngrok** | sim | `1.tcp.ngrok.io:12345`: "random ngrok-controlled hostname and a random port"; endereço fixo só com TCP Address, cobrado à parte | TCP grátis só "after adding a valid payment method"; **1 GB** de tráfego; Hobbyist US$ 10/mês com 5 GB | [TCP endpoints](https://ngrok.com/docs/universal-gateway/tcp/), [Pricing](https://ngrok.com/pricing) |
| **bore** (`bore.pub`) | sim | porta "assigned randomly"; `--port` pede uma, mas falha se estiver ocupada | grátis, sem limites publicados nem garantia; MIT; dá para hospedar o próprio (`bore server`, controle na 7835), o que pede uma VPS. O túnel não cifra, mas o SSH já cifra | [ekzhang/bore](https://github.com/ekzhang/bore) |
| **playit.gg** | sim ("custom TCP/UDP for anything else") | formato do endereço e porta fixa: não conferido | existe grátis; Premium US$ 3/mês; tem ponto em São Paulo | [playit.gg](https://playit.gg/), [pricing](https://playit.gg/pricing) |
| **pinggy** | sim | endereço aleatório, muda a cada reinício | TCP grátis com "60 minutes tunnel timeout"; preço do Pro não apareceu | [pinggy.io](https://pinggy.io/) |
| **Serveo** | sim, com `ssh -J serveo.net <alias>` (OpenSSH 7.3 ou mais novo) | alias privado: `ssh -R myalias:22:localhost:22 serveo.net`; porta pública escolhida só no pago | TCP público anônimo: 10 min, 2 conexões, 256 KiB/s; para alias, os limites não estão na doc. Pro: US$ 6/mês | [Serveo docs](http://serveo.net/docs/), [serveo.net](https://serveo.net/) |
| **Cloudflare Tunnel** | não: `cloudflared` no cliente como `ProxyCommand`, ou o terminal no navegador, ou o Cloudflare One Client | precisa de "an active zone on Cloudflare" (um domínio) | preço do Zero Trust: não conferido | [SSH](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/use-cases/ssh/), [cloudflared](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/use-cases/ssh/ssh-cloudflared-authentication/) |
| **Quick Tunnels** (`trycloudflare.com`) | não: a doc só mostra HTTP | nome muda a cada vez | sem conta nem domínio | [Quick Tunnels](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/do-more-with-tunnels/trycloudflare/) |
| **localhost.run** | não conferido (a página falhou no TLS) | | | |
| **frp, rathole, sish** | sim | o que quem hospeda configurar | autohospedados: exigem uma VPS com IP público, e aí é mais simples rodar o sshiu nela | |

O Serveo é o único que dá `ssh sshiu` para quem está fora sem porta aleatória e sem programa extra: o colega põe `ProxyJump serveo.net` e `HostName <alias>` no bloco. A doc avisa que aliases são "shared routing names rather than an authentication boundary". Para o sshiu tanto faz, porque a chave é a identidade. Só que é um serviço de terceiro, sem limites publicados para alias, e não foi testado.

## 3. PC de casa com port forwarding

- **CGNAT:** as operadoras compartilham um IPv4 entre clientes pela faixa `100.64.0.0/10` ([RFC 6598](https://www.rfc-editor.org/rfc/rfc6598)). Atrás disso, redirecionar a porta no roteador não adianta. Como saber: se o IP WAN que o roteador mostra está em `100.64.0.0/10` ou numa faixa privada (`10/8`, `172.16/12`, `192.168/16`, [RFC 1918](https://www.rfc-editor.org/rfc/rfc1918)), ou é diferente do IP que um site externo mostra, há CGNAT ou NAT duplo. Quais operadoras usam CGNAT, e se e por quanto vendem IP público ou fixo: **não conferido em fonte primária**. O NIC.br tem [material sobre CGNAT](https://semanacap.bcp.nic.br/files/apresentacao/arquivo/1614/Material_apoio_CGNAT.pdf), mas o PDF não pôde ser lido aqui.
- **IPv6 como saída:** com IPv6 dos dois lados não há NAT. O [ipv6.br](https://ipv6.br/) (NIC.br) diz que 60% de quem usa internet no Brasil já tem IPv6 habilitado (data inferida do nome da imagem: 29/10/2026). Os outros 40% não entram, e redes de faculdade e de empresa podem estar entre eles. Os roteadores domésticos costumam bloquear conexão IPv6 de fora por padrão ([RFC 6092](https://www.rfc-editor.org/rfc/rfc6092) recomenda isso), então ainda é preciso liberar a porta para o PC.
- **DDNS para o IP que muda:** [DuckDNS](https://www.duckdns.org/) é grátis e atualiza IPv4 e IPv6 por `https://www.duckdns.org/update?domains=...&token=...[&ipv6=...]` ([spec](https://www.duckdns.org/spec.jsp)); o limite de subdomínios por conta não está na página. [No-IP Free](https://www.noip.com/free): 1 hostname, que precisa ser confirmado a cada 30 dias.
- **Porta 22:** se a operadora bloqueia a entrada na 22, não foi conferido. Dá para redirecionar a porta externa 22 do roteador para a 23234 do PC, ou usar uma porta alta e o `Port` no bloco `Host sshiu`.
- **Firewall do Windows:** precisa de uma regra de entrada para a porta. É o mesmo comando que a Microsoft usa para o sshd, trocando a porta: `New-NetFirewallRule -Name sshiu -DisplayName sshiu -Enabled True -Direction Inbound -Protocol TCP -Action Allow -LocalPort 23234`, num PowerShell de administrador ([OpenSSH no Windows](https://learn.microsoft.com/en-us/windows-server/administration/openssh/openssh_install_firstuse)).
- **Upload e PC ligado:** ver a [conta da banda](#a-conta-da-banda). O Windows ainda precisa do `lfWriter` ([main.go](../../main.go)), um atalho que só existe porque o servidor roda no Windows.

## 4. Como o nome `sshiu` vira um host

Um nome de um rótulo só não resolve no DNS público. O `ssh` primeiro procura o nome no `~/.ssh/config`: "Hostname: Specifies the real host name to log into. The default is the name given on the command line" ([ssh_config(5)](https://man.openbsd.org/ssh_config)). Depois tenta o resolvedor do sistema (arquivo hosts, sufixos de busca como os que o MagicDNS põe).

| Origem | Dá `ssh sshiu`? | Dá `ssh -t sshiu <código>`? | Para quem |
|---|---|---|---|
| bloco `Host sshiu` no `~/.ssh/config` | sim, com qualquer caminho (IP, porta alta, túnel, `ProxyJump`) | sim; com `RequestTTY yes`, até sem o `-t` | cada colega cola uma vez |
| MagicDNS do Tailscale | sim, dentro da tailnet (até 6 pessoas no grátis) | sim | quem está na tailnet; máquina compartilhada só pelo nome completo |
| arquivo hosts (`/etc/hosts`, `C:\Windows\System32\drivers\etc\hosts`) | só com IP fixo e porta 22 (o hosts não guarda porta) | sim | pede administrador no Windows; não serve para túnel |
| domínio (`sshiu.duckdns.org`, `sshiu.com.br`) | não: dá `ssh sshiu.<algo>`, ou `ssh sshiu` com um `Host sshiu` que aponta para ele | idem | ninguém precisa configurar se for porta 22 |

**Windows:** o cliente OpenSSH lê `%userprofile%\.ssh\config` ([Microsoft, configuração do OpenSSH](https://learn.microsoft.com/en-us/windows-server/administration/openssh/openssh-server-configuration)). Ele existe como recurso opcional desde o Windows 10 1809, e a [tabela da Microsoft](https://learn.microsoft.com/en-us/windows-server/administration/openssh/openssh-overview) diz "Not installed" para o Windows 10 1809+. Quem não tiver o `ssh` (teste com `ssh -V`) instala em Configurações > Recursos opcionais > Cliente OpenSSH, ou com `Add-WindowsCapability -Online -Name OpenSSH.Client~~~~0.0.1.0` como administrador ([Microsoft](https://learn.microsoft.com/en-us/windows-server/administration/openssh/openssh_install_firstuse)).

### O bloco pronto

```
Host sshiu
    HostName 203.0.113.10
    Port 22
    RequestTTY yes
    HostKeyAlias sshiu
```

- `HostName`/`Port`: o IP da VPS, o `sshiu.<tailnet>.ts.net`, ou o `bore.pub` com a porta do túnel. Para o Serveo, troque por `HostName <alias>` e `ProxyJump serveo.net`.
- `RequestTTY yes`: "always request a TTY when standard input is a TTY", e `ssh sshiu 49ae4z` funciona sem o `-t` ([ssh_config(5)](https://man.openbsd.org/ssh_config)).
- `HostKeyAlias sshiu`: "an alias that should be used instead of the real host name when looking up or saving the host key". O `known_hosts` guarda a chave como `sshiu`, então trocar de caminho (IP novo, porta nova do túnel, Tailscale ↔ VPS) não pede a confirmação de novo, desde que a chave do host seja a mesma. Se a chave mudar, o aviso aparece do mesmo jeito.
- Não precisa de `User`: o sshiu ignora o nome e usa a chave como identidade.
- Vale o primeiro valor encontrado ([ssh_config(5)](https://man.openbsd.org/ssh_config)): se o arquivo já tem um `Host *` no topo com `Port` ou `HostName`, ele vence o bloco acrescentado no fim.

PowerShell (Windows):

```powershell
New-Item -ItemType Directory -Force "$HOME\.ssh" | Out-Null; Add-Content "$HOME\.ssh\config" "`nHost sshiu`n    HostName 203.0.113.10`n    Port 22`n    RequestTTY yes`n    HostKeyAlias sshiu"
```

bash/zsh (Mac e Linux):

```bash
mkdir -p ~/.ssh && printf '\nHost sshiu\n    HostName 203.0.113.10\n    Port 22\n    RequestTTY yes\n    HostKeyAlias sshiu\n' >> ~/.ssh/config
```

### Com domínio

- **Grátis:** `sshiu.duckdns.org`, com registro A e AAAA ([DuckDNS spec](https://www.duckdns.org/spec.jsp)). Comando: `ssh sshiu.duckdns.org`, sem `-p` se o servidor estiver na 22.
- **`.com` na Cloudflare Registrar:** "does not mark up domain prices at all" ([Cloudflare Registrar](https://www.cloudflare.com/products/registrar/)). A tarifa da Verisign subiu de US$ 10,26 para US$ 10,97 em 01/11/2026, mais US$ 0,20 da ICANN, o que dá **~US$ 11,17/ano (~R$ 61)**. Esse valor vem só de fontes secundárias ([Domain Name Wire](https://domainnamewire.com/2025/11/19/verisign-can-increase-com-prices-in-2026/), [Openprovider](https://www.openprovider.com/blog/com-price-increase-domain-resellers)): o comunicado da Verisign deu 403 e o preço na busca da Cloudflare não foi conferido.
- **`.com.br` no registro.br:** R$ 40/ano, segundo o [Teletime (10/2026)](https://teletime.com.br/14/10/2026/br-6-milhoes/), que cita o NIC.br. A página do registro.br não carregou, então é **fonte secundária**.
- Um domínio só tira o bloco `Host` da frente se o servidor estiver na **porta 22** e num endereço fixo (VPS com IP reservado, ou DDNS). Com túnel de porta aleatória, não ajuda.

### O convite e o endereço errado

O convite ([main.go:146](../../main.go#L146)) e a mensagem de "crie uma chave" ([main.go:70](../../main.go#L70)) tiram o endereço de onde a sessão chegou. A mensagem de subida ([main.go:91](../../main.go#L91)) lista os IPv4 das placas da máquina, que na Oracle, AWS e GCP também são só os privados. O que o convite mostra em cada caminho:

| Caminho | O convite mostra |
|---|---|
| Vultr, Hetzner, DO (IP público na placa), sshiu direto na 22 | `ssh -t 203.0.113.10 <código>`: funciona, mas não diz `sshiu` |
| Oracle, AWS, GCP (IP público por fora da máquina) | `ssh -t 10.0.0.x <código>`: **quebrado** |
| redirecionamento nftables 22 → 23234 | `ssh -t -p 23234 <ip> <código>`: só funciona se a 23234 também estiver aberta |
| Tailscale | `ssh -t 100.x.y.z <código>`: funciona dentro da tailnet; para máquina compartilhada, o IP "usually matches", mas nem sempre |
| túneis (bore, ngrok, playit, pinggy, Serveo) | `ssh -t 127.0.0.1 -p 23234 <código>` ou parecido: **quebrado** |

Para o convite dizer `ssh -t sshiu <código>` em qualquer caminho, o servidor precisa saber o nome público. Um jeito pequeno: uma flag (por exemplo `-name sshiu`) que, quando preenchida, vale no lugar do `LocalAddr()`. Fica para a #25 decidir.

## A chave do host em cada caminho

- **Onde está:** `.ssh/id_ed25519`, relativo à pasta em que o servidor roda ([main.go:43](../../main.go#L43)). Se o arquivo não existe, o `WithHostKeyPath` do Wish gera um Ed25519 novo sem avisar ([options.go:69-78](https://github.com/charmbracelet/wish/blob/v2.0.5/options.go#L69-L78)). Rodar da pasta errada é o jeito mais fácil de trocar a chave sem querer.
- **Backup e migração:** copiar `.ssh/id_ed25519`, `.ssh/id_ed25519.pub` e `sshiu.db` juntos para a mesma pasta no destino (por exemplo `scp -r .ssh sshiu.db vps:/opt/sshiu/`), deixar a chave legível só pela conta que roda o serviço (`chmod 600`) e apontar o `WorkingDirectory=` do systemd para essa pasta. Vale para PC → VPS, VPS → VPS e VPS → PC.
- **Túneis:** Tailscale, bore, ngrok, playit, pinggy, Serveo (`-J` ou `-L`), o port forwarding do roteador e o redirect do nftables passam o TCP adiante sem abrir. O handshake SSH continua entre o `ssh` do colega e o Wish, e o fingerprint visto é o do Wish. No Serveo com `-J`, o colega vê **também** a chave do Serveo, para o salto. A exceção é o terminal no navegador da Cloudflare: lá quem fala SSH com o Wish é a Cloudflare, o colega nunca vê a chave do host, e a chave do colega não é usada. Isso quebra "a chave é a identidade" ([ADR 0001](../adr/0001-chave-ssh-e-a-identidade.md)). É dedução pelo desenho, não testado.
- **Endereço novo, mesma chave:** o `known_hosts` grava a chave pelo nome e, fora da 22, como `[host]:porta` ([sshd(8), formato do known_hosts](https://man.openbsd.org/sshd#SSH_KNOWN_HOSTS_FILE_FORMAT)). Mudar de endereço ou de porta gera a pergunta "authenticity can't be established" de novo, não o aviso de chave trocada. O `HostKeyAlias sshiu` evita até a pergunta. O contrário também vale: no bore.pub, a porta que hoje é do sshiu pode amanhã ser de outro servidor, e aí quem guardou `[bore.pub]:NNNNN` vê o aviso grande.

## Recomendação

- **MVP e apresentação: rede local.** O servidor no notebook de alguém do grupo, chamado `sshiu`, na porta 22, num roteador ou hotspot próprio. Cada colega cola o bloco `Host sshiu` com `HostName sshiu.local` e entra com `ssh sshiu`; quem não colou entra com `ssh sshiu.local`. Antes da demo: fazer um ensaio com um Mac, um Windows e um Linux entrando, e acrescentar a flag de nome público para o convite mostrar `ssh sshiu <código>`. Se o notebook for Windows e o `.local` não responder, o bloco com o IP do notebook resolve, com um IP fixo reservado no roteador.
- **Quando o app sair da rede local:** uma **VPS em São Paulo** com o sshiu direto na 22 (unit do systemd com `CAP_NET_BIND_SERVICE` e `WorkingDirectory`), mais o bloco `Host sshiu` de uma linha para cada colega colar. Tentar primeiro a **Oracle Always Free** com região de origem em São Paulo ou Vinhedo e IP reservado (R$ 0, 10 TB). Se o cartão for recusado ou der "out of host capacity" até uma semana antes, ir de **Vultr São Paulo** (US$ 5, ~R$ 27 no mês), que cobre a sala parada da turma inteira. Por quê: a turma não instala nada além do `ssh`, funciona de dados móveis e da rede da faculdade, não depende do PC nem do upload de casa, e a latência é a menor possível. O convite precisa da flag de nome público, senão na Oracle ele sai com o IP privado.
- **Para testar fora da rede local antes da VPS (até 6 pessoas):** **Tailscale** com a máquina chamada `sshiu`. `ssh sshiu` sai pronto pelo MagicDNS, sem bloco.
- **Domínio:** só se o bloco `Host` incomodar: `sshiu.duckdns.org` é grátis, e um `.com` sai por ~R$ 61/ano. Antes de abrir para a turma toda no brainrot, rodar a [passada de performance](../../IDEIAS.md#passada-de-performance), porque 2,5 TB/mês só cabe na Oracle e na Hetzner.

## Perguntas abertas para a #25

- De quem é o notebook que hospeda, e que sistema ele tem? Se for Windows, o `sshiu.local` precisa ser testado de um Mac.
- A demo usa a Wi-Fi da faculdade (depois de testar se ela isola os aparelhos) ou um roteador ou hotspot próprio?
- Quem hospeda tem cartão de crédito aceito pela Oracle? Se a Oracle falhar, quem paga a Vultr, e por quantos meses?
- A instância Oracle ociosa vai ser recolhida? Vale passar para Pay As You Go (grátis dentro dos limites, mas com o cartão exposto a cobrança se algo passar)?
- O convite ganha a flag de nome público (`-name sshiu`)? Com ela, a mensagem de subida e a de "crie uma chave" também passam a usar o nome.
- Domínio agora ou nunca? O bloco `Host sshiu` resolve, mas é um passo a mais por colega, e o README teria de ensinar.
- Quem guarda o backup do `.ssh/` e do `sshiu.db`, e onde? Perder a chave do host é perder a confiança de toda a turma de uma vez.
- O brainrot fica liberado para todo mundo na VPS de 1 TB, ou só depois da passada de performance?
- Quem recebe máquina compartilhada no Tailscale conta nas 6 vagas do plano grátis? A doc não diz.

## Não conferido

- Resolução de `.local` no Windows e se o Windows anuncia o próprio nome por mDNS (sem doc primária da Microsoft; só fórum).
- Se a Wi-Fi da faculdade isola os aparelhos; quantos aparelhos um hotspot de celular aceita.
- Preço do Magalu Cloud (a calculadora não carregou).
- Excedente de tráfego da Vultr e do Lightsail em São Paulo (a página da Vultr deu 403; o Lightsail só dá exemplo de Ohio).
- Se o Pay As You Go da Oracle livra a instância da regra de ociosidade (só fontes secundárias).
- Preço atual de `.com` na Cloudflare e de `.com.br` no registro.br (só fontes secundárias).
- Plano grátis do Cloudflare Zero Trust; detalhes do grátis do playit.gg; preço do pinggy Pro; localhost.run (falha de TLS).
- CGNAT, IP público e bloqueio da porta 22 por operadora brasileira (sem fonte primária); o PDF de CGNAT do NIC.br não pôde ser lido.
- Funnel do Tailscale com SSH via `openssl s_client` (não testado).
- Socket activation do sshd no Ubuntu recente ao mudar a porta.
- Latência real a partir do Brasil (só o piso físico calculado).
