### base_deadlock -n 5 -r 3 -pausa 50 -limite 3 (deadlock detectado)
base_deadlock: N=5 R=3 pausa=50ms limite=3s

DEADLOCK detectado: nenhum progresso em 3s
as 4 condicoes de Coffman estao presentes:
  1. exclusao mutua  : cada garfo (channel com buffer 1) tem um dono por vez
  2. hold and wait   : cada filosofo segura o garfo esquerdo e espera o direito
  3. nao preempcao   : nenhum filosofo tira o garfo de outro
  4. espera circular : i espera i+1 e o ultimo espera o garfo 0

filosofo  refeicoes  espera media   estado final
0         0          0s             com o garfo esquerdo (esperando o direito)
1         0          0s             com o garfo esquerdo (esperando o direito)
2         0          0s             com o garfo esquerdo (esperando o direito)
3         0          0s             com o garfo esquerdo (esperando o direito)
4         0          0s             com o garfo esquerdo (esperando o direito)

total de refeicoes: 0
### base_deadlock -n 5 -r 3 -pausa 0 -limite 3 (10 execucoes: 0 deadlocks, 10 terminaram)
terminou sem deadlock (todos os 5 filosofos fizeram 3 refeicoes)
terminou sem deadlock (todos os 5 filosofos fizeram 3 refeicoes)
terminou sem deadlock (todos os 5 filosofos fizeram 3 refeicoes)
terminou sem deadlock (todos os 5 filosofos fizeram 3 refeicoes)
terminou sem deadlock (todos os 5 filosofos fizeram 3 refeicoes)
terminou sem deadlock (todos os 5 filosofos fizeram 3 refeicoes)
terminou sem deadlock (todos os 5 filosofos fizeram 3 refeicoes)
terminou sem deadlock (todos os 5 filosofos fizeram 3 refeicoes)
terminou sem deadlock (todos os 5 filosofos fizeram 3 refeicoes)
terminou sem deadlock (todos os 5 filosofos fizeram 3 refeicoes)
### hierarquia rodada 1
hierarquia: N=5 R=1000 pausa=1ms

terminou sem deadlock em 19.513s (todos os 5 filosofos fizeram 1000 refeicoes)

filosofo  refeicoes  espera media   estado final
0         1000       8.145ms        terminou (nao segura garfo)
1         1000       6.395ms        terminou (nao segura garfo)
2         1000       5.484ms        terminou (nao segura garfo)
3         1000       4.187ms        terminou (nao segura garfo)
4         1000       7.66ms         terminou (nao segura garfo)

total de refeicoes: 5000
### hierarquia rodada 2
hierarquia: N=5 R=1000 pausa=1ms

terminou sem deadlock em 19.468s (todos os 5 filosofos fizeram 1000 refeicoes)

filosofo  refeicoes  espera media   estado final
0         1000       7.869ms        terminou (nao segura garfo)
1         1000       6.397ms        terminou (nao segura garfo)
2         1000       5.492ms        terminou (nao segura garfo)
3         1000       4.19ms         terminou (nao segura garfo)
4         1000       7.856ms        terminou (nao segura garfo)

total de refeicoes: 5000
### hierarquia rodada 3
hierarquia: N=5 R=1000 pausa=1ms

terminou sem deadlock em 19.751s (todos os 5 filosofos fizeram 1000 refeicoes)

filosofo  refeicoes  espera media   estado final
0         1000       8.297ms        terminou (nao segura garfo)
1         1000       6.429ms        terminou (nao segura garfo)
2         1000       5.511ms        terminou (nao segura garfo)
3         1000       4.142ms        terminou (nao segura garfo)
4         1000       8.003ms        terminou (nao segura garfo)

total de refeicoes: 5000
### garcom rodada 1
garcom: N=5 R=1000 pausa=1ms assentos=4

terminou sem deadlock em 25.397s (todos os 5 filosofos fizeram 1000 refeicoes)

