// Code generated by pegomock. DO NOT EDIT.
// Source: github.com/petergtz/bitsgo (interfaces: Blobstore)

package bitsgo_test

import (
	pegomock "github.com/petergtz/pegomock"
	io "io"
	"reflect"
)

type MockBlobstore struct {
	fail func(message string, callerSkip ...int)
}

func NewMockBlobstore() *MockBlobstore {
	return &MockBlobstore{fail: pegomock.GlobalFailHandler}
}

func (mock *MockBlobstore) Exists(path string) (bool, error) {
	if mock == nil {
		panic("mock must not be nil. Use myMock := NewMockMockBlobstore().")
	}
	params := []pegomock.Param{path}
	result := pegomock.GetGenericMockFrom(mock).Invoke("Exists", params, []reflect.Type{reflect.TypeOf((*bool)(nil)).Elem(), reflect.TypeOf((*error)(nil)).Elem()})
	var ret0 bool
	var ret1 error
	if len(result) != 0 {
		if result[0] != nil {
			ret0 = result[0].(bool)
		}
		if result[1] != nil {
			ret1 = result[1].(error)
		}
	}
	return ret0, ret1
}

func (mock *MockBlobstore) HeadOrRedirectAsGet(path string) (string, error) {
	if mock == nil {
		panic("mock must not be nil. Use myMock := NewMockMockBlobstore().")
	}
	params := []pegomock.Param{path}
	result := pegomock.GetGenericMockFrom(mock).Invoke("HeadOrRedirectAsGet", params, []reflect.Type{reflect.TypeOf((*string)(nil)).Elem(), reflect.TypeOf((*error)(nil)).Elem()})
	var ret0 string
	var ret1 error
	if len(result) != 0 {
		if result[0] != nil {
			ret0 = result[0].(string)
		}
		if result[1] != nil {
			ret1 = result[1].(error)
		}
	}
	return ret0, ret1
}

func (mock *MockBlobstore) GetOrRedirect(path string) (io.ReadCloser, string, error) {
	if mock == nil {
		panic("mock must not be nil. Use myMock := NewMockMockBlobstore().")
	}
	params := []pegomock.Param{path}
	result := pegomock.GetGenericMockFrom(mock).Invoke("GetOrRedirect", params, []reflect.Type{reflect.TypeOf((*io.ReadCloser)(nil)).Elem(), reflect.TypeOf((*string)(nil)).Elem(), reflect.TypeOf((*error)(nil)).Elem()})
	var ret0 io.ReadCloser
	var ret1 string
	var ret2 error
	if len(result) != 0 {
		if result[0] != nil {
			ret0 = result[0].(io.ReadCloser)
		}
		if result[1] != nil {
			ret1 = result[1].(string)
		}
		if result[2] != nil {
			ret2 = result[2].(error)
		}
	}
	return ret0, ret1, ret2
}

func (mock *MockBlobstore) Get(path string) (io.ReadCloser, error) {
	if mock == nil {
		panic("mock must not be nil. Use myMock := NewMockMockBlobstore().")
	}
	params := []pegomock.Param{path}
	result := pegomock.GetGenericMockFrom(mock).Invoke("Get", params, []reflect.Type{reflect.TypeOf((*io.ReadCloser)(nil)).Elem(), reflect.TypeOf((*error)(nil)).Elem()})
	var ret0 io.ReadCloser
	var ret1 error
	if len(result) != 0 {
		if result[0] != nil {
			ret0 = result[0].(io.ReadCloser)
		}
		if result[1] != nil {
			ret1 = result[1].(error)
		}
	}
	return ret0, ret1
}

func (mock *MockBlobstore) Put(path string, src io.ReadSeeker) error {
	if mock == nil {
		panic("mock must not be nil. Use myMock := NewMockMockBlobstore().")
	}
	params := []pegomock.Param{path, src}
	result := pegomock.GetGenericMockFrom(mock).Invoke("Put", params, []reflect.Type{reflect.TypeOf((*error)(nil)).Elem()})
	var ret0 error
	if len(result) != 0 {
		if result[0] != nil {
			ret0 = result[0].(error)
		}
	}
	return ret0
}

