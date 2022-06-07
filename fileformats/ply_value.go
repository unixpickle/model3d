package fileformats

import (
	"bytes"
	"encoding/binary"
	"math"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type PLYValue interface {
	EncodeString() string
	EncodeBinary(encoding binary.ByteOrder) []byte
	LengthValue() (int, error)
}

type PLYValueList struct {
	Length PLYValue
	Values []PLYValue
}

func (p PLYValueList) EncodeString() string {
	parts := make([]string, len(p.Values)+1)
	parts[0] = p.Length.EncodeString()
	for i, part := range p.Values {
		parts[i+1] = part.EncodeString()
	}
	return strings.Join(parts, " ")
}

func (p PLYValueList) EncodeBinary(encoding binary.ByteOrder) []byte {
	var buf bytes.Buffer
	buf.Write(p.Length.EncodeBinary(encoding))
	for _, part := range p.Values {
		buf.Write(part.EncodeBinary(encoding))
	}
	return buf.Bytes()
}

func (p PLYValueList) LengthValue() (int, error) {
	return 0, errors.New("PLYValueList cannot be converted to length")
}

type PLYValueInt8 struct {
	Value int8
}

func (p PLYValueInt8) EncodeString() string {
	return strconv.FormatInt(int64(p.Value), 10)
}

func (p PLYValueInt8) EncodeBinary(b binary.ByteOrder) []byte {
	return []byte{byte(p.Value)}
}

func (p PLYValueInt8) LengthValue() (int, error) {
	return int(p.Value), nil
}

type PLYValueUint8 struct {
	Value uint8
}

func (p PLYValueUint8) EncodeString() string {
	return strconv.FormatUint(uint64(p.Value), 10)
}

func (p PLYValueUint8) EncodeBinary(b binary.ByteOrder) []byte {
	return []byte{byte(p.Value)}
}

func (p PLYValueUint8) LengthValue() (int, error) {
	return int(p.Value), nil
}

type PLYValueInt16 struct {
	Value int16
}

func (p PLYValueInt16) EncodeString() string {
	return strconv.FormatInt(int64(p.Value), 10)
}

func (p PLYValueInt16) EncodeBinary(b binary.ByteOrder) []byte {
	var res [2]byte
	b.PutUint16(res[:], uint16(p.Value))
	return res[:]
}

func (p PLYValueInt16) LengthValue() (int, error) {
	return int(p.Value), nil
}

type PLYValueUint16 struct {
	Value uint16
}

func (p PLYValueUint16) EncodeString() string {
	return strconv.FormatUint(uint64(p.Value), 10)
}

func (p PLYValueUint16) EncodeBinary(b binary.ByteOrder) []byte {
	var res [2]byte
	b.PutUint16(res[:], p.Value)
	return res[:]
}

func (p PLYValueUint16) LengthValue() (int, error) {
	return int(p.Value), nil
}

type PLYValueInt32 struct {
	Value int32
}

func (p PLYValueInt32) EncodeString() string {
	return strconv.FormatInt(int64(p.Value), 10)
}

func (p PLYValueInt32) EncodeBinary(b binary.ByteOrder) []byte {
	var res [4]byte
	b.PutUint32(res[:], uint32(p.Value))
	return res[:]
}

func (p PLYValueInt32) LengthValue() (int, error) {
	return int(p.Value), nil
}

type PLYValueUint32 struct {
	Value uint32
}

func (p PLYValueUint32) EncodeString() string {
	return strconv.FormatUint(uint64(p.Value), 10)
}

func (p PLYValueUint32) EncodeBinary(b binary.ByteOrder) []byte {
	var res [4]byte
	b.PutUint32(res[:], p.Value)
	return res[:]
}

func (p PLYValueUint32) LengthValue() (int, error) {
	return int(p.Value), nil
}

type PLYValueInt64 struct {
	Value int64
}

func (p PLYValueInt64) EncodeString() string {
	return strconv.FormatInt(p.Value, 10)
}

func (p PLYValueInt64) EncodeBinary(b binary.ByteOrder) []byte {
	var res [8]byte
	b.PutUint64(res[:], uint64(p.Value))
	return res[:]
}

func (p PLYValueInt64) LengthValue() (int, error) {
	return int(p.Value), nil
}

type PLYValueUint64 struct {
	Value uint64
}

func (p PLYValueUint64) EncodeString() string {
	return strconv.FormatUint(p.Value, 10)
}

func (p PLYValueUint64) EncodeBinary(b binary.ByteOrder) []byte {
	var res [8]byte
	b.PutUint64(res[:], p.Value)
	return res[:]
}

func (p PLYValueUint64) LengthValue() (int, error) {
	return int(p.Value), nil
}

type PLYValueFloat32 struct {
	Value float32
}

func (p PLYValueFloat32) EncodeString() string {
	return strconv.FormatFloat(float64(p.Value), 'f', -1, 32)
}

func (p PLYValueFloat32) EncodeBinary(b binary.ByteOrder) []byte {
	var res [4]byte
	b.PutUint32(res[:], math.Float32bits(p.Value))
	return res[:]
}

func (p PLYValueFloat32) LengthValue() (int, error) {
	return 0, errors.New("PLYValueFloat32 cannot be used as length field")
}

type PLYValueFloat64 struct {
	Value float64
}

func (p PLYValueFloat64) EncodeString() string {
	return strconv.FormatFloat(p.Value, 'f', -1, 64)
}

func (p PLYValueFloat64) EncodeBinary(b binary.ByteOrder) []byte {
	var res [8]byte
	b.PutUint64(res[:], math.Float64bits(p.Value))
	return res[:]
}

func (p PLYValueFloat64) LengthValue() (int, error) {
	return 0, errors.New("PLYValueFloat32 cannot be used as length field")
}
