// Copyright (c) 2013 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcwire_test

import (
	"bytes"
	"github.com/conformal/btcwire"
	"github.com/davecgh/go-spew/spew"
	"io"
	"reflect"
	"testing"
)

// TestInv tests the MsgInv API.
func TestInv(t *testing.T) {
	pver := btcwire.ProtocolVersion

	// Ensure the command is expected value.
	wantCmd := "inv"
	msg := btcwire.NewMsgInv()
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgInv: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is expected value for latest protocol version.
	// Num inventory vectors (varInt) + max allowed inventory vectors.
	wantPayload := uint32(1800009)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}

	// Ensure inventory vectors are added properly.
	hash := btcwire.ShaHash{}
	iv := btcwire.NewInvVect(btcwire.InvTypeBlock, &hash)
	err := msg.AddInvVect(iv)
	if err != nil {
		t.Errorf("AddInvVect: %v", err)
	}
	if msg.InvList[0] != iv {
		t.Errorf("AddInvVect: wrong invvect added - got %v, want %v",
			spew.Sprint(msg.InvList[0]), spew.Sprint(iv))
	}

	// Ensure adding more than the max allowed inventory vectors per
	// message returns an error.
	for i := 0; i < btcwire.MaxInvPerMsg; i++ {
		err = msg.AddInvVect(iv)
	}
	if err == nil {
		t.Errorf("AddInvVect: expected error on too many inventory " +
			"vectors not received")
	}

	return
}

