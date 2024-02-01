package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)


type ServerConfig struct {
	ListenAddr string
}


type Server struct {
	ServerConfig
	peers sync.Map
	addPeer chan *Peer
	delPeer chan *Peer
	msgCh chan *Message
	transport *TCPTransport
}

func NewServer(cfg ServerConfig) *Server {
	s := &Server{
		ServerConfig: cfg,
		addPeer: make(chan *Peer, 100),
		msgCh: make(chan *Message, 100),
		delPeer: make(chan *Peer,100),
	}

	tr := NewTCPTransport(s.ListenAddr)
	s.transport = tr
	tr.AddPeer = s.addPeer
	tr.DelPeer = s.delPeer

	return s
}

func (s *Server) Start() {
	go s.loop()
	logrus.WithFields(logrus.Fields{
		"port": s.ListenAddr,
	}).Info("started new game server")

	s.transport.ListenAndAccept()

	
}
func (s *Server) sendPeerList(p *Peer) error {
	peerList := MessagePeerList{
		Peers: []string{},
	}

	s.peers.Range(func(key, value interface{}) bool {
		addr, ok := key.(string)
		if ok && addr != p.listenAddr {
			peerList.Peers = append(peerList.Peers, addr)
		}
		return true
	})


	msg := NewMessage(s.ListenAddr, peerList)
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	return p.Send(buf.Bytes())
}


func (s *Server) AddPeer(p *Peer) {
	s.peers.Store(p.listenAddr, p)
}

func (s *Server) Peers() []string {
	peers := make([]string, 0)
	s.peers.Range(func(key, _ interface{}) bool {
		addr, ok := key.(string)
		if ok {
			peers = append(peers, addr)
		}
		return true
	})
	return peers
}

func (s *Server) SendHandshake(p *Peer) error {
	hs := &Handshake{
		ListenAddr: s.ListenAddr,
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(hs); err != nil {
		return err
	}
	return p.Send(buf.Bytes())
}


func (s *Server) isInPeerList(addr string) bool {
	_, ok := s.peers.Load(addr)
	return ok
}	


func (s *Server) Connect(addr string) error {
	if s.isInPeerList(addr) {
		return nil
	}
	conn, err := net.DialTimeout("tcp", addr, time.Millisecond * 1000)
	if err != nil {
		return err
	}
	peer := &Peer{
		conn: conn,
		outbound: true,
	}
	s.addPeer <- peer
	
	return s.SendHandshake(peer)			
}

func (s *Server) loop() {
	for {
		select {
		case peer := <-s.delPeer:
			logrus.WithFields(logrus.Fields{
				"addr": peer.conn.RemoteAddr(),
			}).Info("player disconnected")

			s.peers.Delete(peer.listenAddr)
			
		case peer := <-s.addPeer:
			if err := s.handleNewPeer(peer); err != nil {
				logrus.Errorf("handle new peer err: %s", err)
			}

		case msg := <- s.msgCh:
			go func() {
				logrus.WithFields(logrus.Fields{
					"localAddr": s.ListenAddr,
					"msg.From:": msg.From,
				}).Info("new message recieved")
				if err := s.handleMessage(msg); err != nil {
					logrus.Errorf("handle msg err: %s", err)
				}
				
			}()
		}
	}
}

func (s *Server) handleNewPeer(peer *Peer) error {
	_,err := s.handshake(peer)
	if err != nil {
		peer.conn.Close()
		s.peers.Delete(peer.listenAddr)
		return fmt.Errorf("failed handshake %s",err.Error())
	}
	
	go peer.ReadLoop(s.msgCh)

	if !peer.outbound {
		if err := s.SendHandshake(peer); err != nil {
			peer.conn.Close()
			s.peers.Delete(peer.listenAddr)
			return fmt.Errorf("failed send handshake with peer: %v", peer)
		}
		go func() {
			if err := s.sendPeerList(peer); err != nil {
				logrus.Errorf("failed send peer list: %s", err)
			}
		}()
	}
	s.AddPeer(peer)
	//s.gameState.AddPlayer(peer.listenAddr,hs.GameStatus)

	logrus.WithFields(logrus.Fields{
		"peer.Local": peer.conn.LocalAddr(),
		"peer.Remote": peer.conn.RemoteAddr(),
		"listenAddr": peer.listenAddr,
		"we": s.ListenAddr,
		"connectedPeers": len(s.Peers()),
	}).Info("handshake successfull: new player connected.")
	return nil
}

func (s *Server) handshake(p *Peer) (*Handshake,error) {
	hs := &Handshake{}
	if err := gob.NewDecoder(p.conn).Decode(hs); err != nil {
		return nil,err
	}

	p.listenAddr = hs.ListenAddr

	return hs,nil
}


func (s *Server) handleMessage(msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessagePeerList:
		logrus.WithFields(logrus.Fields{
			"s.ListenAddr": s.ListenAddr,
			"msg.From": msg.From,
			"msg.Payload": msg.Payload,
		}).Info("recived peer list")
		return s.handlePeerList(v)
	}
	return nil
}


func (s *Server) handlePeerList(l MessagePeerList) error {
	logrus.WithFields(logrus.Fields{
		"s.ListenAddr": s.ListenAddr,
		"l.Peers": l.Peers,
	}).Info("handle peer list")
	for i := 0; i < len(l.Peers); i++ {
		if err := s.Connect(l.Peers[i]); err != nil {
			logrus.Errorf("failed to dial peer: %s", err)
			continue
		}
	}
	return nil
}


func init() {
	gob.Register(MessagePeerList{})
}

