package codebase_test

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/fabric8-services/fabric8-auth/codebase"
	"github.com/fabric8-services/fabric8-auth/gormsupport/cleaner"
	gormbench "github.com/fabric8-services/fabric8-auth/gormtestsupport/benchmark"
	"github.com/fabric8-services/fabric8-auth/migration"
	"github.com/fabric8-services/fabric8-auth/space"
	testsupport "github.com/fabric8-services/fabric8-auth/test"
)

type BenchCodebaseRepository struct {
	gormbench.DBBenchSuite
	clean func()
	repo  codebase.Repository
	ctx   context.Context
}

func BenchmarkRunCodebaseRepository(b *testing.B) {
	testsupport.Run(b, &BenchCodebaseRepository{DBBenchSuite: gormbench.NewDBBenchSuite("../config.yaml")})
}

// SetupSuite overrides the DBTestSuite's function but calls it before doing anything else
// The SetupSuite method will run before the tests in the suite are run.
// It sets up a database connection for all the tests in this suite without polluting global space.
func (s *BenchCodebaseRepository) SetupSuite() {
	s.DBBenchSuite.SetupSuite()
	s.ctx = migration.NewMigrationContext(context.Background())
	s.DBBenchSuite.PopulateDBBenchSuite(s.ctx)
}

func (s *BenchCodebaseRepository) SetupBenchmark() {
	s.clean = cleaner.DeleteCreatedEntities(s.DB)
	s.repo = codebase.NewCodebaseRepository(s.DB)
}

func (s *BenchCodebaseRepository) TearDownBenchmark() {
	s.clean()
}

func (s *BenchCodebaseRepository) createCodebase(c *codebase.Codebase) {
	err := s.repo.Create(s.ctx, c)
	if err != nil {
		s.B().Fail()
	}
}

func (s *BenchCodebaseRepository) BenchmarkCreateCodebases() {
	s.B().ResetTimer()
	s.B().ReportAllocs()
	for n := 0; n < s.B().N; n++ {
		codebase2 := newCodebase(space.SystemSpace, "python-default", "my-used-last-workspace", "git", "git@github.com:aslakknutsen/almighty-core.git")
		s.createCodebase(codebase2)
	}
}

func (s *BenchCodebaseRepository) BenchmarkListCodebases() {
	// given
	codebase := newCodebase(space.SystemSpace, "java-default", "my-used-last-workspace", "git", "git@github.com:aslakknutsen/almighty-core.git")
	s.createCodebase(codebase)
	// when
	offset := 0
	limit := 1
	s.B().ResetTimer()
	s.B().ReportAllocs()
	for n := 0; n < s.B().N; n++ {
		if codebases, _, err := s.repo.List(s.ctx, space.SystemSpace, &offset, &limit); err != nil || (err == nil && len(codebases) == 0) {
			s.B().Fail()
		}
	}
}

func (s *BenchCodebaseRepository) BenchmarkLoadCodebase() {
	// given
	codebaseTest := newCodebase(space.SystemSpace, "golang-default", "my-used-hector-workspace", "git", "git@github.com:hectorj2f/almighty-core.git")
	s.createCodebase(codebaseTest)
	// when
	s.B().ResetTimer()
	s.B().ReportAllocs()
	for n := 0; n < s.B().N; n++ {
		if loadedCodebase, err := s.repo.Load(s.ctx, codebaseTest.ID); err != nil || (err == nil && loadedCodebase == nil) {
			s.B().Fail()
		}
	}
}
