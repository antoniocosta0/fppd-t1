# T1 — Problemas Clássicos de Concorrência (FPPD)

Trabalho da disciplina de Fundamentos de Processamento Paralelo e Distribuído (PUCRS). Problemas clássicos de sincronização em Go, cada versão como um programa independente. O Problema 1 (Jantar dos Filósofos) e o Problema 2 (Produtor/Consumidor) estão implementados.

**Grupo:** _(Antônio Costa, João Pedro Bianchi, Gustavo Both)_

## Requisitos

- Go 1.21 (ou superior), conforme o `go.mod`.
- Nenhuma biblioteca externa: só a biblioteca padrão.

## Estrutura

```
go.mod
filosofos/
  base_deadlock/   versão base: todos pegam o garfo da esquerda e depois o da direita (trava de propósito)
  hierarquia/      estratégia 1: hierarquia de recursos (menor índice primeiro)
  garcom/          estratégia 2: garçom, no máximo N-1 filósofos à mesa
produtor-consumidor/
  canal/           buffer limitado = channel com capacidade K
  semaforos/       buffer limitado = array protegido por semáforos (notEmpty, notFull) e mutex
```

Cada subpasta tem o seu próprio `main.go`. Por isso cada versão é executada separadamente.

## Como executar

Rode sempre da raiz do projeto (a pasta que contém o `go.mod`). Todos os programas devem rodar sem erro com `-race`.

### Problema 1 — Jantar dos Filósofos

```
go run -race ./filosofos/base_deadlock -n 5 -r 3 -pausa 50 -limite 3
go run -race ./filosofos/hierarquia    -n 5 -r 1000 -pausa 1
go run -race ./filosofos/garcom        -n 5 -r 1000 -pausa 1
go run -race ./filosofos/garcom        -n 5 -r 1000 -pausa 1 -assentos 3
```

| Flag | Significado | Padrão |
|---|---|---|
| `-n` | número de filósofos (e de garfos), mínimo 2 | 5 |
| `-r` | refeições por filósofo | 3 |
| `-pausa` | pausa entre pegar o primeiro e o segundo garfo, em ms | 50 |
| `-limite` | tempo em s. No `base_deadlock` é o tempo **sem progresso** (nenhuma refeição nova) antes de declarar deadlock, padrão 5. Nas outras versões é só uma rede de segurança (tempo total) e 0 desliga | 5 / 0 |
| `-assentos` | só no `garcom`: quantos filósofos podem disputar garfos ao mesmo tempo, de 1 a N-1 | N-1 |

Resultado esperado: o `base_deadlock` com pausa maior que 0 trava, e o programa imprime "DEADLOCK detectado" com o estado de cada filósofo e encerra. `hierarquia` e `garcom` terminam as R refeições de todos e imprimem a tabela com refeições e espera média por filósofo.

### Problema 2 — Produtor/Consumidor

```
go run -race ./produtor-consumidor/canal     -p 4 -c 4 -k 10
go run -race ./produtor-consumidor/semaforos -p 4 -c 4 -k 10
```

| Flag | Significado | Padrão |
|---|---|---|
| `-p` | número de produtores | 4 |
| `-c` | número de consumidores | 4 |
| `-k` | capacidade do buffer | 10 |
| `-itens` | itens produzidos por cada produtor | 500 |

Resultado esperado: os dois imprimem tempo, throughput, ocupação média do buffer, timeouts, itens por consumidor e, no fim, `OK: produzidos = consumidos`. Para comparar o efeito do buffer, rode com `-k 1`, `-k 10` e `-k 100`.

## Verificações de qualidade

```
gofmt -l .
go vet ./...
```