func (mock *MockBlobstore) Copy(src string, dest string) error {
	if mock == nil {
		panic("mock must not be nil. Use myMock := NewMockMockBlobstore().")
	}
	params := []pegomock.Param{src, dest}
	result := pegomock.GetGenericMockFrom(mock).Invoke("Copy", params, []reflect.Type{reflect.TypeOf((*error)(nil)).Elem()})
	var ret0 error
	if len(result) != 0 {
		if result[0] != nil {
			ret0 = result[0].(error)
		}
	}
	return ret0
}

func (mock *MockBlobstore) Delete(path string) error {
	if mock == nil {
		panic("mock must not be nil. Use myMock := NewMockMockBlobstore().")
	}
	params := []pegomock.Param{path}
	result := pegomock.GetGenericMockFrom(mock).Invoke("Delete", params, []reflect.Type{reflect.TypeOf((*error)(nil)).Elem()})
	var ret0 error
	if len(result) != 0 {
		if result[0] != nil {
			ret0 = result[0].(error)
		}
	}
	return ret0
}

func (mock *MockBlobstore) DeleteDir(prefix string) error {
	if mock == nil {
		panic("mock must not be nil. Use myMock := NewMockMockBlobstore().")
	}
	params := []pegomock.Param{prefix}
	result := pegomock.GetGenericMockFrom(mock).Invoke("DeleteDir", params, []reflect.Type{reflect.TypeOf((*error)(nil)).Elem()})
	var ret0 error
	if len(result) != 0 {
		if result[0] != nil {
			ret0 = result[0].(error)
		}
	}
	return ret0
}

func (mock *MockBlobstore) VerifyWasCalledOnce() *VerifierBlobstore {
	return &VerifierBlobstore{mock, pegomock.Times(1), nil}
}

func (mock *MockBlobstore) VerifyWasCalled(invocationCountMatcher pegomock.Matcher) *VerifierBlobstore {
	return &VerifierBlobstore{mock, invocationCountMatcher, nil}
}

func (mock *MockBlobstore) VerifyWasCalledInOrder(invocationCountMatcher pegomock.Matcher, inOrderContext *pegomock.InOrderContext) *VerifierBlobstore {
	return &VerifierBlobstore{mock, invocationCountMatcher, inOrderContext}
}

type VerifierBlobstore struct {
	mock                   *MockBlobstore
	invocationCountMatcher pegomock.Matcher
	inOrderContext         *pegomock.InOrderContext
}

func (verifier *VerifierBlobstore) Exists(path string) *Blobstore_Exists_OngoingVerification {
	params := []pegomock.Param{path}
	methodInvocations := pegomock.GetGenericMockFrom(verifier.mock).Verify(verifier.inOrderContext, verifier.invocationCountMatcher, "Exists", params)
	return &Blobstore_Exists_OngoingVerification{mock: verifier.mock, methodInvocations: methodInvocations}
}

type Blobstore_Exists_OngoingVerification struct {
	mock              *MockBlobstore
	methodInvocations []pegomock.MethodInvocation
}

func (c *Blobstore_Exists_OngoingVerification) GetCapturedArguments() string {
	path := c.GetAllCapturedArguments()
	return path[len(path)-1]
}

func (c *Blobstore_Exists_OngoingVerification) GetAllCapturedArguments() (_param0 []string) {
	params := pegomock.GetGenericMockFrom(c.mock).GetInvocationParams(c.methodInvocations)
	if len(params) > 0 {
		_param0 = make([]string, len(params[0]))
		for u, param := range params[0] {
			_param0[u] = param.(string)
		}
	}
	return
}

func (verifier *VerifierBlobstore) HeadOrRedirectAsGet(path string) *Blobstore_HeadOrRedirectAsGet_OngoingVerification {
	params := []pegomock.Param{path}
	methodInvocations := pegomock.GetGenericMockFrom(verifier.mock).Verify(verifier.inOrderContext, verifier.invocationCountMatcher, "HeadOrRedirectAsGet", params)
	return &Blobstore_HeadOrRedirectAsGet_OngoingVerification{mock: verifier.mock, methodInvocations: methodInvocations}
}

