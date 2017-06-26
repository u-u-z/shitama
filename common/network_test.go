package common

import (
	"bytes"
	"net"
	"testing"
)

func TestPackData(t *testing.T) {

	addr1, err := net.ResolveUDPAddr("udp4", "110.119.120.233:6666")

	if err != nil {
		t.Error(err)
	}

	data1 := []byte("Hello world!")

	buf := make([]byte, len(data1)+6)

	// Comment out the logging makes a huge difference.
	for i := 0; i < 1000000; i++ {

		//t.Log(addr1, []byte(addr1.IP), data1)

		//buf := PackData(addr1, data1)
		PackData2(addr1, data1, buf)

		//t.Log(buf, len(buf))

		addr2, data2 := UnpackData2(buf)

		//t.Log(addr2, []byte(addr2.IP), data2)

		if addr1.String() != addr2.String() {
			t.Error("ERROR_ADDR_MISMATCH")
		}

		if !bytes.Equal(data1, data2) {
			t.Error("ERROR_DATA_MISMATCH")
		}

	}

}

func TestAddrConvert(t *testing.T) {

	addr1, err := net.ResolveUDPAddr("udp4", "110.119.120.233:6666")

	if err != nil {
		t.Error(err)
	}

	buf := make([]byte, 16)

	for i := 0; i < 1000000; i++ {

		//buf := UDPAddrToSockAddr(addr1)
		UDPAddrToSockAddr2(addr1, buf)

		addr2 := SockAddrToUDPAddr(buf)

		if addr1.String() != addr2.String() {
			t.Error("ERROR_ADDR_MISMATCH")
		}

	}

}
