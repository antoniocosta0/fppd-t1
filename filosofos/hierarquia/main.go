// Jantar dos filosofos: hierarquia de recursos, sempre pega o garfo de menor indice primeiro
// go run -race ./filosofos/hierarquia -n 5 -r 1000 -pausa 1
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	pensando     = 0
	comOPrimeiro = 1
	comendo      = 2
	encerrado    = 3
)

func nomeEstado(e int) string {
	switch e {
	case pensando:
		return "pensando"
	case comOPrimeiro:
		return "com o primeiro garfo (esperando o segundo)"
	case comendo:
		return "comendo"
	case encerrado:
		return "terminou (nao segura garfo)"
	default:
		return "desconhecido"
	}
}

type mesa struct {
	mu       sync.Mutex
	estado   []int
	feitas   []int
	esperaNs []time.Duration
}

func novaMesa(n int) *mesa {
	return &mesa{
		estado:   make([]int, n),
		feitas:   make([]int, n),
		esperaNs: make([]time.Duration, n),
	}
}

func (m *mesa) setEstado(id, e int) {
	m.mu.Lock()
	m.estado[id] = e
	m.mu.Unlock()
}

func (m *mesa) registrarRefeicao(id int, espera time.Duration) {
	m.mu.Lock()
	m.feitas[id]++
	m.esperaNs[id] += espera
	m.mu.Unlock()
}

func (m *mesa) imprimirTabela() {
	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Println()
	fmt.Printf("%-9s %-10s %-14s %s\n", "filosofo", "refeicoes", "espera media", "estado final")
	total := 0
	for i := range m.feitas {
		n := m.feitas[i]
		total += n
		media := time.Duration(0)
		if n > 0 {
			media = m.esperaNs[i] / time.Duration(n)
		}
		fmt.Printf("%-9d %-10d %-14s %s\n", i, n, media.Round(time.Microsecond), nomeEstado(m.estado[i]))
	}
	fmt.Printf("\ntotal de refeicoes: %d\n", total)
}

func filosofo(id int, forks []chan struct{}, refeicoes int, pausa time.Duration,
	m *mesa, wg *sync.WaitGroup) {
	defer wg.Done()

	esq := id
	dir := (id + 1) % len(forks)

	// menor indice primeiro, so o ultimo filosofo inverte a ordem
	primeiro, segundo := esq, dir
	if primeiro > segundo {
		primeiro, segundo = segundo, primeiro
	}

	for rodada := 0; rodada < refeicoes; rodada++ {
		m.setEstado(id, pensando)
		time.Sleep(time.Duration(rand.Intn(10)+1) * time.Millisecond)

		t0 := time.Now()

		<-forks[primeiro] // pegar garfo = receber do canal
		m.setEstado(id, comOPrimeiro)

		time.Sleep(pausa)

		<-forks[segundo]
		espera := time.Since(t0)

		m.setEstado(id, comendo)
		time.Sleep(time.Duration(rand.Intn(10)+1) * time.Millisecond)
		m.registrarRefeicao(id, espera)

		forks[primeiro] <- struct{}{}
		forks[segundo] <- struct{}{}
	}

	m.setEstado(id, encerrado)
}

func main() {
	n := flag.Int("n", 5, "numero de filosofos (e de garfos)")
	r := flag.Int("r", 3, "refeicoes por filosofo")
	pausaMs := flag.Int("pausa", 50, "pausa entre o primeiro e o segundo garfo, em ms")
	limiteS := flag.Int("limite", 0, "rede de seguranca em s (0 = sem limite); esta versao nao deve travar")
	flag.Parse()

	if *n < 2 || *r < 1 {
		fmt.Println("erro: -n precisa ser no minimo 2 e -r no minimo 1")
		return
	}

	pausa := time.Duration(*pausaMs) * time.Millisecond

	fmt.Printf("hierarquia: N=%d R=%d pausa=%v\n", *n, *r, pausa)

	forks := make([]chan struct{}, *n)
	for i := range forks {
		forks[i] = make(chan struct{}, 1)
		forks[i] <- struct{}{} // garfo livre = token no canal
	}

	m := novaMesa(*n)
	inicio := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < *n; i++ {
		wg.Add(1)
		go filosofo(i, forks, *r, pausa, m, &wg)
	}

	// wg.Wait numa goroutine pra main nao ficar presa
	pronto := make(chan struct{})
	go func() {
		wg.Wait()
		close(pronto)
	}()

	if *limiteS > 0 {
		limite := time.Duration(*limiteS) * time.Second
		select {
		case <-pronto:
		case <-time.After(limite):
			fmt.Printf("\nTEMPO LIMITE: execucao excedeu %v (nao esperado nesta versao)\n", limite)
			m.imprimirTabela()
			return
		}
	} else {
		<-pronto
	}

	fmt.Printf("\nterminou sem deadlock em %v (todos os %d filosofos fizeram %d refeicoes)\n",
		time.Since(inicio).Round(time.Millisecond), *n, *r)
	m.imprimirTabela()
}
