// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

/*
Package rlp implements the RLP serialization format.

The purpose of RLP (Recursive Linear Prefix) is to encode arbitrarily nested arrays of
binary data, and RLP is the main encoding method used to serialize objects in Ethereum.
The only purpose of RLP is to encode structure; encoding specific atomic data types (eg.
strings, ints, floats) is left up to higher-order protocols. In Ethereum integers must be
represented in big endian binary form with no leading zeroes (thus making the integer
value zero equivalent to the empty string).

RLP values are distinguished by a type tag. The type tag precedes the value in the input
stream and defines the size and kind of the bytes that follow.

# Encoding Rules

Package rlp uses reflection and encodes RLP based on the Go type of the value.

If the type implements the Encoder interface, Encode calls EncodeRLP. It does not
call EncodeRLP on nil pointer values.

To encode a pointer, the value being pointed to is encoded. A nil pointer to a struct
type, slice or array always encodes as an empty RLP list unless the slice or array has
element type byte. A nil pointer to any other value encodes as the empty string.

Struct values are encoded as an RLP list of all their encoded public fields. Recursive
struct types are supported.

To encode slices and arrays, the elements are encoded as an RLP list of the value's
elements. Note that arrays and slices with element type uint8 or byte are always encoded
as an RLP string.

A Go string is encoded as an RLP string.

An unsigned integer value is encoded as an RLP string. Zero always encodes as an empty RLP
string. big.Int values are treated as integers. Signed integers (int, int8, int16, ...)
are not supported and will return an error when encoding.

Boolean values are encoded as the unsigned integers zero (false) and one (true).

An interface value encodes as the value contained in the interface.

Floating point numbers, maps, channels and functions are not supported.

# Decoding Rules

Decoding uses the following type-dependent rules:

If the type implements the Decoder interface, DecodeRLP is called.

To decode into a pointer, the value will be decoded as the element type of the pointer. If
the pointer is nil, a new value of the pointer's element type is allocated. If the pointer
is non-nil, the existing value will be reused. Note that package rlp never leaves a
pointer-type struct field as nil unless one of the "nil" struct tags is present.

To decode into a struct, decoding expects the input to be an RLP list. The decoded
elements of the list are assigned to each public field in the order given by the struct's
definition. The input list must contain an element for each decoded field. Decoding
returns an error if there are too few or too many elements for the struct.

To decode into a slice, the input must be a list and the resulting slice will contain the
input elements in order. For byte slices, the input must be an RLP string. Array types
decode similarly, with the additional restriction that the number of input elements (or
bytes) must match the array's defined length.

To decode into a Go string, the input must be an RLP string. The input bytes are taken
as-is and will not necessarily be valid UTF-8.

To decode into an unsigned integer type, the input must also be an RLP string. The bytes
are interpreted as a big endian representation of the integer. If the RLP string is larger
than the bit size of the type, decoding will return an error. Decode also supports
*big.Int. There is no size limit for big integers.

To decode into a boolean, the input must contain an unsigned integer of value zero (false)
or one (true).

To decode into an interface value, one of these types is stored in the value:

	[]interface{}, for RLP lists
	[]byte, for RLP strings

Non-empty interface types are not supported when decoding.
Signed integers, floating point numbers, maps, channels and functions cannot be decoded into.

# Struct Tags

As with other encoding packages, the "-" tag ignores fields.

	type StructWithIgnoredField struct{
	    Ignored uint `rlp:"-"`
	    Field   uint
	}

Go struct values encode/decode as RLP lists. There are two ways of influencing the mapping
of fields to list elements. The "tail" tag, which may only be used on the last exported
struct field, allows slurping up any excess list elements into a slice.

	type StructWithTail struct{
	    Field   uint
	    Tail    []string `rlp:"tail"`
	}

The "optional" tag says that the field may be omitted if it is zero-valued. If this tag is
used on a struct field, all subsequent public fields must also be declared optional.

When encoding a struct with optional fields, the output RLP list contains all values up to
the last non-zero optional field.

When decoding into a struct, optional fields may be omitted from the end of the input
list. For the example below, this means input lists of one, two, or three elements are
accepted.

	type StructWithOptionalFields struct{
	     Required  uint
	     Optional1 uint `rlp:"optional"`
	     Optional2 uint `rlp:"optional"`
	}

The "nil", "nilList" and "nilString" tags apply to pointer-typed fields only, and change
the decoding rules for the field type. For regular pointer fields without the "nil" tag,
input values must always match the required input length exactly and the decoder does not
produce nil values. When the "nil" tag is set, input values of size zero decode as a nil
pointer. This is especially useful for recursive types.

	type StructWithNilField struct {
	    Field *[3]byte `rlp:"nil"`
	}

In the example above, Field allows two possible input sizes. For input 0xC180 (a list
containing an empty string) Field is set to nil after decoding. For input 0xC483000000 (a
list containing a 3-byte string), Field is set to a non-nil array pointer.

RLP supports two kinds of empty values: empty lists and empty strings. When using the
"nil" tag, the kind of empty value allowed for a type is chosen automatically. A field
whose Go type is a pointer to an unsigned integer, string, boolean or byte array/slice
expects an empty RLP string. Any other pointer field type encodes/decodes as an empty RLP
list.

The choice of null value can be made explicit with the "nilList" and "nilString" struct
tags. Using these tags encodes/decodes a Go nil pointer value as the empty RLP value kind
defined by the tag.
*/
package rlp

