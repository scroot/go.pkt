// Provides packet capturing and injection on live network interfaces via
// libpcap.
package live

// #cgo LDFLAGS: -lpcap
// #include <stdlib.h>
// #include <pcap.h>
import "C"

import "fmt"
import "unsafe"

import "github.com/ghedo/hype/packet"

type Handle struct {
	Device string
	pcap   *C.pcap_t
}

// Create a new capture handle from the given network interface. Noe that this
// may require root privileges.
func Open(dev_name string) (*Handle, error) {
	handle := &Handle{ Device: dev_name }

	dev_str := C.CString(dev_name)
	defer C.free(unsafe.Pointer(dev_str))

	err_str := (*C.char)(C.calloc(256, 1))
	defer C.free(unsafe.Pointer(err_str))

	handle.pcap = C.pcap_create(dev_str, err_str)
	if handle == nil {
		return nil, fmt.Errorf(
			"Could not open device: %s", C.GoString(err_str),
		)
	}

	return handle, nil
}

// Return the link type of the capture handle (that is, the type of packets that
// come out of the packet source).
func (h *Handle) LinkType() packet.Type {
	return packet.LinkType(uint32(C.pcap_datalink(h.pcap)))
}

func (h *Handle) SetMTU(mtu int) error {
	err := C.pcap_set_snaplen(h.pcap, C.int(mtu))
	if err < 0 {
		return fmt.Errorf("Handle already active")
	}

	return nil
}

// Enable/disable promiscuous mode.
func (h *Handle) SetPromiscMode(promisc bool) error {
	var promisc_int C.int

	if promisc {
		promisc_int = 1
	} else {
		promisc_int = 0
	}

	err := C.pcap_set_promisc(h.pcap, promisc_int)
	if err < 0 {
		return fmt.Errorf("Handle already active")
	}

	return nil
}

// Enable/disable monitor mode. This is only relevant to RF-based packet sources
// (e.g. a WiFi or Bluetooth network interface)
func (h *Handle) SetMonitorMode(monitor bool) error {
	var rfmon_int C.int

	if monitor {
		rfmon_int = 1
	} else {
		rfmon_int = 0
	}

	err := C.pcap_set_rfmon(h.pcap, rfmon_int)
	if err < 0 {
		return fmt.Errorf("Handle already active")
	}

	return nil
}

// Compile the given filter and apply it to the packet source. Only packets that
// match this filter will be returned.
func (h *Handle) ApplyFilter(filter string) error {
	var net, mask C.bpf_u_int32
	var bpf       C.struct_bpf_program

	fil_str := C.CString(filter)
	defer C.free(unsafe.Pointer(fil_str))

	err_str := (*C.char)(C.calloc(256, 1))
	defer C.free(unsafe.Pointer(err_str))

	dev_str := C.CString(h.Device)
	defer C.free(unsafe.Pointer(dev_str))

	err := C.pcap_lookupnet(dev_str, &net, &mask, err_str)
	if err < 0 {
		return fmt.Errorf(
			"Could not get device netmask: %s", C.GoString(err_str),
		)
	}

	err = C.pcap_compile(h.pcap, &bpf, fil_str, 0, net)
	if err < 0 {
		return fmt.Errorf("Could not compile filter: %s", h.get_error())
	}

	err = C.pcap_setfilter(h.pcap, &bpf)
	if err < 0 {
		return fmt.Errorf("Could not set filter: %s", h.get_error())
	}

	return nil
}

// Activate the packet source. Note that after calling this method it will not
// be possible to change the packet source configuration (MTU, promiscuous mode,
// monitor mode, ...)
func (h *Handle) Activate() error {
	err := C.pcap_activate(h.pcap)
	if err < 0 {
		return fmt.Errorf("Could not activate: %s", h.get_error())
	}

	return nil
}

// Capture a single packet from the packet source. This will block until a acket
// is received.
func (h *Handle) Capture() ([]byte, error) {
	var raw_pkt *C.u_char
	var pkt_hdr *C.struct_pcap_pkthdr

	for {
		err := C.pcap_next_ex(h.pcap, &pkt_hdr, &raw_pkt)
		switch err {
		case -2:
			return nil, nil

		case -1:
			return nil, fmt.Errorf(
				"Could not read packet: %s", h.get_error(),
			)

		case 0:
			continue

		case 1:
			return C.GoBytes(unsafe.Pointer(raw_pkt),
			                 C.int(pkt_hdr.len)), nil
		}
	}

	return nil, fmt.Errorf("WTF")
}

// Inject a packet in the packet source.
func (h *Handle) Inject(raw_pkt []byte) error {
	buf     := (*C.u_char)(&raw_pkt[0])
	buf_len := C.int(len(raw_pkt))

	err := C.pcap_sendpacket(h.pcap, buf, buf_len)
	if err < 0 {
		return fmt.Errorf("Could not inject packet: %s", h.get_error())
	}

	return nil
}

// Close the packet source.
func (h *Handle) Close() {
	C.pcap_close(h.pcap)
}

func (h *Handle) get_error() error {
	err_str := C.pcap_geterr(h.pcap)
	return fmt.Errorf(C.GoString(err_str))
}