package xxor

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"io"
	rand2 "math/rand"
	"time"
)

type XxorMessage struct {
	// MINO
	Version int
	// Auto
	PaddingSize int
	// XOR = 1
	EncodeType int
	// Auto
	Timestamp int64
	// Auto
	Padding []byte
	// SHA1
	Mac []byte
}

func defaultXxor() *XxorMessage {
	return &XxorMessage{
		Version:    1,
		EncodeType: 1,
	}
}

func (m *XxorMessage) Encoding(key []byte) (data, sessionKey []byte, err error) {
	rdm := make([]byte, headerSize)
	if _, err := io.ReadFull(rand.Reader, rdm); err != nil {
		return nil, nil, err
	}

	key = xor(key, rdm)
	paddingSize := rand2.Intn(randomMaxSize)
	m.PaddingSize = paddingSize
	m.Padding = make([]byte, paddingSize)
	if _, err := io.ReadFull(rand.Reader, m.Padding); err != nil {
		return nil, nil, err
	}

	curTime := time.Now().Unix()
	timeBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBuf, uint64(curTime))

	head := []byte("XXOR")
	head = append(head, byte(m.Version))
	head = append(head, byte(m.PaddingSize))
	head = append(head, byte(m.EncodeType))
	head = append(head, byte(1)) // EnableMac

	//fmt.Println("headKey", hex.EncodeToString(key))
	//fmt.Println("head", hex.EncodeToString(head))

	head = xor(head, key)

	data = append(data, rdm...)
	data = append(data, head...)
	data = append(data, m.Padding...)
	data = append(data, timeBuf...)

	sha1.New().Write(data)
	sha1Sum := sha1.Sum(nil)
	data = append(data, sha1Sum[:]...)

	sessionKey = xor(m.Padding, key)
	return
}

func (m *XxorMessage) Decoding(r io.Reader, key []byte) (sessionKey []byte, err error) {
	rdm := make([]byte, headerSize)
	if _, err := io.ReadFull(r, rdm); err != nil {
		return nil, err
	}

	key = xor(key, rdm)

	head := make([]byte, headerSize)
	if _, err := io.ReadFull(r, head); err != nil {
		return nil, err
	}

	//fmt.Println("headKey", hex.EncodeToString(key))
	head = xor(head, key)
	//fmt.Println("head", hex.EncodeToString(head), string(head))

	if string(head[:4]) != "XXOR" {
		return nil, errors.New("xxor: unknown magic")
	}

	m.PaddingSize = int(head[5])
	m.EncodeType = int(head[6])

	// XXXXXXX1
	hasMac := head[7]&0x1 > 0

	m.Padding = make([]byte, m.PaddingSize)
	if _, err := io.ReadFull(r, m.Padding); err != nil {
		return nil, errors.New("xxor: padding read: " + err.Error())
	}

	sessionKey = xor(m.Padding, key)

	if !hasMac {
		return
	}

	timeBuf := make([]byte, 8)
	if _, err := io.ReadFull(r, timeBuf); err != nil {
		return nil, errors.New("xxor: time read: " + err.Error())
	}

	m.Timestamp = int64(binary.BigEndian.Uint64(timeBuf))
	m.Mac = make([]byte, 20)
	if _, err := io.ReadFull(r, m.Mac); err != nil {
		return nil, errors.New("xxor: mac read: " + err.Error())
	}

	data := append(rdm, head...)
	data = append(data, m.Padding...)
	data = append(data, timeBuf...)
	sha1.New().Write(data)
	sha1Sum := sha1.Sum(nil)
	if string(sha1Sum[:]) != string(m.Mac) {
		return nil, errors.New("xxor: mac error")
	}
	return
}
