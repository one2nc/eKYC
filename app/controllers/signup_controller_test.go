package controllers

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/mkrs2404/eKYC/app/models"
	"github.com/mkrs2404/eKYC/app/rabbitmq"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/mkrs2404/eKYC/app/database"
	"github.com/mkrs2404/eKYC/app/minio_client"
	"github.com/mkrs2404/eKYC/app/redis_client"
	"github.com/mkrs2404/eKYC/app/services"

	"gorm.io/gorm/logger"
)

const signUpUrl = "/api/v1/signup"

var signupTestData = []struct {
	expectedCode int
	body         string
}{
	//Valid request
	{
		body:         `{"name": "bob","email": "bob@one2.in","plan": "basic"}`,
		expectedCode: 201,
	},
	//Duplicate request
	{
		body:         `{"name": "bob","email": "bob@one2.in","plan": "basic"}`,
		expectedCode: 400,
	},
	//Invalid email
	{
		body:         `{"name": "bob","email": "bobone2.in","plan": "basic"}`,
		expectedCode: 400,
	},
	//Invalid plan
	{
		body:         `{"name": "bob","email": "bob@one2.in","plan": "secure"}`,
		expectedCode: 400,
	},
	//Missing plan
	{
		body:         `{"name": "bob","email": "bob@one2.in","plan": ""}`,
		expectedCode: 400,
	},
}

//Setting up DB connection, data seeding and Minio connection
func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error fetching the environment values")
	}
	database.Connect(os.Getenv("TEST_DB_HOST"), os.Getenv("TEST_DB_NAME"), os.Getenv("TEST_DB_USER"), os.Getenv("TEST_DB_PASSWORD"), os.Getenv("TEST_DB_PORT"), logger.Silent)
	//Migrating tables to the database
	database.DB.AutoMigrate(&models.Plan{}, &models.Client{}, &models.File{}, &models.Api_Calls{})
	services.SeedPlanData()
	database.DB.Exec("DELETE FROM files")
	database.DB.Exec("DELETE FROM clients")

	minio_client.InitializeMinio(os.Getenv("TEST_MINIO_SERVER"), os.Getenv("TEST_MINIO_USER"), os.Getenv("TEST_MINIO_PWD"))
	redis_client.InitializeRedis(os.Getenv("REDIS_SERVER"), os.Getenv("REDIS_PASSWORD"))
	rabbitmq.InitializeRabbitMq(os.Getenv("RABBITMQ_USER"), os.Getenv("RABBITMQ_PWD"), os.Getenv("RABBITMQ_SERVER"))
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestSignUpClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/api/v1/signup", SignUpClient)

	for _, data := range signupTestData {
		resRecorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(resRecorder)
		ctx.Request, _ = http.NewRequest(http.MethodPost, signUpUrl, strings.NewReader(data.body))
		SignUpClient(ctx)

		if resRecorder.Code != data.expectedCode {
			t.Errorf("Expected %d, Got %d", data.expectedCode, resRecorder.Code)
		}
	}

}
