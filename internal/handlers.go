package internal

import (
	"encoding/json"
	"github.com/go-pg/pg/v9"
	"github.com/google/uuid"
	"net/http"
	"test_project/internal/db"
	"time"
)

func sleepOneSec() {
	time.Sleep(time.Second)
}

func writeResponse(data interface{}, err_ *Err, status int, writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/json")
	response := new(Response)
	response.Data = data
	if err_ != nil {
		response.Error = err_
	}
	writer.WriteHeader(status)
	responseData, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	if _, err := writer.Write(responseData); err != nil {
		panic(err)
	}
}

func getRequestData(writer http.ResponseWriter, request *http.Request) (*Request, error) {
	requestData := new(Request)
	err := json.NewDecoder(request.Body).Decode(requestData)
	defer request.Body.Close()
	if err != nil {
		writeResponse(nil, ErrInvalidRequestBody, http.StatusBadRequest, writer)
	}
	return requestData, err
}

func validateRequestData(data *Request, writer http.ResponseWriter) error {
	if data.RequestType != "start" && data.RequestType != "stop" {
		writeResponse(nil, ErrInvalidRequestType, http.StatusBadRequest, writer)
	} else if _, err := uuid.Parse(data.RequestId); err != nil {
		writeResponse(nil, ErrInvalidRequestId, http.StatusBadRequest, writer)
	}
	return nil
}

func (s *server) Handler(writer http.ResponseWriter, request *http.Request) {
	requestData, err := getRequestData(writer, request)
	if err != nil {
		return
	}
	if err := validateRequestData(requestData, writer); err != nil {
		return
	}
	task := new(db.Task)
	err = s.db.Model(task).
		Where("request_id = ?", requestData.RequestId).
		First()
	isTaskNew := err == pg.ErrNoRows
	if isTaskNew {
		task = &db.Task{
			Id:        uuid.New().String(),
			RequestId: requestData.RequestId,
			Status:    db.STATUSRUNNING,
		}
	} else if err != nil {
		writeResponse(nil, ErrInternalServer, http.StatusInternalServerError, writer)
		return
	}
	if requestData.RequestType == "start" {
		if task.Status == db.STATUSSTOPPED || task.Status == db.STATUSFINISHED {
			writeResponse(nil, ErrStartStoppedTask, http.StatusBadRequest, writer)
			return
		} else if !isTaskNew {
			writeResponse(nil, ErrStartRunningTask, http.StatusBadRequest, writer)
			return
		}

		if err := s.db.Insert(task); err != nil {
			writeResponse(nil, ErrInternalServer, http.StatusInternalServerError, writer)
			return
		}
		n := 10
		s.db.Conn().PoolStats()
		for i := 1; i <= 10; i++ {
			sleepOneSec()
			err := s.db.Model(task).WherePK().First()
			if err != nil {
				writeResponse(nil, ErrInternalServer, http.StatusInternalServerError, writer)
				return
			}
			if task.Status == db.STATUSSTOPPED {
				writeResponse(nil, nil, http.StatusOK, writer)
				return
			}
			task.StepsCompleted++
			if i == n {
				task.Status = db.STATUSFINISHED
			}
			if err := s.db.Update(task); err != nil {
				writeResponse(nil, ErrInternalServer, http.StatusInternalServerError, writer)
				return
			}
		}
		writeResponse(nil, nil, http.StatusOK, writer)
	} else if requestData.RequestType == "stop" {
		if task.StepsCompleted > 5 {
			writeResponse(nil, ErrTooLateToStop, http.StatusBadRequest, writer)
			return
		}
		task.Status = db.STATUSSTOPPED
		if err := s.db.Update(task); err != nil {
			writeResponse(nil, ErrInternalServer, http.StatusInternalServerError, writer)
		}
	}
}
