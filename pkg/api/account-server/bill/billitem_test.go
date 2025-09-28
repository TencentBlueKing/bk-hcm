package bill

import (
	"testing"
)

func TestBase64String_checkSize(t *testing.T) {
	type args struct {
		expectedSize int
	}
	tests := []struct {
		name    string
		b       Base64String
		args    args
		wantErr bool
	}{
		{
			name: "Hello, World!",
			b:    Base64String("SGVsbG8sIHdvcmxkIQ=="),
			args: args{
				expectedSize: 1,
			},
			wantErr: true,
		},
		{
			name: "Hello, World! 2",
			b:    Base64String("SGVsbG8sIHdvcmxkIQ=="),
			args: args{
				expectedSize: 1024, // 1KB
			},
			wantErr: false,
		},
		{
			name: "Hello, World! 3",
			b:    Base64String("SGVsbG8sIHdvcmxkIQ=="),
			args: args{
				expectedSize: 20,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.b.checkSize(tt.args.expectedSize); (err != nil) != tt.wantErr {
				t.Errorf("checkSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
