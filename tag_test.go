package configutil

import (
	"reflect"
	"testing"
)

func TestParseTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		want    tagMetadata
		wantErr bool
	}{
		{name: "just name", tag: "PORT", want: tagMetadata{Name: "PORT"}},
		{name: "name and required", tag: "PORT,required", want: tagMetadata{Name: "PORT", Required: true}},
		{name: "name and default", tag: "PORT,default=8080", want: tagMetadata{Name: "PORT", Default: "8080"}},
		{name: "name and prefix", tag: "SERVER,prefix=API_", want: tagMetadata{Name: "SERVER", Prefix: "API_"}},
		{name: "all options", tag: "PORT,required,default=8080,prefix=API_", want: tagMetadata{Name: "PORT", Required: true, Default: "8080", Prefix: "API_"}},
		{name: "whitespace", tag: "  PORT  ,  required  ,  default = 8080  ,  prefix = API_  ", want: tagMetadata{Name: "PORT", Required: true, Default: "8080", Prefix: "API_"}},
		{name: "empty tag", tag: "", want: tagMetadata{}},
		{name: "empty part skipped", tag: "PORT,,required", want: tagMetadata{Name: "PORT", Required: true}},
		{name: "unknown option errors", tag: "PORT,unknown=foo", wantErr: true},
		{name: "malformed key-value", tag: "PORT,=8080", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseTag(tc.tag)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseTag() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parseTag()\ngot  %+v\nwant %+v", got, tc.want)
			}
		})
	}
}
