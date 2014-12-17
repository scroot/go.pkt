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

// Provides encoding and decoding for VLAN packets.
package vlan

import "github.com/ghedo/hype/packet"

type Packet struct {
	Priority     uint8         `name:"prio"`
	DropEligible bool          `name:"drop"`
	VLAN         uint16
	Type         packet.Type
	pkt_payload  packet.Packet `name:"skip"`
}

func Make() *Packet {
	return &Packet{ }
}

func (p *Packet) GetType() packet.Type {
	return packet.VLAN
}

func (p *Packet) GetLength() uint16 {
	if p.pkt_payload != nil {
		return p.pkt_payload.GetLength() + 2
	}

	return 2
}

func (p *Packet) Pack(raw_pkt *packet.Buffer) error {
	tci := uint16(p.Priority) << 13 | p.VLAN
	if p.DropEligible {
		tci |= 0x10
	}

	raw_pkt.WriteI(tci)
	raw_pkt.WriteI(p.Type.ToEtherType())

	return nil
}

func (p *Packet) Unpack(raw_pkt *packet.Buffer) error {
	var tci uint16
	raw_pkt.ReadI(&tci)

	p.Priority     = (uint8(tci >> 8) & 0xE0) >> 5
	p.DropEligible = uint8(tci) & 0x10 != 0
	p.VLAN         = tci & 0x0FFF

	var ethertype uint16
	raw_pkt.ReadI(&ethertype)
	p.Type = packet.EtherType(ethertype)

	return nil
}

func (p *Packet) Payload() packet.Packet {
	return p.pkt_payload
}

func (p *Packet) PayloadType() packet.Type {
	return p.Type
}

func (p *Packet) SetPayload(pl packet.Packet) error {
	p.pkt_payload = pl
	p.Type        = pl.GetType()

	return nil
}

func (p *Packet) InitChecksum(csum uint32) {
}

func (p *Packet) String() string {
	return packet.Stringify(p)
}