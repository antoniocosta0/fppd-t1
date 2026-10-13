// T1 - Problema 2: Produtor/Consumidor com buffer limitado - versão CHANNEL
//
// O buffer é um channel com capacidade K: a semântica de bloqueio do próprio
// Go faz o trabalho de `notFull` e `notEmpty` do Ben-Ari 6.8.
//   - produtor bloqueia em `buf <- item` quando o buffer está CHEIO (K itens);
//   - consumidor bloqueia em `<-buf` quando o buffer está VAZIO.
//
// Não há data race no buffer: o channel já é seguro para várias goroutines.
//
// Encerramento ordenado: quando o último produtor termina, a main faz
// close(buf). Fechar um channel NÃO descarta o que já está dentro: os
// consumidores continuam recebendo os itens restantes e só quando o buffer
// esvazia o receive devolve ok == false (como em 03-close-range). Assim
// nenhum item se perde.
//
// Consumo com timeout: o consumidor 0 usa select + time.After. Se nenhum item
// chega dentro de -timeout ms, ele executa um comportamento alternativo
// (conta o timeout) e volta a tentar.
//
// Execução:
//
//	go run -race ./produtor-consumidor/canal -p 4 -c 4 -k 10 -itens 1000
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

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

// produtor `id` gera `itens` itens com valores únicos: id*itens + j.
// Esses valores permitem conferir no final que nenhum item foi perdido nem
// consumido duas vezes.
func produtor(id int, buf chan int, itens int, tprodUs int, s *estatisticas, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := 0; j < itens; j++ {
		dorme(tprodUs)
		buf <- id*itens + j // bloqueia se o buffer está cheio
		s.produziu(len(buf))
	}
}

// consumidor `id` retira itens até o channel ser fechado E esvaziado.
// Se `timeout` > 0 (só o consumidor 0), usa select com time.After.
func consumidor(id int, buf chan int, tconsUs int, timeout time.Duration,
	s *estatisticas, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		var item int
		var ok bool

		if timeout > 0 {
			select {
			case item, ok = <-buf:
			case <-time.After(timeout):
				// Comportamento alternativo: nenhum item chegou a tempo.
				s.houveTimeout(id)
				continue
			}
		} else {
			item, ok = <-buf // bloqueia se o buffer está vazio
		}

		if !ok {
			return // channel fechado e vazio: tudo foi drenado
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
	fmt.Printf("canal: P=%d C=%d K=%d itens/produtor=%d total=%d\n", *p, *c, *k, *itens, total)

	buf := make(chan int, *k)
	s := novasEstatisticas(total, *c)

	inicio := time.Now()

	var wgP, wgC sync.WaitGroup
	for i := 0; i < *p; i++ {
		wgP.Add(1)
		go produtor(i, buf, *itens, *tprod, s, &wgP)
	}
	for i := 0; i < *c; i++ {
		wgC.Add(1)
		t := time.Duration(0)
		if i == 0 {
			t = timeout // so o consumidor 0 usa select com timeout
		}
		go consumidor(i, buf, *tcons, t, s, &wgC)
	}

	// Encerramento ordenado: quando TODOS os produtores terminam, fecha o
	// buffer. Os consumidores drenam o que sobrou e saem.
	go func() {
		wgP.Wait()
		close(buf)
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
