// Produtor/consumidor com canal (buffer = canal com capacidade K)
// go run -race . -p 4 -c 4 -k 10
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// simula o tempo de produzir/consumir, as vezes com uma pausa maior
func trabalho() {
	if rand.Intn(10) == 0 {
		time.Sleep(5 * time.Millisecond)
		return
	}
	time.Sleep(time.Duration(rand.Intn(500)) * time.Microsecond)
}

func produtor(id, itens int, buffer chan<- int, ocupacao *int, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < itens; i++ {
		trabalho()
		buffer <- id*itens + i // bloqueia se o buffer estiver cheio
		*ocupacao += len(buffer)
	}
}

func consumidor(buffer <-chan int, consumidos, timeouts *int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case _, ok := <-buffer:
			if !ok { // canal fechado e vazio
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

	buffer := make(chan int, *k)
	ocupacao := make([]int, *p) // um contador por goroutine pra nao dar data race
	consumidos := make([]int, *c)
	timeouts := make([]int, *c)
	var wgP, wgC sync.WaitGroup

	inicio := time.Now()
	for i := 0; i < *p; i++ {
		wgP.Add(1)
		go produtor(i, *itens, buffer, &ocupacao[i], &wgP)
	}
	for i := 0; i < *c; i++ {
		wgC.Add(1)
		go consumidor(buffer, &consumidos[i], &timeouts[i], &wgC)
	}

	// so fecha o canal depois que todos os produtores terminaram
	wgP.Wait()
	close(buffer)
	wgC.Wait()
	duracao := time.Since(inicio)

	imprimir("CANAL", *p, *c, *k, *itens, duracao, ocupacao, consumidos, timeouts)
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
