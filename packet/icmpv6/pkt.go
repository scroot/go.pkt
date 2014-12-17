/*
 * Network packet analysis framework.
 *
 * Copyright (c) 2014, Alessandro Ghedini
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     * Redistributions of source code must retain the above copyright
 *       notice, this list of conditions and the following disclaimer.
 *
 *     * Redistributions in binary form must reproduce the above copyright
 *       notice, this list of conditions and the following disclaimer in the
 *       documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS
 * IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO,
 * THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
 * PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR
 * CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
 * EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
 * PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
 * PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
 * LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
 * NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

// Provides encoding and decoding for ICMPv6 packets.
package icmpv6

import "fmt"

import "github.com/ghedo/hype/packet"
import "github.com/ghedo/hype/packet/ipv4"

type Packet struct {
	Type      Type
	Code      Code
	Checksum  uint16 `name:"sum"`
	csum_seed uint32 `name:"skip"`
	Body      uint32 `name:"skip"`
}

type Type uint8
type Code uint8

const (
	DstUnreachable Type = 0
	PacketTooBig        = 1
	TimeExceeded        = 2
	ParamProblem        = 4
	Private1            = 100
	Private2            = 101
	Reserved1           = 127
	EchoRequest         = 128
	EchoReply           = 129
	/* TODO: more types */
)

func Make() *Packet {
	return &Packet{
		Type: EchoRequest,
	}
}

func (p *Packet) GetType() packet.Type {
	return packet.ICMPv6
}

func (p *Packet) GetLength() uint16 {
	return 8
}

func (p *Packet) Pack(raw_pkt *packet.Buffer) error {
	raw_pkt.WriteI(byte(p.Type))
	raw_pkt.WriteI(byte(p.Code))
	raw_pkt.WriteI(uint16(0x00))
	raw_pkt.WriteI(p.Body)

	if p.csum_seed != 0 {
		p.Checksum = ipv4.CalculateChecksum(raw_pkt.BytesOff(), p.csum_seed)
		raw_pkt.PutUint16Off(2, p.Checksum)
	}

	return nil
}

func (p *Packet) Unpack(raw_pkt *packet.Buffer) error {
	raw_pkt.ReadI(&p.Type)
	raw_pkt.ReadI(&p.Code)
	raw_pkt.ReadI(&p.Checksum)

	/* TODO: data */
	raw_pkt.ReadI(&p.Body)

	return nil
}

func (p *Packet) Payload() packet.Packet {
	return nil
}

func (p *Packet) PayloadType() packet.Type {
	return packet.None
}

func (p *Packet) SetPayload(pl packet.Packet) error {
	return nil
}

func (p *Packet) InitChecksum(csum uint32) {
	p.csum_seed = csum
}

func (p *Packet) String() string {
	return packet.Stringify(p)
}

func (t Type) String() string {
	switch t {
	case DstUnreachable:    return "dst-unreach"
	case PacketTooBig:      return "too-big"
	case TimeExceeded:      return "timeout"
	case ParamProblem:      return "param-problem"
	case EchoRequest:       return "echo-request"
	case EchoReply:         return "echo-reply"
	default:                return "unknown"
	}
}

func (c Code) String() string {
	if c != 0 {
		return fmt.Sprintf("%x", c)
	}

	return ""
}