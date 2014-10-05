package main

import "os"
import "log"
import "fmt"
import "bytes"

// The beginnings of a tail(1) clone, for the purpose of learning Go.
//
// Currently it tries to find the 10 last lines in a file, and gets 7
// instead. Oops.
func main() {
	filename := os.Args[1]
	cursor, err := newLineCursor(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close()
	cursor.Prev(10)
	text, err := cursor.Read()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(text)
	os.Exit(0)
}

type lineCursor struct {
	file *os.File
	byteOffset int64
	lineOffset int64
	endOffset  int64
}

func newLineCursor(filename string) (cur lineCursor, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	initialOffset, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		file.Close()
		return
	}
	return lineCursor{file, initialOffset, 0, initialOffset}, nil
}

func (cur *lineCursor) Close() {
	cur.file.Close()
}

func (cur *lineCursor) Prev(n int64) (err error){
	step := int64(n * -64)
	targetOffset := cur.lineOffset - n - 1
	for cur.lineOffset > targetOffset && cur.byteOffset > 0 {
		// make the dubious assumption that we are currently sitting on an
		// ending newline and so we need to back up one byte
		chunkEndOffset := cur.byteOffset - 1
		newOffset, err := cur.file.Seek(chunkEndOffset + step, os.SEEK_SET)
		if err != nil {
			return err
		}
		chunk := make([]byte, chunkEndOffset - newOffset)
		_, err = cur.file.Read(chunk)
		if err != nil {
			return err
		}
		for cur.lineOffset > targetOffset {
			newlineIndex := bytes.LastIndex(chunk, []byte("\n"))
			if newlineIndex == -1 {
				break
			}
			cur.byteOffset = newOffset + int64(newlineIndex) + 1
			cur.lineOffset--
			chunk = chunk[:newlineIndex-1]
		}
	}
	return
}

func (cur *lineCursor) Size() (int64) {
	return cur.endOffset - cur.byteOffset
}

func (cur *lineCursor) Read() (text string, err error) {
	data := make([]byte, cur.endOffset - cur.byteOffset)
	size, err := cur.file.Read(data)
	if err != nil {
		return
	}
	return string(data[:size]), nil
}
