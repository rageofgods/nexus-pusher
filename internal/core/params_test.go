package core

import (
	"testing"
	"time"
)

func TestNexusComponentAsset_removeTrailingZeroFromPath(t *testing.T) {
	type fields struct {
		DownloadURL string
		Path        string
		ID          string
		Repository  string
		Format      string
		Checksum    struct {
			AdditionalProp1 struct {
			} `json:"additionalProp1"`
			AdditionalProp2 struct {
			} `json:"additionalProp2"`
			AdditionalProp3 struct {
			} `json:"additionalProp3"`
		}
		ContentType  string
		LastModified time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "test1",
			fields: fields{Path: "@webassemblyjs/utf8/-/utf8-1.8.5.tgz"},
			want:   "@webassemblyjs/utf8/-/utf8-1.8.5.tgz",
		},
		{
			name:   "test2",
			fields: fields{Path: "cache-loader/-/cache-loader-4.1.0.tgz"},
			want:   "cache-loader/-/cache-loader-4.1.0.tgz",
		},
		{
			name:   "test3",
			fields: fields{Path: "Abp.EntityFrameworkCore/6.5.0"},
			want:   "Abp.EntityFrameworkCore/6.5",
		},
		{
			name:   "test4",
			fields: fields{Path: "Abp.EntityFrameworkCore/6.5.0.0.1"},
			want:   "Abp.EntityFrameworkCore/6.5.0.0.1",
		},
		{
			name:   "test4",
			fields: fields{Path: "Abp.EntityFrameworkCore/6.5.0.0.0"},
			want:   "Abp.EntityFrameworkCore/6.5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nca := NexusComponentAsset{
				DownloadURL:  tt.fields.DownloadURL,
				Path:         tt.fields.Path,
				ID:           tt.fields.ID,
				Repository:   tt.fields.Repository,
				Format:       tt.fields.Format,
				Checksum:     tt.fields.Checksum,
				ContentType:  tt.fields.ContentType,
				LastModified: tt.fields.LastModified,
			}
			if got := nca.AssetPathWithoutTrailingZeroes(); got != tt.want {
				t.Errorf("removeTrailingZeroFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
