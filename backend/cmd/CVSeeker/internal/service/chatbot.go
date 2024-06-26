package services

import (
	"CVSeeker/cmd/CVSeeker/internal/cfg"
	"CVSeeker/internal/dtos"
	"CVSeeker/internal/ginLogger"
	"CVSeeker/internal/meta"
	"CVSeeker/internal/models"
	"CVSeeker/internal/repositories"
	"CVSeeker/pkg/db"
	"CVSeeker/pkg/elasticsearch"
	"CVSeeker/pkg/gpt"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"net/http"
	"strings"
)

type IChatbotService interface {
	StartChatSession(c *gin.Context, ids string, threadName string) (*meta.BasicResponse, error)
	SendMessageToChat(c *gin.Context, threadID, message string) (*meta.BasicResponse, error)
	ListMessage(c *gin.Context, request gpt.ListMessageRequest) (*meta.BasicResponse, error)
	GetAllThreads(c *gin.Context) (*meta.BasicResponse, error)
	GetResumesByThreadID(c *gin.Context, threadID string) (*meta.BasicResponse, error)
	UpdateThreadName(c *gin.Context, threadID string, newName string) (*meta.BasicResponse, error)
	DeleteThreadById(c *gin.Context, threadId string) (*meta.BasicResponse, error)
}

type ChatbotService struct {
	db               *db.DB
	assistantClient  gpt.IGptAdaptorClient
	elasticClient    elasticsearch.IElasticsearchClient
	threadRepo       repositories.IThreadRepository
	threadResumeRepo repositories.IThreadResumeRepository
}

type ChatbotServiceArgs struct {
	dig.In
	DB               *db.DB `name:"talentAcquisitionDB"`
	AssistantClient  gpt.IGptAdaptorClient
	ElasticClient    elasticsearch.IElasticsearchClient
	ThreadRepo       repositories.IThreadRepository
	ThreadResumeRepo repositories.IThreadResumeRepository
}

func NewChatbotService(args ChatbotServiceArgs) IChatbotService {
	return &ChatbotService{
		db:               args.DB,
		assistantClient:  args.AssistantClient,
		elasticClient:    args.ElasticClient,
		threadRepo:       args.ThreadRepo,
		threadResumeRepo: args.ThreadResumeRepo,
	}
}

func (_this *ChatbotService) StartChatSession(c *gin.Context, ids string, threadName string) (*meta.BasicResponse, error) {
	elasticDocumentName := viper.GetString(cfg.ElasticsearchDocumentIndex)

	// Parse the IDs from the string
	idArray := strings.Split(ids, ", ")

	// Fetch documents from Elasticsearch
	documents, err := _this.elasticClient.FetchDocumentsByIDs(c, elasticDocumentName, idArray)
	if err != nil {
		ginLogger.Gin(c).Errorf("failed to fetch documents: %v", err)
		return nil, err
	}

	// Format the documents' content from ResumeSummaryDTO
	var fullTextContent strings.Builder
	fullTextContent.WriteString("You will use these information to answer questions from the user while using markdown for clarity: ")
	for _, resume := range documents {
		fullTextContent.WriteString(fmt.Sprintf("Name: %s", resume.BasicInfo.FullName))
		fullTextContent.WriteString(fmt.Sprintf("Summary: %s; Skills: %v; ", resume.Summary, resume.Skills))
		fullTextContent.WriteString(fmt.Sprintf("Education: %s, %s, GPA: %.2f; ", resume.BasicInfo.University, resume.BasicInfo.EducationLevel, resume.BasicInfo.GPA))
		fullTextContent.WriteString("Work Experience: ")
		for _, work := range resume.WorkExperience {
			fullTextContent.WriteString(fmt.Sprintf("%s at %s, %s; ", work.JobTitle, work.Company, work.Duration))
		}
		fullTextContent.WriteString("Projects: ")
		for _, project := range resume.ProjectExperience {
			fullTextContent.WriteString(fmt.Sprintf("%s: %s; ", project.ProjectName, project.ProjectDescription))
		}
		fullTextContent.WriteString("Awards: ")
		for _, award := range resume.Award {
			fullTextContent.WriteString(fmt.Sprintf("%s; ", award.AwardName))
		}
		fullTextContent.WriteString(" | ") // Separator for multiple resumes
	}

	// Create the initial message for the thread
	initMessage := gpt.CreateMessageRequest{
		Role:    "user",
		Content: fullTextContent.String(),
	}

	// Create a new thread with the initial message
	threadRequest := gpt.CreateThreadRequest{
		Messages: []gpt.CreateMessageRequest{initMessage},
	}

	thread, err := _this.assistantClient.CreateThread(threadRequest)
	if err != nil {
		ginLogger.Gin(c).Errorf("failed to create thread: %v", err)
		return nil, err
	}

	// Create a new thread instance in the database
	newThread := &models.Thread{
		ID:   thread.ID,
		Name: threadName,
	}

	_, err = _this.threadRepo.Create(_this.db, newThread)
	if err != nil {
		ginLogger.Gin(c).Errorf("failed to create new thread record: %v", err)
		return nil, err
	}

	var threadResumes []models.ThreadResume

	// Create new thread_resume instances in the database
	for _, id := range idArray {
		threadResume := models.ThreadResume{
			ThreadID: thread.ID,
			ResumeID: id,
		}
		err = _this.threadResumeRepo.Create(_this.db, &threadResume)
		threadResumes = append(threadResumes, threadResume)
	}

	// Prepare the response with the thread information
	response := &meta.BasicResponse{
		Meta: meta.Meta{
			Code:    http.StatusOK,
			Message: "Session started successfully with initial data",
		},
		Data: thread,
	}

	return response, nil
}

