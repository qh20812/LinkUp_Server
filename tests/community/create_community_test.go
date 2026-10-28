package community_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	errorsapp "linkup/errors"
	"linkup/controllers"
	"linkup/models"
	"linkup/services"
)

// ─── Fakes ────────────────────────────────────────────────────────────

type fakeMediaService struct {
	services.MediaService // embedded nil: unimplemented methods panic if called
	uploadFn func(ctx context.Context, userID string, file *multipart.FileHeader) (*models.Media, error)
	deleted  []string
}

func (f *fakeMediaService) AutoApproveUpload(ctx context.Context, userID string, file *multipart.FileHeader) (*models.Media, error) {
	return f.uploadFn(ctx, userID, file)
}

func (f *fakeMediaService) DeleteMedia(ctx context.Context, userID, mediaID string) error {
	f.deleted = append(f.deleted, mediaID)
	return nil
}

func okMedia(id string) *models.Media {
	return &models.Media{ID: id, UserID: "user-1", FileURI: "https://cdn.example.com/" + id + ".png", Status: models.MediaStatusApproved}
}

type fakeCommunityService struct {
	controllers.CommunityService // embedded nil: only CreateCommunity is stubbed
	called       bool
	gotPrivacy   models.CommunityPrivacy
	gotName      string
	createFn     func(ctx context.Context, creatorID, name, description, avatarURI, backgroundURI, avatarMediaID, backgroundMediaID string, privacy models.CommunityPrivacy, autoApprove bool) (*services.CreateCommunityResult, error)
}

func (f *fakeCommunityService) CreateCommunity(ctx context.Context, creatorID, name, description, avatarURI, backgroundURI, avatarMediaID, backgroundMediaID string, privacy models.CommunityPrivacy, autoApprove bool) (*services.CreateCommunityResult, error) {
	f.called = true
	f.gotName = name
	f.gotPrivacy = privacy
	return f.createFn(ctx, creatorID, name, description, avatarURI, backgroundURI, avatarMediaID, backgroundMediaID, privacy, autoApprove)
}

func okCommunityResult() (*services.CreateCommunityResult, error) {
	return &services.CreateCommunityResult{
		Community:  &models.Community{ID: "community-1", CreatorID: "user-1", Name: "Crew", Privacy: models.PrivacyPublic},
		Chat:       &models.Chat{ID: "chat-1", Name: "Crew"},
		InviteCode: "ABC123",
	}, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────

func newCreateRequest(t *testing.T, fields map[string]string, files map[string][]byte) *http.Request {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}
	for name, content := range files {
		fw, err := w.CreateFormFile(name, name+".png")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := fw.Write(content); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/communities", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func serveCreate(communitySvc controllers.CommunityService, mediaSvc services.MediaService, req *http.Request) *httptest.ResponseRecorder {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "user-1")
		c.Next()
	})
	ctrl := controllers.NewCommunityController(communitySvc, mediaSvc)
	router.POST("/api/communities", ctrl.CreateCommunity)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body %q: %v", w.Body.String(), err)
	}
	return body
}

// ─── Tests ────────────────────────────────────────────────────────────

