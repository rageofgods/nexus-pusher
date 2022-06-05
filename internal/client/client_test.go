package client

import (
	"nexus-pusher/internal/comps"
	"reflect"
	"testing"
)

func Test_fileNameFromPath(t *testing.T) {
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
			args: args{path: "https://somedomain.org/url/path/file.jar"},
			want: "file.jar",
		},
		{
			name: "test2",
			args: args{path: "https://somedomain.org/url/path/@file.jar"},
			want: "file.jar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fileNameFromPath(tt.args.path); got != tt.want {
				t.Errorf("fileNameFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_compareComponents(t *testing.T) {
	type args struct {
		src []*comps.NexusComponent
		dst []*comps.NexusComponent
	}
	tests := []struct {
		name string
		args args
		want []*comps.NexusComponent
	}{
		{
			name: "test1",
			args: args{src: []*comps.NexusComponent{
				{
					ID:         "id1",
					Repository: "repo1",
					Format:     "npm",
					Group:      "group1",
					Name:       "name1",
					Version:    "1.0",
					Assets: []*comps.NexusComponentAsset{
						{
							DownloadURL: "https://some.org/file1.tar",
							Path:        "https://some.org/path/file1.tar",
							ID:          "id1",
							Repository:  "repo1",
							Format:      "npm",
						},
						{
							DownloadURL: "https://some2.org/file2.tar",
							Path:        "https://some2.org/path2/file2.tar",
							ID:          "id2",
							Repository:  "repo2",
							Format:      "npm",
						},
					},
				},
				{
					ID:         "id2",
					Repository: "repo2",
					Format:     "npm",
					Group:      "group2",
					Name:       "name2",
					Version:    "1.0",
					Assets: []*comps.NexusComponentAsset{
						{
							DownloadURL: "https://some.org/file3.tar",
							Path:        "https://some.org/path/file3.tar",
							ID:          "id1",
							Repository:  "repo1",
							Format:      "npm",
						},
						{
							DownloadURL: "https://some2.org/file4.tar",
							Path:        "https://some2.org/path2/file4.tar",
							ID:          "id2",
							Repository:  "repo2",
							Format:      "npm",
						},
					},
				},
				{
					ID:         "id3",
					Repository: "repo3",
					Format:     "npm",
					Group:      "group3",
					Name:       "name3",
					Version:    "1.0",
					Assets: []*comps.NexusComponentAsset{
						{
							DownloadURL: "https://some.org/file5.tar",
							Path:        "https://some.org/path/file5.tar",
							ID:          "id1",
							Repository:  "repo1",
							Format:      "npm",
						},
						{
							DownloadURL: "https://some2.org/file6.tar",
							Path:        "https://some2.org/path2/file6.tar",
							ID:          "id2",
							Repository:  "repo2",
							Format:      "npm",
						},
					},
				},
			},
				dst: []*comps.NexusComponent{
					{
						ID:         "id1",
						Repository: "repo1",
						Format:     "npm",
						Group:      "group1",
						Name:       "name1",
						Version:    "1.0",
						Assets: []*comps.NexusComponentAsset{
							{
								DownloadURL: "https://some.org/file1.tar",
								Path:        "https://some.org/path/file1.tar",
								ID:          "id1",
								Repository:  "repo1",
								Format:      "npm",
							},
						},
					},
					{
						ID:         "id2",
						Repository: "repo2",
						Format:     "npm",
						Group:      "group2",
						Name:       "name2",
						Version:    "1.0",
						Assets: []*comps.NexusComponentAsset{
							{
								DownloadURL: "https://some.org/file3.tar",
								Path:        "https://some.org/path/file3.tar",
								ID:          "id1",
								Repository:  "repo1",
								Format:      "npm",
							},
						},
					},
				},
			},
			want: []*comps.NexusComponent{
				{
					ID:         "id1",
					Repository: "repo1",
					Format:     "npm",
					Group:      "group1",
					Name:       "name1",
					Version:    "1.0",
					Assets: []*comps.NexusComponentAsset{
						{
							DownloadURL: "https://some2.org/file2.tar",
							Path:        "https://some2.org/path2/file2.tar",
							ID:          "id2",
							Repository:  "repo2",
							Format:      "npm",
						},
					},
				},
				{
					ID:         "id2",
					Repository: "repo2",
					Format:     "npm",
					Group:      "group2",
					Name:       "name2",
					Version:    "1.0",
					Assets: []*comps.NexusComponentAsset{
						{
							DownloadURL: "https://some2.org/file4.tar",
							Path:        "https://some2.org/path2/file4.tar",
							ID:          "id2",
							Repository:  "repo2",
							Format:      "npm",
						},
					},
				},
				{
					ID:         "id3",
					Repository: "repo3",
					Format:     "npm",
					Group:      "group3",
					Name:       "name3",
					Version:    "1.0",
					Assets: []*comps.NexusComponentAsset{
						{
							DownloadURL: "https://some.org/file5.tar",
							Path:        "https://some.org/path/file5.tar",
							ID:          "id1",
							Repository:  "repo1",
							Format:      "npm",
						},
						{
							DownloadURL: "https://some2.org/file6.tar",
							Path:        "https://some2.org/path2/file6.tar",
							ID:          "id2",
							Repository:  "repo2",
							Format:      "npm",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareComponents(tt.args.src, tt.args.dst); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compareComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}
