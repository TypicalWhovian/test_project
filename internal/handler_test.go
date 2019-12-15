package internal

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

func init() {
	go Run()
	time.Sleep(time.Second * 4)
}

func post(data interface{}) (*http.Response, error) {
	body, err := json.Marshal(data)
	req, err := http.NewRequest(http.MethodPost, "http://0.0.0.0:"+os.Getenv("PORT"), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	return new(http.Client).Do(req)
}

func TestSuccessStartServer_Handler(t *testing.T) {
	data := map[string]string{
		"requestId": uuid.New().String(),
		"type":      "start",
	}
	resp, err := post(data)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, instead got: %d", http.StatusOK, resp.StatusCode)
	}
}

func TestSuccessSeveralStartServer_Handler(t *testing.T) {
	n := 5_000 // 25k was ok and took 112s to run, however at 30k test timed out (time out limit is 10 minutes)
	nFinished := 0
	var m sync.Mutex
	finished := make(chan bool, 1)
	for i := 0; i < n; i++ {
		go func() {
			resp, err := post(map[string]string{
				"requestId": uuid.New().String(),
				"type":      "start",
			})
			if err != nil {
				t.Fatal(err)
			}
			responseData := new(Response)
			defer resp.Body.Close()
			if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
				t.Fatal(err)
			}
			if responseData.Error != nil {
				t.Fatal(responseData.Error)
			}
			m.Lock()
			nFinished++
			m.Unlock()
			if nFinished == n {
				finished <- true
			}
		}()
	}
	<-finished
}

func TestSuccessSeveralStopServer_Handler(t *testing.T) {
	n := 5_000
	var ids []string
	for i := 0; i < n; i++ {
		ids = append(ids, uuid.New().String())
	}
	nFinished := 0
	var m sync.Mutex
	finished := make(chan bool, 1)
	for i := 0; i < n; i++ {
		go func(i int) {
			resp, err := post(map[string]string{
				"requestId": ids[i],
				"type":      "start",
			})
			if err != nil {
				t.Fatal(err)
			}
			responseData := new(Response)
			defer resp.Body.Close()
			if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
				t.Fatal(err)
			}
			if responseData.Error != nil {
				t.Fatal(responseData.Error)
			}
		}(i)
	}
	time.Sleep(time.Second * 8)
	rand.Seed(time.Now().Unix())
	nStop := n / 2
	for i := 0; i < nStop; i++ {
		go func() {
			randomId := ids[rand.Intn(n)]
			resp, err := post(map[string]string{
				"requestId": randomId,
				"type":      "stop",
			})
			if err != nil {
				t.Fatal(err)
			}
			responseData := new(Response)
			defer resp.Body.Close()
			if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
				t.Fatal(err)
			}
			if responseData.Error != nil {
				code := responseData.Error.Code
				if code != ErrTooLateToStop.Code && code != ErrTaskIsNotRunning.Code && code != ErrStartStoppedTask.Code {
					t.Fatal(responseData.Error)
				}
			}
			m.Lock()
			nFinished++
			m.Unlock()
			if nFinished == nStop {
				finished <- true
			}
		}()
	}
	<-finished
}

func TestSuccessStopServer_Handler(t *testing.T) {
	requestId := uuid.New().String()
	data := map[string]string{
		"requestId": requestId,
		"type":      "start",
	}
	go post(data)
	time.Sleep(time.Second * 2)
	data = map[string]string{
		"requestId": requestId,
		"type":      "stop",
	}
	resp, err := post(data)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, instead got: %d", http.StatusOK, resp.StatusCode)
	}
}

func TestFailStopServer_Handler(t *testing.T) {
	requestId := uuid.New().String()
	data := map[string]string{
		"requestId": requestId,
		"type":      "start",
	}
	go post(data)
	time.Sleep(time.Second * 8)
	data = map[string]string{
		"requestId": requestId,
		"type":      "stop",
	}
	resp, err := post(data)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, instead got: %d", http.StatusBadRequest, resp.StatusCode)
	}
	defer resp.Body.Close()
	responseData := new(Response)
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		t.Fatal(err)
	}
	if responseData.Error.Code != ErrTooLateToStop.Code {
		t.Errorf("expected error: %v, instead got: %v", responseData.Error, ErrTooLateToStop)
	}
}

func TestDuplicateIdsServer_Handler(t *testing.T) {
	requestId := uuid.New().String()
	data := map[string]string{
		"requestId": requestId,
		"type":      "start",
	}
	go post(data)
	time.Sleep(time.Second)
	resp, err := post(data)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, instead got: %d", http.StatusBadRequest, resp.StatusCode)
	}
	defer resp.Body.Close()
	responseData := new(Response)
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		t.Fatal(err)
	}
	if responseData.Error.Code != ErrStartRunningTask.Code {
		t.Errorf("expected error: %v, instead got: %v", responseData.Error, ErrStartRunningTask)
	}
}

func TestStartFinishedTaskServer_Handler(t *testing.T) {
	requestId := uuid.New().String()
	data := map[string]string{
		"requestId": requestId,
		"type":      "start",
	}
	resp, err := post(data)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = post(data)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, instead got: %d", http.StatusBadRequest, resp.StatusCode)
	}
	defer resp.Body.Close()
	responseData := new(Response)
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		t.Fatal(err)
	}
	if responseData.Error.Code != ErrStartStoppedTask.Code {
		t.Errorf("expected error: %v, instead got: %v", ErrStartStoppedTask, responseData.Error)
	}
}
