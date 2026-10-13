// T1 - Problema 1: Jantar dos Filósofos - estratégia 1: HIERARQUIA DE RECURSOS
//
// Quebra de simetria: em vez de "esquerdo e depois direito", todo filósofo
// pega sempre o garfo de MENOR índice primeiro e o de MAIOR índice depois.
// O único filósofo que muda de comportamento é o último (N-1), porque para
// ele o garfo 0 é o da direita. Essa assimetria quebra a condição de Coffman
// 4 (espera circular): não existe mais um ciclo i -> i+1 -> ... -> i.
//
// Referências da aula: exemplo 03-canais/filosofos (o filósofo 0 inverte a
// ordem), Ben-Ari (quebra de simetria) e slides da aula 10 v2 (Coffman).
//
// Execução:
//
//	go run -race ./filosofos/hierarquia -n 5 -r 1000 -pausa 1
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

// mesa guarda os dados compartilhados entre os filósofos e a main. Todo
// acesso passa pelo mutex (mesma ideia do monitor PC visto na aula 10).
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

// imprimirTabela são os dados de execução usados na análise de fairness.
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
// Diferença em relação ao base_deadlock: a ORDEM de aquisição dos garfos.
func filosofo(id int, forks []chan struct{}, refeicoes int, pausa time.Duration,
	m *mesa, wg *sync.WaitGroup) {
	defer wg.Done()

	esq := id
	dir := (id + 1) % len(forks)

	// Hierarquia: primeiro o garfo de menor índice, depois o de maior.
	primeiro, segundo := esq, dir
	if primeiro > segundo {
		primeiro, segundo = segundo, primeiro
	}

	for rodada := 0; rodada < refeicoes; rodada++ {
		// a) pensar
		m.setEstado(id, pensando)
		time.Sleep(time.Duration(rand.Intn(10)+1) * time.Millisecond)

		// t0 ANTES de disputar os garfos.
		t0 := time.Now()

		// b) pegar o garfo de MENOR índice
		<-forks[primeiro]
		m.setEstado(id, comOPrimeiro)

		// c) a mesma pausa que trava o base_deadlock. Aqui ela NÃO trava:
		//    é a prova de que a hierarquia elimina o deadlock.
		time.Sleep(pausa)

		// d) pegar o garfo de MAIOR índice
		<-forks[segundo]
		espera := time.Since(t0)

		// e) comer
		m.setEstado(id, comendo)
		time.Sleep(time.Duration(rand.Intn(10)+1) * time.Millisecond)
		m.registrarRefeicao(id, espera)

		// f) devolver os dois garfos
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

	// Garfos: channel com buffer 1, token pré-carregado = garfo livre.
	forks := make([]chan struct{}, *n)
	for i := range forks {
		forks[i] = make(chan struct{}, 1)
		forks[i] <- struct{}{}
	}

	m := novaMesa(*n)
	inicio := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < *n; i++ {
		wg.Add(1)
		go filosofo(i, forks, *r, pausa, m, &wg)
	}

	// Encerramento ordenado: aqui o wg.Wait() SEMPRE retorna, porque não há
	// deadlock. O watchdog fica como rede de segurança para os testes
	// (-limite > 0): se algum dia travar, o programa avisa em vez de ficar mudo.
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
