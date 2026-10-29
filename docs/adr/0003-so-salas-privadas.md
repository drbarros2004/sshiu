# Só salas privadas, sem amigos e sem sala pública

O app é para uma turma, então tudo acontece dentro de salas privadas, em que se entra por código. Não existe sala pública nem relação de amigos (como no GymRats): o "se ver" é a lista de membros da sala, com o status de cada um. Com uma relação só, cada feature social (reações, leaderboard, gráficos, recompensas) pertence à sala, sem decidir se vale para amigos, para a sala ou para os dois.

## Consequências

- O primeiro acesso é um menu ("entrar com código" ou "criar sala"), inclusive na demo, em que a sala demo aparece como uma das opções.
- No MVP, a sala não tem dono com poderes: qualquer membro sai, e a sala some quando o último sai. Trocar o código, remover um membro e passar a sala para outro ficam para depois (o banco continua guardando quem criou a sala).
- Quem quer ver alguém de fora da turma cria uma sala de dois.
