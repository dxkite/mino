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

func sha1Sum(d []byte) []byte {
	v := sha1.New()
	v.Write(d)
	return v.Sum(nil)[:]
}

// Encoding 握手协议
//  1. 获取 8 位的随机数据 rdm
//  2. headerKey = key xor rdm
//  3. 随机生成 0~255 长度的 padding
//  4. 获取当前时间戳（毫秒级别） int64 timestamp
//  5. 生成 header =
//     "XXOR"
//     + byte(version=1)
//     + byte(paddingSize)
//     + byte(encodingType=1)
//     + byte(checkMac=1)
//  6. encodingHeader = header xor headerKey
//  7. checkHeader = rdm + encodingHeader + padding + timestamp
//  8. realHeader = checkHeader + sha1(checkHeader)
//  9. sessionKey = padding xor headerKey
//
// 后续数据使用 sessionKey xor
func (m *XxorMessage) Encoding(key []byte) (data, sessionKey []byte, err error) {
	rdm := make([]byte, headerSize)
	if _, err := io.ReadFull(rand.Reader, rdm); err != nil {
		return nil, nil, err
	}

	key = xor(key, rdm)

	r := rand2.New(rand2.NewSource(time.Now().UnixNano()))

	// fix: 0%0 integer divide by zero
	paddingSize := r.Intn(randomMaxSize-headerSize) + headerSize

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

	//log.Debug("headKey", hex.EncodeToString(key))
	//log.Debug("head", hex.EncodeToString(head))
	//log.Debug("head", "\n"+hex.Dump(head))

	head = xor(head, key)

	//log.Debug("encodingHeader", "\n"+hex.Dump(head))
	//log.Debug("padding", "\n"+hex.Dump(m.Padding))

	data = append(data, rdm...)
	data = append(data, head...)
	data = append(data, m.Padding...)
	data = append(data, timeBuf...)

	//log.Debug("sha1Data", hex.EncodeToString(data))
	//log.Debug("sha1Data", "\n"+hex.Dump(data))

	sha1Sum := sha1Sum(data)
	//log.Debug("sha1Sum", hex.EncodeToString(sha1Sum))

	data = append(data, sha1Sum...)
	sessionKey = xor(m.Padding, key)
	return
}

func (m *XxorMessage) Decoding(r io.Reader, key []byte) (sessionKey []byte, err error) {
	rdm := make([]byte, headerSize)
	if _, err := io.ReadFull(r, rdm); err != nil {
		return nil, err
	}
	//log.Debug("rdm", hex.EncodeToString(rdm))
	//log.Debug("key", hex.EncodeToString(key))

	key = xor(key, rdm)

	head := make([]byte, headerSize)
	if _, err := io.ReadFull(r, head); err != nil {
		return nil, err
	}

	//log.Debug("headKey", hex.EncodeToString(key))
	encodingHead := append([]byte(nil), head...)
	head = xor(head, key)
	//log.Debug("head", hex.EncodeToString(head), string(head))

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

	padding := append([]byte(nil), m.Padding...)
	sessionKey = xor(padding, key)

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

	//log.Debug("padding", "\n"+hex.Dump(m.Padding))
	//log.Debug("timestamp", m.Timestamp)
	//log.Debug("sha1", hex.EncodeToString(m.Mac))

	data := append(rdm, encodingHead...)
	data = append(data, m.Padding...)
	data = append(data, timeBuf...)

	//log.Debug("sha1Data", hex.EncodeToString(data))
	//log.Debug("sha1Data", "\n"+hex.Dump(data))

	sha1Sum := sha1Sum(data)
	//log.Debug("sha1Sum", hex.EncodeToString(sha1Sum))

	if string(sha1Sum) != string(m.Mac) {
		return nil, errors.New("xxor: mac error")
	}
	return
}
