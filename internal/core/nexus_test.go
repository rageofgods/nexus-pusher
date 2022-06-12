package core

import "testing"

func Test_filterAssets(t *testing.T) {
	type args struct {
		nc *NexusComponents
	}
	tests := []struct {
		name            string
		args            args
		wantVersion     []string
		wantPath        []string
		wantAssetsCount int
	}{
		{
			name: "Test1_Nuget",
			args: args{
				nc: &NexusComponents{
					Items: []*NexusComponent{
						{
							Version: "5.2.1-develop.1832",
							Format:  "nuget",
							Assets: []*NexusComponentAsset{
								{
									Path:   "MassTransit/5.2.1-develop.1832",
									Format: "nuget",
								},
								{
									Path:   "MassTransit/5.2.1-develop.1833",
									Format: "nuget",
								},
							},
						},
					},
				},
			},
			wantVersion:     []string{"5.2.1-develop.1832"},
			wantPath:        []string{"MassTransit/5.2.1-develop.1832", "MassTransit/5.2.1-develop.1833"},
			wantAssetsCount: 2,
		},
		{
			name: "Test2_Nuget",
			args: args{
				nc: &NexusComponents{
					Items: []*NexusComponent{
						{
							Version: "5.2.1-develop.1832+sha.19b3cdc",
							Format:  "nuget",
							Assets: []*NexusComponentAsset{
								{
									Path:   "MassTransit/5.2.1-develop.1832+sha.19b3cdc",
									Format: "nuget",
								},
								{
									Path:   "MassTransit/5.2.1-develop.1833+sha.19b3cdc",
									Format: "nuget",
								},
							},
						},
					},
				},
			},
			wantVersion:     []string{"5.2.1-develop.1832"},
			wantPath:        []string{"MassTransit/5.2.1-develop.1832", "MassTransit/5.2.1-develop.1833"},
			wantAssetsCount: 2,
		},
		{
			name: "Test3_Maven2",
			args: args{
				nc: &NexusComponents{
					Items: []*NexusComponent{
						{
							Version: "5.2.1-develop.1832",
							Format:  "maven2",
							Assets: []*NexusComponentAsset{
								{
									Path:   "MassTransit/5.2.1-develop.1833.sha256",
									Format: "maven2",
								},
								{
									Path:   "MassTransit/5.2.1-develop.1832.zip",
									Format: "maven2",
								},
								{
									Path:   "MassTransit/5.2.1-develop.1833.sha1",
									Format: "maven2",
								},
								{
									Path:   "MassTransit/5.2.1-develop.1834.txt",
									Format: "maven2",
								},
								{
									Path:   "MassTransit/5.2.1-develop.1833.md5",
									Format: "maven2",
								},
								{
									Path:   "MassTransit/5.2.1-develop.1834.bin",
									Format: "maven2",
								},
								{
									Path:   "MassTransit/5.2.1-develop.1833.sha512",
									Format: "maven2",
								},
							},
						},
					},
				},
			},
			wantAssetsCount: 3,
			wantPath: []string{"MassTransit/5.2.1-develop.1832.zip",
				"MassTransit/5.2.1-develop.1834.txt",
				"MassTransit/5.2.1-develop.1834.bin",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterHashAssets(tt.args.nc)
			for i, v := range tt.args.nc.Items {
				switch v.Format {
				case "nuget":
					if v.Version != tt.wantVersion[i] {
						t.Errorf("filterHashAssets() = %v, want version %v", v.Version, tt.wantVersion)
					}
					for ii, vv := range v.Assets {
						if vv.Path != tt.wantPath[ii] {
							t.Errorf("filterHashAssets() = %v, want path %v", vv.Path, tt.wantPath)
						}
					}
				case "maven2":
					if len(v.Assets) != tt.wantAssetsCount {
						t.Errorf("filterHashAssets() = len(%v), want path len(%v)", len(v.Assets), len(tt.wantPath))
					}
					for ii, vv := range v.Assets {
						if vv.Path != tt.wantPath[ii] {
							t.Errorf("filterHashAssets() = %v, want path %v", vv.Path, tt.wantPath)
						}
					}
				}
			}
		})
	}
}
