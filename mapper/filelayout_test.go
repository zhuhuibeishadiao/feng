package mapper

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/martinlindhe/feng/template"
	"github.com/martinlindhe/feng/value"
	"github.com/stretchr/testify/assert"
)

func TestFieldPresent(t *testing.T) {

	fl := &FileLayout{}

	test := []struct {
		field    Field
		expected string
	}{
		{Field{Offset: 0x1, Length: 0x2, Value: []uint8{0xff, 0xd8}, Endian: "", Format: value.DataField{Kind: "u8", Range: "2", Slice: false, Label: "U8 array"}}, "  [000001] U8 array                       u8[2]                               ff d8               \n"},
		{Field{Offset: 0x1, Length: 0x4, Value: []uint8{0x52, 0x5e, 0x65, 0xef}, Endian: "little", Format: value.DataField{Kind: "time_t_32", Range: "", Slice: false, Label: "TimeT_32_LE"}}, "  [000001] TimeT_32_LE                    time_t_32 le  2013-10-16T10:09:51Z  52 5e 65 ef         \n"},
	}
	for _, h := range test {
		assert.Equal(t, h.expected, fl.PresentField(&h.field))
	}
}

func TestGetValue(t *testing.T) {
	templateData := `
structs:
  header:
    u8 Number: ??
    u8 Field:
      bit b1000_0000: High bit
    if self.Number in (5):
      u8[FILE_SIZE-self.offset] Extra: ??   # refers to current field offset

layout:
  - header Header
  - header Header2
`

	ds, err := template.UnmarshalTemplateIntoDataStructure([]byte(templateData))
	assert.Equal(t, nil, err)

	data := []byte{
		// Header
		0x04, // Number
		0xff, // Field

		// Header2
		0x05,       // Number
		0x00,       // Field
		0xbb, 0xaa, // Extra
	}

	fl, err := MapReader(bytes.NewReader(data), ds)
	assert.Equal(t, nil, err)

	// Header
	_, val, err := fl.GetValue("Header.Number", &ds.Layout[0])
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte{4}, val)

	_, val, err = fl.GetValue("self.Number", &ds.Layout[0])
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte{4}, val)

	_, val, err = fl.GetValue("self.Field.High bit", &ds.Layout[0])
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte{1}, val)

	_, _, err = fl.GetValue("self.Extra", &ds.Layout[0])
	assert.Equal(t, fmt.Errorf("GetValue: 'Header.Extra' not found"), err)

	// Header2
	_, val, err = fl.GetValue("Header2.Number", &ds.Layout[0])
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte{5}, val)

	_, val, err = fl.GetValue("self.Number", &ds.Layout[1])
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte{5}, val)

	// assert that "u8[FILE_SIZE-self.offset]" evaluates to u8[6-4] == u8[2] (all remaining bytes)
	assert.Equal(t, "  [000004] Extra                          u8[2]                               bb aa               \n", fl.PresentField(&fl.Structs[1].Fields[2]))
	_, val, err = fl.GetValue("self.Extra", &ds.Layout[1])
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte{0xbb, 0xaa}, val)

}

func TestGetFieldOffset(t *testing.T) {
	templateData := `
structs:
  header:
    u8[2] Number: ??
    u8 ID: ??
    u8[self.ID.offset] Padding: ??          # refers to a previous field

layout:
  - header Header
`

	ds, err := template.UnmarshalTemplateIntoDataStructure([]byte(templateData))
	assert.Equal(t, nil, err)

	data := []byte{
		// Header
		0x04, 0x05, // Number
		0x07,       // ID
		0xff, 0xfe, // Padding

	}

	fl, err := MapReader(bytes.NewReader(data), ds)
	assert.Equal(t, nil, err)

	// Header
	_, val, err := fl.GetValue("Header.Number", &ds.Layout[0])
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte{0x04, 0x05}, val)

	_, val, err = fl.GetValue("self.Padding", &ds.Layout[0])
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte{0xff, 0xfe}, val)
}

func TestGetFieldLen(t *testing.T) {
	templateData := `
structs:
  header:
    u8[3] Number: ??
    u8[self.Number.len] Pad1: ??                        # refers to a previous field
    u8[self.Pad1.offset - self.Pad1.len + 1] Pad2: ??   # expression with multiple variables    3 - 3 + 1 = 1

layout:
  - header Header
`

	ds, err := template.UnmarshalTemplateIntoDataStructure([]byte(templateData))
	assert.Equal(t, nil, err)

	data := []byte{
		// Header
		0x04, 0x05, 0x06, // Number
		0xff, 0xfe, 0xfd, // Pad1
		0xaa, // Pad2
	}

	fl, err := MapReader(bytes.NewReader(data), ds)
	assert.Equal(t, nil, err)

	// Header
	_, val, err := fl.GetValue("Header.Number", &ds.Layout[0])
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte{0x04, 0x05, 0x06}, val)

	_, val, err = fl.GetValue("self.Pad1", &ds.Layout[0])
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte{0xff, 0xfe, 0xfd}, val)

	_, val, err = fl.GetValue("self.Pad2", &ds.Layout[0])
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte{0xaa}, val)
}

func TestDataFieldPresentType(t *testing.T) {

	fl := FileLayout{}

	test := []struct {
		field    value.DataField
		expected string
	}{
		{value.DataField{Kind: "u32", Range: "2"}, "u32[2]"},
		{value.DataField{Kind: "u8", Range: "2"}, "u8[2]"},
		{value.DataField{Kind: "u32"}, "u32"},
	}
	for _, h := range test {
		assert.Equal(t, h.expected, fl.PresentType(&h.field))
	}
}
