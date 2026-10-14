// Jantar dos filosofos: versao base, todos pegam o garfo esquerdo e depois o direito (da deadlock)
// go run -race ./filosofos/base_deadlock -n 5 -r 3 -pausa 50 -limite 3
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
	comOEsquerdo = 1
	comendo      = 2
	encerrado    = 3
)

func nomeEstado(e int) string {
	switch e {
	case pensando:
		return "pensando"
	case comOEsquerdo:
		return "com o garfo esquerdo (esperando o direito)"
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

func (m *mesa) totalRefeicoes() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	total := 0
	for _, n := range m.feitas {
		total += n
	}
	return total
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

	for rodada := 0; rodada < refeicoes; rodada++ {
		m.setEstado(id, pensando)
		time.Sleep(time.Duration(rand.Intn(10)+1) * time.Millisecond)

		t0 := time.Now()

		<-forks[esq] // pegar garfo = receber do canal
		m.setEstado(id, comOEsquerdo)

		time.Sleep(pausa) // da tempo de todos pegarem o esquerdo

		<-forks[dir]
		espera := time.Since(t0)

		m.setEstado(id, comendo)
		time.Sleep(time.Duration(rand.Intn(10)+1) * time.Millisecond)
		m.registrarRefeicao(id, espera)

		forks[esq] <- struct{}{}
		forks[dir] <- struct{}{}
	}

	m.setEstado(id, encerrado)
}

func main() {
	n := flag.Int("n", 5, "numero de filosofos (e de garfos)")
	r := flag.Int("r", 3, "refeicoes por filosofo")
	pausaMs := flag.Int("pausa", 50, "pausa entre pegar o garfo esquerdo e o direito, em ms")
	limiteS := flag.Int("limite", 5, "tempo sem progresso (nenhuma refeicao nova) antes de declarar deadlock, em s")
	flag.Parse()

	if *n < 2 || *r < 1 {
		fmt.Println("erro: -n precisa ser no minimo 2 e -r no minimo 1")
		return
	}

	pausa := time.Duration(*pausaMs) * time.Millisecond
	limite := time.Duration(*limiteS) * time.Second

	fmt.Printf("base_deadlock: N=%d R=%d pausa=%v limite=%v\n", *n, *r, pausa, limite)

	forks := make([]chan struct{}, *n)
	for i := range forks {
		forks[i] = make(chan struct{}, 1)
		forks[i] <- struct{}{} // garfo livre = token no canal
	}

	m := novaMesa(*n)

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

	// so e deadlock se ninguem comeu durante um intervalo inteiro
	ultimo := 0
	for {
		select {
		case <-pronto:
			fmt.Printf("\nterminou sem deadlock (todos os %d filosofos fizeram %d refeicoes)\n", *n, *r)
			m.imprimirTabela()
			return

		case <-time.After(limite):
			total := m.totalRefeicoes()
			if total != ultimo {
				ultimo = total
				continue
			}
			fmt.Printf("\nDEADLOCK detectado: nenhum progresso em %v\n", limite)
			fmt.Println("as 4 condicoes de Coffman estao presentes:")
			fmt.Println("  1. exclusao mutua  : cada garfo (channel com buffer 1) tem um dono por vez")
			fmt.Println("  2. hold and wait   : cada filosofo segura o garfo esquerdo e espera o direito")
			fmt.Println("  3. nao preempcao   : nenhum filosofo tira o garfo de outro")
			fmt.Println("  4. espera circular : i espera i+1 e o ultimo espera o garfo 0")
			m.imprimirTabela()
			return
		}
	}
}
