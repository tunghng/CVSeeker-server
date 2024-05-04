package gpt

import (
	"CVSeeker/pkg/cfg"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

type IGptAdaptorClient interface {
	CreateAssistant(request AssistantRequest) (*AssistantResponse, error)
	CreateThread() (*ThreadResponse, error)
	DeleteThread(threadID string) (*DeleteThreadResponse, error)
	ListMessages(threadID string, limit int, order, after, before string) (*ListMessagesResponse, error)
	GetRunDetails(threadID, runID string) (*RunResponse, error)
	CreateRun(threadID string, request CreateRunRequest) (*RunResponse, error)
	SubmitToolOutputs(threadID, runID string, request SubmitToolOutputsRequest) (*RunResponse, error)
	CreateMessage(threadID string, request CreateMessageRequest) (*MessageResponse, error)
	DownloadAndUploadImage(imageURL string) (*UploadFileResponse, error)
}

type gptAdaptorClient struct {
	Client *http.Client
	ApiKey string
}

func NewGptAdaptorClient(cfgReader *viper.Viper) (IGptAdaptorClient, error) {
	return &gptAdaptorClient{
		Client: &http.Client{},
		ApiKey: cfgReader.GetString(cfg.GptApiKey),
	}, nil
}

func (g *gptAdaptorClient) addCommonHeaders(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+g.ApiKey)
	req.Header.Add("OpenAI-Beta", OpenaiAssistantsV1)
}

func (g *gptAdaptorClient) CreateAssistant(request AssistantRequest) (*AssistantResponse, error) {
	url := AssistantEndpoint

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	g.addCommonHeaders(req)

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errMsg string
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg = "Failed to read response body"
		} else {
			errMsg = string(bodyBytes)
		}
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, errMsg)
	}
	var response AssistantResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (g *gptAdaptorClient) CreateThread() (*ThreadResponse, error) {
	url := ThreadEndpoint
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	g.addCommonHeaders(req)

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errMsg string
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg = "Failed to read response body"
		} else {
			errMsg = string(bodyBytes)
		}
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, errMsg)
	}
	// Xử lý response
	var response ThreadResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (g *gptAdaptorClient) DeleteThread(threadID string) (*DeleteThreadResponse, error) {
	url := fmt.Sprintf("%v/%v", ThreadEndpoint, threadID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	g.addCommonHeaders(req)

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errMsg string
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg = "Failed to read response body"
		} else {
			errMsg = string(bodyBytes)
		}
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, errMsg)
	}
	// Xử lý response
	var response DeleteThreadResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (g *gptAdaptorClient) CreateMessage(threadID string, request CreateMessageRequest) (*MessageResponse, error) {
	url := fmt.Sprintf("%v/%v/messages", ThreadEndpoint, threadID)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	g.addCommonHeaders(req)

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errMsg string
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg = "Failed to read response body"
		} else {
			errMsg = string(bodyBytes)
		}
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, errMsg)
	}

	var response MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (g *gptAdaptorClient) ListMessages(threadID string, limit int, order, after, before string) (*ListMessagesResponse, error) {
	urls := fmt.Sprintf("%v/%v/messages", ThreadEndpoint, threadID)

	// Xây dựng các tham số truy vấn
	queryParams := url.Values{}
	if limit > 0 {
		queryParams.Add("limit", strconv.Itoa(limit))
	}
	if order != "" {
		queryParams.Add("order", order)
	}
	if after != "" {
		queryParams.Add("after", after)
	}
	if before != "" {
		queryParams.Add("before", before)
	}
	urls += "?" + queryParams.Encode()
	req, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return nil, err
	}
	g.addCommonHeaders(req)

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errMsg string
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg = "Failed to read response body"
		} else {
			errMsg = string(bodyBytes)
		}
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, errMsg)
	}
	// Xử lý response
	var response ListMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (g *gptAdaptorClient) CreateRun(threadID string, request CreateRunRequest) (*RunResponse, error) {
	urls := fmt.Sprintf("%v/%v/runs", ThreadEndpoint, threadID)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", urls, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	// Sử dụng hàm helper để thêm headers chung
	g.addCommonHeaders(req)

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errMsg string
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg = "Failed to read response body"
		} else {
			errMsg = string(bodyBytes)
		}
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, errMsg)
	}
	// Xử lý response
	var response RunResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (g *gptAdaptorClient) GetRunDetails(threadID, runID string) (*RunResponse, error) {
	urls := fmt.Sprintf("%v/%v/runs/%v", ThreadEndpoint, threadID, runID)

	req, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return nil, err
	}

	// Sử dụng hàm helper để thêm headers chung
	g.addCommonHeaders(req)

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errMsg string
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg = "Failed to read response body"
		} else {
			errMsg = string(bodyBytes)
		}
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, errMsg)
	}
	// Xử lý response
	var response RunResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (g *gptAdaptorClient) SubmitToolOutputs(threadID, runID string, request SubmitToolOutputsRequest) (*RunResponse, error) {
	urls := fmt.Sprintf("%v/%v/runs/%v/submit_tool_outputs", ThreadEndpoint, threadID, runID)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", urls, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	// Sử dụng hàm helper để thêm headers chung
	g.addCommonHeaders(req)

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errMsg string
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg = "Failed to read response body"
		} else {
			errMsg = string(bodyBytes)
		}
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, errMsg)
	}
	// Xử lý response
	var response RunResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (g *gptAdaptorClient) DownloadAndUploadImage(imageURL string) (*UploadFileResponse, error) {
	// Bước 1: Tải ảnh từ URL
	resp, err := http.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("Error downloading image: %v", err)
	}
	defer resp.Body.Close()

	// Bước 2: Lấy tên file từ URL và lưu ảnh vào hệ thống file
	fileName := path.Base(imageURL) // Lấy tên file từ URL
	file, err := os.Create(fileName)
	if err != nil {
		return nil, fmt.Errorf("Error creating file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error saving image: %v", err)
	}

	// Bước 3: Upload ảnh lên OpenAI
	uploadResponse, err := g.UploadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("Error uploading file to OpenAI: %v", err)
	}

	// Bước 4: Xoá ảnh khỏi hệ thống file cục bộ
	err = os.Remove(fileName)
	if err != nil {
		return nil, fmt.Errorf("Error deleting file: %v", err)
	}

	return uploadResponse, nil
}

func (g *gptAdaptorClient) UploadFile(filePath string) (*UploadFileResponse, error) {
	// Mở file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// Tạo multipart form request
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// Thêm field 'purpose'
	_ = writer.WriteField("purpose", "assistants")

	// Thêm file vào form
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("error creating form file: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("Error copying file to form file: %v", err)
	}

	// Đóng writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("Error closing writer: %v", err)
	}

	// Tạo và gửi request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/files", &buffer)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+g.ApiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Kiểm tra response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Upload failed with status code: %d", resp.StatusCode)
	}

	// Decode response
	var uploadResp UploadFileResponse
	err = json.NewDecoder(resp.Body).Decode(&uploadResp)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response JSON: %v", err)
	}

	return &uploadResp, nil
}
