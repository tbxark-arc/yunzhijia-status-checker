package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	HasRule    bool `json:"hasRule"`
	TimePoints []struct {
		TimePointType ClockInTimeType `json:"timePointType"` // START_WORK/END_WORK

		WorkTime int64 `json:"workTime"` // 签到期限

		ClockInTime int64 `json:"clockInTime"` // 签到时间
		HasTravel   int   `json:"hasTravel"`   // 休假
		HasGoOut    int   `json:"hasGoOut"`    // 外出
		HasLeave    bool  `json:"hasLeave"`    // 请假
	} `json:"timePoints"`
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
	if !flow.HasRule {
		return true, nil
	}
	for _, v := range flow.TimePoints {
		if v.TimePointType == t {
			if v.HasTravel == 1 || v.HasGoOut == 1 || v.HasLeave { // 休假或外出
				return true, nil
			}
			switch v.TimePointType {
			case ClockInTimeTypeStart:
				if v.ClockInTime == 0 { // 未打卡
					return false, nil
				}
				return v.ClockInTime <= v.WorkTime, nil
			case ClockInTimeTypeEnd:
				if time.Now().UnixMilli() < v.WorkTime { // 未到下班时间
					return true, nil
				}
				if v.ClockInTime == 0 { // 未打卡
					return false, nil
				}
				return v.ClockInTime >= v.WorkTime, nil
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

	respRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data := YzjResponse[ClockInFlowFlow]{}
	err = json.Unmarshal(respRaw, &data)
	if err != nil {
		return nil, err
	}
	return &data.Data, nil
}