/*
包 rlp 实现了 RLP 序列化格式。

RLP（递归线性前缀）的目的是编码任意嵌套数组的二进制数据，
而 RLP 是在以太坊中序列化对象时使用的主要编码方法。
RLP 唯一的目的就是编码结构；具体原子数据类型（例如字符串、整数、浮点数）的编码则留给更高级别的协议处理。在以太坊中，整数必须以大端二进制形式表示，并且不能有前导零（因此使得整数值零等同于空字符串）。

RLP 值通过类型标签进行区分。类型标签位于输入流中的值之前，并定义后续字节的大小和种类。

编码规则

包 rlp 使用反射并基于 Go 值的数据类型进行 RLP 编码。

如果该类型实现了 Encoder 接口，则 Encode 调用 EncodeRLP。它不会对 nil 指针值调用 EncodeRLP。

要编码指针，将被指向的值进行编码。对于结构体、切片或数组类型，如果指向的是 nil 指针，则始终将其编码为空 RLP 列表，除非切片或数组具有元素类型 byte。对于其他任何值，nil 指针将被编码为空字符串。

结构体值作为所有已编解码公有字段组成的 RLP 列表进行编码。支持递归结构体类型。

要对切片和数组进行编码，其元素作为该值元素组成的 RLP 列表进行编码。请注意，对于元素类型为 uint8 或 byte 的数组和切片，总是将其作为 RLP 字符串进行编码。

Go 字符串被编为一个 RLP 字符串。

无符号整数值被编为一个 RLP 字符串。零总是被编为一个空 R LP 字符串。big.Int 值视作整数处理。不支持带符号整型（int, int8, int16 等），在编解码时会返回错误。

布尔值分别被编为无符号整数零（假）和一（真）。

接口值按接口中包含的实际值得到编解码结果。

不支持浮点数、映射、通道及函数等数据结构。


解码规则

解码使用以下依赖于数据类型的方法：

如果该类型实现了 Decoder 接口，则调用 DecodeR LP 。

要解码到指针，该值将按照指针所指向的数据元素来解码。如果指针为 nil，将分配新的对应元素类别的新实例。如果指针对应已有实例，将重用现存价值。但需要注意的是，在 package rlp 中，不会让 pointer 类型 struct 字段保持 nil 除非存在某个 "nil" struct 标签.

要解码到结构体，需要期望输入为一个 R LP 列表。这些列表中的每个已解析项都按顺序赋给相应 public field 。若输入列表缺少某个字段或者多出多余字段，会返回错误提示.

要解码成切片，输入必须是一组列表，而生成后的 slice 将包含这些 input 元素。有关于字节 slice ， 输入必须是一组 RL P string 。 数组也类似地执行 decode ，但额外限制条件要求 input 元素数量 ( 或 bytes ) 必须匹配 array 定义长度 .

当 decoding 到 Go string 时 , 输入需满足 RL P string 格式 . 输入 bytes 会直接取用 , 不一定符合有效 UTF - 8 标准 .

当 decoding 到 unsigned integer type 时 , 输入也需满足 RL P string 格式 . Bytes 被解释成大端表示法. 如果 RL P string 超过 bit size 范围, 解密过程会报错. Decode 同样支持 * big.Int . 对于 big integers 没有限制条件.

当 decoding 为 boolean 时 , 输入需包含无签名整型0(假)或者1(真).

当 decoding 为 interface value 时，其中一种如下所示的数据存储在其中:

[]interface{} 用于RL P lists
[]byte 用于RL P strings

非空接口 types 在 decod ing过程中是不受支持.
带签名整型 、 浮点数 、 映射 、 通道以及函数均无法完成 decod ing 操作.


Struct Tags

与其他 encoding 包一样，“-" 标签用于忽略字段.

type StructWithIgnoredField struct{
Ignored uint ￼
Field   uint
}

Go struct values 编译/解析成为 RL P lists 。影响 fields mapping 有两种方式
*/
