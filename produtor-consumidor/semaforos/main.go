// T1 - Problema 2: Produtor/Consumidor com buffer limitado - versão SEMÁFOROS
//
// O buffer é uma estrutura compartilhada (array de K posições) protegida
// explicitamente, reproduzindo o algoritmo clássico do Ben-Ari 6.8 estendido
// para VÁRIOS produtores e VÁRIOS consumidores:
//
//	notFull  : semáforo contador, valor inicial K (posições livres)
//	notEmpty : semáforo contador, valor inicial 0 (itens disponíveis)
//	mutexEsc : exclusão mútua na posição de escrita (`in`)
//	mutexLei : exclusão mútua na posição de leitura (`out`)
//
// Por que os dois mutex: notFull/notEmpty só CONTAM. Com 2 produtores
// passando por notFull ao mesmo tempo, os dois leriam o mesmo `in` e
// escreveriam na mesma posição. Com 2 consumidores, os dois retirariam o mesmo
// item. Os mutex impedem isso; usar dois (um por ponta) permite que um
// produtor e um consumidor trabalhem ao mesmo tempo, em posições diferentes.
//
// Semáforo contador em Go (aula 10 v2): channel de struct{} com buffer K.
// Nº de tokens dentro = valor do semáforo. wait (P) = receber, signal (V) =
// enviar. notFull nasce CHEIO de tokens (K); notEmpty nasce VAZIO (0).
//
// Encerramento ordenado: quando os produtores terminam, a main insere C
// sentinelas (-1), uma por consumidor, pelo mesmo caminho dos itens normais.
// Como o buffer é FIFO, os sentinelas só são retirados DEPOIS de todos os
// itens reais: o consumidor que pega um sentinela encerra, e nenhum item
// se perde.
//
// Consumo com timeout: o consumidor 0 espera notEmpty com select +
// time.After; se nada chega a tempo, executa o comportamento alternativo
// (conta o timeout) e tenta de novo.
//
// Execução:
//
//	go run -race ./produtor-consumidor/semaforos -p 4 -c 4 -k 10 -itens 1000
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const sentinela = -1

// buffer limitado com semáforos.
type buffer struct {
	dados    []int
	in, out  int // próxima posição de escrita / de leitura
	notFull  chan struct{}
	notEmpty chan struct{}
	mutexEsc sync.Mutex
	mutexLei sync.Mutex
}

func novoBuffer(k int) *buffer {
	b := &buffer{
		dados:    make([]int, k),
		notFull:  make(chan struct{}, k),
		notEmpty: make(chan struct{}, k),
	}
	for i := 0; i < k; i++ {
		b.notFull <- struct{}{} // valor inicial de notFull = K
	}
	return b // notEmpty começa vazio = valor 0
}

// inserir é o `append` do produtor. Devolve a ocupação logo após inserir,
// que é o valor de notEmpty (quantos itens há no buffer).
func (b *buffer) inserir(item int) int {
	<-b.notFull // wait(notFull): bloqueia se não há posição livre

	b.mutexEsc.Lock() // só UM produtor por vez escolhe a posição
	b.dados[b.in] = item
	b.in = (b.in + 1) % len(b.dados)
	b.mutexEsc.Unlock()

	b.notEmpty <- struct{}{} // signal(notEmpty): há um item a mais
	return len(b.notEmpty)
}

// retirar é o `take` do consumidor (sem timeout).
func (b *buffer) retirar() int {
	<-b.notEmpty // wait(notEmpty): bloqueia se não há item
	return b.retirarDepoisDoWait()
}

// retirarDepoisDoWait faz o resto do `take`, depois que o consumidor já
// passou por notEmpty (seja pelo caminho normal, seja pelo select).
func (b *buffer) retirarDepoisDoWait() int {
	b.mutexLei.Lock() // só UM consumidor por vez escolhe o item
	item := b.dados[b.out]
	b.out = (b.out + 1) % len(b.dados)
	b.mutexLei.Unlock()

	b.notFull <- struct{}{} // signal(notFull): há uma posição livre a mais
	return item
}

// estatisticas guarda os contadores compartilhados entre as goroutines. Todo
// acesso passa pelo mutex (mesma ideia do monitor PC visto na aula 10).
type estatisticas struct {
	mu         sync.Mutex
	visto      []int // quantas vezes cada item foi consumido (deve ser 1)
	consumidos []int // itens consumidos por consumidor
	timeouts   []int // timeouts por consumidor
	produzidos int
	somaOcup   int // soma das ocupações amostradas
	amostras   int
}

func novasEstatisticas(total, c int) *estatisticas {
	return &estatisticas{
		visto:      make([]int, total),
		consumidos: make([]int, c),
		timeouts:   make([]int, c),
	}
}

// produziu registra um item produzido e a ocupação do buffer logo após inserir.
func (s *estatisticas) produziu(ocupacao int) {
	s.mu.Lock()
	s.produzidos++
	s.somaOcup += ocupacao
	s.amostras++
	s.mu.Unlock()
}

func (s *estatisticas) consumiu(id, item int) {
	s.mu.Lock()
	s.visto[item]++
	s.consumidos[id]++
	s.mu.Unlock()
}

