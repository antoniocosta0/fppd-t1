// T1 - Problema 1: Jantar dos Filósofos - versão BASE (demonstra deadlock)
//
// Objetivo desta versão: provar que o deadlock ACONTECE.
// Todos os filósofos pegam primeiro o garfo da esquerda e depois o da direita,
// com uma pausa entre os dois: isso fecha a espera circular.
//
// Referências da aula: exemplos 03-canais/filosofos (garfo = chan struct{}
// com buffer 1 e token) e Ben-Ari, Algoritmo 6.10 / 8.5 (todos pegam
// forks[i] antes de forks[i+1] -> a espera circular continua possível).
//
// Execução:
//
//	go run -race ./filosofos/base_deadlock -n 5 -r 3 -pausa 50 -limite 3
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Estados do filósofo.
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

// mesa guarda os dados compartilhados entre os filósofos e a main. Todo
// acesso passa pelo mutex (mesma ideia do monitor PC visto na aula 10): os
// filósofos escrevem e a main lê, então sem o lock haveria data race.
type mesa struct {
	mu       sync.Mutex
	estado   []int           // estado atual de cada filósofo
	feitas   []int           // refeições concluídas por filósofo
	esperaNs []time.Duration // tempo total de espera pelos garfos
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

// registrarRefeicao conta uma refeição concluída e soma a espera dela.
func (m *mesa) registrarRefeicao(id int, espera time.Duration) {
	m.mu.Lock()
	m.feitas[id]++
	m.esperaNs[id] += espera
	m.mu.Unlock()
}

// imprimirTabela mostra os dados de execução. É chamada nos dois finais
// possíveis: deadlock detectado ou término normal.
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

// filosofo é o processo do filósofo `id`.
// Sequência do Ben-Ari: pensar -> pegar esquerdo -> pegar direito -> comer ->
// devolver os dois garfos.
//
//   - id        : índice do filósofo
//   - forks     : os N garfos (garfo livre = token dentro do channel)
//   - refeicoes : quantas vezes repetir (o "R" do enunciado)
//   - pausa     : tempo entre pegar o esquerdo e tentar o direito
//   - m         : dados compartilhados (estado, refeições, espera)
//   - wg        : avisa a main que este filósofo acabou
func filosofo(id int, forks []chan struct{}, refeicoes int, pausa time.Duration,
	m *mesa, wg *sync.WaitGroup) {
	defer wg.Done()

	// Mesa circular: o garfo da direita do último filósofo é o garfo 0.
	esq := id
	dir := (id + 1) % len(forks)

	for rodada := 0; rodada < refeicoes; rodada++ {
		// a) pensar
		m.setEstado(id, pensando)
		time.Sleep(time.Duration(rand.Intn(10)+1) * time.Millisecond)

		// t0 marcado ANTES de disputar os garfos: é isso que o enunciado
		// chama de "tempo de espera pelos garfos".
		t0 := time.Now()

		// b) pegar o garfo ESQUERDO. Pegar = receber do channel; se o garfo
		//    está na mão de outro filósofo, o receive bloqueia.
		<-forks[esq]
		m.setEstado(id, comOEsquerdo)

		// c) a PAUSA é a peça que provoca o deadlock: dá tempo de todos os N
		//    filósofos pegarem o seu garfo esquerdo antes de qualquer um
		//    tentar o direito. Com -pausa 0 o deadlock vira "às vezes".
		time.Sleep(pausa)

		// d) pegar o garfo DIREITO. Quando há deadlock, TODOS ficam parados
		//    exatamente aqui, cada um segurando um garfo (hold and wait) e
		//    esperando o garfo do vizinho (espera circular).
		<-forks[dir]
		espera := time.Since(t0)

		// e) comer
		m.setEstado(id, comendo)
		time.Sleep(time.Duration(rand.Intn(10)+1) * time.Millisecond)
		m.registrarRefeicao(id, espera)

		// f) devolver os dois garfos (enviar o token de volta). A ordem de
		//    devolução não muda o deadlock: ele acontece na AQUISIÇÃO.
		forks[esq] <- struct{}{}
		forks[dir] <- struct{}{}
	}

	m.setEstado(id, encerrado)
}

func main() {
	n := flag.Int("n", 5, "numero de filosofos (e de garfos)")
	r := flag.Int("r", 3, "refeicoes por filosofo")
	pausaMs := flag.Int("pausa", 50, "pausa entre pegar o garfo esquerdo e o direito, em ms")
	limiteS := flag.Int("limite", 5, "tempo maximo de execucao antes de declarar deadlock, em s")
	flag.Parse()

	if *n < 2 || *r < 1 {
		fmt.Println("erro: -n precisa ser no minimo 2 e -r no minimo 1")
		return
	}

	pausa := time.Duration(*pausaMs) * time.Millisecond
	limite := time.Duration(*limiteS) * time.Second

	fmt.Printf("base_deadlock: N=%d R=%d pausa=%v limite=%v\n", *n, *r, pausa, limite)

	// Garfos: channel com buffer 1. Buffer 1 = um dono por vez (exclusão
	// mútua). Token pré-carregado = garfo LIVRE (há algo para receber).
	forks := make([]chan struct{}, *n)
	for i := range forks {
		forks[i] = make(chan struct{}, 1)
		forks[i] <- struct{}{}
	}

	m := novaMesa(*n)

	var wg sync.WaitGroup
	for i := 0; i < *n; i++ {
		wg.Add(1) // sempre ANTES do `go`
		go filosofo(i, forks, *r, pausa, m, &wg)
	}

	// Watchdog: a main NÃO pode chamar wg.Wait() direto, senão fica presa no
	// deadlock e o terminal fica mudo. O Wait roda em outra goroutine e
	// fecha `pronto`; a main escolhe entre `pronto` e o estouro do tempo.
	pronto := make(chan struct{})
	go func() {
		wg.Wait()
		close(pronto)
	}()

	select {
	case <-pronto:
		// Acontece com pausa 0 ou pequena: o deadlock é POSSÍVEL, não certo.
		fmt.Printf("\nterminou sem deadlock (todos os %d filosofos fizeram %d refeicoes)\n", *n, *r)
		m.imprimirTabela()

	case <-time.After(limite):
		fmt.Printf("\nDEADLOCK detectado: nenhum progresso em %v\n", limite)
		fmt.Println("as 4 condicoes de Coffman estao presentes:")
		fmt.Println("  1. exclusao mutua  : cada garfo (channel com buffer 1) tem um dono por vez")
		fmt.Println("  2. hold and wait   : cada filosofo segura o garfo esquerdo e espera o direito")
		fmt.Println("  3. nao preempcao   : nenhum filosofo tira o garfo de outro")
		fmt.Println("  4. espera circular : i espera i+1 e o ultimo espera o garfo 0")
		m.imprimirTabela()
		// Ao retornar da main o programa termina, mesmo com as goroutines
		// dos filósofos bloqueadas no deadlock.
	}
}
