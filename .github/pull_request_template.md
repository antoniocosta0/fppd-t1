## Descrição da Mudança

> Resumo do que esta alteração faz e por quê.

**Programa(s) afetado(s):**

- [ ] `filosofos/base_deadlock`
- [ ] `filosofos/hierarquia`
- [ ] `filosofos/garcom`
- [ ] `produtor-consumidor/canal`
- [ ] `produtor-consumidor/semaforos`
- [ ] Outro (README, relatório, `go.mod`, ...)

## Tipo de Mudança

- [ ] Feature - Nova versão ou funcionalidade de um programa
- [ ] Bugfix - Correção de bug (deadlock indevido, data race, item perdido, ...)
- [ ] Refactoring - Refatoração sem mudar comportamento
- [ ] Dados - Coleta ou ajuste de métricas de execução
- [ ] Documentation - README, relatório
- [ ] Chore - Estrutura, configuração, manutenção

## Requisitos do Enunciado

> Marque só o que se aplica a este PR.

- [ ] Programa recebe os parâmetros por flag (`-n`/`-r` ou `-p`/`-c`/`-k`).
- [ ] Encerra de forma ordenada, sem goroutines bloqueadas (exceto a versão base de deadlock, que deve detectar e avisar).
- [ ] Coleta e imprime os dados de execução pedidos.
- [ ] Sem bibliotecas que abstraiam a concorrência.

## Checklist

- [ ] PR aberto contra `main`.
- [ ] Branch sem conflitos com `main`.
- [ ] `gofmt -l .` não lista nenhum arquivo.
- [ ] `go vet ./...` passando.
- [ ] `go run -race` passando no(s) programa(s) alterado(s) (regra 3 do enunciado).
- [ ] Commits no formato `<tipo>: <descrição>` (`feat`, `fix`, `docs`, `chore`).
- [ ] O diff tem só arquivos desta mudança.
- [ ] Sem binários, `.DS_Store` ou arquivos de editor no diff.
- [ ] Fontes consultadas anotadas para a seção de Referências do relatório.

## Evidências

> Saída do programa (tabelas de refeições, espera média, throughput) e do `-race`.
