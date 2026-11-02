# Caminhos para a turma entrar com `ssh sshiu`

> **É uma sugestão, não uma decisão.** Esta pesquisa serve de ponto de partida: a decisão de onde hospedar fica para a issue [#25](https://github.com/drbarros2004/ssh-focus-app/issues/25). Preços e planos mudam rápido: quem for contratar deve conferir de novo as páginas citadas.

Pesquisa da issue [#24](https://github.com/drbarros2004/ssh-focus-app/issues/24), feita em 02/11/2026, com fontes oficiais (docs, páginas de preço, RFCs, man pages). O que não deu para conferir está marcado. Reais pela PTAX de 30/10/2026: US$ 1 = R$ 5,42.

O teste com outro computador funcionou na mesma Wi-Fi e não nos dados móveis porque o IP do servidor é privado. Para o MVP, a turma fica na rede local; os caminhos de fora ficam para depois.

## Resumo

- **MVP: rede local.** O notebook que hospeda se chama `sshiu` e roda o servidor na porta 22. Cada colega cola uma vez um bloco `Host sshiu` que aponta para `sshiu.local` e entra com `ssh sshiu`; sem o bloco, `ssh sshiu.local`.
- **O nome `sshiu` sozinho não resolve** em lugar nenhum sem configuração: nem no DNS público, nem no mDNS da rede local. Por isso o bloco `Host`.
- **Risco da demo:** Wi-Fi de faculdade costuma isolar os aparelhos. O mais seguro é um roteador ou o hotspot de um celular.
- **Depois do MVP:** VPS em São Paulo (Oracle Always Free, ou Vultr a US$ 5/mês). Tailscale serve para até 6 pessoas.
- **Dois achados no código** valem para qualquer caminho: o convite mostra um IP, não `sshiu`; e o Wish troca a chave do host sem avisar se o servidor rodar de outra pasta ([achados no código](#achados-no-código)).

## Rede local: o caminho do MVP

### O comando

Na rede local, cada máquina anuncia `<nome>.local` por mDNS (o Bonjour), e nomes terminados em `.local` "MUST be sent to the mDNS IPv4 link-local multicast address" ([RFC 6762, seção 3](https://www.rfc-editor.org/rfc/rfc6762#section-3)). Sem o `.local`, os sistemas não perguntam:

| Quem entra | `ssh sshiu.local` | `ssh sshiu` sem configurar |
|---|---|---|
| Mac | sim | não (testado no macOS 26.3) |
| Linux com `nss-mdns` | sim | não: "will always deny to resolve host names that don't end in `.local`" ([nss-mdns](https://github.com/lathiat/nss-mdns)) |
| Windows 10/11 | provavelmente; **não conferido** em doc da Microsoft | só de Windows para Windows (LLMNR/NetBIOS) |

O bloco, colado uma vez por cada colega:

```
Host sshiu
    HostName sshiu.local
    RequestTTY yes
    HostKeyAlias sshiu
```

- `HostName sshiu.local` continua valendo quando o roteador dá outro IP ao notebook. Com o bloco, a porta tanto faz (`Port 23234` também serve).
- `RequestTTY yes` dispensa o `-t`: o convite vira `ssh sshiu <código>` ([ssh_config(5)](https://man.openbsd.org/ssh_config)).
- `HostKeyAlias sshiu` guarda a chave do host pelo nome, então trocar de IP ou de caminho (rede local → VPS) não pede a confirmação de novo, desde que a chave seja a mesma.
- Vale o primeiro valor encontrado: um `Host *` no topo do arquivo vence o bloco.

Como o `ssh terminal.shop`, que publica a linha de `known_hosts` dele, a colagem leva junto a chave do servidor. Com o `HostKeyAlias sshiu`, a linha é `sshiu` seguido das duas primeiras palavras do `.ssh/id_ed25519.pub` de quem hospeda (`awk '{print "sshiu "$1" "$2}' .ssh/id_ed25519.pub`). Com ela, a primeira conexão não pergunta "Are you sure you want to continue connecting?", e um servidor falso na rede é recusado (testado: sem a linha e com checagem estrita, "Host key verification failed"; com ela, conecta). Troque `AAAA...` pela chave de verdade.

PowerShell (Windows):

```powershell
New-Item -ItemType Directory -Force "$HOME\.ssh" | Out-Null; Add-Content "$HOME\.ssh\config" "`nHost sshiu`n    HostName sshiu.local`n    RequestTTY yes`n    HostKeyAlias sshiu"; Add-Content "$HOME\.ssh\known_hosts" "sshiu ssh-ed25519 AAAA..."
```

bash/zsh (Mac e Linux):

```bash
mkdir -p ~/.ssh && printf '\nHost sshiu\n    HostName sshiu.local\n    RequestTTY yes\n    HostKeyAlias sshiu\n' >> ~/.ssh/config && echo 'sshiu ssh-ed25519 AAAA...' >> ~/.ssh/known_hosts
```

### O que cada colega digita

| Situação | Comandos |
|---|---|
| já tem chave SSH (quem usa o GitHub por SSH, por exemplo) | `ssh sshiu.local`: um comando, como o `ssh terminal.shop`, mais o "yes" da primeira conexão |
| ainda não tem chave | o sshiu responde com a instrução ([main.go:72](../../main.go#L72)): `ssh-keygen -t ed25519` (enter em tudo), depois `ssh sshiu.local`. Só na primeira vez; o `ssh-keygen` vem com o cliente OpenSSH do Windows |
| quer `ssh sshiu` | colar a linha acima uma vez; daí em diante, `ssh sshiu`, sem a pergunta da primeira conexão |

O que fica diferente do terminal.shop: a chave é obrigatória (ela é o login, [ADR 0001](../adr/0001-chave-ssh-e-a-identidade.md)), e só funciona na mesma rede.

No Windows, o cliente lê `%userprofile%\.ssh\config` ([Microsoft](https://learn.microsoft.com/en-us/windows-server/administration/openssh/openssh-server-configuration)). Quem não tiver o `ssh` (`ssh -V`) instala o Cliente OpenSSH em Configurações > Recursos opcionais ([Microsoft](https://learn.microsoft.com/en-us/windows-server/administration/openssh/openssh_install_firstuse)).

### O que quem hospeda faz

- **Nome do notebook:** no Mac, Ajustes do Sistema > Geral > Compartilhamento > "Nome do host local" ([Apple](https://support.apple.com/guide/mac-help/change-computers-local-hostname-mac-mchlp2322/mac)); no Windows, `Rename-Computer -NewName sshiu` ([Microsoft](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.management/rename-computer)); no Linux, `hostnamectl set-hostname sshiu`, com o `avahi-daemon` rodando. Se o Windows anuncia `sshiu.local` para um Mac, **não foi conferido**: testar antes.
- **Porta 22** (para `ssh sshiu.local` sem `-p`): `-host "" -port 22`. No Mac sobe sem `sudo` (testado no macOS 26.3) se o "Login remoto" estiver desligado. No Windows a 22 costuma estar livre e o firewall pede uma regra: `New-NetFirewallRule -Name sshiu -DisplayName sshiu -Enabled True -Direction Inbound -Protocol TCP -Action Allow -LocalPort 22`. No Linux, `sudo setcap cap_net_bind_service=+ep ./sshiu` ([capabilities(7)](https://man7.org/linux/man-pages/man7/capabilities.7.html)).
- **Rodar sempre da mesma pasta**, a que tem o `.ssh/` e o `sshiu.db`.

### O que dá errado

- **Rede que isola aparelhos**, comum em campus e eduroam: nada funciona, nem com o IP. Testar na sala, ou levar roteador/hotspot. O mDNS também não atravessa redes diferentes ([RFC 6762](https://www.rfc-editor.org/rfc/rfc6762#section-3)).
- **Hotspot:** o tráfego fica entre os aparelhos, sem gastar dados; o limite é quantos aparelhos ele aceita (varia; não conferido).
- **Notebook dormindo** derruba o servidor no meio da demo.
- **Windows hospedando:** se o `.local` não responder, o bloco usa o IP do notebook, reservado no roteador.

## Para depois: fora da rede local

| Caminho | Custo | Quem hospeda | Cada colega | Banda | No ar com o PC desligado? | O que dá errado |
|---|---|---|---|---|---|---|
| **VPS Oracle Always Free** (São Paulo ou Vinhedo) | R$ 0 | conta com cartão; região de origem no Brasil; VM; liberar a porta na security list e no firewall da imagem; systemd na 22 | o mesmo bloco, com o IP | 10 TB/mês | sim | "out of host capacity"; cartão virtual/pré-pago recusado; instância ociosa por 7 dias pode ser recolhida |
| **VPS Vultr São Paulo** | US$ 5/mês (~R$ 27) | conta; VM; systemd na 22 | o mesmo bloco, com o IP | 1 TB/mês | sim | 1 TB não cobre 30 pessoas no brainrot |
| **Tailscale** | R$ 0 até 6 pessoas | instalar; `tailscale set --hostname=sshiu`; convidar | instalar, criar conta, aceitar o convite | upload de quem hospeda | não | passa de 6 pessoas: o resto entra por compartilhamento, só pelo nome completo `sshiu.<tailnet>.ts.net` |

**Banda.** Pelas medidas do [`IDEIAS.md`](../../IDEIAS.md#passada-de-performance), 30 pessoas × 2 h/dia dão ~583 GB/mês numa sala parada (90 KB/s) e ~2,5 TB/mês se todo mundo ficar no brainrot (380 KB/s).

**VPS.**
- Oracle ([Always Free](https://docs.oracle.com/en-us/iaas/Content/FreeTier/resourceref.htm), [FAQ](https://www.oracle.com/cloud/free/faq/)): a instância grátis fica na região de origem, escolhida no cadastro; o A1 grátis hoje é "equivalent to 2 OCPUs and 12 GB of memory"; "10 TB per month of outbound data"; instância com CPU, rede e memória abaixo de 20% por 7 dias pode ser recolhida. O IP público fica fora da máquina ([Public IP Addresses](https://docs.oracle.com/en-us/iaas/Content/Network/Tasks/managingpublicIPs.htm)): usar um IP reservado.
- Vultr: `vc2-1c-1gb`, US$ 5/mês, 1 TB, em São Paulo ([API: plans](https://api.vultr.com/v2/plans), [regions](https://api.vultr.com/v2/regions)).
- **Wish na 22:** mover o sshd da VPS para outra porta (abrir a nova antes de reiniciar) e rodar o sshiu pelo systemd com a capability ([systemd.exec(5)](https://www.freedesktop.org/software/systemd/man/latest/systemd.exec.html)). O `WorkingDirectory` é obrigatório, porque o `.ssh/` e o `sshiu.db` são lidos da pasta atual:

```ini
[Service]
User=sshiu
WorkingDirectory=/opt/sshiu
ExecStart=/opt/sshiu/sshiu -host "" -port 22
AmbientCapabilities=CAP_NET_BIND_SERVICE
Restart=always
```

**Tailscale** ([Pricing](https://tailscale.com/pricing), [MagicDNS](https://tailscale.com/kb/1081/magicdns), [Sharing](https://tailscale.com/kb/1084/sharing)): plano Personal com "Up to 6 users"; dentro da tailnet, `ssh sshiu` funciona pelo MagicDNS sem bloco; "Shared machines can only be reached by using their fully qualified domain name". Bom para quem desenvolve testar de fora da rede antes da VPS.

**Domínio:** só se o bloco `Host` incomodar, e só ajuda com servidor fixo na porta 22. Dá `ssh sshiu.<algo>`, nunca `ssh sshiu`. `sshiu.duckdns.org` é grátis ([DuckDNS](https://www.duckdns.org/spec.jsp)); um `.com` sai por ~US$ 11/ano na Cloudflare (preço de fonte secundária).

### Descartados

- **Túneis TCP públicos** (ngrok, bore.pub, playit.gg, pinggy): porta aleatória que muda ao reiniciar e plano grátis curto demais (ngrok: 1 GB e cartão, [pricing](https://ngrok.com/pricing); pinggy: 60 min, [pinggy.io](https://pinggy.io/)).
- **Cloudflare Tunnel:** SSH exige o `cloudflared` instalado em cada colega, ou o terminal no navegador, que não usa a chave de quem entra ([docs](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/use-cases/ssh/)).
- **Tailscale Funnel:** só HTTPS/TLS nas portas 443, 8443 e 10000 ([docs](https://tailscale.com/kb/1223/funnel)); o `ssh` puro não fala TLS.
- **Serveo com alias:** dá nome fixo com `ProxyJump`, mas é serviço de terceiro sem limites publicados ([docs](http://serveo.net/docs/)).
- **PC de casa com port forwarding:** o CGNAT das operadoras ([RFC 6598](https://www.rfc-editor.org/rfc/rfc6598)) impede, o upload doméstico limita e o PC fica ligado.
- **AWS e Google Cloud grátis:** o grátis acaba em 6 meses (AWS) ou tem 1 GB/mês de saída (GCP).
- **Hetzner e DigitalOcean:** sem datacenter no Brasil; Vultr São Paulo custa o mesmo.

## Achados no código

- **O convite mostra um IP, não `sshiu`.** Ele e a mensagem de "crie uma chave" montam o endereço a partir de `s.LocalAddr()` ([main.go:146](../../main.go#L146), [main.go:70](../../main.go#L70)). Na rede local sai o IP do notebook (funciona, mas não é bonito); na Oracle, o IP privado (quebrado). Uma flag de nome público (por exemplo `-name sshiu`) no lugar do `LocalAddr()` resolve os dois.
- **A chave do host pode trocar sem aviso.** Se `.ssh/id_ed25519` não existe na pasta atual ([main.go:43](../../main.go#L43)), o Wish gera uma nova ([wish v2.0.5, options.go:69-78](https://github.com/charmbracelet/wish/blob/v2.0.5/options.go#L69-L78)), e a turma inteira vê "REMOTE HOST IDENTIFICATION HAS CHANGED". Guardar backup do `.ssh/` e do `sshiu.db` e levá-los juntos em qualquer migração.

## Perguntas abertas para a #25

- De quem é o notebook que hospeda, e que sistema ele tem? Se for Windows, testar `sshiu.local` de um Mac.
- A demo usa a Wi-Fi da faculdade (testada antes) ou um roteador/hotspot próprio?
- O convite ganha a flag de nome público?
- Quem guarda o backup do `.ssh/` e do `sshiu.db`?

## Não conferido

- Resolução de `.local` no Windows e se o Windows anuncia o próprio nome por mDNS.
- Se a Wi-Fi da faculdade isola os aparelhos; quantos aparelhos um hotspot aceita.
- Excedente de tráfego da Vultr; preço do `.com` direto na Cloudflare.
