package codebase_test

import (
	"context"
	"testing"

	"github.com/fabric8-services/fabric8-auth/codebase"
	"github.com/fabric8-services/fabric8-auth/errors"
	"github.com/fabric8-services/fabric8-auth/gormsupport/cleaner"
	"github.com/fabric8-services/fabric8-auth/gormtestsupport"
	"github.com/fabric8-services/fabric8-auth/resource"
	"github.com/fabric8-services/fabric8-auth/space"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestCodebaseToMap(t *testing.T) {
	branch := "task-101"
	repo := "golang-project"
	file := "main.go"
	line := 200
	cb := codebase.Content{
		Branch:     branch,
		Repository: repo,
		FileName:   file,
		LineNumber: line,
	}

	codebaseMap := cb.ToMap()
	require.NotNil(t, codebaseMap)
	assert.Equal(t, repo, codebaseMap[codebase.RepositoryKey])
	assert.Equal(t, branch, codebaseMap[codebase.BranchKey])
	assert.Equal(t, file, codebaseMap[codebase.FileNameKey])
	assert.Equal(t, line, codebaseMap[codebase.LineNumberKey])
}

func TestNewCodebase(t *testing.T) {
	// Test for empty map
	codebaseMap := map[string]interface{}{}
	cb, err := codebase.NewCodebaseContent(codebaseMap)
	require.NotNil(t, err)
	assert.Equal(t, "", cb.Repository)
	assert.Equal(t, "", cb.Branch)
	assert.Equal(t, "", cb.FileName)
	assert.Equal(t, 0, cb.LineNumber)

	// test for all values in codebase
	branch := "task-101"
	repo := "golang-project"
	file := "main.go"
	line := 200
	codebaseMap = map[string]interface{}{
		codebase.RepositoryKey: repo,
		codebase.BranchKey:     branch,
		codebase.FileNameKey:   file,
		codebase.LineNumberKey: line,
	}
	cb, err = codebase.NewCodebaseContent(codebaseMap)
	require.Nil(t, err)
	assert.Equal(t, repo, cb.Repository)
	assert.Equal(t, branch, cb.Branch)
	assert.Equal(t, file, cb.FileName)
	assert.Equal(t, line, cb.LineNumber)
}

func TestIsValid(t *testing.T) {
	cb := codebase.Content{
		Repository: "hello",
	}
	assert.Nil(t, cb.IsValid())

	cb = codebase.Content{}
	assert.NotNil(t, cb.IsValid())
}

type TestCodebaseRepository struct {
	gormtestsupport.DBTestSuite

	clean func()
}

func TestRunCodebaseRepository(t *testing.T) {
	resource.Require(t, resource.Database)
	suite.Run(t, &TestCodebaseRepository{DBTestSuite: gormtestsupport.NewDBTestSuite("../config.yaml")})
}

func (test *TestCodebaseRepository) SetupTest() {
	test.clean = cleaner.DeleteCreatedEntities(test.DB)
}

func (test *TestCodebaseRepository) TearDownTest() {
	test.clean()
}

func newCodebase(spaceID uuid.UUID, stackID, lastUsedWorkspace, repotype, url string) *codebase.Codebase {
	return &codebase.Codebase{
		SpaceID:           spaceID,
		Type:              repotype,
		URL:               url,
		StackID:           &stackID,
		LastUsedWorkspace: lastUsedWorkspace,
	}
}

func (test *TestCodebaseRepository) createCodebase(c *codebase.Codebase) {
	repo := codebase.NewCodebaseRepository(test.DB)
	err := repo.Create(context.Background(), c)
	require.Nil(test.T(), err)
}

func (test *TestCodebaseRepository) TestListCodebases() {
	// given
	spaceID := space.SystemSpace
	repo := codebase.NewCodebaseRepository(test.DB)
	codebase1 := newCodebase(spaceID, "golang-default", "my-used-last-workspace", "git", "git@github.com:fabric8-services/fabric8-auth.git")
	codebase2 := newCodebase(spaceID, "python-default", "my-used-last-workspace", "git", "git@github.com:aslakknutsen/fabric8-wit.git")

	test.createCodebase(codebase1)
	test.createCodebase(codebase2)
	// when
	offset := 0
	limit := 1
	codebases, _, err := repo.List(context.Background(), spaceID, &offset, &limit)
	// then
	require.Nil(test.T(), err)
	require.Equal(test.T(), 1, len(codebases))
	assert.Equal(test.T(), codebase1.URL, codebases[0].URL)
}

func (test *TestCodebaseRepository) TestExistsCodebase() {
	t := test.T()
	resource.Require(t, resource.Database)

	t.Run("codebase exists", func(t *testing.T) {
		// given
		spaceID := space.SystemSpace
		repo := codebase.NewCodebaseRepository(test.DB)
		codebase := newCodebase(spaceID, "lisp-default", "my-used-lisp-workspace", "git", "git@github.com:hectorj2f/fabric8-wit.git")
		test.createCodebase(codebase)
		// when
		exists, err := repo.Exists(context.Background(), codebase.ID.String())
		// then
		require.Nil(t, err)
		assert.True(t, exists)
	})

	t.Run("codebase doesn't exist", func(t *testing.T) {
		// given
		repo := codebase.NewCodebaseRepository(test.DB)
		// when
		exists, err := repo.Exists(context.Background(), uuid.NewV4().String())
		// then
		require.IsType(t, errors.NotFoundError{}, err)
		assert.False(t, exists)
	})

}

func (test *TestCodebaseRepository) TestLoadCodebase() {
	// given
	spaceID := space.SystemSpace
	repo := codebase.NewCodebaseRepository(test.DB)
	codebase := newCodebase(spaceID, "golang-default", "my-used-last-workspace", "git", "git@github.com:aslakknutsen/fabric8-wit.git")
	test.createCodebase(codebase)
	// when
	loadedCodebase, err := repo.Load(context.Background(), codebase.ID)
	require.Nil(test.T(), err)
	assert.Equal(test.T(), codebase.ID, loadedCodebase.ID)
	assert.Equal(test.T(), "golang-default", *loadedCodebase.StackID)
	assert.Equal(test.T(), "my-used-last-workspace", loadedCodebase.LastUsedWorkspace)
}
