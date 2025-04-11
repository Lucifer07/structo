// Package structo provides functionality for encoding and decoding Go structures
// to and from binary data and string representations, including optional encryption.
package structo

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"sync"

	"github.com/Lucifer07/Structo/cha"
	"github.com/Lucifer07/Structo/errdefs"
)

// Converter defines the interface for encoding and decoding structures.
type Converter interface {
	StructToBinary(data interface{}) ([]byte, error)
	BinaryToStruct(data []byte, result interface{}) error
	EncodeToString(data interface{}) (string, error)
	DecodeFromString(data string, result interface{}) error
	EncodeToStringSafe(data interface{}) (string, error)
	DecodeFromStringSafe(data string, result interface{}) error
}

// ConverterImpl is the concrete implementation of the Converter interface.
type converterImpl struct {
	bufferPool sync.Pool
	encryptor  *cha.EncryptData
}

// NewConverter creates and returns a new instance of ConverterImpl.
func NewConverter() Converter {
	return &converterImpl{
		bufferPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		encryptor: cha.NewEncryptor(),
	}
}

// StructToBinary converts a struct to its binary representation.
func (c *converterImpl) StructToBinary(data interface{}) ([]byte, error) {
	buf := c.getBuffer()
	defer c.putBuffer(buf)

	if err := c.encodeValue(buf, reflect.ValueOf(data)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// BinaryToStruct converts binary data back into the provided struct pointer.
func (c *converterImpl) BinaryToStruct(data []byte, result interface{}) error {
	if reflect.TypeOf(result).Kind() != reflect.Ptr {
		return errdefs.ErrNotPointerToStruct
	}

	reader := bytes.NewReader(data)
	return c.decodeValue(reader, reflect.ValueOf(result).Elem())
}

// EncodeToString converts a struct to a raw string representation.
func (c *converterImpl) EncodeToString(data interface{}) (string, error) {
	bin, err := c.StructToBinary(data)
	if err != nil {
		return "", err
	}
	return string(bin), nil
}

// DecodeFromString decodes a raw string back into a struct.
func (c *converterImpl) DecodeFromString(data string, result interface{}) error {
	return c.BinaryToStruct([]byte(data), result)
}

// EncodeToStringSafe converts a struct to an encrypted string representation.
func (c *converterImpl) EncodeToStringSafe(data interface{}) (string, error) {
	bin, err := c.StructToBinary(data)
	if err != nil {
		return "", err
	}
	return c.encryptor.Encrypt(bin)
}

// DecodeToStringSafe decrypts and decodes an encrypted string into a struct.
func (c *converterImpl) DecodeFromStringSafe(data string, result interface{}) error {
	plain, err := c.encryptor.Decrypt(data)
	if err != nil {
		return err
	}
	return c.BinaryToStruct([]byte(plain), result)
}

// Internal helpers

func (c *converterImpl) getBuffer() *bytes.Buffer {
	buf := c.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func (c *converterImpl) putBuffer(buf *bytes.Buffer) {
	c.bufferPool.Put(buf)
}

func (c *converterImpl) encodeValue(buf *bytes.Buffer, v reflect.Value) error {
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
			if err := c.encodeValue(buf, v.Field(i)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice, reflect.Array:
		length := v.Len()
		if err := binary.Write(buf, binary.LittleEndian, int32(length)); err != nil {
			return err
		}
		for i := 0; i < length; i++ {
			if err := c.encodeValue(buf, v.Index(i)); err != nil {
				return err
			}
		}
		return nil

	case reflect.Map:
		if v.IsNil() {
			if err := binary.Write(buf, binary.LittleEndian, int32(-1)); err != nil {
				return err
			}
			return nil
		}
		keys := v.MapKeys()
		if err := binary.Write(buf, binary.LittleEndian, int32(len(keys))); err != nil {
			return err
		}
		for _, key := range keys {
			if err := c.encodeValue(buf, key); err != nil {
				return err
			}
			if err := c.encodeValue(buf, v.MapIndex(key)); err != nil {
				return err
			}
		}
		return nil

	case reflect.Interface:
		if v.IsNil() {
			if err := binary.Write(buf, binary.LittleEndian, int32(-1)); err != nil {
				return err
			}
			return nil
		}
		return c.encodeValue(buf, v.Elem())

	default:
		return errdefs.ErrUnsupportedKind
	}
}

func (c *converterImpl) decodeValue(buf *bytes.Reader, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return c.decodeValue(buf, v.Elem())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return readAndSetInt(buf, v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return readAndSetUint(buf, v)
	case reflect.Float32, reflect.Float64:
		return readAndSetFloat(buf, v)
	case reflect.Bool:
		var b byte
		if err := binary.Read(buf, binary.LittleEndian, &b); err != nil {
			return err
		}
		v.SetBool(b == 1)
		return nil
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
		return nil
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if err := c.decodeValue(buf, v.Field(i)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		var length int32
		if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
			return err
		}
		if length == -1 {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		slice := reflect.MakeSlice(v.Type(), int(length), int(length))
		for i := 0; i < int(length); i++ {
			if err := c.decodeValue(buf, slice.Index(i)); err != nil {
				return err
			}
		}
		v.Set(slice)
		return nil

	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := c.decodeValue(buf, v.Index(i)); err != nil {
				return err
			}
		}
		return nil

	case reflect.Map:
		var length int32
		if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
			return err
		}
		if length == -1 {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		mapType := v.Type()
		newMap := reflect.MakeMap(mapType)
		for i := 0; i < int(length); i++ {
			key := reflect.New(mapType.Key()).Elem()
			val := reflect.New(mapType.Elem()).Elem()
			if err := c.decodeValue(buf, key); err != nil {
				return err
			}
			if err := c.decodeValue(buf, val); err != nil {
				return err
			}
			newMap.SetMapIndex(key, val)
		}
		v.Set(newMap)
		return nil

	case reflect.Interface:
		var length int32
		if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
			return err
		}
		if length == -1 {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		elem := reflect.New(v.Type()).Elem()
		if err := c.decodeValue(buf, elem); err != nil {
			return err
		}
		v.Set(elem)
		return nil
	default:
		return errdefs.ErrUnsupportedKind
	}
}

func readAndSetInt(buf *bytes.Reader, v reflect.Value) error {
	var val int64
	if err := binary.Read(buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	v.SetInt(val)
	return nil
}

func readAndSetUint(buf *bytes.Reader, v reflect.Value) error {
	var val uint64
	if err := binary.Read(buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	v.SetUint(val)
	return nil
}

func readAndSetFloat(buf *bytes.Reader, v reflect.Value) error {
	var val float64
	if err := binary.Read(buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	v.SetFloat(val)
	return nil
}
