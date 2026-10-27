package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	errorsapp "linkup/errors"
	"linkup/services"
)

// fakePostService chỉ override 2 method cần test; phần còn lại của
// services.PostService được nhúng để không phải implement cả interface.
type fakePostService struct {
	services.PostService

	hideErr   error
	unhideErr error

	hiddenUser   string
	hiddenPost   string
	unhiddenUser string
	unhiddenPost string
}

func (f *fakePostService) HidePost(_ context.Context, userID, postID string) error {
	f.hiddenUser, f.hiddenPost = userID, postID
	return f.hideErr
}

func (f *fakePostService) UnhidePost(_ context.Context, userID, postID string) error {
	f.unhiddenUser, f.unhiddenPost = userID, postID
	return f.unhideErr
}

func newPostParamContext(method, postID, userID string) (*gin.Context, *httptest.ResponseRecorder) {
	c, w := newTestContext(method)
	c.Params = gin.Params{{Key: "id", Value: postID}}
	if userID != "" {
		c.Set("userID", userID)
	}
	return c, w
}

func TestHidePost_Success(t *testing.T) {
	svc := &fakePostService{}
	ctrl := NewPostController(svc)

	c, w := newPostParamContext(http.MethodPost, "post-1", "user-1")
	ctrl.HidePost(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	if svc.hiddenUser != "user-1" || svc.hiddenPost != "post-1" {
		t.Fatalf("unexpected service call: user=%q post=%q", svc.hiddenUser, svc.hiddenPost)
	}
	if body := decodeBody(t, w); body["hidden"] != true {
		t.Fatalf("expected hidden=true, got %v", body["hidden"])
	}
}

func TestHidePost_MissingUserID_NoPanic(t *testing.T) {
	ctrl := NewPostController(&fakePostService{})

	c, w := newPostParamContext(http.MethodPost, "post-1", "")
	ctrl.HidePost(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if body := decodeBody(t, w); body["error"] != errorsapp.ErrCodeInternal {
		t.Fatalf("expected error code %q, got %v", errorsapp.ErrCodeInternal, body["error"])
	}
}

func TestHidePost_MissingPostID(t *testing.T) {
	ctrl := NewPostController(&fakePostService{})

	c, w := newPostParamContext(http.MethodPost, "", "user-1")
	ctrl.HidePost(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHidePost_CannotHideOwnMapsTo400(t *testing.T) {
	svc := &fakePostService{hideErr: errorsapp.New(errorsapp.ErrCodePostCannotHideOwn)}
	ctrl := NewPostController(svc)

	c, w := newPostParamContext(http.MethodPost, "post-1", "user-1")
	ctrl.HidePost(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (%s)", w.Code, w.Body.String())
	}
	if body := decodeBody(t, w); body["error"] != errorsapp.ErrCodePostCannotHideOwn {
		t.Fatalf("expected error code %q, got %v", errorsapp.ErrCodePostCannotHideOwn, body["error"])
	}
}

func TestHidePost_UnknownErrorMapsTo500(t *testing.T) {
	svc := &fakePostService{hideErr: errors.New("db down")}
	ctrl := NewPostController(svc)

	c, w := newPostParamContext(http.MethodPost, "post-1", "user-1")
	ctrl.HidePost(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUnhidePost_Success(t *testing.T) {
	svc := &fakePostService{}
	ctrl := NewPostController(svc)

	c, w := newPostParamContext(http.MethodDelete, "post-2", "user-2")
	ctrl.UnhidePost(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	if svc.unhiddenUser != "user-2" || svc.unhiddenPost != "post-2" {
		t.Fatalf("unexpected service call: user=%q post=%q", svc.unhiddenUser, svc.unhiddenPost)
	}
	if body := decodeBody(t, w); body["hidden"] != false {
		t.Fatalf("expected hidden=false, got %v", body["hidden"])
	}
}

func TestUnhidePost_NotFoundMapsTo404(t *testing.T) {
	svc := &fakePostService{unhideErr: errorsapp.New(errorsapp.ErrCodePostNotFound)}
	ctrl := NewPostController(svc)

	c, w := newPostParamContext(http.MethodDelete, "post-2", "user-2")
	ctrl.UnhidePost(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d (%s)", w.Code, w.Body.String())
	}
}
