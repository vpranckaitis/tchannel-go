package typed

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimple(t *testing.T) {
	buf := make([]byte, 200)

	var r ReadBuffer
	var w WriteBuffer

	{
		w.Wrap(buf)
		w.WriteByte(0xFC)
		r.Wrap(buf)
		assert.Equal(t, byte(0xFC), r.ReadByte())
	}

	{
		w.Wrap(buf)
		w.WriteUint16(0xDEAD)
		r.Wrap(buf)
		assert.Equal(t, uint16(0xDEAD), r.ReadUint16())
	}

	{
		w.Wrap(buf)
		w.WriteUint32(0xBEEFDEAD)
		r.Wrap(buf)
		assert.Equal(t, uint32(0xBEEFDEAD), r.ReadUint32())
	}

}

func TestReadWrite(t *testing.T) {
	s := "the small brown fix"
	bslice := []byte("jumped over the lazy dog")

	w := NewWriteBufferWithSize(1024)
	w.WriteUint64(0x0123456789ABCDEF)
	w.WriteUint32(0xABCDEF01)
	w.WriteUint16(0x2345)
	w.WriteByte(0xFF)
	w.WriteString(s)
	w.WriteBytes(bslice)
	require.NoError(t, w.Err())

	var b bytes.Buffer
	w.FlushTo(&b)

	r := NewReadBufferWithSize(1024)
	r.FillFrom(bytes.NewReader(b.Bytes()), len(b.Bytes()))

	{
		n := r.ReadUint64()
		assert.Equal(t, n, uint64(0x0123456789ABCDEF), "mismatched uint64")
	}
	{
		n := r.ReadUint32()
		assert.Equal(t, n, uint32(0xABCDEF01), "mismatched uint32")
	}
	{
		n := r.ReadUint16()
		assert.Equal(t, n, uint16(0x2345), "mismatched uint16")
	}
	{
		n := r.ReadByte()
		assert.Equal(t, n, byte(0xFF), "mismatched byte")
	}
	{

		rs := r.ReadString(len(s))
		assert.Equal(t, rs, s, "mismatched string")
	}
	{
		rbslice := r.ReadBytes(len(bslice))
		assert.Equal(t, rbslice, bslice, "mismatched byte slices")
	}

	require.NoError(t, r.Err())
}

func TestSeek(t *testing.T) {
	w := NewWriteBufferWithSize(1024)
	pos := w.CurrentPos()
	w.WriteUint16(0)
	w.WriteString("Hello NYC")
	endPos := w.CurrentPos()

	require.Nil(t, w.Seek(pos))
	w.WriteUint16(uint16(len("Hello NYC")))
	require.Nil(t, w.Seek(endPos))

	pos = w.CurrentPos()
	w.WriteUint16(0) // We'll come back to this
	w.WriteString("The quick brown fox")
	endPos = w.CurrentPos()

	require.Nil(t, w.Seek(pos))
	w.WriteUint16(uint16(len("The quick brown fox")))
	require.Nil(t, w.Seek(endPos))

	var b bytes.Buffer
	_, err := w.FlushTo(&b)
	require.Nil(t, err)

	r := NewReadBufferWithSize(1024)
	_, err = r.FillFrom(bytes.NewReader(b.Bytes()), w.BytesWritten())
	require.Nil(t, err)

	n := r.ReadUint16()
	s := r.ReadString(int(n))
	assert.Equal(t, "Hello NYC", s)

	n = r.ReadUint16()
	s = r.ReadString(int(n))
	assert.Equal(t, "The quick brown fox", s)
}

func TestDeferredWrites(t *testing.T) {
	w := NewWriteBufferWithSize(1024)
	u16ref := w.DeferUint16()
	require.NotNil(t, u16ref)

	u32ref := w.DeferUint32()
	require.NotNil(t, u32ref)

	u64ref := w.DeferUint64()
	require.NotNil(t, u64ref)

	bref := w.DeferBytes(5)
	require.NotNil(t, bref)

	sref := w.DeferBytes(5)
	require.NotNil(t, sref)

	byteref := w.DeferByte()
	require.NotNil(t, byteref)

	assert.Equal(t, 2+4+8+5+5+1, w.BytesWritten())

	u16ref.Update(2040)
	u32ref.Update(495404)
	u64ref.Update(0x40950459)
	bref.Update([]byte{0x30, 0x12, 0x45, 0x55, 0x65})
	sref.UpdateString("where")
	byteref.Update(0x44)

	var buf bytes.Buffer
	w.FlushTo(&buf)

	r := NewReadBuffer(buf.Bytes())

	u16 := r.ReadUint16()
	assert.Equal(t, uint16(2040), u16)

	u32 := r.ReadUint32()
	assert.Equal(t, uint32(495404), u32)

	u64 := r.ReadUint64()
	assert.Equal(t, uint32(0x40950459), u64)

	b := r.ReadBytes(5)
	assert.Equal(t, []byte{0x30, 0x12, 0x45, 0x55, 0x65}, b)

	s := r.ReadString(5)
	assert.Equal(t, "where", s)

	u8 := r.ReadByte()
	assert.Equal(t, byte(0x44), u8)
	assert.NoError(t, r.Err())
}
