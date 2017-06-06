package shard

import "testing"
import "net"
import "bytes"

func TestPackData(t *testing.T) {

	l := new(UDPLink)

	addr1, err := net.ResolveUDPAddr("udp4", "110.119.120.233:6666")

	if err != nil {
		t.Error(err)
	}

	data1 := []byte("Hello world!")

	t.Log(addr1, []byte(addr1.IP), data1)

	buf := l.packData(addr1, data1)

	t.Log(buf, len(buf))

	addr2, data2 := l.unpackData(buf)

	t.Log(addr2, []byte(addr2.IP), data2)

	if addr1.String() != addr2.String() {
		t.Error("ERROR_ADDR_MISMATCH")
	}

	if !bytes.Equal(data1, data2) {
		t.Error("ERROR_DATA_MISMATCH")
	}

}
