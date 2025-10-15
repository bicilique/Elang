package dto

import (
	"elang-backend/internal/repository"
)

// BasicRepositories groups all repository interfaces needed for basic operations
type BasicRepositories struct {
	AppRepository              repository.ApplicationRepository
	DepedencyRepository        repository.DependencyRepository
	AppToDepedencyRepository   repository.AppDependencyRepository
	DepedencyVersionRepository repository.DependencyVersionRepository
	RunTimeRepository          repository.RuntimeRepository
	FrameWorkRepository        repository.FrameworkRepository
	AuditTrailRepository       repository.AuditTrailRepository
}

// BasicServices groups all service interfaces needed for basic operations
// type CommonBasicServices struct {
// 	DepedencyParserService helper.DependencyParser
// 	GithubApiService       services.GitHubAPIInterface
// 	ObjectStorageService   services.ObjectStorageInterface
// }