func (_this *ChatbotService) SendMessageToChat(c *gin.Context, threadID, message string) (*meta.BasicResponse, error) {
	DefaultAssistant := viper.GetString(cfg.DefaultOpenAIAssistant)

	// Create message and add to thread
	messageRequest := gpt.CreateMessageRequest{
		Content: message,
		Role:    "user",
	}

	_, err := _this.assistantClient.CreateMessage(threadID, messageRequest)
	if err != nil {
		ginLogger.Gin(c).Errorf("failed to send message: %v", err)
		return nil, err
	}

	// Create run for assistant and thread with streaming enabled
	runRequest := gpt.CreateRunRequest{
		AssistantID: DefaultAssistant,
		Stream:      true,
	}

	// Collect and process streamed responses
	var messages []string
	values, err := _this.assistantClient.CreateRunAndStreamResponse(threadID, runRequest)
	if err != nil {
		ginLogger.Gin(c).Errorf("error streaming responses: %v", err)
		return nil, err
	}
	for value := range values {
		messages = append(messages, value)
	}

	// Prepare the final response
	response := &meta.BasicResponse{
		Meta: meta.Meta{
			Code:    http.StatusOK,
			Message: "Response retrieved successfully",
		},
		Data: messages, // Assuming you want to return the collected messages
	}

	return response, nil
}

func (_this *ChatbotService) ListMessage(c *gin.Context, request gpt.ListMessageRequest) (*meta.BasicResponse, error) {
	resp, err := _this.assistantClient.ListMessages(request.ThreadId, request.Limit, request.Order, request.After, request.Before)
	if err != nil {
		ginLogger.Gin(c).Errorf("Error when create assistant: %v", err)
		return nil, err
	}

	response := &meta.BasicResponse{
		Meta: meta.Meta{
			Code:    http.StatusOK,
			Message: "Success",
		},
		Data: resp,
	}
	return response, nil
}

func (_this *ChatbotService) GetAllThreads(c *gin.Context) (*meta.BasicResponse, error) {
	modelThreads, err := _this.threadRepo.GetAllThreads(_this.db)
	if err != nil {
		ginLogger.Gin(c).Errorf("failed to get all threads: %v", err)
		return nil, err
	}

	// Map model threads to DTOs
	threadDTOs := make([]dtos.Thread, len(modelThreads))
	for i, modelThread := range modelThreads {
		threadDTOs[i] = dtos.Thread{
			ID:        modelThread.ID,
			Name:      modelThread.Name,
			UpdatedAt: modelThread.UpdatedAt.Unix(),
		}
	}

	response := &meta.BasicResponse{
		Meta: meta.Meta{
			Code:    http.StatusOK,
			Message: "All threads retrieved successfully",
		},
		Data: threadDTOs,
	}
	return response, nil
}

func (_this *ChatbotService) DeleteThreadById(c *gin.Context, threadId string) (*meta.BasicResponse, error) {
	err := _this.threadRepo.Delete(_this.db, threadId)
	if err != nil {
		ginLogger.Gin(c).Errorf("failed to get all threads: %v", err)
		return nil, err
	}

	response := &meta.BasicResponse{
		Meta: meta.Meta{
			Code:    http.StatusOK,
			Message: "Thread deleted successfully",
		},
		Data: nil,
	}

	return response, nil
}

func (_this *ChatbotService) GetResumesByThreadID(c *gin.Context, threadID string) (*meta.BasicResponse, error) {
	elasticDocumentName := viper.GetString(cfg.ElasticsearchDocumentIndex)

	resumeIDs, err := _this.threadResumeRepo.GetResumeIDsByThreadID(_this.db, threadID)
	if err != nil {
		ginLogger.Gin(c).Errorf("failed to fetch resume IDs by thread ID: %v", err)
		return nil, err
	}

	var documents []*elasticsearch.ResumeSummaryDTO
	for _, resumeID := range resumeIDs {
		document, err := _this.elasticClient.GetDocumentByID(c, elasticDocumentName, resumeID)
		if err != nil {
			ginLogger.Gin(c).Errorf("failed to fetch document by ID: %v", err)
			continue // or return nil, err if you prefer to fail on the first error
		}
		documents = append(documents, document)
	}

	response := &meta.BasicResponse{
		Meta: meta.Meta{
			Code:    http.StatusOK,
			Message: "Resume IDs retrieved successfully for the thread",
		},
		Data: documents,
	}
	return response, nil
}

func (_this *ChatbotService) UpdateThreadName(c *gin.Context, threadID string, newName string) (*meta.BasicResponse, error) {
	// Attempt to update the thread name
	err := _this.threadRepo.UpdateThreadName(_this.db, threadID, newName)
	if err != nil {
		ginLogger.Gin(c).Errorf("failed to update thread name: %v", err)
		return nil, err
	}

	// Retrieve the updated thread to confirm the change
	updatedThread, err := _this.threadRepo.FindByID(_this.db, threadID)
	if err != nil {
		ginLogger.Gin(c).Errorf("failed to fetch updated thread: %v", err)
		return nil, err
	}

	// Prepare the response
	response := &meta.BasicResponse{
		Meta: meta.Meta{
			Code:    http.StatusOK,
			Message: "Thread name updated successfully",
		},
		Data: updatedThread,
	}
	return response, nil
}
