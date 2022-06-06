package core

import "testing"

func Test_removeLastSlash(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{path: "https://some-server.org/api/"},
			want: "https://some-server.org/api",
		},
		{
			name: "test2",
			args: args{path: "https://some-server.org/api"},
			want: "https://some-server.org/api",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotP := removeLastSlash(tt.args.path); gotP != tt.want {
				t.Errorf("removeLastSlash() = %v, want %v", gotP, tt.want)
			}
		})
	}
}
