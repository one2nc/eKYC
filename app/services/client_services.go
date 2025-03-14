package services

import (
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/mkrs2404/eKYC/app/auth"
	"github.com/mkrs2404/eKYC/app/database"
	"github.com/mkrs2404/eKYC/app/models"
	"github.com/mkrs2404/eKYC/app/resources"
)

/*SaveClient saves the client to the DB and returns back the client and an error(if there's email duplicacy)*/
func SaveClient(signUpRequest resources.SignUpRequest) (models.Client, error) {

	//Create a client object from the provided request data
	var client models.Client
	client.Name = signUpRequest.Name
	client.Email = signUpRequest.Email
	client.Plan = GetPlanId(signUpRequest.Plan)

	err := database.DB.Create(&client).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr); pgErr.Code == "23505" {
			err = errors.New("client already exists")
		}
	}
	return client, err
}

//SetupClient creates a client in DB and returns the Auth header
func SetupClient() (string, models.Client, error) {

	dummyClient := resources.SignUpRequest{
		Name:  "bob",
		Email: "bob@one2n.in",
		Plan:  "basic",
	}

	client, err := SaveClient(dummyClient)
	if err != nil {
		return "", client, err
	}

	token, err := auth.GenerateToken(client.ID)
	//Creating the auth header
	token = fmt.Sprintf("Bearer %s", token)

	return token, client, err
}

//GetClient finds the client with the given ID
func GetClient(clientID int) (models.Client, error) {
	var client models.Client
	err := database.DB.First(&client, clientID).Error
	return client, err
}
