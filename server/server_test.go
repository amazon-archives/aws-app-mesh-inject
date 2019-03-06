package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/awslabs/aws-app-mesh-inject/config"
	"k8s.io/api/admission/v1beta1"
)

func TestAdmissionResponseError(t *testing.T) {
	err := errors.New("test error")
	ar := admissionResponseError(err)
	if ar.Result.Message != err.Error() {
		t.Fatal("Expected error and admission response message to match")
	}
}

func TestServerHandler(t *testing.T) {
	s := AppMeshHandler{config.Config{}}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/healthz", nil)
	s.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatal("Expecting 200 for healthz check")
	}
	if w.Body == nil {
		t.Fatal("Expecting body for healthz check")
	}
	data, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data, healthResponse) {
		t.Fatal("Expected health response body")
	}
	r = httptest.NewRequest("POST", "/", nil)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatal("Expecting 415 for wrong content type")
	}
	if w.Body == nil {
		t.Fatal("Expecting body for wrong content type")
	}
	data, err = ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data, wrongContentResponse) {
		t.Fatal("Expected wrong content type response body")
	}
	r.Header.Add("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.ServeHTTP(w, r)
	rarobject := v1beta1.AdmissionReview{}
	data, err = ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = Decode(data, &rarobject)
	if err != nil {
		t.Fatal(err)
	}
	if rarobject.Response.Result.Message != ErrNoObject.Error() {
		t.Fatal("Expected no object error")
	}
	ar := v1beta1.AdmissionRequest{UID: ""}
	sar := v1beta1.AdmissionReview{
		Request: &ar,
	}
	sarbytes, err := json.Marshal(sar)
	if err != nil {
		t.Fatal(err)
	}
	sarbytesreader := bytes.NewReader(sarbytes)
	r = httptest.NewRequest("POST", "/", sarbytesreader)
	r.Header.Add("Content-Type", "application/json")
	s.ServeHTTP(w, r)
	data, err = ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	raruid := v1beta1.AdmissionReview{}
	err = Decode(data, &raruid)
	if err != nil {
		t.Fatal(err)
	}
	if raruid.Response.Result.Message != ErrNoUID.Error() {
		fmt.Println(raruid.Response.Result.Message)
		t.Fatal("Expected no uid error")
	}
}

func TestMutate(t *testing.T) {

}

func TestNewServer(t *testing.T) {
	_, err := NewServer(config.Config{})
	if err == nil {
		t.Fatal("expected pem error")
	}
}
