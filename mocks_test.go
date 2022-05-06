// Code generated by MockGen. DO NOT EDIT.
// Source: s2kv (interfaces: Command,Writer)

// Package s2kv_test is a generated GoMock package.
package s2kv_test

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockCommand is a mock of Command interface.
type MockCommand struct {
	ctrl     *gomock.Controller
	recorder *MockCommandMockRecorder
}

// MockCommandMockRecorder is the mock recorder for MockCommand.
type MockCommandMockRecorder struct {
	mock *MockCommand
}

// NewMockCommand creates a new mock instance.
func NewMockCommand(ctrl *gomock.Controller) *MockCommand {
	mock := &MockCommand{ctrl: ctrl}
	mock.recorder = &MockCommandMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCommand) EXPECT() *MockCommandMockRecorder {
	return m.recorder
}

// ArgCount mocks base method.
func (m *MockCommand) ArgCount() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ArgCount")
	ret0, _ := ret[0].(int)
	return ret0
}

// ArgCount indicates an expected call of ArgCount.
func (mr *MockCommandMockRecorder) ArgCount() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ArgCount", reflect.TypeOf((*MockCommand)(nil).ArgCount))
}

// Get mocks base method.
func (m *MockCommand) Get(arg0 int) []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].([]byte)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockCommandMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCommand)(nil).Get), arg0)
}

// MockWriter is a mock of Writer interface.
type MockWriter struct {
	ctrl     *gomock.Controller
	recorder *MockWriterMockRecorder
}

// MockWriterMockRecorder is the mock recorder for MockWriter.
type MockWriterMockRecorder struct {
	mock *MockWriter
}

// NewMockWriter creates a new mock instance.
func NewMockWriter(ctrl *gomock.Controller) *MockWriter {
	mock := &MockWriter{ctrl: ctrl}
	mock.recorder = &MockWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWriter) EXPECT() *MockWriterMockRecorder {
	return m.recorder
}

// Write mocks base method.
func (m *MockWriter) Write(arg0 []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockWriterMockRecorder) Write(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockWriter)(nil).Write), arg0)
}

// WriteBulk mocks base method.
func (m *MockWriter) WriteBulk(arg0 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteBulk", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteBulk indicates an expected call of WriteBulk.
func (mr *MockWriterMockRecorder) WriteBulk(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteBulk", reflect.TypeOf((*MockWriter)(nil).WriteBulk), arg0)
}

// WriteBulkString mocks base method.
func (m *MockWriter) WriteBulkString(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteBulkString", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteBulkString indicates an expected call of WriteBulkString.
func (mr *MockWriterMockRecorder) WriteBulkString(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteBulkString", reflect.TypeOf((*MockWriter)(nil).WriteBulkString), arg0)
}

// WriteBulkStrings mocks base method.
func (m *MockWriter) WriteBulkStrings(arg0 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteBulkStrings", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteBulkStrings indicates an expected call of WriteBulkStrings.
func (mr *MockWriterMockRecorder) WriteBulkStrings(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteBulkStrings", reflect.TypeOf((*MockWriter)(nil).WriteBulkStrings), arg0)
}

// WriteBulks mocks base method.
func (m *MockWriter) WriteBulks(arg0 ...[]byte) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "WriteBulks", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteBulks indicates an expected call of WriteBulks.
func (mr *MockWriterMockRecorder) WriteBulks(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteBulks", reflect.TypeOf((*MockWriter)(nil).WriteBulks), arg0...)
}

// WriteError mocks base method.
func (m *MockWriter) WriteError(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteError", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteError indicates an expected call of WriteError.
func (mr *MockWriterMockRecorder) WriteError(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteError", reflect.TypeOf((*MockWriter)(nil).WriteError), arg0)
}

// WriteInt mocks base method.
func (m *MockWriter) WriteInt(arg0 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteInt", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteInt indicates an expected call of WriteInt.
func (mr *MockWriterMockRecorder) WriteInt(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteInt", reflect.TypeOf((*MockWriter)(nil).WriteInt), arg0)
}

// WriteSimpleString mocks base method.
func (m *MockWriter) WriteSimpleString(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteSimpleString", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteSimpleString indicates an expected call of WriteSimpleString.
func (mr *MockWriterMockRecorder) WriteSimpleString(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteSimpleString", reflect.TypeOf((*MockWriter)(nil).WriteSimpleString), arg0)
}