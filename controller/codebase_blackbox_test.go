package controller_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/fabric8-services/fabric8-auth/app/test"
	"github.com/fabric8-services/fabric8-auth/application"
	"github.com/fabric8-services/fabric8-auth/codebase"

	"github.com/fabric8-services/fabric8-auth/account"
	"github.com/fabric8-services/fabric8-auth/controller"
	. "github.com/fabric8-services/fabric8-auth/controller"
	"github.com/fabric8-services/fabric8-auth/gormapplication"
	"github.com/fabric8-services/fabric8-auth/gormsupport/cleaner"
	"github.com/fabric8-services/fabric8-auth/gormtestsupport"
	"github.com/fabric8-services/fabric8-auth/resource"
	"github.com/fabric8-services/fabric8-auth/space"
	testsupport "github.com/fabric8-services/fabric8-auth/test"
	almtoken "github.com/fabric8-services/fabric8-auth/token"
	"github.com/goadesign/goa"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// a normal test function that will kick off TestSuiteCodebases
func TestRunCodebasesTest(t *testing.T) {
	resource.Require(t, resource.Database)
	suite.Run(t, &TestCodebaseREST{DBTestSuite: gormtestsupport.NewDBTestSuite("../config.yaml")})
}

// ========== TestCodebaseREST struct that implements SetupSuite, TearDownSuite, SetupTest, TearDownTest ==========
type TestCodebaseREST struct {
	gormtestsupport.DBTestSuite

	db      *gormapplication.GormDB
	clean   func()
	testDir string
}

func (s *TestCodebaseREST) SetupTest() {
	s.db = gormapplication.NewGormDB(s.DB)
	s.clean = cleaner.DeleteCreatedEntities(s.DB)
	s.testDir = filepath.Join("test-files", "codebase")
}

func (s *TestCodebaseREST) TearDownTest() {
	s.clean()
}

func (s *TestCodebaseREST) UnsecuredController() (*goa.Service, *CodebaseController) {
	svc := goa.New("Codebases-service")
	return svc, NewCodebaseController(svc, s.db, s.Configuration)
}

func (s *TestCodebaseREST) SecuredControllers(identity account.Identity) (*goa.Service, *CodebaseController) {
	pub, _ := almtoken.ParsePublicKey([]byte(almtoken.RSAPublicKey))

	svc := testsupport.ServiceAsUser("Codebase-Service", almtoken.NewManager(pub), identity)
	return svc, controller.NewCodebaseController(svc, s.db, s.Configuration)
}

func (s *TestCodebaseREST) TestSuccessShowCodebaseWithoutAuth() {
	resetFn := s.DisableGormCallbacks()
	defer resetFn()

	s.T().Run("success without auth", func(t *testing.T) {
		resource.Require(t, resource.Database)

		// Create space and codebase with sticky IDs
		spaceID := uuid.FromStringOrNil("a8bee527-12d2-4aff-9823-3511c1c8e6b9")
		codebaseID := uuid.FromStringOrNil("d7a282f6-1c10-459e-bb44-55a1a6d48bdd")
		cb := requireSpaceAndCodebase(t, s.db, codebaseID, spaceID)

		svc, ctrl := s.UnsecuredController()
		_, cbresp := test.ShowCodebaseOK(t, svc.Context, svc, ctrl, cb.ID)
		require.NotNil(t, cbresp)
		compareWithGolden(t, filepath.Join(s.testDir, "show", "ok_without_auth.golden.json"), cbresp)
	})
}

func requireSpaceAndCodebase(t *testing.T, db *gormapplication.GormDB, ID, spaceID uuid.UUID) *codebase.Codebase {
	var c *codebase.Codebase
	application.Transactional(db, func(appl application.Application) error {

		s := &space.Space{
			ID:   spaceID,
			Name: "Test Space " + spaceID.String(),
		}
		_, err := appl.Spaces().Create(context.Background(), s)
		require.Nil(t, err)
		stackId := "golang-default"
		c = &codebase.Codebase{
			ID:                ID,
			SpaceID:           spaceID,
			Type:              "git",
			URL:               "https://github.com/fabric8-services/fabric8-auth.git",
			StackID:           &stackId,
			LastUsedWorkspace: "my-last-used-workspace",
		}
		err = appl.Codebases().Create(context.Background(), c)
		require.Nil(t, err)
		return nil
	})
	return c
}