filosofo  refeicoes  espera media   estado final
0         1000       13.569ms       terminou (nao segura garfo nem assento)
1         1000       13.778ms       terminou (nao segura garfo nem assento)
2         1000       13.725ms       terminou (nao segura garfo nem assento)
3         1000       13.49ms        terminou (nao segura garfo nem assento)
4         1000       13.993ms       terminou (nao segura garfo nem assento)

total de refeicoes: 5000
### garcom rodada 2
garcom: N=5 R=1000 pausa=1ms assentos=4

terminou sem deadlock em 25.082s (todos os 5 filosofos fizeram 1000 refeicoes)

filosofo  refeicoes  espera media   estado final
0         1000       13.511ms       terminou (nao segura garfo nem assento)
1         1000       13.306ms       terminou (nao segura garfo nem assento)
2         1000       13.707ms       terminou (nao segura garfo nem assento)
3         1000       13.401ms       terminou (nao segura garfo nem assento)
4         1000       13.349ms       terminou (nao segura garfo nem assento)

total de refeicoes: 5000
### garcom rodada 3
garcom: N=5 R=1000 pausa=1ms assentos=4

terminou sem deadlock em 26.304s (todos os 5 filosofos fizeram 1000 refeicoes)

filosofo  refeicoes  espera media   estado final
0         1000       14.582ms       terminou (nao segura garfo nem assento)
1         1000       14.71ms        terminou (nao segura garfo nem assento)
2         1000       14.539ms       terminou (nao segura garfo nem assento)
3         1000       14.604ms       terminou (nao segura garfo nem assento)
4         1000       14.905ms       terminou (nao segura garfo nem assento)

total de refeicoes: 5000

### produtor-consumidor canal -p 4 -c 4 -k 1
VERSÃO CANAL  P=4 C=4 K=1 itens/produtor=500
tempo: 658ms | throughput: 3041 itens/s | ocupação média: 0.50 de 1 | timeouts: 64
itens por consumidor: [519 552 446 483]
OK: produzidos = consumidos = 2000
### produtor-consumidor canal -p 4 -c 4 -k 10
VERSÃO CANAL  P=4 C=4 K=10 itens/produtor=500
tempo: 665ms | throughput: 3009 itens/s | ocupação média: 3.87 de 10 | timeouts: 31
itens por consumidor: [522 493 522 463]
OK: produzidos = consumidos = 2000
### produtor-consumidor canal -p 4 -c 4 -k 100
VERSÃO CANAL  P=4 C=4 K=100 itens/produtor=500
tempo: 616ms | throughput: 3246 itens/s | ocupação média: 14.95 de 100 | timeouts: 34
itens por consumidor: [487 491 468 554]
OK: produzidos = consumidos = 2000
### produtor-consumidor semaforos -p 4 -c 4 -k 1
VERSÃO SEMÁFOROS  P=4 C=4 K=1 itens/produtor=500
tempo: 692ms | throughput: 2892 itens/s | ocupação média: 0.56 de 1 | timeouts: 29
itens por consumidor: [539 495 481 485]
OK: produzidos = consumidos = 2000
### produtor-consumidor semaforos -p 4 -c 4 -k 10
VERSÃO SEMÁFOROS  P=4 C=4 K=10 itens/produtor=500
tempo: 621ms | throughput: 3220 itens/s | ocupação média: 5.39 de 10 | timeouts: 40
itens por consumidor: [516 504 495 485]
OK: produzidos = consumidos = 2000
### produtor-consumidor semaforos -p 4 -c 4 -k 100
VERSÃO SEMÁFOROS  P=4 C=4 K=100 itens/produtor=500
tempo: 618ms | throughput: 3234 itens/s | ocupação média: 55.93 de 100 | timeouts: 31
itens por consumidor: [481 529 509 481]
OK: produzidos = consumidos = 2000
