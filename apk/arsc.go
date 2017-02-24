// Package apk reference article
// http://blog.csdn.net/mldxs/article/details/44956911
package apk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
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
	log.Printf("starts start %04x %04x", stringsStart, 0x7471+0x0c)

	body.Seek(int64(stringsStart)+0x0c-1, io.SeekStart)
	log.Printf("%0x", int64(stringsStart)+0x0c)
	var strSize uint8
	// binRead(body, &totalCount)
	for i := 0; i < int(2); i++ {
		binRead(body, &strSize, &strSize, &strSize)
		// log.Println(totalCount, strSize)
		chars := make([]byte, int(strSize))
		io.ReadAtLeast(body, chars, int(strSize))
		log.Println(string(chars))

	}
	return nil
}

func binRead(rd io.Reader, vs ...interface{}) {
	for _, v := range vs {
		binary.Read(rd, binary.LittleEndian, v)
	}
}

// func compXmlStringAt(arr io.ReaderAt, meta stringsMeta, strOff uint32) string {
// 	if strOff == 0xffffffff {
// 		return ""
// 	}
// 	length := make([]byte, 2)
// 	off := meta.StringDataOffset + meta.DataOffset[strOff]
// 	arr.ReadAt(length, int64(off))
// 	strLen := int(length[1]<<8 + length[0])

// 	chars := make([]byte, int64(strLen))
// 	ii := 0
// 	for i := 0; i < strLen; i++ {
// 		c := make([]byte, 1)
// 		arr.ReadAt(c, int64(int(off)+2+ii))

// 		if c[0] == 0 {
// 			i--
// 		} else {
// 			chars[i] = c[0]
// 		}
// 		ii++
// 	}

// 	return string(chars)
// } // end of compXmlStringAt
