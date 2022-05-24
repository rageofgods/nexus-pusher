package comps

import "testing"

func TestNuget_assetDownloadURL(t *testing.T) {
	type fields struct {
		Server   string
		FileName string
		Name     string
		Version  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test1",
			fields: fields{
				Server:  "stub/",
				Name:    "Microsoft.NET.Sdk.iOS.Manifest",
				Version: "6.0.200-15.2.301-preview.13.7+sha.53134097a",
			},
			want: "stub/microsoft.net.sdk.ios.manifest/6.0.200-15.2.301-preview.13.7/" +
				"microsoft.net.sdk.ios.manifest.6.0.200-15.2.301-preview.13.7.nupkg",
		},
		{
			name: "test2",
			fields: fields{
				Server:  "stub/",
				Name:    "Microsoft.NET.Sdk.iOS.Manifest",
				Version: "6.0.200-15.2.301-preview.13.7",
			},
			want: "stub/microsoft.net.sdk.ios.manifest/6.0.200-15.2.301-preview.13.7/" +
				"microsoft.net.sdk.ios.manifest.6.0.200-15.2.301-preview.13.7.nupkg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Nuget{
				Server:   tt.fields.Server,
				FileName: tt.fields.FileName,
				Name:     tt.fields.Name,
				Version:  tt.fields.Version,
			}
			if got := n.assetDownloadURL(); got != tt.want {
				t.Errorf("assetDownloadURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
