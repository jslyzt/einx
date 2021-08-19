package network

import (
	"fmt"
	"math"

	"github.com/jslyzt/einx/slog"
)

func RpcMarshal(b []byte, val interface{}) []byte {
	var buffer []byte = nil
	switch v := val.(type) {
	case nil:
		buffer = append(b, 'z')
	case bool:
		if v == true {
			buffer = append(b, 't')
		} else {
			buffer = append(b, 'f')
		}
	case string:
		slen := uint32(len(v))
		buffer = append(b, 's', byte(slen), byte(slen>>8), byte(slen>>16), byte(slen>>24))
		buffer = append(buffer, v...)
	case float32, float64:
		n := math.Float64bits(v.(float64))
		buffer = append(b, 'd', '2', byte(n), byte(n>>8), byte(n>>16), byte(n>>24), byte(n>>32), byte(n>>40), byte(n>>48), byte(n>>56))
	case int, int16, uint16, int32, int64, uint32, uint64:
		I64i, Itype := convertInteger(v)
		buffer = append(b, Itype)
		ux := uint64(I64i) << 1
		if I64i < 0 {
			ux = ^ux
		}
		for ux >= 0x80 {
			buffer = append(buffer, byte(ux)|0x80)
			ux >>= 7
		}
		buffer = append(buffer, byte(ux))
	case []byte:
		slen := uint32(len(v))
		buffer = append(b, '!', byte(slen), byte(slen>>8), byte(slen>>16), byte(slen>>24))
		buffer = append(buffer, v...)
	case map[string]interface{}:
		buffer = append(b, '[')
		for key, value := range v {
			buffer = RpcMarshal(RpcMarshal(buffer, key), value)
		}
		buffer = append(buffer, ']')
	case []interface{}:
		buffer = append(b, '{')
		for _, m := range v {
			buffer = RpcMarshal(buffer, m)
		}
		buffer = append(buffer, '}')
	default:
		panic("unsopported rpc type")
	}
	return buffer
}

func convertInteger(val interface{}) (int64, byte) {
	switch v := val.(type) {
	case int:
		return int64(v), 'i'
	case int16:
		return int64(v), 'w'
	case uint16:
		return int64(v), 'm'
	case int32:
		return int64(v), 'q'
	case uint32:
		return int64(v), 'p'
	case int64:
		return v, 'l'
	case uint64:
		return int64(v), 'n'
	default:
		panic("unknown integer type.")
	}
}

func makeInteger(x int64, t byte) interface{} {
	switch t {
	case 'i':
		return int(x)
	case 'w':
		return int16(x)
	case 'm':
		return uint16(x)
	case 'q':
		return int32(x)
	case 'p':
		return uint32(x)
	case 'l':
		return x
	case 'n':
		return uint64(x)
	default:
		panic(fmt.Sprintf("unknown integer type [%v]", t))
	}
}

func RpcUnMarshal(b []byte) (interface{}, []byte) {
	if len(b) < 1 {
		return nil, b
	}
	t := b[0]
	switch t {
	case 'z':
		return nil, b[1:]
	case 't':
		return true, b[1:]
	case 'f':
		return false, b[1:]
	case 's':
		if len(b) < 5 {
			slog.LogWarning("rpc_unmarshal", "error:unknow unmarshal string")
			return nil, b
		}
		slen := uint32(b[1]) | uint32(b[2])<<8 | uint32(b[3])<<16 | uint32(b[4])<<24
		return string(b[5 : 5+slen]), b[5+slen:]
	case '!':
		if len(b) < 5 {
			slog.LogWarning("rpc_unmarshal", "error:unknow unmarshal []byte")
			return nil, b
		}
		slen := uint32(b[1]) | uint32(b[2])<<8 | uint32(b[3])<<16 | uint32(b[4])<<24
		newBytes := make([]byte, slen)
		copy(newBytes, b[5:5+slen])
		return newBytes, b[5+slen:]
	case 'd':
		if len(b) < 9 {
			slog.LogWarning("rpc_unmarshal", "error:unknow unmarshal number")
			return nil, b
		}
		n := uint64(b[1]) | uint64(b[2])<<8 | uint64(b[3])<<16 | uint64(b[4])<<24 |
			uint64(b[5])<<32 | uint64(b[6])<<40 | uint64(b[7])<<48 | uint64(b[8])<<56
		return math.Float64frombits(n), b[9:]
	case 'i', 'w', 'm', 'q', 'p', 'l', 'n':
		length := len(b)
		if length < 2 {
			slog.LogWarning("rpc_unmarshal", "error:unknow unmarshal varint")
			return nil, b
		}
		var ux uint64
		var s uint32
		i := 1
		for ; i < 9; i++ {
			if i >= length {
				slog.LogWarning("rpc_unmarshal", "error:unknow unmarshal varint:overflow")
				return nil, b
			}
			m := b[i]
			if m < 0x80 {
				if i == 9 && m > 1 {
					slog.LogWarning("rpc_unmarshal", "error:unknow unmarshal varint:overflow")
					return nil, b
				}
				ux |= uint64(m) << s
				break
			}
			ux |= uint64(m&0x7f) << s
			s += 7
		}
		x := int64(ux >> 1)
		if ux&1 != 0 {
			x = ^x
		}
		return makeInteger(x, t), b[i+1:]
	case '[':
		var key interface{}
		var val interface{}
		tb := b[1:]
		m := make(map[interface{}]interface{})
		for tb[0] != ']' {
			key, tb = RpcUnMarshal(tb)
			val, tb = RpcUnMarshal(tb)
			m[key] = val
		}
		return m, tb[1:]
	case '{':
		var val interface{}
		index := 1
		tb := b[1:]
		lt := make([]interface{}, 0, 8)
		for tb[0] != '}' {
			val, tb = RpcUnMarshal(tb)
			lt = append(lt, val)
			index++
		}
		return lt, tb[1:]
	default:
		slog.LogError("rpc_unmarshal", "error rpc type %v", t)
		panic("error rpc type")
	}
	return nil, b
}