type Blobstore_HeadOrRedirectAsGet_OngoingVerification struct {
	mock              *MockBlobstore
	methodInvocations []pegomock.MethodInvocation
}

func (c *Blobstore_HeadOrRedirectAsGet_OngoingVerification) GetCapturedArguments() string {
	path := c.GetAllCapturedArguments()
	return path[len(path)-1]
}

func (c *Blobstore_HeadOrRedirectAsGet_OngoingVerification) GetAllCapturedArguments() (_param0 []string) {
	params := pegomock.GetGenericMockFrom(c.mock).GetInvocationParams(c.methodInvocations)
	if len(params) > 0 {
		_param0 = make([]string, len(params[0]))
		for u, param := range params[0] {
			_param0[u] = param.(string)
		}
	}
	return
}

func (verifier *VerifierBlobstore) GetOrRedirect(path string) *Blobstore_GetOrRedirect_OngoingVerification {
	params := []pegomock.Param{path}
	methodInvocations := pegomock.GetGenericMockFrom(verifier.mock).Verify(verifier.inOrderContext, verifier.invocationCountMatcher, "GetOrRedirect", params)
	return &Blobstore_GetOrRedirect_OngoingVerification{mock: verifier.mock, methodInvocations: methodInvocations}
}

type Blobstore_GetOrRedirect_OngoingVerification struct {
	mock              *MockBlobstore
	methodInvocations []pegomock.MethodInvocation
}

func (c *Blobstore_GetOrRedirect_OngoingVerification) GetCapturedArguments() string {
	path := c.GetAllCapturedArguments()
	return path[len(path)-1]
}

func (c *Blobstore_GetOrRedirect_OngoingVerification) GetAllCapturedArguments() (_param0 []string) {
	params := pegomock.GetGenericMockFrom(c.mock).GetInvocationParams(c.methodInvocations)
	if len(params) > 0 {
		_param0 = make([]string, len(params[0]))
		for u, param := range params[0] {
			_param0[u] = param.(string)
		}
	}
	return
}

func (verifier *VerifierBlobstore) Get(path string) *Blobstore_Get_OngoingVerification {
	params := []pegomock.Param{path}
	methodInvocations := pegomock.GetGenericMockFrom(verifier.mock).Verify(verifier.inOrderContext, verifier.invocationCountMatcher, "Get", params)
	return &Blobstore_Get_OngoingVerification{mock: verifier.mock, methodInvocations: methodInvocations}
}

type Blobstore_Get_OngoingVerification struct {
	mock              *MockBlobstore
	methodInvocations []pegomock.MethodInvocation
}

func (c *Blobstore_Get_OngoingVerification) GetCapturedArguments() string {
	path := c.GetAllCapturedArguments()
	return path[len(path)-1]
}

func (c *Blobstore_Get_OngoingVerification) GetAllCapturedArguments() (_param0 []string) {
	params := pegomock.GetGenericMockFrom(c.mock).GetInvocationParams(c.methodInvocations)
	if len(params) > 0 {
		_param0 = make([]string, len(params[0]))
		for u, param := range params[0] {
			_param0[u] = param.(string)
		}
	}
	return
}

func (verifier *VerifierBlobstore) Put(path string, src io.ReadSeeker) *Blobstore_Put_OngoingVerification {
	params := []pegomock.Param{path, src}
	methodInvocations := pegomock.GetGenericMockFrom(verifier.mock).Verify(verifier.inOrderContext, verifier.invocationCountMatcher, "Put", params)
	return &Blobstore_Put_OngoingVerification{mock: verifier.mock, methodInvocations: methodInvocations}
}

type Blobstore_Put_OngoingVerification struct {
	mock              *MockBlobstore
	methodInvocations []pegomock.MethodInvocation
}

func (c *Blobstore_Put_OngoingVerification) GetCapturedArguments() (string, io.ReadSeeker) {
	path, src := c.GetAllCapturedArguments()
	return path[len(path)-1], src[len(src)-1]
}

