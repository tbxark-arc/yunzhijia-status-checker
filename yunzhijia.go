package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type YzjResponse[T any] struct {
	Data      T    `json:"data"`
	ErrorCode int  `json:"errorCode"`
	Success   bool `json:"success"`
}

type TicketContainer struct {
	Appids string `json:"appids"`
}

type ClockInFlowFlow struct {
	HasRule      bool `json:"hasRule"`
	Rest         bool `json:"rest"`
	SignDataList []struct {
		TimePoint struct {
			WorkTime      int64  `json:"workTime"`
			ClockInTime   int64  `json:"clockInTime"`
			TimePointType string `json:"timePointType"`
		} `json:"timePoint"`
	} `json:"signDataList"`
	WorkHoursStr string `json:"workHoursStr"`
}

type YunZhiJia struct {
	token string
	oid   string
	appid string
}

type ClockInTimeType string

const (
	ClockInTimeTypeStart = ClockInTimeType("START_WORK")
	ClockInTimeTypeEnd   = ClockInTimeType("END_WORK")
)

func NewYunZhiJia(token, oid, appid string) *YunZhiJia {
	return &YunZhiJia{token: token, oid: oid, appid: appid}
}

func (y *YunZhiJia) ClockInFlowForDate(date string) (*ClockInFlowFlow, error) {
	ticket, err := fetchTicket(y.token)
	if err != nil {
		log.Printf("fetch ticket error: %v", err)
		return nil, err
	}
	flow, err := fetchClockInFlow(y.oid, y.appid, ticket, date)
	if err != nil {
		log.Printf("fetch clock in flow error: %v", err)
		return nil, err
	}
	return flow, nil
}

func (y *YunZhiJia) ClockInFlow() (*ClockInFlowFlow, error) {
	today := time.Now()
	date := fmt.Sprintf("%d-%02d-%02d", today.Year(), today.Month(), today.Day())
	return y.ClockInFlowForDate(date)
}

func (y *YunZhiJia) IsClockInToday(t ClockInTimeType) (bool, error) {
	flow, err := y.ClockInFlow()
	if err != nil {
		return false, err
	}
	//if !flow.HasRule {
	//	return true, nil
	//}
	//if flow.Rest {
	//	return true, nil
	//}
	//log.Printf("type: %s, flow: %+v", t, flow)
	for _, v := range flow.SignDataList {
		if v.TimePoint.TimePointType == string(t) {
			if v.TimePoint.ClockInTime == 0 {
				return false, nil
			}
			//log.Printf("%s: %+v", t, v.TimePoint)
			if t == ClockInTimeTypeStart {
				return v.TimePoint.ClockInTime <= v.TimePoint.WorkTime, nil
			} else {
				return v.TimePoint.ClockInTime >= v.TimePoint.WorkTime, nil
			}
		}
	}
	return true, nil
}

func fetchTicket(token string) (string, error) {

	req, err := http.NewRequest("POST", "https://do.yunzhijia.com/cloudwork/batchticket/tickets", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("openToken", token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data := YzjResponse[TicketContainer]{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}
	if data.Data.Appids == "" {
		return "", fmt.Errorf("fetch ticket error: %v", data)
	}
	return data.Data.Appids, nil
}

func fetchClockInFlow(oid, appid, ticket, date string) (*ClockInFlowFlow, error) {

	body := map[string]string{
		"date": date,
		"oid":  oid,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://www.yunzhijia.com/gateway/smartatt-core/mobile/statistics/getClockInFlow?appId=%s&ticket=%s", appid, ticket)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data := YzjResponse[ClockInFlowFlow]{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data.Data, nil
}
