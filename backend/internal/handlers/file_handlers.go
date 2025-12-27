package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/sanitize"
	"react-golang-starter/internal/services"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// AllowedMimeTypes defines the MIME types allowed for file uploads
var AllowedMimeTypes = map[string]bool{
	// Images
	"image/jpeg":    true,
	"image/png":     true,
	"image/gif":     true,
	"image/webp":    true,
	"image/svg+xml": true,
	// Documents
	"application/pdf":    true,
	"text/plain":         true,
	"text/csv":           true,
	"application/json":   true,
	"application/xml":    true,
	"text/xml":           true,
	"text/markdown":      true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
}

// isAllowedMimeType checks if the given MIME type is in the allowed list
func isAllowedMimeType(mimeType string) bool {
	// Normalize MIME type (remove parameters like charset)
	normalizedType := strings.Split(mimeType, ";")[0]
	normalizedType = strings.TrimSpace(strings.ToLower(normalizedType))
	return AllowedMimeTypes[normalizedType]
}

// FileHandler handles file-related HTTP requests
type FileHandler struct {
	fileService *services.FileService
}

// NewFileHandler creates a new file handler instance
func NewFileHandler(fileService *services.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// UploadFile handles file upload requests
// @Summary Upload a file
// @Description Upload a file to the server. Uses S3 if configured, otherwise stores in database.
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 200 {object} models.SuccessResponse{data=models.FileResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /files/upload [post]
func (fh *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Failed to parse multipart form",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Failed to get file from form",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer file.Close()

	// Validate file size (optional - you can set your own limits)
	if header.Size > 10<<20 { // 10MB limit
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: "File size exceeds 10MB limit",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate file type using magic bytes (content sniffing)
	// Read first 512 bytes to detect actual content type
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Failed to read file content",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Detect actual content type from file content (magic bytes)
	detectedType := http.DetectContentType(buf[:n])

	// Reset file reader for subsequent operations
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to process file",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get claimed content type from header
	claimedType := header.Header.Get("Content-Type")
	if claimedType == "" {
		claimedType = detectedType
	}

	// Use detected type for validation (more secure than trusting headers)
	if !isAllowedMimeType(detectedType) {
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: fmt.Sprintf("File type '%s' is not allowed. Detected type: '%s'. Allowed types: images (jpeg, png, gif, webp, svg), documents (pdf, txt, csv, json, xml, md, docx, xlsx, pptx)", claimedType, detectedType),
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Sanitize filename to prevent path traversal and other attacks
	sanitizedFilename := sanitize.Filename(header.Filename)
	if sanitizedFilename == "" || sanitizedFilename == "unnamed" {
		sanitizedFilename = fmt.Sprintf("file_%d", header.Size)
	}
	header.Filename = sanitizedFilename

	// Upload file using service
	uploadedFile, err := fh.fileService.UploadFile(r.Context(), file, header)
	if err != nil {
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: fmt.Sprintf("Failed to upload file: %v", err),
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.SuccessResponse{
		Success: true,
		Message: "File uploaded successfully",
		Data:    uploadedFile.ToFileResponse(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DownloadFile handles file download requests
// @Summary Download a file
// @Description Download a file by its ID. For S3 files, redirects to the S3 URL.
// @Tags files
// @Produce octet-stream
// @Param id path int true "File ID"
// @Success 200 {file} binary
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /files/{id}/download [get]
func (fh *FileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	fileIDStr := chi.URLParam(r, "id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid file ID",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check ownership - get user from context
	userID, hasUser := auth.GetUserIDFromContext(r.Context())
	user, _ := auth.GetUserFromContext(r.Context())
	isAdmin := hasUser && user != nil && (user.Role == models.RoleAdmin || user.Role == models.RoleSuperAdmin)

	// Verify ownership before allowing download
	_, err = fh.fileService.GetFileByIDForUser(uint(fileID), userID, isAdmin)
	if err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			response := models.ErrorResponse{
				Error:   "Forbidden",
				Message: "You do not have permission to access this file",
				Code:    http.StatusForbidden,
			}
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response)
			return
		}
		response := models.ErrorResponse{
			Error:   "Not Found",
			Message: "File not found",
			Code:    http.StatusNotFound,
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get file content and metadata
	content, file, err := fh.fileService.DownloadFile(r.Context(), uint(fileID))
	if err != nil {
		// If it's an S3 file, redirect to the URL
		if file != nil && file.StorageType == "s3" {
			url, urlErr := fh.fileService.GetFileURL(r.Context(), uint(fileID))
			if urlErr != nil {
				response := models.ErrorResponse{
					Error:   "Internal Server Error",
					Message: "Failed to get file URL",
					Code:    http.StatusInternalServerError,
				}
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			return
		}
		response := models.ErrorResponse{
			Error:   "Not Found",
			Message: "File not found",
			Code:    http.StatusNotFound,
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.FileName))
	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(file.FileSize, 10))

	// Write file content
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(content); err != nil {
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to write file content",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
}

// GetFileInfo handles requests for file information
// @Summary Get file information
// @Description Get metadata for a file by its ID
// @Tags files
// @Produce json
// @Param id path int true "File ID"
// @Success 200 {object} models.SuccessResponse{data=models.FileResponse}
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /files/{id} [get]
func (fh *FileHandler) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	fileIDStr := chi.URLParam(r, "id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid file ID",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check ownership - get user from context
	userID, hasUser := auth.GetUserIDFromContext(r.Context())
	user, _ := auth.GetUserFromContext(r.Context())
	isAdmin := hasUser && user != nil && (user.Role == models.RoleAdmin || user.Role == models.RoleSuperAdmin)

	file, err := fh.fileService.GetFileByIDForUser(uint(fileID), userID, isAdmin)
	if err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			response := models.ErrorResponse{
				Error:   "Forbidden",
				Message: "You do not have permission to access this file",
				Code:    http.StatusForbidden,
			}
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response)
			return
		}
		response := models.ErrorResponse{
			Error:   "Not Found",
			Message: "File not found",
			Code:    http.StatusNotFound,
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.SuccessResponse{
		Success: true,
		Message: "File information retrieved successfully",
		Data:    file.ToFileResponse(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetFileURL handles requests for file URLs
// @Summary Get file URL
// @Description Get the URL for accessing a file
// @Tags files
// @Produce json
// @Param id path int true "File ID"
// @Success 200 {object} models.SuccessResponse{data=string}
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /files/{id}/url [get]
func (fh *FileHandler) GetFileURL(w http.ResponseWriter, r *http.Request) {
	fileIDStr := chi.URLParam(r, "id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid file ID",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check ownership - get user from context
	userID, hasUser := auth.GetUserIDFromContext(r.Context())
	user, _ := auth.GetUserFromContext(r.Context())
	isAdmin := hasUser && user != nil && (user.Role == models.RoleAdmin || user.Role == models.RoleSuperAdmin)

	// Verify ownership before allowing access
	_, err = fh.fileService.GetFileByIDForUser(uint(fileID), userID, isAdmin)
	if err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			response := models.ErrorResponse{
				Error:   "Forbidden",
				Message: "You do not have permission to access this file",
				Code:    http.StatusForbidden,
			}
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response)
			return
		}
		response := models.ErrorResponse{
			Error:   "Not Found",
			Message: "File not found",
			Code:    http.StatusNotFound,
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	url, err := fh.fileService.GetFileURL(r.Context(), uint(fileID))
	if err != nil {
		response := models.ErrorResponse{
			Error:   "Not Found",
			Message: "File not found",
			Code:    http.StatusNotFound,
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.SuccessResponse{
		Success: true,
		Message: "File URL retrieved successfully",
		Data:    url,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteFile handles file deletion requests
// @Summary Delete a file
// @Description Delete a file by its ID
// @Tags files
// @Produce json
// @Param id path int true "File ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /files/{id} [delete]
func (fh *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	fileIDStr := chi.URLParam(r, "id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid file ID",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check ownership - get user from context
	userID, hasUser := auth.GetUserIDFromContext(r.Context())
	user, _ := auth.GetUserFromContext(r.Context())
	isAdmin := hasUser && user != nil && (user.Role == models.RoleAdmin || user.Role == models.RoleSuperAdmin)

	// Verify ownership before allowing deletion
	_, err = fh.fileService.GetFileByIDForUser(uint(fileID), userID, isAdmin)
	if err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			response := models.ErrorResponse{
				Error:   "Forbidden",
				Message: "You do not have permission to delete this file",
				Code:    http.StatusForbidden,
			}
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response)
			return
		}
		response := models.ErrorResponse{
			Error:   "Not Found",
			Message: "File not found",
			Code:    http.StatusNotFound,
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	err = fh.fileService.DeleteFile(r.Context(), uint(fileID))
	if err != nil {
		response := models.ErrorResponse{
			Error:   "Not Found",
			Message: "File not found",
			Code:    http.StatusNotFound,
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.SuccessResponse{
		Success: true,
		Message: "File deleted successfully",
		Data:    nil,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ListFiles handles requests for listing files
// @Summary List files
// @Description Get a list of uploaded files with pagination. Users see only their own files unless admin.
// @Tags files
// @Produce json
// @Param limit query int false "Number of files to return (default: 10)"
// @Param offset query int false "Number of files to skip (default: 0)"
// @Success 200 {object} models.SuccessResponse{data=[]models.FileResponse}
// @Failure 500 {object} models.ErrorResponse
// @Router /files [get]
func (fh *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default
	offset := 0 // default

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get user from context for ownership filtering
	userID, hasUser := auth.GetUserIDFromContext(r.Context())
	user, _ := auth.GetUserFromContext(r.Context())
	isAdmin := hasUser && user != nil && (user.Role == models.RoleAdmin || user.Role == models.RoleSuperAdmin)

	files, err := fh.fileService.ListFilesForUser(userID, isAdmin, limit, offset)
	if err != nil {
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to list files",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	var fileResponses []models.FileResponse
	for _, file := range files {
		fileResponses = append(fileResponses, file.ToFileResponse())
	}

	response := models.SuccessResponse{
		Success: true,
		Message: "Files retrieved successfully",
		Data:    fileResponses,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetStorageStatus handles requests for storage status
// @Summary Get storage status
// @Description Get information about the current storage configuration
// @Tags files
// @Produce json
// @Success 200 {object} models.SuccessResponse{data=map[string]interface{}}
// @Router /files/storage/status [get]
func (fh *FileHandler) GetStorageStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"storage_type": fh.fileService.GetStorageType(),
		"message":      "Storage status retrieved successfully",
	}

	response := models.SuccessResponse{
		Success: true,
		Message: "Storage status retrieved successfully",
		Data:    status,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
