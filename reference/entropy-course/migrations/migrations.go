package migrations

import (
	"embed"
	"fmt"
)

type Migration struct {
	Version string
	UpSQL   string
	DownSQL string
}

const (
	courseSchemaVersion   = "000001_create_course_schema"
	contentBlocksVersion  = "000002_add_content_blocks"
	quizSchemaVersion     = "000003_add_quizzes"
	practiceSchemaVersion = "000004_add_practices"
	testSchemaVersion     = "000005_add_tests"
)

//go:embed *.sql
var files embed.FS

func Registered() ([]Migration, error) {
	versions := []string{
		courseSchemaVersion,
		contentBlocksVersion,
		quizSchemaVersion,
		practiceSchemaVersion,
		testSchemaVersion,
	}

	registered := make([]Migration, 0, len(versions))
	for _, version := range versions {
		migration, err := migration(version)
		if err != nil {
			return nil, err
		}
		registered = append(registered, migration)
	}

	return registered, nil
}

func migration(version string) (Migration, error) {
	upSQL, err := readSQL(version + ".up.sql")
	if err != nil {
		return Migration{}, err
	}
	downSQL, err := readSQL(version + ".down.sql")
	if err != nil {
		return Migration{}, err
	}

	return Migration{
		Version: version,
		UpSQL:   upSQL,
		DownSQL: downSQL,
	}, nil
}

func readSQL(name string) (string, error) {
	contents, err := files.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("read migration %s: %w", name, err)
	}

	return string(contents), nil
}
