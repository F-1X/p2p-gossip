package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/F-1X/p2p-gossip/p2p"
)

func spawnServer(addr string) *p2p.Server {
	cfg := p2p.ServerConfig{
		ListenAddr: addr,
	}
	srv := p2p.NewServer(cfg)
	go srv.Start()
	time.Sleep(100*time.Millisecond)
	return srv
}

func main(){
	// playerA := spawnServer(":3000")
	// playerB := spawnServer(":4000")
	// playerC := spawnServer(":5000")
	// // playerD := spawnServer(":6000")

	// // //_ = playerA
	// // time.Sleep(500*time.Millisecond)
	// playerB.Connect(playerA.ListenAddr)

	// time.Sleep(300*time.Millisecond)
	// playerC.Connect(playerB.ListenAddr)
	// time.Sleep(300*time.Millisecond)

	// fmt.Println("A:",playerA.Peers(),"B:",playerB.Peers(),"C:",playerC.Peers())

	// playerD.Connect(playerA.ListenAddr)
	// time.Sleep(200*time.Millisecond)

	// fmt.Println(playerA.Peers(),playerB.Peers(),playerC.Peers(),playerD.Peers())

	// testPeerConnected(playerA)
	// testPeerConnected(playerB)
	// testPeerConnected(playerC)
	// testPeerConnected(playerD)

	

	s, err := spawnManyServers(10)
	if err != nil {
		fmt.Println(err)
	}

	err = ConnectAllServersSequentially(s)
	if err != nil {
		fmt.Println(err)
	}


	// count := 2
	// randomIndexes := generateRandomIndexes(len(s), count)

	// //deleteSomePeers(s,randomIndexes)
	

	// fmt.Println(randomIndexes)


	ticker := time.NewTicker(time.Second * 5)

	go func() {
		for {
			select {
			case <- ticker.C:
				printAllPeers(s)
				fmt.Println()
			default:

			}
		}
	}()

	select {}

}


func printAllPeers(s []*p2p.Server) {
	for i :=0; i < len(s); i++{
		fmt.Println(s[i].ListenAddr,": => ", len(s[i].Peers()))
	}
}

func spawnManyServers(count int) ([]*p2p.Server,error) {
	port := 3000
	s := make([]*p2p.Server,count)
	for i := 0 ; i < count; i++ {
		portStr := ":" + strconv.Itoa(port)
		s[i] = spawnServer(portStr)
		port++

	}
	return s,nil
	
}

func ConnectAllServersSequentially(s []*p2p.Server) error {
	for i := 1; i<len(s); i++ {
		err := s[i].Connect(s[i-1].ListenAddr)
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond*300)
	}
	return nil
}

type Servers []p2p.Server



// func shutdown(s []*p2p.Server, scope []int) {

// 	for i:= 0; i< len(scope); i++ {
// 		if err := s[scope[i]].Close(); err != nil {
// 			continue
// 		}
// 	}
	
// }

// func deleteSomePeers(s []*p2p.Server, randomIndexes []int) {
	
// 	fmt.Println(randomIndexes)
// 	shutdown(s,randomIndexes)
// }

func init() {

}

func generateRandomIndexes(s int, count int) []int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]int, count)
	selected := make(map[int]bool)

	for i := 0; i < count; i++ {
		var index int
		for {
			index = r.Intn(s)
			if !selected[index] {
				selected[index] = true
				result[i] = index
				break
			}
			
		}
	}
	return result
}