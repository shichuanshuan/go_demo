package main

import (
	"fmt"
	"golang.org/x/net/ipv4"
	"net"
	"time"
)

type NetPacketConn interface {
	JoinGroup(ifi *net.Interface, group net.Addr) error
	LeaveGroup(ifi *net.Interface, group net.Addr) error
	SetMulticastInterface(ini *net.Interface) error
	SetMulticastTTL(int) error
	ReadFrom(buf []byte) (int, net.Addr, error)
	WriteTo(buf []byte, dst net.Addr) (int, error)
}

type PacketConn4 struct {
	*ipv4.PacketConn
}

// ReadFrom wraps the ipv4 ReadFrom without a control message
func (pc4 PacketConn4) ReadFrom(buf []byte) (int, net.Addr, error) {
	n, _, addr, err := pc4.PacketConn.ReadFrom(buf)
	return n, addr, err
}

// WriteTo wraps the ipv4 WriteTo without a control message
func (pc4 PacketConn4) WriteTo(buf []byte, dst net.Addr) (int, error) {
	return pc4.PacketConn.WriteTo(buf, nil, dst)
}

func listen() (recievedBytes []byte, err error) {
	portNum := 5100
	MulticastAddress := "233.255.255.216"

	portNum = 5200
	MulticastAddress = "233.255.255.211"

	Port := fmt.Sprintf("%d", portNum)
	address := net.JoinHostPort(MulticastAddress, Port)
	fmt.Printf("Port[%v]\n", Port)

	ethDevName := "ge0-0"
	ifaces, err := net.InterfaceByName(ethDevName)
	if err != nil {
		return nil, err
	}

	// Open up a connection
	c, err := net.ListenPacket(fmt.Sprintf("udp%d", 4), address)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	group := net.ParseIP(MulticastAddress)
	var p2 NetPacketConn

	p2 = PacketConn4{ipv4.NewPacketConn(c)}

	p2.JoinGroup(ifaces, &net.UDPAddr{IP: group, Port: portNum})

	err = p2.SetMulticastInterface(ifaces)
	if err != nil {
		fmt.Println("err= ", err)
	}

	// Loop forever reading from the socket
	for {
		buffer := make([]byte, 4200)
		var (
			n       int
			src     net.Addr
			errRead error
		)
		fmt.Println("start read")
		n, src, errRead = p2.ReadFrom(buffer)
		fmt.Printf("n [%v] src[%v] errRead[%v] buffer[%s]\n", n, src, errRead, buffer[:n])
		if errRead != nil {
			err = errRead
			return
		}
		group := net.ParseIP(MulticastAddress)

		p2.WriteTo([]byte("i'm noah"), &net.UDPAddr{IP: group, Port: portNum})

		time.Sleep(time.Second)

	}

	return
}

// 从组播读取数据
func main() {
	listen()
}
