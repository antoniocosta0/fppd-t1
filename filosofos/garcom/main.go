// T1 - Problema 1: Jantar dos Filósofos - estratégia 2: GARÇOM (arbitragem)
//
// Em vez de cada filósofo sentar à mesa quando quiser, ele precisa de
// PERMISSÃO do garçom. O garçom é um channel de "assentos" com capacidade
// N-1: no máximo N-1 filósofos disputam garfos ao mesmo tempo.
//
// Por que isso elimina o deadlock: a espera circular só fecha quando os N
// filósofos seguram um garfo cada. Com no máximo N-1 na disputa, pelo menos
// um deles tem os dois garfos livres, come e devolve. Quebra a condição de
// Coffman 4 (espera circular). A ordem de aquisição continua sendo a mesma
// do base_deadlock (esquerdo e depois direito), inclusive com a mesma pausa:
// é isso que prova que o ganho vem do garçom, não da ordem dos garfos.
//
// Referências da aula: slides da aula 10 v2 ("limitar participantes,
// seats = N-1") e o enunciado do T1 ("arbitragem centralizada por um
// processo garçom").
//
// Execução:
//
//	go run -race ./filosofos/garcom -n 5 -r 1000 -pausa 1
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
	pensando         = 0
	esperandoAssento = 1
	comOEsquerdo     = 2
	comendo          = 3
	encerrado        = 4
)

func nomeEstado(e int) string {
	switch e {
	case pensando:
		return "pensando"
	case esperandoAssento:
		return "esperando assento (garcom)"
	case comOEsquerdo:
		return "com o garfo esquerdo (esperando o direito)"
	case comendo:
		return "comendo"
	case encerrado:
		return "terminou (nao segura garfo nem assento)"
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
	esperaNs []time.Duration // tempo total de espera (assento + garfos)
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
//
// Diferença em relação ao base_deadlock: antes de tocar em qualquer garfo,
// o filósofo ocupa um assento (pede permissão ao garçom) e, depois de
// devolver os garfos, libera o assento.
//
//   - assentos : channel com buffer N-1. Ocupar assento = ENVIAR (bloqueia
//     quando já há N-1 sentados); liberar = RECEBER.
func filosofo(id int, forks []chan struct{}, assentos chan struct{}, refeicoes int,
	pausa time.Duration, m *mesa, wg *sync.WaitGroup) {
	defer wg.Done()

	esq := id
	dir := (id + 1) % len(forks)

	for rodada := 0; rodada < refeicoes; rodada++ {
		// a) pensar
		m.setEstado(id, pensando)
		time.Sleep(time.Duration(rand.Intn(10)+1) * time.Millisecond)

		// b) t0 ANTES de pedir o assento: a espera medida inclui a fila do
		//    garçom, que é parte do custo desta estratégia.
		t0 := time.Now()

		// c) pedir permissão ao garçom (ocupar um assento)
		m.setEstado(id, esperandoAssento)
		assentos <- struct{}{}

		// d) pegar o garfo esquerdo, pausar, pegar o direito
		<-forks[esq]
		m.setEstado(id, comOEsquerdo)
		time.Sleep(pausa)
		<-forks[dir]

		// e) espera concluída: já tem os DOIS garfos
		espera := time.Since(t0)

		// f) comer
		m.setEstado(id, comendo)
		time.Sleep(time.Duration(rand.Intn(10)+1) * time.Millisecond)
		m.registrarRefeicao(id, espera)

		// g) devolver os dois garfos e LIBERAR o assento. Esquecer de
		//    liberar o assento criaria um deadlock novo, causado pelo
		//    próprio garçom.
		forks[esq] <- struct{}{}
		forks[dir] <- struct{}{}
		<-assentos
	}

	m.setEstado(id, encerrado)
}

func main() {
	n := flag.Int("n", 5, "numero de filosofos (e de garfos)")
	r := flag.Int("r", 3, "refeicoes por filosofo")
	pausaMs := flag.Int("pausa", 50, "pausa entre pegar o garfo esquerdo e o direito, em ms")
	limiteS := flag.Int("limite", 0, "rede de seguranca em s (0 = sem limite); esta versao nao deve travar")
	flag.Parse()

	if *n < 2 || *r < 1 {
		fmt.Println("erro: -n precisa ser no minimo 2 e -r no minimo 1")
		return
	}

	pausa := time.Duration(*pausaMs) * time.Millisecond

	fmt.Printf("garcom: N=%d R=%d pausa=%v assentos=%d\n", *n, *r, pausa, *n-1)

	// Garfos: channel com buffer 1, token pré-carregado = garfo livre.
	forks := make([]chan struct{}, *n)
	for i := range forks {
		forks[i] = make(chan struct{}, 1)
		forks[i] <- struct{}{}
	}

	// Assentos: channel com buffer N-1 e VAZIO no início (não há token
	// inicial). Ocupar = enviar, liberar = receber. O buffer cheio é o que
	// faz o N-ésimo filósofo esperar.
	assentos := make(chan struct{}, *n-1)

	m := novaMesa(*n)
	inicio := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < *n; i++ {
		wg.Add(1)
		go filosofo(i, forks, assentos, *r, pausa, m, &wg)
	}

	// Encerramento ordenado: todos os filósofos terminam as R refeições,
	// devolvem garfos e assento, e o wg.Wait() retorna. Nenhuma goroutine
	// fica bloqueada. O watchdog (-limite > 0) é só rede de segurança.
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
			fmt.Printf("\nTRAVOU: nenhum progresso em %v (nao esperado nesta versao)\n", limite)
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
