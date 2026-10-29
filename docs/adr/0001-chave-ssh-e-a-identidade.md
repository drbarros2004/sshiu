# A chave SSH é a identidade

O usuário é o fingerprint da chave pública com que se conecta: o servidor aceita qualquer chave, sem senha, cadastro ou e-mail. É o que permite entrar só com `ssh <nome>`, sem configurar nada, e o servidor não guarda segredo nenhum. Quem chega sem chave recebe a instrução para criar uma.

## Consequências

- Outra máquina, com outra chave, é outro usuário (outro apelido, outro pet). Copiar a mesma chave para as duas máquinas resolve hoje; o login pelo GitHub em [`IDEIAS.md`](../../IDEIAS.md) é o conserto definitivo.
- O fingerprint é a chave primária dos usuários no banco: trocar de identidade depois exige migrar os dados.
