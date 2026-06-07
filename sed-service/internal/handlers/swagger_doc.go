package handlers

// Файл содержит аннотации swag для маршрутов из internal/server/router.go.

// Health godoc
// @Summary Liveness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func swaggerSEDHealth() {}

// Ready godoc
// @Summary Readiness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]interface{}
// @Router /ready [get]
func swaggerSEDReady() {}

// ListDocumentTypes godoc
// @Summary Список типов документов
// @Tags catalog
// @Security BearerAuth
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/document-types [get]
func swaggerSEDListDocumentTypes() {}

// CreateDocumentType godoc
// @Summary Создать тип документа
// @Tags catalog
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/document-types [post]
func swaggerSEDCreateDocumentType() {}

// GetDocumentType godoc
// @Summary Тип документа по ID
// @Tags catalog
// @Security BearerAuth
// @Produce json
// @Param id path string true "UUID типа"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/document-types/{id} [get]
func swaggerSEDGetDocumentType() {}

// UpdateDocumentType godoc
// @Summary Обновить тип документа
// @Tags catalog
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "UUID типа"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/document-types/{id} [put]
func swaggerSEDUpdateDocumentType() {}

// DeleteDocumentType godoc
// @Summary Удалить тип документа
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID типа"
// @Success 204
// @Router /api/v1/document-types/{id} [delete]
func swaggerSEDDeleteDocumentType() {}

// ListWorkflows godoc
// @Summary Список маршрутов согласования
// @Tags catalog
// @Security BearerAuth
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/workflows [get]
func swaggerSEDListWorkflows() {}

// CreateWorkflow godoc
// @Summary Создать маршрут согласования
// @Tags catalog
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/workflows [post]
func swaggerSEDCreateWorkflow() {}

// UpdateWorkflow godoc
// @Summary Обновить маршрут
// @Tags catalog
// @Security BearerAuth
// @Accept json
// @Param id path string true "UUID маршрута"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflows/{id} [put]
func swaggerSEDUpdateWorkflow() {}

// DeleteWorkflow godoc
// @Summary Удалить маршрут
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID маршрута"
// @Success 204
// @Router /api/v1/workflows/{id} [delete]
func swaggerSEDDeleteWorkflow() {}

// ListWorkflowSteps godoc
// @Summary Шаги маршрута
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID маршрута"
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/workflows/{id}/steps [get]
func swaggerSEDListWorkflowSteps() {}

// AddWorkflowStep godoc
// @Summary Добавить шаг маршрута
// @Tags catalog
// @Security BearerAuth
// @Accept json
// @Param id path string true "UUID маршрута"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/workflows/{id}/steps [post]
func swaggerSEDAddWorkflowStep() {}

// DeleteWorkflowStep godoc
// @Summary Удалить шаг маршрута
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID шага"
// @Success 204
// @Router /api/v1/workflow-steps/{id} [delete]
func swaggerSEDDeleteWorkflowStep() {}

// ListDocuments godoc
// @Summary Список документов
// @Tags documents
// @Security BearerAuth
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/documents [get]
func swaggerSEDListDocuments() {}

// GetDocument godoc
// @Summary Документ по ID
// @Tags documents
// @Security BearerAuth
// @Param id path string true "UUID документа"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/documents/{id} [get]
func swaggerSEDGetDocument() {}

// DocumentHistory godoc
// @Summary История документа
// @Tags documents
// @Security BearerAuth
// @Param id path string true "UUID документа"
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/documents/{id}/history [get]
func swaggerSEDDocumentHistory() {}

// ListDocumentFiles godoc
// @Summary Файлы документа
// @Tags files
// @Security BearerAuth
// @Param id path string true "UUID документа"
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/documents/{id}/files [get]
func swaggerSEDListDocumentFiles() {}

// DownloadDocumentFile godoc
// @Summary Скачать файл документа
// @Tags files
// @Security BearerAuth
// @Param id path string true "UUID документа"
// @Param file_id path string true "UUID файла"
// @Success 200 {file} file
// @Router /api/v1/documents/{id}/files/{file_id} [get]
func swaggerSEDDownloadDocumentFile() {}

