package tran

import (
	"fmt"
	ssh "github.com/flynn-archive/go-crypto-ssh"
	"io"
	"log"
	"net"
)

type Board struct {
	account   string
	pwd       string
	addr      string
	sqlAddr   string
	conn      *ssh.Client
	localPort int
	remote    net.Conn
	local     net.Listener
}

func NewBoard(addr, account, pwd, sqlAddr string, localPort int) *Board {
	return &Board{
		account:   account,
		pwd:       pwd,
		addr:      addr,
		sqlAddr:   sqlAddr,
		localPort: localPort,
	}
}

func (b *Board) Start() error {

	config := &ssh.ClientConfig{
		User: b.account,
		Auth: []ssh.AuthMethod{
			ssh.Password(b.pwd),
		},
	}

	conn, err := ssh.Dial("tcp", b.addr, config)
	if err != nil {
		log.Fatalf("Unable to connect %s", err)
		return err
	}

	b.conn = conn

	addr := fmt.Sprintf("localhost:%d", b.localPort)

	local, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Unable to connect %s", err)
		return err
	}

	b.local = local

	go func() {
		for {

			lc, err := local.Accept()
			log.Printf("accept new package lc is %#v", lc)
			if err != nil {
				log.Fatalf("listen Accept failed %s", err)
			}

			remote, err := conn.Dial("tcp", b.sqlAddr)
			if err != nil {
				log.Fatalf("Unable to connect %s", err)
			}

			b.remote = remote

			cp(lc, remote)

			cp(remote, lc)

		}
	}()

	return nil
}

func cp(conn net.Conn, conn2 net.Conn) {
	go func() {
		_, err := io.Copy(conn, conn2)
		if err != nil {
			log.Fatalf("io.Copy failed: %v", err)
		}
		log.Printf("cp %#v,%#v to %#v,%#v", conn2.LocalAddr(), conn2.RemoteAddr(), conn.LocalAddr(), conn.RemoteAddr())
	}()
}

func (b *Board) Close() error {
	err := b.remote.Close()
	if err != nil {
		return err
	}
	err = b.local.Close()
	if err != nil {
		return err
	}
	err = b.conn.Close()
	if err != nil {
		return err
	}
	return nil
}
