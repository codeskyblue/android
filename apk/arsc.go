// Package apk reference article
// http://blog.csdn.net/mldxs/article/details/44956911
// http://www.freebuf.com/articles/terminal/75944.html
package apk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"unicode/utf16"
	"unicode/utf8"
)

const (
	RES_NULL_TYPE                = 0x0000
	RES_STRING_POOL_TYPE         = 0x0001
	RES_TABLE_TYPE               = 0x0002
	RES_XML_FIRST_CHUNK_TYPE     = 0x0100
	RES_XML_START_NAMESPACE_TYPE = 0x0100
	RES_XML_END_NAMESPACE_TYPE   = 0x0101
	RES_XML_START_ELEMENT_TYPE   = 0x0102
	RES_XML_END_ELEMENT_TYPE     = 0x0103
	RES_XML_CDATA_TYPE           = 0x0104
	RES_XML_LAST_CHUNK_TYPE      = 0x017f
	RES_XML_RESOURCE_MAP_TYPE    = 0x0180
	RES_TABLE_PACKAGE_TYPE       = 0x0200
	RES_TABLE_TYPE_TYPE          = 0x0201
	RES_TABLE_TYPE_SPEC_TYPE     = 0x0202

	RES_FLAG_UTF16  = 0x0000
	RES_FLAG_UTF8   = 0x0100
	RES_FLAG_SORTED = 0x0001
)

func UnmarshalArsc(data []byte) error {
	body := bytes.NewReader(data)
	var header, headerSize uint16
	var chunkSize, packageCount uint32
	binary.Read(body, binary.LittleEndian, &header)
	if header != RES_TABLE_TYPE {
		return errors.New("ARSC file has wrong header")
	}
	// ignore headerSize
	binary.Read(body, binary.LittleEndian, &headerSize)
	binary.Read(body, binary.LittleEndian, &chunkSize)
	if int(chunkSize) != len(data) {
		return errors.New("ARSC file has the wrong size")
	}
	binary.Read(body, binary.LittleEndian, &packageCount)
	if err := readGlobalStringPool(body); err != nil {
		return err
	}
	return nil
}

func readGlobalStringPool(body *bytes.Reader) error {
	var blockType, headerSize uint16
	var blockSize, stringsCount, stylesCount, flags, stringsStart, stylesStart uint32
	binary.Read(body, binary.LittleEndian, &blockType)
	if blockType != RES_STRING_POOL_TYPE {
		return fmt.Errorf("Error: expect block 0x%04x", RES_STRING_POOL_TYPE)
	}
	// ignore header size
	binRead(body, &headerSize, &blockSize, &stringsCount, &stylesCount, &flags, &stringsStart, &stylesStart)
	// binary.Read(body, binary.LittleEndian, &headerSize)
	// binary.Read(body, binary.LittleEndian, &blockSize)
	// binary.Read(body, binary.LittleEndian, &stringsCount)
	// binary.Read(body, binary.LittleEndian, &stylesCount)
	log.Printf("0x%0x", blockSize)
	log.Printf("strings count %d", stringsCount)
	log.Printf("strings start 0x%04x", stringsStart)
	log.Printf("flags: %04x", flags)
	log.Printf("%d", stylesCount)

	stringsStart += 0x0C // add header offset

	var delim []byte
	if flags & ^uint32(RES_FLAG_SORTED) == RES_FLAG_UTF8 {
		delim = []byte{0x00}
	} else {
		delim = []byte{0x00, 0x00}
	}

	var offset uint32
	for i := 0; i < int(stringsCount); i++ {
		if err := binRead(body, &offset); err != nil {
			return err
		}
		chars := extractChars(body, stringsStart+offset, delim)
		if len(delim) == 2 { // utf16
			chars, _ = decodeUTF16(chars)
		}
		log.Println(string(chars))
	}
	return nil
}

func binRead(rd io.Reader, vs ...interface{}) error {
	for _, v := range vs {
		err := binary.Read(rd, binary.LittleEndian, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func extractChars(rd io.ReaderAt, offset uint32, delim []byte) []byte {
	length := make([]byte, 2)
	rd.ReadAt(length, int64(offset))
	// actually should be
	// len = (((hbyte & 0x7F) << 8)) | lbyte;
	// but in reality it does not

	strLen := int(length[0])
	log.Println("L=", strLen)
	chars := make([]byte, 0, strLen*len(delim))
	c := make([]byte, len(delim))
	for i := 0; i < strLen; i++ {
		startOff := int(offset) + 2 + len(delim)*i
		_, err := rd.ReadAt(c, int64(startOff))
		if err != nil {
			break
		}
		if bytes.Equal(c, delim) {
			break
		}
		chars = append(chars, c...)
	}
	return chars
}

// Convert UTF16 -> UTF8
func decodeUTF16(b []byte) ([]byte, error) {
	if len(b)%2 != 0 {
		return nil, fmt.Errorf("Must have even length byte slice")
	}

	u16s := make([]uint16, 1)
	ret := &bytes.Buffer{}
	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}
	return ret.Bytes(), nil
}