// TestInvWire tests the MsgInv wire encode and decode for various numbers
// of inventory vectors and protocol versions.
func TestInvWire(t *testing.T) {
	// Block 203707 hash.
	hashStr := "3264bc2ac36a60840790ba1d475d01367e7c723da941069e9dc"
	blockHash, err := btcwire.NewShaHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewShaHashFromStr: %v", err)
	}

	// Transation 1 of Block 203707 hash.
	hashStr = "d28a3dc7392bf00a9855ee93dd9a81eff82a2c4fe57fbd42cfe71b487accfaf0"
	txHash, err := btcwire.NewShaHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewShaHashFromStr: %v", err)
	}

	iv := btcwire.NewInvVect(btcwire.InvTypeBlock, blockHash)
	iv2 := btcwire.NewInvVect(btcwire.InvTypeTx, txHash)

	// Empty inv message.
	NoInv := btcwire.NewMsgInv()
	NoInvEncoded := []byte{
		0x00, // Varint for number of inventory vectors
	}

	// Inv message with multiple inventory vectors.
	MultiInv := btcwire.NewMsgInv()
	MultiInv.AddInvVect(iv)
	MultiInv.AddInvVect(iv2)
	MultiInvEncoded := []byte{
		0x02,                   // Varint for number of inv vectors
		0x02, 0x00, 0x00, 0x00, // InvTypeBlock
		0xdc, 0xe9, 0x69, 0x10, 0x94, 0xda, 0x23, 0xc7,
		0xe7, 0x67, 0x13, 0xd0, 0x75, 0xd4, 0xa1, 0x0b,
		0x79, 0x40, 0x08, 0xa6, 0x36, 0xac, 0xc2, 0x4b,
		0x26, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Block 203707 hash
		0x01, 0x00, 0x00, 0x00, // InvTypeTx
		0xf0, 0xfa, 0xcc, 0x7a, 0x48, 0x1b, 0xe7, 0xcf,
		0x42, 0xbd, 0x7f, 0xe5, 0x4f, 0x2c, 0x2a, 0xf8,
		0xef, 0x81, 0x9a, 0xdd, 0x93, 0xee, 0x55, 0x98,
		0x0a, 0xf0, 0x2b, 0x39, 0xc7, 0x3d, 0x8a, 0xd2, // Tx 1 of block 203707 hash
	}

	tests := []struct {
		in   *btcwire.MsgInv // Message to encode
		out  *btcwire.MsgInv // Expected decoded message
		buf  []byte          // Wire encoding
		pver uint32          // Protocol version for wire encoding
	}{
		// Latest protocol version with no inv vectors.
		{
			NoInv,
			NoInv,
			NoInvEncoded,
			btcwire.ProtocolVersion,
		},

		// Latest protocol version with multiple inv vectors.
		{
			MultiInv,
			MultiInv,
			MultiInvEncoded,
			btcwire.ProtocolVersion,
		},

		// Protocol version BIP0035Version no inv vectors.
		{
			NoInv,
			NoInv,
			NoInvEncoded,
			btcwire.BIP0035Version,
		},

		// Protocol version BIP0035Version with multiple inv vectors.
		{
			MultiInv,
			MultiInv,
			MultiInvEncoded,
			btcwire.BIP0035Version,
		},

		// Protocol version BIP0031Version no inv vectors.
		{
			NoInv,
			NoInv,
			NoInvEncoded,
			btcwire.BIP0031Version,
		},

		// Protocol version BIP0031Version with multiple inv vectors.
		{
			MultiInv,
			MultiInv,
			MultiInvEncoded,
			btcwire.BIP0031Version,
		},

		// Protocol version NetAddressTimeVersion no inv vectors.
		{
			NoInv,
			NoInv,
			NoInvEncoded,
			btcwire.NetAddressTimeVersion,
		},

		// Protocol version NetAddressTimeVersion with multiple inv vectors.
		{
			MultiInv,
			MultiInv,
			MultiInvEncoded,
			btcwire.NetAddressTimeVersion,
		},

		// Protocol version MultipleAddressVersion no inv vectors.
		{
			NoInv,
			NoInv,
			NoInvEncoded,
			btcwire.MultipleAddressVersion,
		},

		// Protocol version MultipleAddressVersion with multiple inv vectors.
		{
			MultiInv,
			MultiInv,
			MultiInvEncoded,
			btcwire.MultipleAddressVersion,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode the message to wire format.
		var buf bytes.Buffer
		err := test.in.BtcEncode(&buf, test.pver)
		if err != nil {
			t.Errorf("BtcEncode #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("BtcEncode #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Decode the message from wire format.
		var msg btcwire.MsgInv
		rbuf := bytes.NewBuffer(test.buf)
		err = msg.BtcDecode(rbuf, test.pver)
		if err != nil {
			t.Errorf("BtcDecode #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&msg, test.out) {
			t.Errorf("BtcDecode #%d\n got: %s want: %s", i,
				spew.Sdump(msg), spew.Sdump(test.out))
			continue
		}
	}
}

// TestInvWireErrors performs negative tests against wire encode and decode
// of MsgInv to confirm error paths work correctly.
func TestInvWireErrors(t *testing.T) {
	pver := btcwire.ProtocolVersion
	btcwireErr := &btcwire.MessageError{}

	// Block 203707 hash.
	hashStr := "3264bc2ac36a60840790ba1d475d01367e7c723da941069e9dc"
	blockHash, err := btcwire.NewShaHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewShaHashFromStr: %v", err)
	}

	iv := btcwire.NewInvVect(btcwire.InvTypeBlock, blockHash)

	// Base inv message used to induce errors.
	baseInv := btcwire.NewMsgInv()
	baseInv.AddInvVect(iv)
	baseInvEncoded := []byte{
		0x02,                   // Varint for number of inv vectors
		0x02, 0x00, 0x00, 0x00, // InvTypeBlock
		0xdc, 0xe9, 0x69, 0x10, 0x94, 0xda, 0x23, 0xc7,
		0xe7, 0x67, 0x13, 0xd0, 0x75, 0xd4, 0xa1, 0x0b,
		0x79, 0x40, 0x08, 0xa6, 0x36, 0xac, 0xc2, 0x4b,
		0x26, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Block 203707 hash
	}

	// Inv message that forces an error by having more than the max allowed
	// inv vectors.
	maxInv := btcwire.NewMsgInv()
	for i := 0; i < btcwire.MaxInvPerMsg; i++ {
		maxInv.AddInvVect(iv)
	}
	maxInv.InvList = append(maxInv.InvList, iv)
	maxInvEncoded := []byte{
		0xfd, 0x51, 0xc3, // Varint for number of inv vectors (50001)
	}

	tests := []struct {
		in       *btcwire.MsgInv // Value to encode
		buf      []byte          // Wire encoding
		pver     uint32          // Protocol version for wire encoding
		max      int             // Max size of fixed buffer to induce errors
		writeErr error           // Expected write error
		readErr  error           // Expected read error
	}{
		// Latest protocol version with intentional read/write errors.
		// Force error in inventory vector count
		{baseInv, baseInvEncoded, pver, 0, io.ErrShortWrite, io.EOF},
		// Force error in inventory list.
		{baseInv, baseInvEncoded, pver, 1, io.ErrShortWrite, io.EOF},
		// Force error with greater than max inventory vectors.
		{maxInv, maxInvEncoded, pver, 3, btcwireErr, btcwireErr},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		w := newFixedWriter(test.max)
		err := test.in.BtcEncode(w, test.pver)
		if reflect.TypeOf(err) != reflect.TypeOf(test.writeErr) {
			t.Errorf("BtcEncode #%d wrong error got: %v, want: %v",
				i, err, test.writeErr)
			continue
		}

		// For errors which are not of type btcwire.MessageError, check
		// them for equality.
		if _, ok := err.(*btcwire.MessageError); !ok {
			if err != test.writeErr {
				t.Errorf("BtcEncode #%d wrong error got: %v, "+
					"want: %v", i, err, test.writeErr)
				continue
			}
		}

		// Decode from wire format.
		var msg btcwire.MsgInv
		r := newFixedReader(test.max, test.buf)
		err = msg.BtcDecode(r, test.pver)
		if reflect.TypeOf(err) != reflect.TypeOf(test.readErr) {
			t.Errorf("BtcDecode #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}

		// For errors which are not of type btcwire.MessageError, check
		// them for equality.
		if _, ok := err.(*btcwire.MessageError); !ok {
			if err != test.readErr {
				t.Errorf("BtcDecode #%d wrong error got: %v, "+
					"want: %v", i, err, test.readErr)
				continue
			}
		}

	}
}
