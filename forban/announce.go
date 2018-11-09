package forban

import (
	"bytes"
	"net"

	log "github.com/sirupsen/logrus"
)

// Announce send announce packet to broadcast address
func Announce() {
	UpdateFileIndex()

	var buffer bytes.Buffer

	buffer.WriteString("forban;name;")
	buffer.WriteString(MyName)
	buffer.WriteString(";uuid;")
	buffer.WriteString(MyUuid)
	buffer.WriteString(";hmac;")
	buffer.WriteString(GetIndexHmac())

	if DisableIPv4 == false {
		dst, err := net.ResolveUDPAddr("udp", "255.255.255.255:12555")
		if err != nil {
			log.Fatal("NET ", err)
		}

		log.Debug("NET Sending IPv4 announcement")
		if _, err := ServerConn.WriteTo(buffer.Bytes(), dst); err != nil {
			log.Fatal("NET ", err)
		}
	}

	if DisableIPv6 == false {
		//ifaces, _ := net.Interfaces()
		for _, iface := range Interfaces {
			dst6, err := net.ResolveUDPAddr("udp", "[ff02::1%"+iface+"]:12555")
			if err != nil {
				log.Fatal("NET ", err)
			}

			log.Debug("NET Sending IPv6 announcement")
			if _, err := ServerConn.WriteTo(buffer.Bytes(), dst6); err != nil {
				log.Fatal("NET ", err)
			}
		}
	}
}
