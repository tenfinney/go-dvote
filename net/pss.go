package net

import (
	"fmt"
	"time"
	"errors"

	"github.com/vocdoni/go-dvote/swarm"
	"github.com/vocdoni/go-dvote/types"
)

type PSSHandle struct {
	c *types.Connection
	s *swarm.SimpleSwarm
}

func (p *PSSHandle) Init(c *types.Connection) error {
	p.c = c
	sn := new(swarm.SimpleSwarm)
	err := sn.Init()
	if err != nil {
		return err
	}
	err = sn.SetLog("crit")
	if err != nil {
		return err
	}
	sn.PssSub(p.c.Encryption, p.c.Key, p.c.Topic, p.c.Address)
	p.s = sn
	fmt.Println("pss init")
	fmt.Println("%v", p)
	return nil
}

func (p *PSSHandle) Listen(reciever chan<- types.Message, errorReciever chan<- error) {
	fmt.Println("%v", p)
	var msg types.Message
	for {
		select {
		case pssMessage := <-p.s.PssTopics[p.c.Topic].Delivery:
			msg.Topic = p.c.Topic
			msg.Data = pssMessage.Msg
			msg.Address = pssMessage.Peer.String()
			msg.TimeStamp = time.Now()
			reciever <- msg
		default:
			errorReciever <- errors.New("no msg")
		}

	}
}

func (p *PSSHandle) Send(msg []byte, errors chan<- error) {

	err := p.s.PssPub(p.c.Encryption, p.c.Key, p.c.Topic, string(msg), p.c.Address)
	if err != nil {
		errors <- err
	}
}