func (c *Blobstore_Put_OngoingVerification) GetAllCapturedArguments() (_param0 []string, _param1 []io.ReadSeeker) {
	params := pegomock.GetGenericMockFrom(c.mock).GetInvocationParams(c.methodInvocations)
	if len(params) > 0 {
		_param0 = make([]string, len(params[0]))
		for u, param := range params[0] {
			_param0[u] = param.(string)
		}
		_param1 = make([]io.ReadSeeker, len(params[1]))
		for u, param := range params[1] {
			_param1[u] = param.(io.ReadSeeker)
		}
	}
	return
}

func (verifier *VerifierBlobstore) Copy(src string, dest string) *Blobstore_Copy_OngoingVerification {
	params := []pegomock.Param{src, dest}
	methodInvocations := pegomock.GetGenericMockFrom(verifier.mock).Verify(verifier.inOrderContext, verifier.invocationCountMatcher, "Copy", params)
	return &Blobstore_Copy_OngoingVerification{mock: verifier.mock, methodInvocations: methodInvocations}
}

type Blobstore_Copy_OngoingVerification struct {
	mock              *MockBlobstore
	methodInvocations []pegomock.MethodInvocation
}

func (c *Blobstore_Copy_OngoingVerification) GetCapturedArguments() (string, string) {
	src, dest := c.GetAllCapturedArguments()
	return src[len(src)-1], dest[len(dest)-1]
}

func (c *Blobstore_Copy_OngoingVerification) GetAllCapturedArguments() (_param0 []string, _param1 []string) {
	params := pegomock.GetGenericMockFrom(c.mock).GetInvocationParams(c.methodInvocations)
	if len(params) > 0 {
		_param0 = make([]string, len(params[0]))
		for u, param := range params[0] {
			_param0[u] = param.(string)
		}
		_param1 = make([]string, len(params[1]))
		for u, param := range params[1] {
			_param1[u] = param.(string)
		}
	}
	return
}

func (verifier *VerifierBlobstore) Delete(path string) *Blobstore_Delete_OngoingVerification {
	params := []pegomock.Param{path}
	methodInvocations := pegomock.GetGenericMockFrom(verifier.mock).Verify(verifier.inOrderContext, verifier.invocationCountMatcher, "Delete", params)
	return &Blobstore_Delete_OngoingVerification{mock: verifier.mock, methodInvocations: methodInvocations}
}

type Blobstore_Delete_OngoingVerification struct {
	mock              *MockBlobstore
	methodInvocations []pegomock.MethodInvocation
}

func (c *Blobstore_Delete_OngoingVerification) GetCapturedArguments() string {
	path := c.GetAllCapturedArguments()
	return path[len(path)-1]
}

func (c *Blobstore_Delete_OngoingVerification) GetAllCapturedArguments() (_param0 []string) {
	params := pegomock.GetGenericMockFrom(c.mock).GetInvocationParams(c.methodInvocations)
	if len(params) > 0 {
		_param0 = make([]string, len(params[0]))
		for u, param := range params[0] {
			_param0[u] = param.(string)
		}
	}
	return
}

func (verifier *VerifierBlobstore) DeleteDir(prefix string) *Blobstore_DeleteDir_OngoingVerification {
	params := []pegomock.Param{prefix}
	methodInvocations := pegomock.GetGenericMockFrom(verifier.mock).Verify(verifier.inOrderContext, verifier.invocationCountMatcher, "DeleteDir", params)
	return &Blobstore_DeleteDir_OngoingVerification{mock: verifier.mock, methodInvocations: methodInvocations}
}

type Blobstore_DeleteDir_OngoingVerification struct {
	mock              *MockBlobstore
	methodInvocations []pegomock.MethodInvocation
}

func (c *Blobstore_DeleteDir_OngoingVerification) GetCapturedArguments() string {
	prefix := c.GetAllCapturedArguments()
	return prefix[len(prefix)-1]
}

func (c *Blobstore_DeleteDir_OngoingVerification) GetAllCapturedArguments() (_param0 []string) {
	params := pegomock.GetGenericMockFrom(c.mock).GetInvocationParams(c.methodInvocations)
	if len(params) > 0 {
		_param0 = make([]string, len(params[0]))
		for u, param := range params[0] {
			_param0[u] = param.(string)
		}
	}
	return
}
