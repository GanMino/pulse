package service

import (
	"github.com/pulse/pulse/internal/repository/sqlite"
)

// Container 是服务的依赖注入容器
// 持有所有 Repository 和 Service 实例,作为单一真相源
type Container struct {
	// Repositories
	UserRepo     *sqlite.UserRepository
	ProjectRepo  *sqlite.ProjectRepository
	ScenarioRepo *sqlite.ScenarioRepository
	TestRunRepo  *sqlite.TestRunRepository
	ReportRepo   *sqlite.ReportRepository
	AgentRepo    *sqlite.AgentRepository

	// Services
	UserService     *UserService
	ProjectService  *ProjectService
	ScenarioService *ScenarioService
	TestRunService  *TestRunService
	RunService      *RunService
	ReportService   *ReportService
	AgentService    *AgentService
	ImportService   *ImportService
}

// NewContainer 从 DB 创建 Container
func NewContainer(db *sqlite.DB) *Container {
	c := &Container{
		UserRepo:     sqlite.NewUserRepository(db),
		ProjectRepo:  sqlite.NewProjectRepository(db),
		ScenarioRepo: sqlite.NewScenarioRepository(db),
		TestRunRepo:  sqlite.NewTestRunRepository(db),
		ReportRepo:   sqlite.NewReportRepository(db),
		AgentRepo:    sqlite.NewAgentRepository(db),
	}

	c.UserService = NewUserService(c.UserRepo, c.ProjectRepo)
	c.ProjectService = NewProjectService(c.ProjectRepo, c.ScenarioRepo, c.UserRepo)
	c.ScenarioService = NewScenarioService(c.ScenarioRepo, c.ProjectRepo, c.UserRepo)
	c.TestRunService = NewTestRunService(c.TestRunRepo, c.ScenarioRepo, c.ProjectRepo, c.UserRepo)
	c.RunService = NewRunService(c.ScenarioRepo, c.ProjectRepo, c.TestRunRepo, c.ReportRepo, c.UserRepo, c.ScenarioService)
	c.ReportService = NewReportService(c.ReportRepo, c.TestRunRepo)
	c.AgentService = NewAgentService(c.AgentRepo)
	c.ImportService = NewImportService(c.ScenarioService, c.ProjectService)

	return c
}