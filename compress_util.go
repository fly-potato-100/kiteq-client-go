package kiteq_client_go

//snappy解压缩
import "github.com/golang/snappy"

func Decompress(body []byte) ([]byte, error) {

	l, err := snappy.DecodedLen(body)
	if nil != err {
		return nil, err
	}
	if l%256 != 0 {
		l = (l/256 + 1) * 256
	}
	dest := make([]byte, l)
	decompressData, err := snappy.Decode(dest, body)
	if nil != err {
		return nil, err
	}
	return decompressData, nil
}

//snapp压缩
func Compress(body []byte) ([]byte, error) {

	l := snappy.MaxEncodedLen(len(body))
	if l%256 != 0 {
		l = (l/256 + 1) * 256
	}

	dest := make([]byte, l)
	return snappy.Encode(dest, body), nil
}
