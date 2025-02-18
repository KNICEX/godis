package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"godis/pkg/logx"
	"godis/resp/protocol"
	"io"
	"strconv"
)

type Payload struct {
	Data protocol.Reply
	Err  error
}

func ParseStream(reader io.Reader) <-chan *Payload {
	return parse0(reader)
}

func ParseOne(data []byte) *Payload {
	return <-parse0(bytes.NewReader(data))
}

func parse0(readerRaw io.Reader) <-chan *Payload {
	// TODO 是否使用有缓冲的channel
	ch := make(chan *Payload)
	reader := bufio.NewReader(readerRaw)
	go func() {
		defer func() {
			if er := recover(); er != nil {
				logx.L().Error(er)
			}
			close(ch)
		}()
		for {
			// 读取header
			header, err := reader.ReadBytes('\n')
			if err != nil {
				ch <- &Payload{Err: err}
				return
			}
			length := len(header)
			if length <= 2 || header[length-2] != '\r' {
				continue
			}

			header = bytes.TrimSuffix(header, []byte{'\r', '\n'})
			switch header[0] {
			case '+':
				// 简单字符串，用来表示状态 例如: +OK\r\n 非二进制安全
				content := string(header[1:])
				ch <- &Payload{Data: protocol.NewStatusReply(content)}
			case '-':
				// 错误信息 例如: -ERR unknown command 'foobar'\r\n 非二进制安全
				content := string(header[1:])
				ch <- &Payload{Data: protocol.NewErrReply(content)}
			case ':':
				// 整数值 例如: :1000\r\n
				value, err := strconv.ParseInt(string(header[1:]), 10, 64)
				if err != nil {
					ch <- &Payload{Err: err}
				} else {
					ch <- &Payload{Data: protocol.NewIntReply(value)}
				}
			case '$':
				// 字符串值 例如: $6\r\nfoobar\r\n 表示一个长度为6的字符串"foobar" 二进制安全
				// 长度为-1时表示空值
				ch <- parseBulkString(header, reader)
			case '*':
				// 数组 例如: *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
				// 表示一个包含3个元素的数组，每个元素都是一个bulk string
				ch <- parseArr(header, reader)
			default:
				ch <- &Payload{Err: fmt.Errorf("invalid header: %s", header)}
			}
		}
	}()
	return ch
}

func parseBulkString(header []byte, reader *bufio.Reader) *Payload {
	strLen, err := strconv.ParseInt(string(header[1:]), 10, 64)
	if err != nil {
		return &Payload{Err: err}
	} else if strLen == -1 {
		return &Payload{Data: protocol.NewNullBulkReply()}
	} else if strLen == 0 {
		return &Payload{Data: protocol.NewEmptyBulkReply()}
	}

	body := make([]byte, strLen+2)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return &Payload{Err: err}
	}
	return &Payload{
		Data: protocol.NewBulkReply(body[:strLen]),
	}
}

func parseArr(header []byte, reader *bufio.Reader) *Payload {
	arrLen, err := strconv.ParseInt(string(header[1:]), 10, 64)
	if err != nil || arrLen < 0 {
		return &Payload{Err: fmt.Errorf("invalid array header %s", header)}
	} else if arrLen == 0 {
		return &Payload{Data: protocol.NewEmptyMultiBulkReply()}
	}

	lines := make([][]byte, 0, arrLen)
	for i := 0; i < int(arrLen); i++ {
		// 读取header
		var line []byte
		line, err = reader.ReadBytes('\n')
		if err != nil {
			return &Payload{Err: err}
		}
		length := len(line)
		if length < 4 || line[length-2] != '\r' || line[0] != '$' {
			return &Payload{Err: fmt.Errorf("invalid bulk string header %s", line)}
		}

		// 读取body
		strLen, err := strconv.ParseInt(string(line[1:length-2]), 10, 64)
		if err != nil || strLen < -1 {
			return &Payload{Err: fmt.Errorf("invalid bulk string length %s", line)}
		} else if strLen == -1 {
			// 空值
			lines = append(lines, nil)
		} else {
			body := make([]byte, strLen+2)
			_, err = io.ReadFull(reader, body)
			if err != nil {
				return &Payload{Err: err}
			}
			lines = append(lines, body[:strLen])
		}
	}
	return &Payload{Data: protocol.NewMultiBulkReply(lines)}
}
