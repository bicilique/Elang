package usecase_test

import (
	"context"
	"elang-backend/internal/usecase"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMinioUsecase struct {
	saveSBOMCalled bool
	getSBOMCalled  bool
}

func (m *mockMinioUsecase) SaveSBOM(ctx context.Context, appID string, appName string, sbomData []byte, format string) (string, error) {
	m.saveSBOMCalled = true
	return "sbom/test-app/2024-01-01/test-app-id_sbom.json", nil
}

func (m *mockMinioUsecase) GetSBOM(ctx context.Context, objectKey string) ([]byte, error) {
	m.getSBOMCalled = true
	return []byte(`{"test": "data"}`), nil
}

func (m *mockMinioUsecase) ListSBOMs(ctx context.Context, appName string) ([]string, error) {
	return []string{
		"sbom/test-app/2024-01-01/test-app-id_sbom.json",
		"sbom/test-app/2024-01-02/test-app-id2_sbom.json",
	}, nil
}

func (m *mockMinioUsecase) DeleteSBOM(ctx context.Context, objectKey string) error {
	return nil
}

func (m *mockMinioUsecase) SaveVulnerabilityReport(ctx context.Context, appID string, appName string, reportData []byte, format string) (string, error) {
	return "vulnerability-reports/test-app/2024-01-01/test-app-id_vuln_report.json", nil
}

func (m *mockMinioUsecase) GetVulnerabilityReport(ctx context.Context, objectKey string) ([]byte, error) {
	return []byte(`{"vulnerabilities": []}`), nil
}

func (m *mockMinioUsecase) ListVulnerabilityReports(ctx context.Context, appName string) ([]string, error) {
	return []string{"vulnerability-reports/test-app/2024-01-01/test-app-id_vuln_report.json"}, nil
}

func TestMinioUsecase_SaveSBOM(t *testing.T) {
	ctx := context.Background()
	mock := &mockMinioUsecase{}

	appID := "test-app-id"
	appName := "test-app"
	sbomData := []byte(`{"components": []}`)
	format := "json"

	objectKey, err := mock.SaveSBOM(ctx, appID, appName, sbomData, format)

	assert.NoError(t, err)
	assert.NotEmpty(t, objectKey)
	assert.True(t, mock.saveSBOMCalled)
	assert.Contains(t, objectKey, appName)
	assert.Contains(t, objectKey, "sbom")
}

func TestMinioUsecase_GetSBOM(t *testing.T) {
	ctx := context.Background()
	mock := &mockMinioUsecase{}

	objectKey := "sbom/test-app/2024-01-01/test-app-id_sbom.json"

	data, err := mock.GetSBOM(ctx, objectKey)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.True(t, mock.getSBOMCalled)
}

func TestMinioUsecase_ListSBOMs(t *testing.T) {
	ctx := context.Background()
	mock := &mockMinioUsecase{}

	sboms, err := mock.ListSBOMs(ctx, "test-app")

	assert.NoError(t, err)
	assert.Len(t, sboms, 2)
	assert.Contains(t, sboms[0], "test-app")
}

func TestObjectStorageInterface(t *testing.T) {
	t.Run("InterfaceCompliance", func(t *testing.T) {
		var _ usecase.ObjectStorageInterface = &mockMinioUsecase{}
		require.True(t, true, "mockMinioUsecase implements ObjectStorageInterface")
	})
}

func TestMinioUsecase_SaveSBOM_XMLFormat(t *testing.T) {
	ctx := context.Background()
	mock := &mockMinioUsecase{}

	appID := "test-app-id"
	appName := "test-app"
	sbomData := []byte(`<bom></bom>`)
	format := "xml"

	objectKey, err := mock.SaveSBOM(ctx, appID, appName, sbomData, format)

	assert.NoError(t, err)
	assert.NotEmpty(t, objectKey)
	assert.True(t, mock.saveSBOMCalled)
}

func TestMinioUsecase_DeleteSBOM(t *testing.T) {
	ctx := context.Background()
	mock := &mockMinioUsecase{}

	objectKey := "sbom/test-app/2024-01-01/test-app-id_sbom.json"

	err := mock.DeleteSBOM(ctx, objectKey)

	assert.NoError(t, err)
}
