package structo

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/Lucifer07/Structo/cha"
)

type struckto struct {
	bufferPool *sync.Pool
	encryptor  *cha.EncryptData
}

func NewStrackto() *struckto {
	buff := sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	enc := cha.NewEncryptor()
	return &struckto{
		bufferPool: &buff,
		encryptor:  enc,
	}
}

func (s *struckto) StructToBinary(data interface{}) ([]byte, error) {
	buf := s.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer s.bufferPool.Put(buf)

	if err := encodeValue(buf, reflect.ValueOf(data)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeValue(buf *bytes.Buffer, v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return binary.Write(buf, binary.LittleEndian, v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return binary.Write(buf, binary.LittleEndian, v.Uint())
	case reflect.Float32, reflect.Float64:
		return binary.Write(buf, binary.LittleEndian, v.Float())
	case reflect.Bool:
		var b byte
		if v.Bool() {
			b = 1
		}
		return buf.WriteByte(b)
	case reflect.String:
		str := v.String()
		if err := binary.Write(buf, binary.LittleEndian, int32(len(str))); err != nil {
			return err
		}
		_, err := buf.WriteString(str)
		return err
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if err := encodeValue(buf, v.Field(i)); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported type: %s", v.Kind())
	}
}

func (s *struckto) BinaryToStruct(data []byte, result interface{}) error {
	if reflect.ValueOf(result).Kind() != reflect.Ptr {
		return errors.New("must be a pointer")
	}
	buf := bytes.NewReader(data)
	return decodeValue(buf, reflect.ValueOf(result).Elem())
}

func decodeValue(buf *bytes.Reader, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var value int64
		if err := binary.Read(buf, binary.LittleEndian, &value); err != nil {
			return err
		}
		v.SetInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var value uint64
		if err := binary.Read(buf, binary.LittleEndian, &value); err != nil {
			return err
		}
		v.SetUint(value)
	case reflect.Float32, reflect.Float64:
		var value float64
		if err := binary.Read(buf, binary.LittleEndian, &value); err != nil {
			return err
		}
		v.SetFloat(value)
	case reflect.Bool:
		var b byte
		if err := binary.Read(buf, binary.LittleEndian, &b); err != nil {
			return err
		}
		v.SetBool(b == 1)
	case reflect.String:
		var length int32
		if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
			return err
		}
		strBuf := make([]byte, length)
		if _, err := buf.Read(strBuf); err != nil {
			return err
		}
		v.SetString(string(strBuf))
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if err := decodeValue(buf, v.Field(i)); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported type: %s", v.Kind())
	}
	return nil
}

// Fungsi Encode dan Decode hasilnya string
func (s *struckto) EncodeToString(data interface{}) (string, error) {
	binaryData, err := s.StructToBinary(data)
	if err != nil {
		return "", err
	}
	return string(binaryData), nil
}

func (s *struckto) DecodeFromString(data string, result interface{}) error {
	return s.BinaryToStruct([]byte(data), result)
}

func (s *struckto) EncodeToStringSafe(data interface{}) (string, error) {
	binaryData, err := s.StructToBinary(data)
	if err != nil {
		return "", err
	}
	str, err := s.encryptor.Encrypt(binaryData)
	if err != nil {
		return "", err
	}
	return str, nil
}

func (s *struckto) DecodeToStringSafe(data string, result interface{}) (error) {
	dataText, err := s.encryptor.Decrypt(data)
	if err != nil {
		return err
	}
	return s.BinaryToStruct([]byte(dataText), result)
}
