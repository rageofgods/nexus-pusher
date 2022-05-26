package server

import "testing"

func Test_isValidNexusRepoName(t *testing.T) {
	type args struct {
		param string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{param: "sdf@sdf"},
			want: false,
		},
		{
			name: "test2",
			args: args{param: "sdf3S._-df"},
			want: true,
		},
		{
			name: "test3",
			args: args{param: "sdf3S._-df$"},
			want: false,
		},
		{
			name: "test4",
			args: args{param: "repo-1"},
			want: true,
		},
		{
			name: "test5",
			args: args{param: "repo_2.test"},
			want: true,
		},
		{
			name: "test6",
			args: args{param: ""},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidNexusRepoName(tt.args.param); got != tt.want {
				t.Errorf("isValidNexusRepoName() = %v, want %v", got, tt.want)
			}
		})
	}
}
