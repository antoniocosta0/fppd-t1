// Produtor/consumidor com semaforos 
// go run -race . -p 4 -c 4 -k 10
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// semaforo feito com canal
type Sem chan struct{}

func NewSem(n int) Sem {
	s := make(Sem, n)
	for i := 0; i < n; i++ {
		s <- struct{}{}
	}
	return s
}
func (s Sem) Acquire() { <-s }             // wait
func (s Sem) Release() { s <- struct{}{} } // signal

const FIM = -1 // avisar o consumidor que acabou

type Buffer struct {
	itens             []int
	in, out           int
	notFull, notEmpty Sem
	mutex             Sem
}

func NovoBuffer(k int) *Buffer {
	// notEmpty tem que comecar em 0 mas aceitar ate K, entao cria com K e esvazia
	notEmpty := NewSem(k)
	for i := 0; i < k; i++ {
		notEmpty.Acquire()
	}
	return &Buffer{
		itens:    make([]int, k),
		notFull:  NewSem(k),
		notEmpty: notEmpty,
		mutex:    NewSem(1),
	}
}

func (b *Buffer) Inserir(v int) {
	b.notFull.Acquire()
	b.mutex.Acquire()
	b.itens[b.in] = v
	b.in = (b.in + 1) % len(b.itens)
	b.mutex.Release()
	b.notEmpty.Release()
}

func (b *Buffer) retirar() int {
	b.mutex.Acquire()
	v := b.itens[b.out]
	b.out = (b.out + 1) % len(b.itens)
	b.mutex.Release()
	b.notFull.Release()
	return v
}

// simular o tempo de produzir/consumir, as vezes com uma pausa maior
func trabalho() {
	if rand.Intn(10) == 0 {
		time.Sleep(5 * time.Millisecond)
		return
	}
	time.Sleep(time.Duration(rand.Intn(500)) * time.Microsecond)
}

func produtor(id, itens int, b *Buffer, ocupacao *int, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < itens; i++ {
		trabalho()
		b.Inserir(id*itens + i)
		*ocupacao += len(b.notEmpty) // itens no buffer agora
	}
}

func consumidor(b *Buffer, consumidos, timeouts *int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-b.notEmpty: // Acquire com timeout
			v := b.retirar()
			if v == FIM {
				return
			}
			trabalho()
			*consumidos++
		case <-time.After(2 * time.Millisecond):
			*timeouts++
		}
	}
}

func main() {
	p := flag.Int("p", 4, "produtores (P)")
	c := flag.Int("c", 4, "consumidores (C)")
	k := flag.Int("k", 10, "capacidade do buffer (K)")
	itens := flag.Int("itens", 500, "itens por produtor")
	flag.Parse()

	b := NovoBuffer(*k)
	ocupacao := make([]int, *p)
	consumidos := make([]int, *c)
	timeouts := make([]int, *c)
	var wgP, wgC sync.WaitGroup

	inicio := time.Now()
	for i := 0; i < *p; i++ {
		wgP.Add(1)
		go produtor(i, *itens, b, &ocupacao[i], &wgP)
	}
	for i := 0; i < *c; i++ {
		wgC.Add(1)
		go consumidor(b, &consumidos[i], &timeouts[i], &wgC)
	}

	// um FIM por consumidor, depois dos itens reais
	wgP.Wait()
	for i := 0; i < *c; i++ {
		b.Inserir(FIM)
	}
	wgC.Wait()
	duracao := time.Since(inicio)

	imprimir("SEMÁFOROS", *p, *c, *k, *itens, duracao, ocupacao, consumidos, timeouts)
}

func imprimir(versao string, p, c, k, itens int, d time.Duration, ocup, cons, tmo []int) {
	total, somaOcup, somaTmo := 0, 0, 0
	for i := range cons {
		total += cons[i]
		somaTmo += tmo[i]
	}
	for _, o := range ocup {
		somaOcup += o
	}
	fmt.Printf("VERSÃO %s  P=%d C=%d K=%d itens/produtor=%d\n", versao, p, c, k, itens)
	fmt.Printf("tempo: %v | throughput: %.0f itens/s | ocupação média: %.2f de %d | timeouts: %d\n",
		d.Round(time.Millisecond), float64(total)/d.Seconds(),
		float64(somaOcup)/float64(p*itens), k, somaTmo)
	fmt.Println("itens por consumidor:", cons)
	if total == p*itens {
		fmt.Printf("OK: produzidos = consumidos = %d\n", total)
	} else {
		fmt.Printf("ERRO: produzidos = %d, consumidos = %d\n", p*itens, total)
	}
}
