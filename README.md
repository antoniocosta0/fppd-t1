# T1 — Problemas Clássicos de Concorrência (FPPD)

Trabalho da disciplina de Fundamentos de Processamento Paralelo e Distribuído (PUCRS). Dois problemas clássicos de sincronização implementados em Go, cada versão como um programa independente.

**Grupo:** _(preencher nomes dos integrantes)_

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
```

| Flag | Significado | Padrão |
|---|---|---|
| `-n` | número de filósofos (e de garfos), mínimo 2 | 5 |
| `-r` | refeições por filósofo | 3 |
| `-pausa` | pausa entre pegar o primeiro e o segundo garfo, em ms | 50 |
| `-limite` | tempo máximo em s antes de declarar deadlock. No `base_deadlock` o padrão é 5. Nas outras versões é só uma rede de segurança e 0 desliga | 5 / 0 |

Resultado esperado: o `base_deadlock` com pausa maior que 0 trava, e o programa imprime "DEADLOCK detectado" com o estado de cada filósofo e encerra. `hierarquia` e `garcom` terminam as R refeições de todos e imprimem a tabela com refeições e espera média por filósofo.

### Problema 2 — Produtor/Consumidor

```
go run -race ./produtor-consumidor/canal     -p 4 -c 4 -k 10 -itens 1000
go run -race ./produtor-consumidor/semaforos -p 4 -c 4 -k 10 -itens 1000
```

| Flag | Significado | Padrão |
|---|---|---|
| `-p` | número de produtores (P) | 4 |
| `-c` | número de consumidores (C) | 4 |
| `-k` | capacidade do buffer (K) | 10 |
| `-itens` | itens produzidos por cada produtor | 1000 |
| `-tprod` | tempo médio de produção, em µs | 0 |
| `-tcons` | tempo médio de consumo, em µs | 100 |
| `-timeout` | timeout do consumidor 0 (`select` + `time.After`), em ms; 0 desliga | 50 |
| `-limite` | rede de segurança em s | 60 |

Para ver o consumo com timeout em ação: `go run -race ./produtor-consumidor/canal -p 1 -c 2 -k 5 -itens 20 -tprod 80000 -tcons 0 -timeout 30`.

No final, cada programa imprime itens e timeouts por consumidor, tempo total, throughput, ocupação média do buffer e uma verificação (total produzido igual ao total consumido, cada item exatamente uma vez).

## Verificações de qualidade

```
gofmt -l .
go vet ./...
```

## Relatório e uso de IA

O relatório em PDF acompanha a entrega e contém as análises de deadlock, starvation e fairness, o efeito de K e a seção sobre o uso de ferramentas de IA.
