package client

import (
	"nexus-pusher/internal/core"
	"reflect"
	"testing"
)

func Test_genNexExpCompFromNexComp(t *testing.T) {
	type args struct {
		artifactsSource string
		c               []*core.NexusComponent
	}
	tests := []struct {
		name string
		args args
		want *core.NexusExportComponents
	}{
		{
			name: "test1",
			args: args{artifactsSource: "some_source", c: []*core.NexusComponent{
				{
					ID:         "id1",
					Repository: "repo1",
					Format:     "npm",
					Group:      "group1",
					Name:       "name1",
					Version:    "1.0",
					Assets: []*core.NexusComponentAsset{
						{
							DownloadURL: "https://some.org/file1.tar",
							Path:        "https://some.org/path/file1.tar",
							ID:          "id1",
							Repository:  "repo1",
							Format:      "npm",
							ContentType: "type1",
						},
						{
							DownloadURL: "https://some2.org/file2.tar",
							Path:        "https://some2.org/path2/file2.tar",
							ID:          "id2",
							Repository:  "repo2",
							Format:      "npm",
							ContentType: "type1",
						},
					},
				},
			},
			}, want: &core.NexusExportComponents{
				Items: []*core.NexusExportComponent{
					{
						Name:            "name1",
						Version:         "1.0",
						Repository:      "repo1",
						Format:          "npm",
						Group:           "group1",
						ArtifactsSource: "some_source",
						Assets: []*core.NexusExportComponentAsset{
							{
								Name:        "name1",
								FileName:    "file1.tar",
								Version:     "1.0",
								Path:        "https://some.org/path/file1.tar",
								ContentType: "type1",
							},
							{
								Name:        "name1",
								FileName:    "file2.tar",
								Version:     "1.0",
								Path:        "https://some2.org/path2/file2.tar",
								ContentType: "type1",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genNexExpCompFromNexComp(tt.args.artifactsSource, tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("genNexExpCompFromNexComp() = %v, want %v", got, tt.want)
			}
		})
	}
}
