package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luxeave/entropy-course/internal/course/adapter/importsource"
	"github.com/luxeave/entropy-course/internal/course/adapter/postgres"
	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/usecase"
)

var ErrMissingDatabaseURL = errors.New("database url is required")

type CLI struct {
	Course   core.CourseService
	Lesson   core.LessonService
	Quiz     core.QuizService
	Practice core.PracticeService
	Test     core.TestService
	Import   core.ImportService
	Config   Config

	pool *pgxpool.Pool
}

func BuildContainer(ctx context.Context, cfg Config) (*CLI, error) {
	pool, err := newPostgresPool(ctx, cfg.DBURL)
	if err != nil {
		return nil, err
	}

	courseRepo := postgres.NewPostgresCourseRepository(pool)
	lessonRepo := postgres.NewPostgresLessonRepository(pool)
	quizRepo := postgres.NewPostgresQuizRepository(pool)
	practiceRepo := postgres.NewPostgresPracticeRepository(pool)
	testRepo := postgres.NewPostgresTestRepository(pool)
	ids := NewUUIDGenerator()
	clock := NewSystemClock()
	importSource := importsource.NewZipImportSource()

	courseService := usecase.NewCourseService(courseRepo, lessonRepo, quizRepo, ids, clock, practiceRepo, testRepo)
	lessonService := usecase.NewLessonService(courseRepo, lessonRepo, quizRepo, ids, clock, practiceRepo)
	quizService := usecase.NewQuizService(courseRepo, lessonRepo, quizRepo, ids, clock)
	practiceService := usecase.NewPracticeService(courseRepo, lessonRepo, practiceRepo, ids, clock)
	testService := usecase.NewTestService(courseRepo, testRepo, ids, clock)
	importService := usecase.NewImportService(
		importSource,
		clock,
		courseRepo,
		lessonRepo,
		quizRepo,
		practiceRepo,
		testRepo,
		courseService,
		lessonService,
		quizService,
		practiceService,
		testService,
	)

	return &CLI{
		Course:   courseService,
		Lesson:   lessonService,
		Quiz:     quizService,
		Practice: practiceService,
		Test:     testService,
		Import:   importService,
		Config:   cfg,
		pool:     pool,
	}, nil
}

func (cli *CLI) Close() {
	if cli != nil && cli.pool != nil {
		cli.pool.Close()
	}
}

func newPostgresPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	if strings.TrimSpace(dbURL) == "" {
		return nil, fmt.Errorf("connect db: %w", ErrMissingDatabaseURL)
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}

	return pool, nil
}
