package handlers

import (
	services "CVSeeker/cmd/CVSeeker/internal/service"
	"CVSeeker/cmd/CVSeeker/pkg/utils"
	"CVSeeker/internal/dtos"
	"CVSeeker/internal/errors"
	"CVSeeker/pkg/gpt"
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
	"strings"
)

type ChatbotHandler struct {
	BaseHandler
	chatbotService services.IChatbotService
}

type ChatbotHandlerParams struct {
	dig.In
	BaseHandler    BaseHandler
	ChatbotService services.IChatbotService
}

func NewChatbotHandler(params ChatbotHandlerParams) *ChatbotHandler {
	return &ChatbotHandler{
		BaseHandler:    params.BaseHandler,
		chatbotService: params.ChatbotService,
	}
}

// StartChatSession
// @Summary Start a new chat session
// @Description Starts a new chat session by creating an assistant and a thread, using specified documents.
// @Tags Chatbot
// @Accept json
// @Produce json
// @Param body body dtos.StartChatRequest true "Comma-separated list of document IDs"
// @Success 200 {object} meta.BasicResponse{data=gpt.ThreadResponse}
// @Failure 400,500 {object} meta.Error
// @Router /cvseeker/resumes/thread/start [POST]
func (_this *ChatbotHandler) StartChatSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		var chatRequest dtos.StartChatRequest
		if err := c.ShouldBindJSON(&chatRequest); err != nil {
			_this.RespondError(c, errors.NewCusErr(errors.ErrCommonInvalidRequest))
			return
		}

		if strings.TrimSpace(chatRequest.Ids) == "" {
			_this.RespondError(c, errors.NewCusErr(errors.ErrCommonInvalidRequest))
			return
		}

		resp, err := _this.chatbotService.StartChatSession(c, chatRequest.Ids, chatRequest.ThreadName)
		_this.HandleResponse(c, resp, err)
	}
}

// SendMessage
// @Summary Send a message to a chat session
// @Description Sends a message to the specified chat session using message content provided in the request body.
// @Tags Chatbot
// @Accept json
// @Produce json
// @Param threadId path string true "Thread ID"
// @Param body body dtos.QueryRequest true "Message content"
// @Success 200 {object} meta.BasicResponse{data=gpt.ListMessagesResponse}
// @Failure 400,500 {object} meta.Error
// @Router /cvseeker/resumes/thread/{threadId}/send [POST]
func (_this *ChatbotHandler) SendMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		threadID := strings.TrimSpace(c.Param("threadId"))
		if threadID == "" {
			_this.RespondError(c, errors.NewCusErr(errors.ErrCommonInternalServer))
			return
		}

		var msgContent dtos.QueryRequest
		if err := c.ShouldBindJSON(&msgContent); err != nil {
			_this.RespondError(c, errors.NewCusErr(errors.ErrCommonInvalidRequest))
			return
		}

		if strings.TrimSpace(msgContent.Content) == "" {
			_this.RespondError(c, errors.NewCusErr(errors.ErrCommonInvalidRequest))
			return
		}

		resp, err := _this.chatbotService.SendMessageToChat(c, threadID, msgContent.Content)
		_this.HandleResponse(c, resp, err)
	}
}

// ListMessage
// @Summary List messages belonging to a thread
// @Description Get a list of messages for a thread.
// @Tags Chatbot
// @Accept json
// @Produce json
// @Param threadId path string true "Thread ID"
// @Param limit query int false "Maximum number of messages to return"
// @Param after query string false "Cursor for pagination, specifying an exclusive start point for the list (ID of a message)"
// @Param before query string false "Cursor for pagination, specifying an exclusive end point for the list (ID of a message)"
// @Success  200  {object}  meta.BasicResponse{data=gpt.ListMessagesResponse}
// @Failure   400,401,404,500  {object}  meta.Error
// @Security  BearerAuth
// @Router /cvseeker/resumes/thread/{threadId}/messages [GET]
func (_this *ChatbotHandler) ListMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		threadId := strings.TrimSpace(c.Param("threadId"))
		if threadId == "" {
			_this.RespondError(c, errors.NewCusErr(errors.ErrCommonInternalServer))
			return
		}
		limit := utils.Str2StrInt64(c.Query("limit"), true)
		after := strings.TrimSpace(c.Query("after"))
		before := strings.TrimSpace(c.Query("before"))

		var request gpt.ListMessageRequest
		request.ThreadId = threadId
		request.Limit = int(limit)
		request.After = after
		request.Before = before
		resp, err := _this.chatbotService.ListMessage(c, request)
		_this.HandleResponse(c, resp, err)
	}
}