func (s *estatisticas) houveTimeout(id int) {
	s.mu.Lock()
	s.timeouts[id]++
	s.mu.Unlock()
}

// dorme simula trabalho: de 0 até 2*mediaUs microssegundos (média = mediaUs).
func dorme(mediaUs int) {
	if mediaUs > 0 {
		time.Sleep(time.Duration(rand.Intn(2*mediaUs+1)) * time.Microsecond)
	}
}

func produtor(id int, b *buffer, itens int, tprodUs int, s *estatisticas, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := 0; j < itens; j++ {
		dorme(tprodUs)
		oc := b.inserir(id*itens + j)
		s.produziu(oc)
	}
}

func consumidor(id int, b *buffer, tconsUs int, timeout time.Duration,
	s *estatisticas, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		var item int

		if timeout > 0 {
			select {
			case <-b.notEmpty: // wait(notEmpty) dentro do select
				item = b.retirarDepoisDoWait()
			case <-time.After(timeout):
				// Comportamento alternativo: nenhum item chegou a tempo.
				s.houveTimeout(id)
				continue
			}
		} else {
			item = b.retirar()
		}

		if item == sentinela {
			return // sentinela: fim da produção, buffer já drenado
		}

		dorme(tconsUs)
		s.consumiu(id, item)
	}
}

func main() {
	p := flag.Int("p", 4, "numero de produtores (P)")
	c := flag.Int("c", 4, "numero de consumidores (C)")
	k := flag.Int("k", 10, "capacidade do buffer (K)")
	itens := flag.Int("itens", 1000, "itens produzidos por cada produtor")
	tprod := flag.Int("tprod", 0, "tempo medio de producao de um item, em microssegundos")
	tcons := flag.Int("tcons", 100, "tempo medio de consumo de um item, em microssegundos")
	timeoutMs := flag.Int("timeout", 50, "timeout do consumidor 0 (select + time.After), em ms; 0 desliga")
	limiteS := flag.Int("limite", 60, "rede de seguranca em s: se estourar, o programa avisa e sai")
	flag.Parse()

	if *p < 1 || *c < 1 || *k < 1 || *itens < 1 {
		fmt.Println("erro: -p, -c, -k e -itens precisam ser no minimo 1")
		return
	}

	total := *p * *itens
	timeout := time.Duration(*timeoutMs) * time.Millisecond
	fmt.Printf("semaforos: P=%d C=%d K=%d itens/produtor=%d total=%d\n", *p, *c, *k, *itens, total)

	b := novoBuffer(*k)
	s := novasEstatisticas(total, *c)

	inicio := time.Now()

	var wgP, wgC sync.WaitGroup
	for i := 0; i < *p; i++ {
		wgP.Add(1)
		go produtor(i, b, *itens, *tprod, s, &wgP)
	}
	for i := 0; i < *c; i++ {
		wgC.Add(1)
		t := time.Duration(0)
		if i == 0 {
			t = timeout // so o consumidor 0 usa select com timeout
		}
		go consumidor(i, b, *tcons, t, s, &wgC)
	}

	// Encerramento ordenado: depois que TODOS os produtores terminam, insere
	// um sentinela por consumidor. Como o buffer é FIFO, eles ficam atrás de
	// todos os itens reais.
	go func() {
		wgP.Wait()
		for i := 0; i < *c; i++ {
			b.inserir(sentinela)
		}
	}()

	pronto := make(chan struct{})
	go func() {
		wgC.Wait()
		close(pronto)
	}()

	select {
	case <-pronto:
	case <-time.After(time.Duration(*limiteS) * time.Second):
		fmt.Printf("\nTRAVOU: sem progresso em %ds (nao esperado)\n", *limiteS)
		return
	}
	dur := time.Since(inicio)

	// Relatório final (sob o lock) e verificação: total produzido == total
	// consumido e cada item consumido exatamente 1 vez.
	s.mu.Lock()
	defer s.mu.Unlock()

	totalCons := 0
	for _, n := range s.consumidos {
		totalCons += n
	}
	erros := 0
	for _, v := range s.visto {
		if v != 1 {
			erros++
		}
	}

	fmt.Printf("\n%-11s %-10s %s\n", "consumidor", "itens", "timeouts")
	for i := range s.consumidos {
		fmt.Printf("%-11d %-10d %d\n", i, s.consumidos[i], s.timeouts[i])
	}

	ocupMedia := 0.0
	if s.amostras > 0 {
		ocupMedia = float64(s.somaOcup) / float64(s.amostras)
	}
	fmt.Printf("\ntempo total        : %v\n", dur.Round(time.Millisecond))
	fmt.Printf("throughput         : %.0f itens/s\n", float64(totalCons)/dur.Seconds())
	fmt.Printf("ocupacao media     : %.2f de %d (medida logo apos cada insercao)\n", ocupMedia, *k)
	fmt.Printf("produzidos         : %d\n", s.produzidos)
	fmt.Printf("consumidos         : %d\n", totalCons)

	if s.produzidos == totalCons && totalCons == total && erros == 0 {
		fmt.Println("verificacao        : OK (produzido == consumido, nenhum item perdido ou repetido)")
	} else {
		fmt.Printf("verificacao        : FALHOU (itens com contagem != 1: %d)\n", erros)
	}
}
