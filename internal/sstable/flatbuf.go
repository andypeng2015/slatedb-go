package sstable

import (
	"github.com/google/flatbuffers/go"
	"github.com/samber/mo"
	"github.com/slatedb/slatedb-go/gen"
	"github.com/slatedb/slatedb-go/internal/compress"
)

type SSTableIndexData struct {
	data []byte
}

func NewSSTableIndexData(data []byte) *SSTableIndexData {
	return &SSTableIndexData{data: data}
}

func (info *SSTableIndexData) SsTableIndex() *flatbuf.SsTableIndex {
	return flatbuf.GetRootAsSsTableIndex(info.data, 0)
}

func (info *SSTableIndexData) Size() int {
	return len(info.data)
}

func (info *SSTableIndexData) Clone() *SSTableIndexData {
	data := make([]byte, len(info.data))
	copy(data, info.data)
	return &SSTableIndexData{
		data: data,
	}
}

// FlatBufferSSTableIndexCodec defines how we
// encode SsTableIndex to byte slice and decode byte slice back to SSTableIndex
type FlatBufferSSTableIndexCodec struct{}

func (f FlatBufferSSTableIndexCodec) Encode(index flatbuf.SsTableIndexT) []byte {
	builder := flatbuffers.NewBuilder(0)
	offset := index.Pack(builder)
	builder.Finish(offset)
	return builder.FinishedBytes()
}

func (f FlatBufferSSTableIndexCodec) Decode(data []byte) *flatbuf.SsTableIndexT {
	indexData := NewSSTableIndexData(data)
	return indexData.SsTableIndex().UnPack()
}

// FlatBufferSSTableInfoCodec implements SsTableInfoCodec and defines how we
// encode sstable.Info to byte slice and decode byte slice back to sstable.Info
type FlatBufferSSTableInfoCodec struct{}

func (f FlatBufferSSTableInfoCodec) Encode(info *Info) []byte {
	fbSSTInfo := SstInfoToFlatBuf(info)
	builder := flatbuffers.NewBuilder(0)
	offset := fbSSTInfo.Pack(builder)
	builder.Finish(offset)
	return builder.FinishedBytes()
}

func (f FlatBufferSSTableInfoCodec) Decode(data []byte) *Info {
	info := flatbuf.GetRootAsSsTableInfo(data, 0)
	return SstInfoFromFlatBuf(info)
}

func SstInfoFromFlatBuf(info *flatbuf.SsTableInfo) *Info {
	firstKey := mo.None[[]byte]()
	keyBytes := info.FirstKeyBytes()
	if keyBytes != nil {
		firstKey = mo.Some(keyBytes)
	}

	return &Info{
		FirstKey:         firstKey,
		IndexOffset:      info.IndexOffset(),
		IndexLen:         info.IndexLen(),
		FilterOffset:     info.FilterOffset(),
		FilterLen:        info.FilterLen(),
		CompressionCodec: compress.CodecFromFlatBuf(info.CompressionFormat()),
	}
}

func SstInfoToFlatBuf(info *Info) *flatbuf.SsTableInfoT {
	var firstKey []byte
	if info.FirstKey.IsPresent() {
		firstKey, _ = info.FirstKey.Get()
	}

	return &flatbuf.SsTableInfoT{
		FirstKey:          firstKey,
		IndexOffset:       info.IndexOffset,
		IndexLen:          info.IndexLen,
		FilterOffset:      info.FilterOffset,
		FilterLen:         info.FilterLen,
		CompressionFormat: compress.CodecToFlatBuf(info.CompressionCodec),
	}
}
