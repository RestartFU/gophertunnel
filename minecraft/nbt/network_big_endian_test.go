package nbt

import (
	"bytes"
	"reflect"
	"testing"
)

// TestNetworkBigEndian_UnnamedRoot checks that NetworkBigEndian omits the root name for every
// root tag type, as Java Edition does for network NBT since 1.20.3, where a text component
// may be sent as a bare string.
func TestNetworkBigEndian_UnnamedRoot(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		value any
		// anyValue is the value decoded into an any, when it differs from value.
		anyValue any
		want     []byte
	}{
		{name: "string", value: "hi", want: []byte{byte(tagString), 0x00, 0x02, 'h', 'i'}},
		{name: "int", value: int32(7), want: []byte{byte(tagInt32), 0x00, 0x00, 0x00, 0x07}},
		{name: "list", value: []string{"a"}, anyValue: []any{"a"}, want: []byte{byte(tagSlice), byte(tagString), 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 'a'}},
		{name: "compound", value: map[string]any{"text": "a"}, want: []byte{
			byte(tagStruct), byte(tagString), 0x00, 0x04, 't', 'e', 'x', 't', 0x00, 0x01, 'a', byte(tagEnd),
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := MarshalEncoding(test.value, NetworkBigEndian)
			if err != nil {
				t.Fatalf("MarshalEncoding() error = %v", err)
			}
			if !bytes.Equal(got, test.want) {
				t.Fatalf("MarshalEncoding() = %x, want %x", got, test.want)
			}
			decoded := reflect.New(reflect.TypeOf(test.value))
			if err := UnmarshalEncoding(test.want, decoded.Interface(), NetworkBigEndian); err != nil {
				t.Fatalf("UnmarshalEncoding() error = %v", err)
			}
			if !reflect.DeepEqual(decoded.Elem().Interface(), test.value) {
				t.Fatalf("UnmarshalEncoding() = %#v, want %#v", decoded.Elem().Interface(), test.value)
			}
			var anyValue any
			if err := UnmarshalEncoding(test.want, &anyValue, NetworkBigEndian); err != nil {
				t.Fatalf("UnmarshalEncoding(any) error = %v", err)
			}
			wantAny := test.anyValue
			if wantAny == nil {
				wantAny = test.value
			}
			if !reflect.DeepEqual(anyValue, wantAny) {
				t.Fatalf("UnmarshalEncoding(any) = %#v, want %#v", anyValue, wantAny)
			}
			if _, err := ReadRaw(bytes.NewReader(test.want), NetworkBigEndian, false); err != nil {
				t.Fatalf("ReadRaw() error = %v", err)
			}
		})
	}
}
