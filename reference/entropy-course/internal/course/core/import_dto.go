package core

import "github.com/luxeave/entropy-course/internal/course/domain"

type PlanImportInput struct {
	ZipPath      string
	InstructorID string
}

type PlanImportOutput struct {
	Plan domain.ImportPlan
}

type ApplyPlanInput struct {
	ZipPath          string
	InstructorID     string
	ResolvedPlanJSON []byte
	ConflictStrategy string
}

type ApplyPlanOutput struct {
	Result domain.ApplyResult
}