// GetAllThreads
// @Summary Get all thread IDs
// @Description Retrieves all thread IDs from the database.
// @Tags Chatbot
// @Accept json
// @Produce json
// @Success 200 {object} meta.BasicResponse{data=[]dtos.Thread}
// @Failure 400,500 {object} meta.Error
// @Router /cvseeker/resumes/thread [GET]
func (_this *ChatbotHandler) GetAllThreads() gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := _this.chatbotService.GetAllThreads(c)
		_this.HandleResponse(c, resp, err)
	}
}

// GetResumesByThreadID
// @Summary Get resume IDs by thread ID
// @Description Retrieves all resume IDs associated with a given thread ID.
// @Tags Chatbot
// @Accept json
// @Produce json
// @Param threadId path string true "Thread ID"
// @Success 200 {object} meta.BasicResponse
// @Failure 400,500 {object} meta.Error
// @Router /cvseeker/resumes/thread/{threadId} [GET]
func (_this *ChatbotHandler) GetResumesByThreadID() gin.HandlerFunc {
	return func(c *gin.Context) {
		threadID := strings.TrimSpace(c.Param("threadId"))
		if threadID == "" {
			_this.RespondError(c, errors.NewCusErr(errors.ErrCommonInternalServer))
			return
		}
		resp, err := _this.chatbotService.GetResumesByThreadID(c, threadID)
		_this.HandleResponse(c, resp, err)
	}
}

// UpdateThreadName
// @Summary Update a thread's name
// @Description Updates the name of an existing thread by thread ID.
// @Tags Chatbot
// @Accept json
// @Produce json
// @Param threadId path string true "Thread ID"
// @Param newName body string true "New Name for the Thread"
// @Success 200 {object} meta.BasicResponse{data=[]elasticsearch.ResumeSummaryDTO}
// @Failure 400,500 {object} meta.Error
// @Router /cvseeker/resumes/thread/{threadId}/updateName [POST]
func (_this *ChatbotHandler) UpdateThreadName() gin.HandlerFunc {
	return func(c *gin.Context) {
		threadID := strings.TrimSpace(c.Param("threadId"))
		if threadID == "" {
			_this.RespondError(c, errors.NewCusErr(errors.ErrCommonInvalidRequest))
			return
		}

		var newNameRequest struct {
			NewName string `json:"newName"`
		}
		if err := c.ShouldBindJSON(&newNameRequest); err != nil {
			_this.RespondError(c, errors.NewCusErr(errors.ErrCommonInvalidRequest))
			return
		}

		if strings.TrimSpace(newNameRequest.NewName) == "" {
			_this.RespondError(c, errors.NewCusErr(errors.ErrCommonInvalidRequest))
			return
		}

		resp, err := _this.chatbotService.UpdateThreadName(c, threadID, newNameRequest.NewName)
		_this.HandleResponse(c, resp, err)
	}
}

// DeleteThreadById
// @Summary Delete a thread by ID
// @Description Deletes the specified thread by its ID.
// @Tags Chatbot
// @Accept json
// @Produce json
// @Param threadId path string true "Thread ID to be deleted"
// @Success 200 {object} meta.BasicResponse
// @Failure 400,404,500 {object} meta.Error
// @Router /cvseeker/resumes/threads/{threadId} [DELETE]
func (_this *ChatbotHandler) DeleteThreadById() gin.HandlerFunc {
	return func(c *gin.Context) {
		threadID := strings.TrimSpace(c.Param("threadId"))
		if threadID == "" {
			_this.RespondError(c, errors.NewCusErr(errors.ErrCommonInvalidRequest))
			return
		}
		resp, err := _this.chatbotService.DeleteThreadById(c, threadID)
		if err != nil {
			_this.RespondError(c, err)
			return
		}
		_this.HandleResponse(c, resp, nil)
	}
}
