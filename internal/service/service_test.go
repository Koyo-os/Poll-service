package service_test

import (
	"testing"

	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/internal/service"
	"github.com/Koyo-os/Poll-service/test/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPollService_Add_Success(t *testing.T) {
	repoMock := new(mocks.PollRepository)
	pubMock := new(mocks.Publisher)
	cashMock := new(mocks.Casher)

	pollService := service.Init(repoMock, pubMock, cashMock)

	newPoll := &entity.Poll{
		ID:   uuid.New(),
		Desc: "about cat",
	}

	repoMock.On("Add", newPoll).Return(nil)
	pubMock.On("Publish", newPoll, "polls.add").Return(nil)
	cashMock.On("DoCashing", mock.Anything, "poll:"+newPoll.ID.String(), newPoll).Return(nil)

	err := pollService.Add(newPoll)

	assert.NoError(t, err)
	repoMock.AssertExpectations(t)
	pubMock.AssertExpectations(t)
	cashMock.AssertExpectations(t)
}

func TestPollService_Add_RepositoryError(t *testing.T) {
	repoMock := new(mocks.PollRepository)
	pubMock := new(mocks.Publisher)
	cashMock := new(mocks.Casher)

	pollService := service.Init(repoMock, pubMock, cashMock)

	newPoll := &entity.Poll{ID: uuid.New()}

	repoMock.On("Add", newPoll).Return(assert.AnError)

	err := pollService.Add(newPoll)

	assert.Error(t, err)
	pubMock.AssertNotCalled(t, "Publish")
	cashMock.AssertNotCalled(t, "DoCashing")
}

func TestPollService_Update_Success(t *testing.T) {
	repoMock := new(mocks.PollRepository)
	pubMock := new(mocks.Publisher)
	cashMock := new(mocks.Casher)

	pollService := service.Init(repoMock, pubMock, cashMock)

	pollID := uuid.New().String()
	updatedPoll := &entity.Poll{
		ID:   uuid.MustParse(pollID),
		Desc: "Updated Question",
	}
	repoMock.On("Update", uuid.MustParse(pollID), updatedPoll).Return(nil)
	pubMock.On("Publish", updatedPoll, "polls.update").Return(nil)
	cashMock.On("DoCashing", mock.Anything, "poll:"+pollID, updatedPoll).Return(nil)

	err := pollService.Update(pollID, updatedPoll)

	assert.NoError(t, err)
	repoMock.AssertExpectations(t)
	pubMock.AssertExpectations(t)
	cashMock.AssertExpectations(t)
}

func TestPollService_Update_InvalidID(t *testing.T) {
	repoMock := new(mocks.PollRepository)
	pubMock := new(mocks.Publisher)
	cashMock := new(mocks.Casher)

	pollService := service.Init(repoMock, pubMock, cashMock)

	err := pollService.Update("invalid-uuid", &entity.Poll{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID")
	repoMock.AssertNotCalled(t, "Update")
	pubMock.AssertNotCalled(t, "Publish")
	cashMock.AssertNotCalled(t, "DoCashing")
}

func TestPollService_Update_CasherError(t *testing.T) {
	repoMock := new(mocks.PollRepository)
	pubMock := new(mocks.Publisher)
	cashMock := new(mocks.Casher)

	pollService := service.Init(repoMock, pubMock, cashMock)

	pollID := uuid.New().String()
	updatedPoll := &entity.Poll{ID: uuid.MustParse(pollID)}

	repoMock.On("Update", uuid.MustParse(pollID), updatedPoll).Return(nil)
	pubMock.On("Publish", updatedPoll, "polls.update").Return(nil)
	cashMock.On("DoCashing", mock.Anything, "poll:"+pollID, updatedPoll).Return(assert.AnError)

	err := pollService.Update(pollID, updatedPoll)

	assert.NoError(t, err)
	cashMock.AssertExpectations(t)
}