// CreateDocument godoc
// @Summary Создать документ
// @Tags documents
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/documents [post]
func swaggerSEDCreateDocument() {}

// PatchDocument godoc
// @Summary Изменить черновик документа
// @Tags documents
// @Security BearerAuth
// @Accept json
// @Param id path string true "UUID документа"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/documents/{id} [patch]
func swaggerSEDPatchDocument() {}

// SubmitDocument godoc
// @Summary Отправить документ на согласование
// @Tags documents
// @Security BearerAuth
// @Param id path string true "UUID документа"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/documents/{id}/submit [post]
func swaggerSEDSubmitDocument() {}

// SignDocument godoc
// @Summary Подписать документ
// @Tags documents
// @Security BearerAuth
// @Param id path string true "UUID документа"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/documents/{id}/sign [post]
func swaggerSEDSignDocument() {}

// CancelDocument godoc
// @Summary Отменить документ
// @Tags documents
// @Security BearerAuth
// @Param id path string true "UUID документа"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/documents/{id}/cancel [post]
func swaggerSEDCancelDocument() {}

// UploadDocumentFile godoc
// @Summary Загрузить файл к документу
// @Tags files
// @Security BearerAuth
// @Accept multipart/form-data
// @Param id path string true "UUID документа"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/documents/{id}/files [post]
func swaggerSEDUploadDocumentFile() {}

// DeleteDocumentFile godoc
// @Summary Удалить файл документа
// @Tags files
// @Security BearerAuth
// @Param id path string true "UUID документа"
// @Param file_id path string true "UUID файла"
// @Success 204
// @Router /api/v1/documents/{id}/files/{file_id} [delete]
func swaggerSEDDeleteDocumentFile() {}

// ListTasks godoc
// @Summary Задачи согласования
// @Tags tasks
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/tasks [get]
func swaggerSEDListTasks() {}

// ApproveDocument godoc
// @Summary Согласовать документ
// @Tags tasks
// @Security BearerAuth
// @Param id path string true "UUID документа"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/documents/{id}/approve [post]
func swaggerSEDApproveDocument() {}

// RejectDocument godoc
// @Summary Отклонить документ
// @Tags tasks
// @Security BearerAuth
// @Param id path string true "UUID документа"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/documents/{id}/reject [post]
func swaggerSEDRejectDocument() {}

// GetRagCorpus godoc
// @Summary Корпус документов для RAG (internal)
// @Tags internal-rag
// @Produce json
// @Param X-Service-Secret header string true "служебный секрет"
// @Param X-Tenant-Id header string true "код тенанта"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/internal/rag/corpus [get]
func swaggerSEDGetRagCorpus() {}

// RagDownloadFile godoc
// @Summary Скачать файл для RAG (internal)
// @Tags internal-rag
// @Param X-Service-Secret header string true "служебный секрет"
// @Param X-Tenant-Id header string true "код тенанта"
// @Param id path string true "UUID документа"
// @Param file_id path string true "UUID файла"
// @Success 200 {file} file
// @Router /api/v1/internal/rag/documents/{id}/files/{file_id} [get]
func swaggerSEDRagDownloadFile() {}

// UpdateDocumentFixture godoc
// @Summary Обновить fixture документа (internal)
// @Tags internal-rag
// @Accept json
// @Param X-Service-Secret header string true "служебный секрет"
// @Param X-Tenant-Id header string true "код тенанта"
// @Param id path string true "UUID документа"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/internal/rag/documents/{id}/fixture [put]
func swaggerSEDUpdateDocumentFixture() {}

// SetDocumentRagContent godoc
// @Summary Задать rag_content документа (internal)
// @Tags internal-rag
// @Accept json
// @Param X-Service-Secret header string true "служебный секрет"
// @Param X-Tenant-Id header string true "код тенанта"
// @Param id path string true "UUID документа"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/internal/rag/documents/{id}/content [put]
func swaggerSEDSetDocumentRagContent() {}