func TestCreateCommunity_Success201(t *testing.T) {
	communitySvc := &fakeCommunityService{
		createFn: func(ctx context.Context, creatorID, name, description, avatarURI, backgroundURI, avatarMediaID, backgroundMediaID string, privacy models.CommunityPrivacy, autoApprove bool) (*services.CreateCommunityResult, error) {
			return okCommunityResult()
		},
	}
	mediaSvc := &fakeMediaService{
		uploadFn: func(ctx context.Context, userID string, file *multipart.FileHeader) (*models.Media, error) {
			t.Error("no files sent, AutoApproveUpload must not be called")
			return nil, errors.New("unexpected upload")
		},
	}

	req := newCreateRequest(t, map[string]string{
		"name": "Crew", "description": "Desc",
		"privacy": "invitation_only", "auto_approve": "true",
	}, nil)
	w := serveCreate(communitySvc, mediaSvc, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["community_id"] != "community-1" {
		t.Errorf("community_id = %v", body["community_id"])
	}
	community, ok := body["community"].(map[string]any)
	if !ok || community["name"] != "Crew" {
		t.Errorf("community payload = %v", body["community"])
	}
	chat, ok := body["default_group_chat"].(map[string]any)
	if !ok || chat["id"] != "chat-1" {
		t.Errorf("default_group_chat = %v", body["default_group_chat"])
	}
	if body["invite_code"] != "ABC123" {
		t.Errorf("invite_code = %v, want starter code", body["invite_code"])
	}
	if communitySvc.gotPrivacy != models.PrivacyInvitationOnly {
		t.Errorf("privacy passed to service = %q, want invitation_only", communitySvc.gotPrivacy)
	}
}

func TestCreateCommunity_PrivacyTypoRejected(t *testing.T) {
	communitySvc := &fakeCommunityService{
		createFn: func(ctx context.Context, creatorID, name, description, avatarURI, backgroundURI, avatarMediaID, backgroundMediaID string, privacy models.CommunityPrivacy, autoApprove bool) (*services.CreateCommunityResult, error) {
			// Mirrors service validation: no silent fallback to public.
			if privacy != "" && privacy != models.PrivacyPublic && privacy != models.PrivacyInvitationOnly {
				return nil, errorsapp.New(errorsapp.ErrCodeCommunityPrivacyInvalid)
			}
			return okCommunityResult()
		},
	}
	mediaSvc := &fakeMediaService{}

	req := newCreateRequest(t, map[string]string{"name": "Crew", "privacy": "publik"}, nil)
	w := serveCreate(communitySvc, mediaSvc, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("community.PRIVACY_INVALID")) {
		t.Fatalf("body = %q, want PRIVACY_INVALID", w.Body.String())
	}
	if communitySvc.gotPrivacy == models.PrivacyPublic {
		t.Error("controller silently coerced invalid privacy to public")
	}
}

func TestCreateCommunity_UploadErrorPreserved(t *testing.T) {
	communitySvc := &fakeCommunityService{
		createFn: func(ctx context.Context, creatorID, name, description, avatarURI, backgroundURI, avatarMediaID, backgroundMediaID string, privacy models.CommunityPrivacy, autoApprove bool) (*services.CreateCommunityResult, error) {
			t.Error("upload failed, service must not be called")
			return okCommunityResult()
		},
	}
	mediaSvc := &fakeMediaService{
		uploadFn: func(ctx context.Context, userID string, file *multipart.FileHeader) (*models.Media, error) {
			return nil, errorsapp.New(errorsapp.ErrCodeMediaFileTooLarge)
		},
	}

	req := newCreateRequest(t, map[string]string{"name": "Crew"},
		map[string][]byte{"avatar": []byte("fake-image")})
	w := serveCreate(communitySvc, mediaSvc, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("media.FILE_TOO_LARGE")) {
		t.Fatalf("body = %q, want original media error, not masked", w.Body.String())
	}
	if communitySvc.called {
		t.Error("service called despite upload failure")
	}
}

func TestCreateCommunity_CleanupOnServiceFailure(t *testing.T) {
	communitySvc := &fakeCommunityService{
		createFn: func(ctx context.Context, creatorID, name, description, avatarURI, backgroundURI, avatarMediaID, backgroundMediaID string, privacy models.CommunityPrivacy, autoApprove bool) (*services.CreateCommunityResult, error) {
			if avatarURI == "" || backgroundURI == "" {
				t.Errorf("expected both URIs, got avatar=%q background=%q", avatarURI, backgroundURI)
			}
			return nil, errorsapp.New(errorsapp.ErrCodeCommunityNameExists)
		},
	}
	mediaSvc := &fakeMediaService{
		uploadFn: func(ctx context.Context, userID string, file *multipart.FileHeader) (*models.Media, error) {
			return okMedia("media-" + file.Filename), nil
		},
	}

	req := newCreateRequest(t, map[string]string{"name": "Crew"},
		map[string][]byte{"avatar": []byte("a"), "background": []byte("b")})
	w := serveCreate(communitySvc, mediaSvc, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusConflict, w.Body.String())
	}
	if len(mediaSvc.deleted) != 2 {
		t.Fatalf("deleted = %v, want both uploaded media cleaned up", mediaSvc.deleted)
	}
}

func TestCreateCommunity_CleanupOnSecondUploadFailure(t *testing.T) {
	communitySvc := &fakeCommunityService{
		createFn: func(ctx context.Context, creatorID, name, description, avatarURI, backgroundURI, avatarMediaID, backgroundMediaID string, privacy models.CommunityPrivacy, autoApprove bool) (*services.CreateCommunityResult, error) {
			t.Error("service must not be called when background upload fails")
			return okCommunityResult()
		},
	}
	mediaSvc := &fakeMediaService{
		uploadFn: func(ctx context.Context, userID string, file *multipart.FileHeader) (*models.Media, error) {
			if strings.Contains(file.Filename, "background") {
				return nil, errors.New("cloudinary down")
			}
			return okMedia("media-avatar"), nil
		},
	}

	req := newCreateRequest(t, map[string]string{"name": "Crew"},
		map[string][]byte{"avatar": []byte("a"), "background": []byte("b")})
	w := serveCreate(communitySvc, mediaSvc, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusInternalServerError, w.Body.String())
	}
	if len(mediaSvc.deleted) != 1 || mediaSvc.deleted[0] != "media-avatar" {
		t.Fatalf("deleted = %v, want avatar cleaned up", mediaSvc.deleted)
	}
}

func TestCreateCommunity_BodyTooLarge(t *testing.T) {
	communitySvc := &fakeCommunityService{}
	mediaSvc := &fakeMediaService{}

	big := bytes.Repeat([]byte("x"), 46<<20)
	req := newCreateRequest(t, map[string]string{"name": "Crew"},
		map[string][]byte{"avatar": big})
	w := serveCreate(communitySvc, mediaSvc, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("media.FILE_TOO_LARGE")) {
		t.Fatalf("body = %q, want FILE_TOO_LARGE", w.Body.String())
	}
	if communitySvc.called {
		t.Error("service called despite oversized body")
	}
}

func TestCreateCommunity_PlainErrorNoLeak(t *testing.T) {
	communitySvc := &fakeCommunityService{
		createFn: func(ctx context.Context, creatorID, name, description, avatarURI, backgroundURI, avatarMediaID, backgroundMediaID string, privacy models.CommunityPrivacy, autoApprove bool) (*services.CreateCommunityResult, error) {
			return nil, errors.New("sql: connection exploded SECRET")
		},
	}
	mediaSvc := &fakeMediaService{}

	req := newCreateRequest(t, map[string]string{"name": "Crew"}, nil)
	w := serveCreate(communitySvc, mediaSvc, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusInternalServerError, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("common.INTERNAL_ERROR")) {
		t.Fatalf("body = %q, want generic internal error", w.Body.String())
	}
	if bytes.Contains(w.Body.Bytes(), []byte("SECRET")) {
		t.Fatalf("body = %q, leaks internal detail", w.Body.String())
	}
}
