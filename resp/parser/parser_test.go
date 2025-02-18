package parser

import (
	"github.com/stretchr/testify/assert"
	"godis/resp/protocol"
	"testing"
)

func TestParseOne(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected protocol.Reply
		wantErr  bool
	}{
		{
			name:  "simple string",
			input: []byte("+OK\r\n"),
			expected: &protocol.StatusReply{
				Status: "OK",
			},
		},
		{
			name:  "error",
			input: []byte("-ERR unknown command 'foobar'\r\n"),
			expected: &protocol.ErrReply{
				Err: "ERR unknown command 'foobar'",
			},
		},
		{
			name:  "integer positive",
			input: []byte(":1000\r\n"),
			expected: &protocol.IntReply{
				Value: 1000,
			},
		},
		{
			name:  "integer negative",
			input: []byte(":-1000\r\n"),
			expected: &protocol.IntReply{
				Value: -1000,
			},
		},
		{
			name:  "bulk string",
			input: []byte("$6\r\nfoobar\r\n"),
			expected: &protocol.BulkReply{
				Value: []byte("foobar"),
			},
		},
		{
			name:  "null bulk string",
			input: []byte("$-1\r\n"),
			expected: &protocol.BulkReply{
				Value: nil,
			},
		},
		{
			name:  "empty bulk string",
			input: []byte("$0\r\n\r\n"),
			expected: &protocol.BulkReply{
				Value: []byte(""),
			},
		},
		{
			name:  "multi line bulk string",
			input: []byte("$12\r\nhello\r\nworld\r\n"),
			expected: &protocol.BulkReply{
				Value: []byte("hello\r\nworld"),
			},
		},
		{
			name:  "array",
			input: []byte("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"),
			expected: &protocol.MultiBulkReply{
				Values: [][]byte{
					[]byte("foo"),
					[]byte("bar"),
				},
			},
		},
		{
			name:  "empty array",
			input: []byte("*0\r\n"),
			expected: &protocol.MultiBulkReply{
				Values: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload := ParseOne(tc.input)
			if payload.Err != nil {
				if tc.wantErr {
					return
				}
				t.Fatalf("unexpected error: %v", payload.Err)
			}
			assert.Equal(t, tc.expected, payload.Data)
		})
	}
}
