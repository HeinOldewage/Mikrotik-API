package Mikrotik

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

type Word []byte

type Sentence []Word

func (s *Sentence) Add(w Word) {
	*s = append(*s, w)
}

func (s *Sentence) Contains(w Word) bool {
	for _, ww := range *s {
		if strings.Index(string(ww), string(w)) >= 0 {
			return true
		}
	}
	return false
}

func (s *Sentence) Index(ss string) int {
	for k, ww := range *s {
		if strings.Index(string(ww), ss) >= 0 {
			return k
		}
	}
	return -1
}

func (s *Sentence) Get(ss string) (Word,bool) {
	ss = "="+ss+"="
	for _, ww := range *s {
		if strings.Index(string(ww), ss) >= 0 {
			return ww[len(ss):],true
		}
	}
	return nil,false
}

func Command(s string) Word {
	//Assume that command is in correct format for now.
	return []byte(s)
}

func Attribute(key, val string) Word {
	//Assume that command is in correct format for now.
	return []byte("=" + key + "=" + val)
}

func APIAttribute(key, val string) Word {
	//Assume that command is in correct format for now.
	return []byte("." + key + "=" + val)
}

func Query(key, val string) Word {
	//Assume that command is in correct format for now.
	return []byte("?" + key + "=" + val)
}

func (w Word) String() string {
	return string(w)
}

func encodelength(i int) []byte {
	buf := new(bytes.Buffer)
	var toWrite uint32 = uint32(i)

	if (0 <= i) && (i <= 0x7F) {
		err := binary.Write(buf, binary.LittleEndian, toWrite)
		if err != nil {
			panic(err)
		}
		return buf.Bytes()[:1]
	} else if (0x80 <= i) && (i <= 0x3FFF) {
		err := binary.Write(buf, binary.LittleEndian, toWrite|0x8000)
		if err != nil {
			panic(err)
		}
		return buf.Bytes()[:2]
	} else if (0x4000 <= i) && (i <= 0x1FFFFF) {
		err := binary.Write(buf, binary.LittleEndian, toWrite|0xC00000)
		if err != nil {
			panic(err)
		}
		return buf.Bytes()[:3]
	} else if (0x200000 <= i) && (i <= 0xFFFFFFF) {
		err := binary.Write(buf, binary.LittleEndian, toWrite|0xE0000000)
		if err != nil {
			panic(err)
		}
		return buf.Bytes()[:4]
	} else {
		//This case is ill defined in the docs (states both 4 and 5 bytes required?).
		//Should not be a problem as messages of this length(268,435,456 bytes) is unlikely
		//The example python code on the web page did clear things up.
		buf.WriteByte(0xF0)
		err := binary.Write(buf, binary.LittleEndian, toWrite)
		if err != nil {
			panic(err)
		}
		return buf.Bytes()[:5]
	}
}

func decodelength(b []byte) (length int, bytesUsed int, err error) {
	if len(b) == 0 {
		return -1, -1, fmt.Errorf("Could not decode length")
	}
	if b[0]&0x80 == 0x00 {
		return int(b[0]), 1, nil
	} else
	{
		if (b[0] & 0xC0) == 0x80 {
			if len(b) < 2 {
				return -1, -1, fmt.Errorf("Could not decode length")
			}
			return (int(b[0])& ^0xC0) <<8 + int(b[1]) , 2, nil
		} else if (b[0] & 0xE0) == 0xC0 {
			fmt.Println("3 byte case")
			if len(b) < 3 {
				return -1, -1, fmt.Errorf("Could not decode length")
			}
			return  int(b[0])& ^0xE0<<8+int(b[1])<<8 + int(b[2]) , 3, nil
		} else if b[0]&0x80 == 0x00 {
			fmt.Println("4 byte case")
			if len(b) < 4 {
				return -1, -1, fmt.Errorf("Could not decode length")
			}
			return (int(b[0])& ^0x80) <<8 +int(b[1])<<8+int(b[2])<<8 + int(b[3])  , 4, nil
		} else {
			fmt.Println("5 byte case")
			if len(b) < 5 {
				return -1, -1, fmt.Errorf("Could not decode length")
			}
			return int(((b[1]<<8+b[2])<<8+b[3])<<8 + b[4]), 5, nil
		}
	}
}

func (s Sentence) Encode() []byte {
	res := make([]byte, 0)
	for _, w := range s {
		//send length
		res = append(res, encodelength(len(w))...)
		//send data
		res = append(res, w...)
	}
	res = append(res, 0)
	return res
}

/*
Will return an error if insufficient bytes are available for the next word
Returns the number of bytes used to create the word. This is not equal to word length as the word does not contain the length bytes
*/

func Decode(b []byte) (Word, int, error) {
	length, bu, err := decodelength(b)
	if err != nil {
		return nil, 0, err
	}
	if len(b) < length+bu {
		return nil, 0, fmt.Errorf("Could not decode word")
	}
	res := make(Word, length)
	copy(res, b[bu:])
	return res, length + bu, nil
}
