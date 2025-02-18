package protocol

import (
	"errors"
	"strconv"
)

type Reply interface {
	ToBytes() []byte
}

const CRLF = "\r\n"

type StatusReply struct {
	Status string
}

func NewStatusReply(status string) *StatusReply {
	return &StatusReply{
		Status: status,
	}
}

func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

type ErrReply struct {
	Err string
}

func NewErrReply(status string) *ErrReply {
	return &ErrReply{
		Err: status,
	}
}

func (r *ErrReply) ToBytes() []byte {
	return []byte("-" + r.Err + CRLF)
}

func (r *ErrReply) Error() error {
	return errors.New(r.Err)
}

func IsErrorReply(reply Reply) bool {
	return reply.ToBytes()[0] == '-'
}

const nullBulkBytes = "$-1" + CRLF
const emptyBulkBytes = "$0" + CRLF + CRLF

type BulkReply struct {
	Value []byte
}

func NewBulkReply(value []byte) *BulkReply {
	return &BulkReply{
		Value: value,
	}
}

func NewEmptyBulkReply() *BulkReply {
	return &BulkReply{
		Value: []byte{},
	}
}

func NewNullBulkReply() *BulkReply {
	return &BulkReply{
		Value: nil,
	}
}

func (r *BulkReply) ToBytes() []byte {
	if r.Value == nil {
		return []byte(nullBulkBytes)
	}
	if len(r.Value) == 0 {
		return []byte(emptyBulkBytes)
	}
	return []byte("$" + strconv.Itoa(len(r.Value)) + CRLF + string(r.Value) + CRLF)
}

type IntReply struct {
	Value int64
}

func NewIntReply(value int64) *IntReply {
	return &IntReply{
		Value: value,
	}
}

func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Value, 10) + CRLF)
}

type MultiBulkReply struct {
	Values [][]byte
}

var emptyMultiBulkBytes = []byte("*0" + CRLF)

func NewMultiBulkReply(values [][]byte) *MultiBulkReply {
	return &MultiBulkReply{
		Values: values,
	}
}

func NewEmptyMultiBulkReply() *MultiBulkReply {
	return &MultiBulkReply{
		Values: nil,
	}
}

func (r *MultiBulkReply) ToBytes() []byte {
	if r.Values == nil || len(r.Values) == 0 {
		return emptyMultiBulkBytes
	}
	buf := "*" + strconv.Itoa(len(r.Values)) + CRLF
	for _, value := range r.Values {
		buf += "$" + strconv.Itoa(len(value)) + CRLF + string(value) + CRLF
	}
	return []byte(buf)
}
