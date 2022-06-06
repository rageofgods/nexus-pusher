package comps

import "testing"

func TestAssetFileNameFromURI(t *testing.T) {
	type args struct {
		assetPath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{assetPath: "/some/asset/path/file.tar.gz"},
			want: "file.tar.gz",
		},
		{
			name: "test2",
			args: args{assetPath: "/some/Asset/path/File2.tar.gz"},
			want: "File2.tar.gz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AssetFileNameFromURI(tt.args.assetPath); got != tt.want {
				t.Errorf("AssetFileNameFromURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileExtensionFromFile(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{fileName: "someFile.zip"},
			want: "zip",
		},
		{
			name: "test2",
			args: args{fileName: "someFile-test.bz2"},
			want: "bz2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileExtensionFromFile(tt.args.fileName); got != tt.want {
				t.Errorf("FileExtensionFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pathSplit(t *testing.T) {
	type args struct {
		assetPath string
		index     int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{assetPath: "/some/asset/path/file.tar.gz", index: 1},
			want: "file.tar.gz",
		},
		{
			name: "test2",
			args: args{assetPath: "/some/asset/path/file.tar.gz", index: 2},
			want: "path",
		},
		{
			name: "test3",
			args: args{assetPath: "/some/asset/path/file.tar.gz", index: 3},
			want: "asset",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathSplit(tt.args.assetPath, tt.args.index); got != tt.want {
				t.Errorf("pathSplit() = %v, want %v", got, tt.want)
			}
		})
	}
}
