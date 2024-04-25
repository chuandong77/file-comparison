package main

import (
	"fmt"
	"github.com/adrium/goheif"
	"image/jpeg"
	"io"
	"os"
)

func HeicToJpg(fin string, fout string) error {
	fi, err := os.Open(fin)
	if err != nil {
		fmt.Println("打开源文件失败")
		return err
	}
	defer fi.Close()

	exif, err := goheif.ExtractExif(fi)
	if err != nil {
		fmt.Println("读取源文件 exif 失败")
		return err
	}

	img, err := goheif.Decode(fi)
	if err != nil {
		fmt.Println("读取源文件失败")
		return err
	}

	fo, err := os.OpenFile(fout, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("读取转换文件失败" + err.Error())
		return err
	}
	defer fo.Close()

	w, _ := newWriterExif(fo, exif)
	err = jpeg.Encode(w, img, nil)
	if err != nil {
		fmt.Println("转换格式失败")
		return err
	}

	fmt.Println("转换格式成功！")

	return nil
}

// Skip Writer for exif writing
type writerSkipper struct {
	w           io.Writer
	bytesToSkip int
}

func (w *writerSkipper) Write(data []byte) (int, error) {
	if w.bytesToSkip <= 0 {
		return w.w.Write(data)
	}

	if dataLen := len(data); dataLen < w.bytesToSkip {
		w.bytesToSkip -= dataLen
		return dataLen, nil
	}

	if n, err := w.w.Write(data[w.bytesToSkip:]); err == nil {
		n += w.bytesToSkip
		w.bytesToSkip = 0
		return n, nil
	} else {
		return n, err
	}
}

func newWriterExif(w io.Writer, exif []byte) (io.Writer, error) {
	writer := &writerSkipper{w, 2}
	soi := []byte{0xff, 0xd8}
	if _, err := w.Write(soi); err != nil {
		return nil, err
	}

	if exif != nil {
		app1Marker := 0xe1
		markerlen := 2 + len(exif)
		marker := []byte{0xff, uint8(app1Marker), uint8(markerlen >> 8), uint8(markerlen & 0xff)}
		if _, err := w.Write(marker); err != nil {
			return nil, err
		}

		if _, err := w.Write(exif); err != nil {
			return nil, err
		}
	}

	return writer, nil
}