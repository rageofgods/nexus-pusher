package core

import "testing"

func TestMaven2_assetClassifier(t *testing.T) {
	componentAsset1 := &NexusExportComponent{
		Name:    "quarkus-bom-quarkus-platform-descriptor",
		Version: "1.12.2.Final",
	}
	componentAsset2 := &NexusExportComponent{
		Name:    "webdrivermanager",
		Version: "3.6.2",
	}

	const fileName1 = "quarkus-bom-quarkus-platform-descriptor-1.12.2.Final.pom"
	const fileName2 = "quarkus-bom-quarkus-platform-descriptor-1.12.2.Final-1.12.2.Final.json"
	const fileName3 = "webdrivermanager-3.6.2-sources.jar"

	type fields struct {
		Server    string
		Component *NexusExportComponent
	}
	type args struct {
		fileName      string
		fileExtension string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{name: "test1", fields: fields{Server: "stub", Component: componentAsset1},
			args: args{fileName: fileName1, fileExtension: "pom"}, want: ""},
		{name: "test2", fields: fields{Server: "stub", Component: componentAsset1},
			args: args{fileName: fileName2, fileExtension: "json"}, want: "1.12.2.Final"},
		{name: "test3", fields: fields{Server: "stub", Component: componentAsset2},
			args: args{fileName: fileName3, fileExtension: "jar"}, want: "sources"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Maven2{
				Server:    tt.fields.Server,
				Component: tt.fields.Component,
			}
			if got := m.assetClassifier(tt.args.fileName, tt.args.fileExtension); got != tt.want {
				t.Errorf("assetClassifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaven2_pomInComponent(t *testing.T) {
	componentAsset1 := &NexusExportComponent{Assets: []*NexusExportComponentAsset{
		{FileName: "webdrivermanager-3.6.2-sources.jar"},
		{FileName: "webdrivermanager-3.6.2-sources.jar.pom"},
		{FileName: "webdrivermanager-3.6.2-sources.jar.sha1"}},
	}
	componentAsset2 := &NexusExportComponent{Assets: []*NexusExportComponentAsset{
		{FileName: "webdrivermanager-3.6.2-sources.jar"},
		{FileName: "webdrivermanager-3.6.2-sources.jar.pom.sha1"},
		{FileName: "webdrivermanager-3.6.2-sources.jar.sha1"}},
	}

	type fields struct {
		Server    string
		Component *NexusExportComponent
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "test1", fields: fields{Server: "stub", Component: componentAsset1}, want: true},
		{name: "test2", fields: fields{Server: "stub", Component: componentAsset2}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Maven2{
				Server:    tt.fields.Server,
				Component: tt.fields.Component,
			}
			if got := m.pomInComponent(); got != tt.want {
				t.Errorf("pomInComponent() = %v, want %v", got, tt.want)
			}
		})
	}
}
