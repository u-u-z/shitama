package common

import (
	"encoding/binary"
	"net"
)

func UDPAddrToSockAddr(addr *net.UDPAddr) []byte {

	buf := make([]byte, 8)

	binary.BigEndian.PutUint16(buf[:2], 0x200)
	binary.BigEndian.PutUint16(buf[2:4], uint16(addr.Port))
	copy(buf[4:8], addr.IP[len(addr.IP)-4:])

	return buf

}

func SockAddrToUDPAddr(buf []byte) *net.UDPAddr {

	addr := new(net.UDPAddr)
	addr.IP = make([]byte, 16)
	addr.IP[10] = 255
	addr.IP[11] = 255
	copy(addr.IP[len(addr.IP)-4:], buf[4:8])
	addr.Port = int(binary.BigEndian.Uint16(buf[2:4]))

	return addr

}

func PackData(addr *net.UDPAddr, data []byte) []byte {

	buf := make([]byte, len(data)+6)

	copy(buf[:4], addr.IP[len(addr.IP)-4:])
	binary.BigEndian.PutUint16(buf[4:6], uint16(addr.Port))
	copy(buf[6:], data)

	return buf

}

func UnpackData(buf []byte) (addr *net.UDPAddr, data []byte) {

	addr = new(net.UDPAddr)
	addr.IP = make([]byte, 16)
	addr.IP[10] = 255
	addr.IP[11] = 255
	copy(addr.IP[len(addr.IP)-4:], buf[:4])
	addr.Port = int(binary.BigEndian.Uint16(buf[4:6]))

	data = make([]byte, len(buf)-6)
	copy(data, buf[6:])
	return addr, data

}

func UDPAddrToSockAddr2(addr *net.UDPAddr, outBuf []byte) []byte {

	binary.BigEndian.PutUint16(outBuf[:2], 0x200)
	binary.BigEndian.PutUint16(outBuf[2:4], uint16(addr.Port))
	copy(outBuf[4:8], addr.IP[len(addr.IP)-4:])

	return outBuf

}

func PackData2(addr *net.UDPAddr, data []byte, outBuf []byte) []byte {

	copy(outBuf[:4], addr.IP[len(addr.IP)-4:])
	binary.BigEndian.PutUint16(outBuf[4:6], uint16(addr.Port))
	copy(outBuf[6:], data)

	return outBuf

}

func UnpackData2(buf []byte) (addr *net.UDPAddr, data []byte) {

	addr = new(net.UDPAddr)
	addr.IP = make([]byte, 16)
	addr.IP[10] = 255
	addr.IP[11] = 255
	copy(addr.IP[len(addr.IP)-4:], buf[:4])
	addr.Port = int(binary.BigEndian.Uint16(buf[4:6]))

	return addr, buf[6:]

}
