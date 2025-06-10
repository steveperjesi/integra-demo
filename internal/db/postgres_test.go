package db_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/steveperjesi/integra-demo/internal/db"
)

func TestDB(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DB Suite")
}

var _ = ginkgo.Describe("Connect", func() {
	var (
		conn *sql.DB
		err  error
	)

	ginkgo.BeforeEach(func() {
		// Set fake or test DB credentials (use real ones in a CI or Docker-based test DB setup)
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_USER", "postgres")
		os.Setenv("DB_PASSWORD", "password")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_PORT", "5432")
	})

	ginkgo.AfterEach(func() {
		if conn != nil {
			conn.Close()
		}
	})

	ginkgo.Context("when environment variables are set correctly", func() {
		ginkgo.It("returns a *sql.DB without error", func() {
			conn, err = db.Connect()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(conn).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("when environment variables are missing", func() {
		ginkgo.It("returns a *sql.DB with a bad DSN", func() {
			os.Unsetenv("DB_HOST")
			conn, err = db.Connect()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(conn).ToNot(gomega.BeNil())
		})
	})
})
