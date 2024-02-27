package main

import (
	"context"
	pb "converter/converter"
	"database/sql"

	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestConvertCurrency_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	server := &Server{db: db}

	mock.ExpectQuery("SELECT conversion_value FROM currency_table WHERE currency = .*").
		WithArgs("INR").
		WillReturnRows(sqlmock.NewRows([]string{"conversion_value"}).AddRow(1.0))

	mock.ExpectQuery("SELECT conversion_value FROM currency_table WHERE currency = .*").
		WithArgs("EURO").
		WillReturnRows(sqlmock.NewRows([]string{"conversion_value"}).AddRow(90.0))

	request := &pb.ConvertRequest{
		Amount:       90,
		FromCurrency: "INR",
		ToCurrency:   "EURO",
	}

	response, err := server.ConvertCurrency(context.Background(), request)

	assert.NoError(t, err)
	assert.Equal(t, float64(1), response.Amount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConvertCurrency_InvalidFromCurrency(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	server := &Server{db: db}

	mock.ExpectQuery("SELECT conversion_value FROM currency_table WHERE currency = .*").
		WithArgs("UNKNOWN").
		WillReturnError(sql.ErrNoRows)

	request := &pb.ConvertRequest{
		Amount:       100,
		FromCurrency: "UNKNOWN",
		ToCurrency:   "EURO",
	}

	_, err = server.ConvertCurrency(context.Background(), request)

	assert.Error(t, err)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, statusErr.Code())
}

