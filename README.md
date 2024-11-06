# Transação Distribuída

* Implementando em Go o clássico exemplo de Passagens, Hotel & Pagamento.
* Uso do padrão SAGAS para coreografar os eventos
  * Confirmou: dispara confirmações nos serviços em sequência
  * Em caso de qualquer erro, dispara reversões na ordem contrária
  * Tempo esgotado: dispara o cancelamento das pré-reservas
* Uso de lock distribuído para as atividades de confirmação e cancelamentos por timeout não interferirem na mesma sessão de usuário

## Tecnologias

* RabbitMQ como broker para coordenação das transações distribuídas
* Redis como mecanismo de lock distribuído e armazenamento de sessão
* PostgreSQL para persistência dos dados de negócio
* Go como linguagem de programação