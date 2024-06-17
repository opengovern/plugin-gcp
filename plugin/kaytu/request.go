package kaytu

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var ErrLogin = errors.New("your session is expired, please login")

func Ec2InstanceWastageRequest(reqBody GcpComputeInstanceWastageRequest, token string) (*GcpComputeInstanceWastageResponse, error) {
	payloadEncoded, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://api.kaytu.io/kaytu/wastage/api/v1/wastage/gcp-compute", bytes.NewBuffer(payloadEncoded))
	if err != nil {
		return nil, fmt.Errorf("[ec2-instance]: %v", err)
	}
	req.Header.Add("content-type", "application/json")
	if len(token) > 0 {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[gcp-compute]: %v", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("[gcp-compute]: %v", err)
	}
	err = res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("[gcp-compute]: %v", err)
	}

	if res.StatusCode == 401 {
		return nil, ErrLogin
	}

	if res.StatusCode >= 300 || res.StatusCode < 200 {
		return nil, fmt.Errorf("server returned status code %d, [requestAbout] : %s", res.StatusCode, string(body))
	}

	response := GcpComputeInstanceWastageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("[requestAbout] : %v", err)
	}
	return &response, nil
}
